package duckdb

import (
	"context"
	"errors"
	"fmt"
	"io"
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

	// Backward compatibility for the old duckdb SQL:
	// It was possible to set a duckdb SQL which ingests data from an object store without setting the object store credentials.
	// We did rewriting for path to rewrite object store paths to paths of locally downloaded files.
	// The handling can now be done with duckdb's native connectors by setting a SQL that creates secret to access the object store.
	// However duckdb does not support GCS's native credentials(google_application_credentials) so we still maintain the hack for the same.
	// We expect to remove this rewriting once all users start using GCS's s3 compatibility API support.
	if scheme, secretSQL, ast, ok := objectStoreRef(ctx, inputProps, opts); ok {
		if secretSQL != "" {
			inputProps.PreExec = secretSQL
		} else if scheme == "gcs" || scheme == "gs" {
			// rewrite duckdb sql with locally downloaded files
			handle, release, err := opts.Env.AcquireConnector(ctx, scheme)
			if err != nil {
				return nil, err
			}
			defer release()
			rawProps := maps.Clone(opts.InputProperties)
			rawProps["path"] = ast.GetTableRefs()[0].Paths[0]
			rawProps["batch_size"] = -1
			release, err = rewriteDuckDBSQL(ctx, inputProps, handle, rawProps, ast)
			if err != nil {
				return nil, err
			}
			defer release()
		} else {
			rewrittenSQL, err := rewriteLocalPaths(ast, opts.Env.RepoRoot, opts.Env.AllowHostAccess)
			if err != nil {
				return nil, fmt.Errorf("invalid local path: %w", err)
			}
			inputProps.SQL = rewrittenSQL
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

func objectStoreRef(ctx context.Context, props *ModelInputProperties, opts *drivers.ModelExecuteOptions) (string, string, *duckdbsql.AST, bool) {
	// We take an assumption that if there is a pre_exec query, the user has already set the secret SQL.
	if props.PreExec != "" || opts.InputConnector != "duckdb" {
		return "", "", nil, false
	}
	// Parse AST
	ast, err := duckdbsql.Parse(props.SQL)
	if err != nil {
		// If we can't parse the SQL just let duckdb run on it and give a sql parse error.
		return "", "", nil, false
	}

	// If there is a single table reference check if it is an object store reference.
	refs := ast.GetTableRefs()
	if len(refs) != 1 {
		return "", "", nil, false
	}
	ref := refs[0]
	// Parse the path as a URL (also works for local paths)
	if len(ref.Paths) == 0 {
		return "", "", nil, false
	}
	uri, err := url.Parse(ref.Paths[0])
	if err != nil {
		return "", "", nil, false
	}

	if uri.Scheme == "s3" || uri.Scheme == "azure" || uri.Scheme == "gcs" || uri.Scheme == "gs" {
		// for s3 and azure we can just set a duckdb secret and ingest data using duckdb's native support for s3 and azure
		handle, release, err := opts.Env.AcquireConnector(ctx, uri.Scheme)
		if err != nil {
			return "", "", nil, false
		}
		defer release()
		secretSQL, err := objectStoreSecretSQL(ctx, ref.Paths[0], opts.ModelName, opts.InputConnector, handle, opts.InputProperties)
		if err != nil {
			if errors.Is(err, errGCSUsesNativeCreds) {
				return uri.Scheme, "", ast, true
			}
			return "", "", nil, false
		}
		return uri.Scheme, secretSQL, ast, true
	}
	if uri.Scheme == "" && uri.Host == "" {
		// local file reference
		return "local_file", "", ast, true
	}
	return "", "", nil, false
}

func rewriteDuckDBSQL(ctx context.Context, props *ModelInputProperties, inputHandle drivers.Handle, rawProps map[string]any, ast *duckdbsql.AST) (release func(), retErr error) {
	fs, ok := inputHandle.AsObjectStore()
	if !ok {
		return nil, fmt.Errorf("internal error: expected object store connector")
	}

	var files []string
	iter, err := fs.DownloadFiles(ctx, rawProps)
	if err != nil {
		return nil, err
	}
	defer func() {
		if retErr != nil {
			_ = iter.Close()
		}
	}()
	for {
		localFiles, err := iter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		files = append(files, localFiles...)
	}

	// Rewrite the SQL
	props.SQL, err = rewriteSQL(ast, files)
	return func() { _ = iter.Close() }, err
}
