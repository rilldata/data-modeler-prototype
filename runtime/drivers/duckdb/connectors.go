package duckdb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/localfile"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

const (
	_iteratorBatch        = 8
	_defaultIngestTimeout = 60 * time.Minute
)

// Ingest data from a source with a timeout
func (c *connection) Ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) (*drivers.IngestionSummary, error) {
	// Wraps c.ingest with timeout handling

	timeout := _defaultIngestTimeout
	if source.Timeout > 0 {
		timeout = time.Duration(source.Timeout) * time.Second
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	summary, err := c.ingest(ctxWithTimeout, env, source)
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		return nil, fmt.Errorf("ingestion timeout exceeded (source=%q, timeout=%s)", source.Name, timeout.String())
	}

	return summary, err
}

func (c *connection) ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) (*drivers.IngestionSummary, error) {
	// Driver-specific overrides
	if source.Connector == "local_file" {
		return c.ingestLocalFiles(ctx, env, source)
	}

	iterator, err := connectors.ConsumeAsIterator(ctx, env, source, c.logger)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	appendToTable := false
	summary := &drivers.IngestionSummary{}

	format, formatDefined := source.Properties["format"].(string)
	if formatDefined {
		format = fmt.Sprintf(".%s", format)
	}

	allowSchemaRelaxation, err := schemaRelaxationProperty(source.Properties)
	if err != nil {
		return nil, err
	}

	var ingestionProps map[string]any
	if duckDBProps, ok := source.Properties["duckdb"].(map[string]any); ok {
		ingestionProps = duckDBProps
	} else {
		ingestionProps = map[string]any{}
	}
	if _, ok := ingestionProps["union_by_name"]; !ok && allowSchemaRelaxation {
		// set union_by_name to unify the schema of the files
		ingestionProps["union_by_name"] = true
	}

	a := newAppender(c, source, ingestionProps, allowSchemaRelaxation)

	for iterator.HasNext() {
		files, err := iterator.NextBatch(_iteratorBatch)
		if err != nil {
			return nil, err
		}

		if !formatDefined {
			format = fileutil.FullExt(files[0])
			formatDefined = true
		}

		st := time.Now()
		c.logger.Info("ingesting files", zap.String("source", source.Name), zap.Strings("files", files), observability.ZapCtx(ctx))
		if appendToTable {
			if err := a.appendData(ctx, files, format); err != nil {
				return nil, err
			}
		} else {
			from, err := sourceReader(files, format, ingestionProps)
			if err != nil {
				return nil, err
			}

			query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", source.Name, from)
			if err := c.Exec(ctx, &drivers.Statement{Query: query, Priority: 1}); err != nil {
				return nil, c.checkErr(err)
			}
		}

		size := fileSize(files)
		summary.BytesIngested += size
		c.logger.Info("ingested files", zap.String("source", source.Name), zap.Strings("files", files), zap.Int64("bytes_ingested", size), zap.Duration("duration", time.Since(st)), observability.ZapCtx(ctx))
		appendToTable = true
	}
	return summary, nil
}

// local files
func (c *connection) ingestLocalFiles(ctx context.Context, env *connectors.Env, source *connectors.Source) (*drivers.IngestionSummary, error) {
	conf, err := localfile.ParseConfig(source.Properties)
	if err != nil {
		return nil, err
	}

	path, err := resolveLocalPath(env, conf.Path, source.Name)
	if err != nil {
		return nil, err
	}

	// get all files in case glob passed
	localPaths, err := doublestar.FilepathGlob(path)
	if err != nil {
		return nil, err
	}
	if len(localPaths) == 0 {
		return nil, fmt.Errorf("file does not exist at %s", conf.Path)
	}

	var format string
	if conf.Format != "" {
		format = fmt.Sprintf(".%s", conf.Format)
	} else {
		format = fileutil.FullExt(localPaths[0])
	}

	var ingestionProps map[string]any
	if duckDBProps, ok := source.Properties["duckdb"].(map[string]any); ok {
		ingestionProps = duckDBProps
	} else {
		ingestionProps = map[string]any{}
	}

	// Ingest data
	from, err := sourceReader(localPaths, format, ingestionProps)
	if err != nil {
		return nil, err
	}
	qry := fmt.Sprintf("CREATE OR REPLACE TABLE %q AS (SELECT * FROM %s)", source.Name, from)
	err = c.Exec(ctx, &drivers.Statement{Query: qry, Priority: 1})
	if err != nil {
		return nil, c.checkErr(err)
	}

	bytesIngested := fileSize(localPaths)
	return &drivers.IngestionSummary{BytesIngested: bytesIngested}, nil
}

