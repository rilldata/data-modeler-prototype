package duckdb

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/rduckdb"
)

type selfToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &selfToSelfExecutor{}

func (e *selfToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *selfToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	olap, ok := e.c.AsOLAP(e.c.instanceID)
	if !ok {
		return nil, fmt.Errorf("output connector is not OLAP")
	}

	inputProps := &ModelInputProperties{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if err := outputProps.Validate(opts); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
	}

	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = opts.ModelName
		usedModelName = true
	}

	materialize := opts.Env.DefaultMaterialize
	if outputProps.Materialize != nil {
		materialize = *outputProps.Materialize
	}

	asView := !materialize
	tableName := outputProps.Table

	// check if the SQL is ingesting data from an object store
	if scheme, ref, ok := objectStoreRef(inputProps, opts); ok {
		if scheme == "s3" || scheme == "azure" {
			// for s3 and azure we can just set a duckdb secret and ingest data using duckdb's native support for s3 and azure
			handle, release, err := opts.Env.AcquireConnector(ctx, scheme)
			if err != nil {
				return nil, err
			}
			defer release()
			secretSQL, err := objectStoreSecretSQL(opts.ModelName, opts.InputConnector, handle, opts.InputProperties)
			if err != nil {
				return nil, err
			}
			inputProps.PreExec = secretSQL
		} else { // gcs, gs, local
			// for gcs and gcs duckdb we need to hook into our object store connector to download the files and then ingest them into duckdb
			// this is a rudimentary rewrite and only cover simple use cases like SELECT * FROM read_xxx(path, union_by_name=true...)
			// and does not cover other SQL features like filter and limits
			handle, release, err := opts.Env.AcquireConnector(ctx, scheme)
			if err != nil {
				return nil, err
			}
			defer release()

			clone := *opts
			clone.InputConnector = scheme
			clone.InputHandle = handle

			props := maps.Clone(opts.InputProperties)
			if _, ok := props["format"].(string); !ok {
				switch ref.Function {
				case "read_csv_auto", "read_csv":
					props["format"] = "csv"
				case "read_json", "read_json_auto", "read_json_objects", "read_json_objects_auto", "read_ndjson_objects", "read_ndjson", "read_ndjson_auto":
					props["format"] = "json"
				case "read_parquet":
					props["format"] = "parquet"
				}
			}
			props["duckdb"] = ref.Properties
			if scheme == "local_file" {
				resolved, err := fileutil.ResolveLocalPath(ref.Paths[0], opts.Env.RepoRoot, opts.Env.AllowHostAccess)
				if err != nil {
					return nil, err
				}
				props["path"] = resolved
				clone.InputProperties = props
				filestore, ok := handle.AsFileStore()
				if !ok {
					return nil, fmt.Errorf("internal error: expected file store connector")
				}
				executor := &localFileToSelfExecutor{
					c:    e.c,
					from: filestore,
				}
				return executor.Execute(ctx, &clone)
			}
			// gcs
			props["path"] = ref.Paths[0]
			clone.InputProperties = props
			// call the objectStoreToSelfExecutor which has logic to download files based on path and then call selfToSelfExecutor
			executor := &objectStoreToSelfExecutor{c: e.c}
			return executor.Execute(ctx, &clone)
		}
	}

	if !opts.IncrementalRun {
		// Prepare for ingesting into the staging view/table.
		// NOTE: This intentionally drops the end table if not staging changes.
		stagingTableName := tableName
		if opts.Env.StageChanges {
			stagingTableName = stagingTableNameFor(tableName)
		}
		_ = olap.DropTable(ctx, stagingTableName)

		// Create the table
		if inputProps.DB != "" {
			// special handling for ingesting from an external database
			// not handling incremental use cases since ingesting from an external database is mostly for small,experimental use cases
			var err error
			inputProps.DB, err = fileutil.ResolveLocalPath(inputProps.DB, opts.Env.RepoRoot, opts.Env.AllowHostAccess)
			if err != nil {
				return nil, err
			}
			err = e.createFromExternalDuckDB(ctx, inputProps, stagingTableName)
			if err != nil {
				_ = olap.DropTable(ctx, stagingTableName)
				return nil, fmt.Errorf("failed to create model: %w", err)
			}
		} else {
			createTableOpts := &drivers.CreateTableOptions{
				View:         asView,
				BeforeCreate: inputProps.PreExec,
				AfterCreate:  inputProps.PostExec,
			}
			err := olap.CreateTableAsSelect(ctx, stagingTableName, inputProps.SQL, createTableOpts)
			if err != nil {
				_ = olap.DropTable(ctx, stagingTableName)
				return nil, fmt.Errorf("failed to create model: %w", err)
			}
		}

		// Rename the staging table to the final table name
		if stagingTableName != tableName {
			err := olapForceRenameTable(ctx, olap, stagingTableName, asView, tableName)
			if err != nil {
				return nil, fmt.Errorf("failed to rename staged model: %w", err)
			}
		}
	} else {
		// Insert into the table
		insertTableOpts := &drivers.InsertTableOptions{
			BeforeInsert: inputProps.PreExec,
			AfterInsert:  inputProps.PostExec,
			ByName:       false,
			InPlace:      true,
			Strategy:     outputProps.IncrementalStrategy,
			UniqueKey:    outputProps.UniqueKey,
		}
		err := olap.InsertTableAsSelect(ctx, tableName, inputProps.SQL, insertTableOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to incrementally insert into table: %w", err)
		}
	}

	// Build result props
	resultProps := &ModelResultProperties{
		Table:         tableName,
		View:          asView,
		UsedModelName: usedModelName,
	}
	resultPropsMap := map[string]interface{}{}
	err := mapstructure.WeakDecode(resultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	// Done
	return &drivers.ModelResult{
		Connector:  opts.OutputConnector,
		Properties: resultPropsMap,
		Table:      tableName,
	}, nil
}

