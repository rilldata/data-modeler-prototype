package druid

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/druid/druidsqldriver"
	"go.uber.org/zap"
)

const (
	numRetries = 3
	retryWait  = 300 * time.Millisecond
)

var _ drivers.OLAPStore = &connection{}

// AddTableColumn implements drivers.OLAPStore.
func (c *connection) AddTableColumn(ctx context.Context, tableName, columnName, typ string) error {
	return fmt.Errorf("druid: data transformation not yet supported")
}

// AlterTableColumn implements drivers.OLAPStore.
func (c *connection) AlterTableColumn(ctx context.Context, tableName, columnName, newType string) error {
	return fmt.Errorf("druid: data transformation not yet supported")
}

// CreateTableAsSelect implements drivers.OLAPStore.
func (c *connection) CreateTableAsSelect(ctx context.Context, name, sql string, opts *drivers.CreateTableOptions) error {
	return fmt.Errorf("druid: data transformation not yet supported")
}

// InsertTableAsSelect implements drivers.OLAPStore.
func (c *connection) InsertTableAsSelect(ctx context.Context, name, sql string, opts *drivers.InsertTableOptions) error {
	return fmt.Errorf("druid: data transformation not yet supported")
}

// DropTable implements drivers.OLAPStore.
func (c *connection) DropTable(ctx context.Context, name string) error {
	return fmt.Errorf("druid: data transformation not yet supported")
}

// RenameTable implements drivers.OLAPStore.
func (c *connection) RenameTable(ctx context.Context, name, newName string) error {
	return fmt.Errorf("druid: data transformation not yet supported")
}

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectDruid
}

func (c *connection) WithConnection(ctx context.Context, priority int, longRunning bool, fn drivers.WithConnectionFunc) error {
	return fmt.Errorf("druid: WithConnection not supported")
}

func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	res, err := c.Execute(ctx, stmt)
	if err != nil {
		return err
	}
	if stmt.DryRun {
		return nil
	}
	return res.Close()
}

func (c *connection) Execute(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	// Log query if enabled (usually disabled)
	if c.config.LogQueries {
		c.logger.Info("druid query", zap.String("sql", stmt.Query), zap.Any("args", stmt.Args))
	}

	if stmt.DryRun {
		rows, err := c.db.QueryxContext(ctx, "EXPLAIN PLAN FOR "+stmt.Query, stmt.Args...)
		if err != nil {
			return nil, err
		}

		return nil, rows.Close()
	}

	var cancelFunc context.CancelFunc
	if stmt.ExecutionTimeout != 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, stmt.ExecutionTimeout)
	}

	var queryCfg *druidsqldriver.QueryConfig
	if stmt.UseCache != nil {
		queryCfg = &druidsqldriver.QueryConfig{
			UseCache: stmt.UseCache,
		}
	}
	if stmt.PopulateCache != nil {
		if queryCfg == nil {
			queryCfg = &druidsqldriver.QueryConfig{}
		}
		queryCfg.PopulateCache = stmt.PopulateCache
	}

	if queryCfg != nil {
		ctx = druidsqldriver.WithQueryConfig(ctx, queryCfg)
	}

	var rows *sqlx.Rows
	var err error

	re := retrier.New(retrier.ExponentialBackoff(numRetries, retryWait), retryErrClassifier{})
	err = re.RunCtx(ctx, func(ctx2 context.Context) error {
		rows, err = c.db.QueryxContext(ctx2, stmt.Query, stmt.Args...)
		return err
	})
	if err != nil {
		if cancelFunc != nil {
			cancelFunc()
		}
		return nil, err
	}

	schema, err := rowsToSchema(rows)
	if err != nil {
		rows.Close()
		if cancelFunc != nil {
			cancelFunc()
		}

		return nil, err
	}

	r := &drivers.Result{Rows: rows, Schema: schema}
	r.SetCleanupFunc(func() error {
		if cancelFunc != nil {
			cancelFunc()
		}

		return nil
	})

	return r, nil
}

func (c *connection) MayBeScaledToZero(ctx context.Context) bool {
	return false
}

func rowsToSchema(r *sqlx.Rows) (*runtimev1.StructType, error) {
	if r == nil {
		return nil, nil
	}

	cts, err := r.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := make([]*runtimev1.StructType_Field, len(cts))
	for i, ct := range cts {
		nullable, ok := ct.Nullable()
		if !ok {
			nullable = true
		}

		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: databaseTypeToPB(ct.DatabaseTypeName(), nullable),
		}
	}

	return &runtimev1.StructType{Fields: fields}, nil
}

// In druid there are multiple schemas but all user tables are in druid schema.
// Other useful schema is INFORMATION_SCHEMA for metadata.
// There are 2 more schemas - sys (internal things) and lookup (druid specific lookup).
// While querying druid does not support db name just use schema.table
//
// Since all user tables are in `druid` schema so we hardcode schema as `druid` and does not query database
type informationSchema struct {
	c *connection
}