func (c *connection) scanSchemaFromQuery(ctx context.Context, qry string) (map[string]string, error) {
	result, err := c.Execute(ctx, &drivers.Statement{Query: qry, Priority: 1})
	if err != nil {
		return nil, c.checkErr(err)
	}
	defer result.Close()

	schema := make(map[string]string)
	for result.Next() {
		var s duckDBTableSchemaResult
		if err := result.StructScan(&s); err != nil {
			return nil, c.checkErr(err)
		}
		schema[s.ColumnName] = s.ColumnType
	}

	if err := result.Err(); err != nil {
		return nil, c.checkErr(err)
	}

	return schema, nil
}

type appender struct {
	*connection
	source                *connectors.Source
	ingestionProps        map[string]any
	allowSchemaRelaxation bool
	tableSchema           map[string]string
}

func newAppender(c *connection, source *connectors.Source, ingestionProps map[string]any, allowSchemaRelaxation bool) *appender {
	return &appender{
		connection:            c,
		source:                source,
		ingestionProps:        ingestionProps,
		allowSchemaRelaxation: allowSchemaRelaxation,
		tableSchema:           nil,
	}
}

func (a *appender) appendData(ctx context.Context, files []string, format string) error {
	from, err := sourceReader(files, format, a.ingestionProps)
	if err != nil {
		return err
	}

	var query string
	if a.allowSchemaRelaxation {
		query = fmt.Sprintf("INSERT INTO %q BY NAME (SELECT * FROM %s);", a.source.Name, from)
	} else {
		query = fmt.Sprintf("INSERT INTO %q (SELECT * FROM %s);", a.source.Name, from)
	}
	a.logger.Debug("generated query", zap.String("query", query), observability.ZapCtx(ctx))
	err = a.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
	if err == nil || !a.allowSchemaRelaxation || !containsAny(err.Error(), []string{"binder error", "conversion error"}) {
		return a.connection.checkErr(err)
	}

	// error is of type binder error (more or less columns than current table schema)
	// or of type conversion error (datatype changed or column sequence changed)
	err = a.updateSchema(ctx, from, files)
	if err != nil {
		return fmt.Errorf("failed to update schema %w", err)
	}

	query = fmt.Sprintf("INSERT INTO %q BY NAME (SELECT * FROM %s);", a.source.Name, from)
	a.logger.Debug("generated query", zap.String("query", query), observability.ZapCtx(ctx))
	err = a.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
	return a.connection.checkErr(err)
}