func (e *selfToSelfExecutor) createFromExternalDuckDB(ctx context.Context, inputProps *ModelInputProperties, tbl string) error {
	safeDBName := safeName(tbl + "_external_db_")
	safeTempTable := safeName(tbl + "__temp__")
	beforeCreateFn := func(ctx context.Context, conn *sqlx.Conn) error {
		if inputProps.PreExec != "" {
			if _, err := conn.ExecContext(ctx, inputProps.PreExec); err != nil {
				return err
			}
		}

		if _, err := conn.ExecContext(ctx, fmt.Sprintf("ATTACH %s AS %s (READ_ONLY)", safeSQLString(inputProps.DB), safeDBName)); err != nil {
			return err
		}

		var localDB, localSchema string
		if err := conn.QueryRowxContext(ctx, "SELECT current_database(),current_schema();").Scan(&localDB, &localSchema); err != nil {
			return err
		}

		if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE %s;", safeDBName)); err != nil {
			return err
		}

		userQuery := strings.TrimSpace(inputProps.SQL)
		userQuery, _ = strings.CutSuffix(userQuery, ";") // trim trailing semi colon
		query := fmt.Sprintf("CREATE OR REPLACE TABLE %s.%s.%s AS (%s\n);", safeName(localDB), safeName(localSchema), safeTempTable, userQuery)
		_, execErr := conn.ExecContext(ctx, query)
		// revert to localdb and schema before returning
		_, err := conn.ExecContext(ctx, fmt.Sprintf("USE %s.%s;", safeName(localDB), safeName(localSchema)))
		return errors.Join(execErr, err)
	}
	afterCreateFn := func(ctx context.Context, conn *sqlx.Conn) error {
		if inputProps.PostExec != "" {
			if _, err := conn.ExecContext(ctx, inputProps.PostExec); err != nil {
				return err
			}
		}
		_, err := conn.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", safeTempTable))
		return err
	}
	db, release, err := e.c.acquireDB()
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()
	return db.CreateTableAsSelect(ctx, tbl, fmt.Sprintf("SELECT * FROM %s", safeTempTable), &rduckdb.CreateTableOptions{
		BeforeCreateFn: beforeCreateFn,
		AfterCreateFn:  afterCreateFn,
	})
}

// Backward compatibility: It was possible to set a duckdb SQL which ingests data from an object store without setting the object store credentials.
// We did some rewriting for path to rewrite object store paths to paths of locally downloaded files.
// This function rewrites the source properties to use object store connector so that a model executor that ingests data from object store to duckdb can work.
func objectStoreRef(props *ModelInputProperties, opts *drivers.ModelExecuteOptions) (string, *duckdbsql.TableRef, bool) {
	// We take an assumption that if there is a pre_exec query, the user has already set the secret SQL.
	if props.PreExec != "" || opts.InputConnector != "duckdb" {
		return "", nil, false
	}
	// Parse AST
	ast, err := duckdbsql.Parse(props.SQL)
	if err != nil {
		// If we can't parse the SQL just let duckdb run on it and give a sql parse error.
		return "", nil, false
	}

	// If there is a single table reference check if it is an object store reference.
	refs := ast.GetTableRefs()
	if len(refs) != 1 {
		return "", nil, false
	}
	ref := refs[0]
	// Parse the path as a URL (also works for local paths)
	if len(ref.Paths) == 0 {
		return "", nil, false
	}
	uri, err := url.Parse(ref.Paths[0])
	if err != nil {
		return "", nil, false
	}

	if uri.Scheme == "s3" || uri.Scheme == "azure" || uri.Scheme == "gcs" || uri.Scheme == "gs" {
		return uri.Scheme, ref, true
	}
	if uri.Scheme == "" && uri.Host == "" {
		// local file reference
		return "local_file", ref, true
	}
	return "", nil, false
}