func (c *connection) InformationSchema() drivers.InformationSchema {
	return informationSchema{c: c}
}

func (i informationSchema) All(ctx context.Context, like string) ([]*drivers.Table, error) {
	var likeClause string
	var args []any
	if like != "" {
		likeClause = "AND LOWER(T.TABLE_NAME) LIKE LOWER(?)"
		args = []any{like}
	}

	q := fmt.Sprintf(`
		SELECT
			T.TABLE_SCHEMA AS SCHEMA,
			T.TABLE_NAME AS NAME,
			T.TABLE_TYPE AS TABLE_TYPE, 
			C.COLUMN_NAME AS COLUMNS,
			C.DATA_TYPE AS COLUMN_TYPE,
			C.IS_NULLABLE = 'YES' AS IS_NULLABLE
		FROM INFORMATION_SCHEMA.TABLES T 
		JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
		WHERE T.TABLE_SCHEMA = 'druid'
		%s
		ORDER BY SCHEMA, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	`, likeClause)

	rows, err := i.c.db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

func (i informationSchema) Lookup(ctx context.Context, db, schema, name string) (*drivers.Table, error) {
	// Ensure Coordinator is ready.
	// The issues is that the request
	//	SELECT ...
	//	FROM INFORMATION_SCHEMA.TABLES T
	//	JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
	//	WHERE T.TABLE_SCHEMA = 'druid' AND T.TABLE_NAME = ?
	//	ORDER BY SCHEMA, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	// returns false-negative if the Coordinator is being restarted. Retrier is a more abstract component and it doesn't check
	// if SQL tries to retrieve the dynamic schema and the will be no error from Druid Router
	// (because if the dynamic schema is empty - it's considered OK by the Druid cluster).
	q := "SELECT * FROM sys.segments LIMIT 1"
	rows, err := i.c.db.QueryxContext(ctx, q, name)
	if err != nil {
		return nil, err
	}
	rows.Close()

	q = `
		SELECT
			T.TABLE_SCHEMA AS SCHEMA,
			T.TABLE_NAME AS NAME,
			T.TABLE_TYPE AS TABLE_TYPE, 
			C.COLUMN_NAME AS COLUMN_NAME,
			C.DATA_TYPE AS COLUMN_TYPE,
			C.IS_NULLABLE = 'YES' AS IS_NULLABLE
		FROM INFORMATION_SCHEMA.TABLES T 
		JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
		WHERE T.TABLE_SCHEMA = 'druid' AND T.TABLE_NAME = ?
		ORDER BY SCHEMA, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	`

	rows, err = i.c.db.QueryxContext(ctx, q, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, drivers.ErrNotFound
	}

	return tables[0], nil
}

func (i informationSchema) scanTables(rows *sqlx.Rows) ([]*drivers.Table, error) {
	var res []*drivers.Table

	for rows.Next() {
		var schema string
		var name string
		var tableType string
		var columnName string
		var columnType string
		var nullable bool

		err := rows.Scan(&schema, &name, &tableType, &columnName, &columnType, &nullable)
		if err != nil {
			return nil, err
		}

		// set t to res[len(res)-1] if it's the same table, else set t to a new table and append it
		var t *drivers.Table
		if len(res) > 0 {
			t = res[len(res)-1]
			if !(t.DatabaseSchema == schema && t.Name == name) {
				t = nil
			}
		}
		if t == nil {
			t = &drivers.Table{
				DatabaseSchema:          schema,
				IsDefaultDatabaseSchema: true,
				Name:                    name,
				Schema:                  &runtimev1.StructType{},
			}
			res = append(res, t)
		}

		// append column
		t.Schema.Fields = append(t.Schema.Fields, &runtimev1.StructType_Field{
			Name: columnName,
			Type: databaseTypeToPB(columnType, nullable),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func databaseTypeToPB(dbt string, nullable bool) *runtimev1.Type {
	t := &runtimev1.Type{Nullable: nullable}
	switch dbt {
	case "BOOLEAN":
		t.Code = runtimev1.Type_CODE_BOOL
	case "TINYINT":
		t.Code = runtimev1.Type_CODE_INT8
	case "SMALLINT":
		t.Code = runtimev1.Type_CODE_INT16
	case "INTEGER":
		t.Code = runtimev1.Type_CODE_INT32
	case "BIGINT":
		t.Code = runtimev1.Type_CODE_INT64
	case "FLOAT":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "DOUBLE":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "REAL":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "DECIMAL":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "CHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "VARCHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "TIMESTAMP":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "OTHER":
		t.Code = runtimev1.Type_CODE_JSON
	}

	return t
}

// retryErrClassifier classifies 429 errors as retryable and all other errors as non retryable
type retryErrClassifier struct{}

func (retryErrClassifier) Classify(err error) retrier.Action {
	if err == nil {
		return retrier.Succeed
	}

	if strings.Contains(err.Error(), "QueryCapacityExceededException") {
		return retrier.Retry
	}

	return retrier.Fail
}