// updateSchema updates the schema of the table in case new file adds a new column or
// updates the datatypes of an existing columns with a wider datatype.
func (a *appender) updateSchema(ctx context.Context, from string, fileNames []string) error {
	// schema of new files
	srcSchema, err := a.scanSchemaFromQuery(ctx, fmt.Sprintf("DESCRIBE (SELECT * FROM %s LIMIT 0);", from))
	if err != nil {
		return err
	}

	// combined schema
	qry := fmt.Sprintf("DESCRIBE ((SELECT * FROM %s limit 0) UNION ALL BY NAME (SELECT * FROM %s limit 0));", a.source.Name, from)
	unionSchema, err := a.scanSchemaFromQuery(ctx, qry)
	if err != nil {
		return err
	}

	// current schema
	if a.tableSchema == nil {
		a.tableSchema, err = a.scanSchemaFromQuery(ctx, fmt.Sprintf("DESCRIBE %q;", a.source.Name))
		if err != nil {
			return err
		}
	}

	newCols := make(map[string]string)
	colTypeChanged := make(map[string]string)
	for colName, colType := range unionSchema {
		oldType, ok := a.tableSchema[colName]
		if !ok {
			newCols[colName] = colType
		} else if oldType != colType {
			colTypeChanged[colName] = colType
		}
	}

	if !a.allowSchemaRelaxation {
		if len(srcSchema) < len(unionSchema) {
			fileNames := strings.Join(names(fileNames), ",")
			columns := strings.Join(missingMapKeys(a.tableSchema, srcSchema), ",")
			return fmt.Errorf("new files %q are missing columns %q and schema relaxation not allowed", fileNames, columns)
		}

		if len(colTypeChanged) != 0 {
			fileNames := strings.Join(names(fileNames), ",")
			columns := strings.Join(keys(colTypeChanged), ",")
			return fmt.Errorf("new files %q change datatypes of some columns %q and schema relaxation not allowed", fileNames, columns)
		}
	}

	if len(newCols) != 0 && !a.allowSchemaRelaxation {
		fileNames := strings.Join(names(fileNames), ",")
		columns := strings.Join(missingMapKeys(srcSchema, a.tableSchema), ",")
		return fmt.Errorf("new files %q have new columns %q and schema relaxation not allowed", fileNames, columns)
	}

	for colName, colType := range newCols {
		a.tableSchema[colName] = colType
		qry := fmt.Sprintf("ALTER TABLE %q ADD COLUMN %q %s", a.source.Name, colName, colType)
		if err := a.Exec(ctx, &drivers.Statement{Query: qry}); err != nil {
			return a.connection.checkErr(err)
		}
	}

	for colName, colType := range colTypeChanged {
		a.tableSchema[colName] = colType
		qry := fmt.Sprintf("ALTER TABLE %q ALTER COLUMN %q SET DATA TYPE %s", a.source.Name, colName, colType)
		if err := a.Exec(ctx, &drivers.Statement{Query: qry}); err != nil {
			return a.connection.checkErr(err)
		}
	}

	return nil
}

func resolveLocalPath(env *connectors.Env, path, sourceName string) (string, error) {
	path, err := fileutil.ExpandHome(path)
	if err != nil {
		return "", err
	}

	repoRoot := env.RepoRoot
	finalPath := path
	if !filepath.IsAbs(path) {
		finalPath = filepath.Join(repoRoot, path)
	}

	if !env.AllowHostAccess && !strings.HasPrefix(finalPath, repoRoot) {
		// path is outside the repo root
		return "", fmt.Errorf("file connector cannot ingest source '%s': path is outside repo root", sourceName)
	}
	return finalPath, nil
}

func sourceReader(paths []string, format string, ingestionProps map[string]any) (string, error) {
	// Generate a "read" statement
	if containsAny(format, []string{".csv", ".tsv", ".txt"}) {
		// CSV reader
		return generateReadCsvStatement(paths, ingestionProps)
	} else if strings.Contains(format, ".parquet") {
		// Parquet reader
		return generateReadParquetStatement(paths, ingestionProps)
	} else if containsAny(format, []string{".json", ".ndjson"}) {
		// JSON reader
		return generateReadJSONStatement(paths, ingestionProps)
	} else {
		return "", fmt.Errorf("file type not supported : %s", format)
	}
}

func generateReadCsvStatement(paths []string, properties map[string]any) (string, error) {
	ingestionProps := copyMap(properties)
	// set sample_size to 200000 by default
	if _, sampleSizeDefined := ingestionProps["sample_size"]; !sampleSizeDefined {
		ingestionProps["sample_size"] = 200000
	}
	// auto_detect (enables auto-detection of parameters) is true by default, it takes care of params/schema
	return fmt.Sprintf("read_csv_auto(%s)", convertToStatementParamsStr(paths, ingestionProps)), nil
}

func generateReadParquetStatement(paths []string, properties map[string]any) (string, error) {
	ingestionProps := copyMap(properties)
	// set hive_partitioning to true by default
	if _, hivePartitioningDefined := ingestionProps["hive_partitioning"]; !hivePartitioningDefined {
		ingestionProps["hive_partitioning"] = true
	}
	return fmt.Sprintf("read_parquet(%s)", convertToStatementParamsStr(paths, ingestionProps)), nil
}

func generateReadJSONStatement(paths []string, properties map[string]any) (string, error) {
	ingestionProps := copyMap(properties)
	// auto_detect is false by default so setting it to true simplifies the ingestion
	// if columns are defined then DuckDB turns the auto-detection off so no need to check this case here
	if _, autoDetectDefined := ingestionProps["auto_detect"]; !autoDetectDefined {
		ingestionProps["auto_detect"] = true
	}
	// set sample_size to 200000 by default
	if _, sampleSizeDefined := ingestionProps["sample_size"]; !sampleSizeDefined {
		ingestionProps["sample_size"] = 200000
	}
	// set format to auto by default
	if _, formatDefined := ingestionProps["format"]; !formatDefined {
		ingestionProps["format"] = "auto"
	}
	return fmt.Sprintf("read_json(%s)", convertToStatementParamsStr(paths, ingestionProps)), nil
}

func convertToStatementParamsStr(paths []string, properties map[string]any) string {
	ingestionParamsStr := make([]string, 0, len(properties)+1)
	// The first parameter is a source path
	ingestionParamsStr = append(ingestionParamsStr, fmt.Sprintf("['%s']", strings.Join(paths, "','")))
	for key, value := range properties {
		ingestionParamsStr = append(ingestionParamsStr, fmt.Sprintf("%s=%v", key, value))
	}
	return strings.Join(ingestionParamsStr, ",")
}

type duckDBTableSchemaResult struct {
	ColumnName string  `db:"column_name"`
	ColumnType string  `db:"column_type"`
	Nullable   *string `db:"null"`
	Key        *string `db:"key"`
	Default    *string `db:"default"`
	Extra      *string `db:"extra"`
}

func schemaRelaxationProperty(prop map[string]interface{}) (bool, error) {
	allowSchemaRelaxation, defined := prop["allow_schema_relaxation"].(bool)
	val, ok := prop["union_by_name"].(bool)
	if ok && !val && allowSchemaRelaxation {
		// if union_by_name is set as false addition can't be done
		return false, fmt.Errorf("if `union_by_name` is set `allow_schema_relaxation` must be disabled")
	}

	if hasKey(prop, "columns", "types", "dtypes") && allowSchemaRelaxation {
		return false, fmt.Errorf("if any of `columns`,`types`,`dtypes` is set `allow_schema_relaxation` must be disabled")
	}

	// set default values
	if !defined {
		allowSchemaRelaxation = true
	}

	return allowSchemaRelaxation, nil
}

// utility functions
func hasKey(m map[string]interface{}, key ...string) bool {
	for _, k := range key {
		if _, ok := m[k]; ok {
			return true
		}
	}
	return false
}

func missingMapKeys(src, lookup map[string]string) []string {
	keys := make([]string, 0)
	for k := range src {
		if _, ok := lookup[k]; !ok {
			keys = append(keys, k)
		}
	}
	return keys
}

func keys(src map[string]string) []string {
	keys := make([]string, 0, len(src))
	for k := range src {
		keys = append(keys, k)
	}
	return keys
}

func names(filePaths []string) []string {
	names := make([]string, len(filePaths))
	for i, f := range filePaths {
		names[i] = filepath.Base(f)
	}
	return names
}

// copyMap does a shallow copy of the map
func copyMap(originalMap map[string]any) map[string]any {
	newMap := make(map[string]any, len(originalMap))
	for key, value := range originalMap {
		newMap[key] = value
	}
	return newMap
}

func containsAny(s string, targets []string) bool {
	source := strings.ToLower(s)
	for _, target := range targets {
		if strings.Contains(source, target) {
			return true
		}
	}
	return false
}

func fileSize(paths []string) int64 {
	var size int64
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil { // ignoring error since only error possible is *PathError
			size += info.Size()
		}
	}
	return size
}
