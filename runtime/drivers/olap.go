package drivers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"

	// Load IANA time zone data
	_ "time/tzdata"
)

// ErrUnsupportedConnector is returned from Ingest for unsupported connectors.
var ErrUnsupportedConnector = errors.New("drivers: connector not supported")

// WithConnectionFunc is a callback function that provides a context to be used in further OLAP store calls to enforce affinity to a single connection.
// It also provides pointers to the actual database/sql and database/sql/driver connections.
// It's called with two contexts: wrappedCtx wraps the input context (including cancellation),
// and ensuredCtx wraps a background context (ensuring it can never be cancelled).
type WithConnectionFunc func(wrappedCtx context.Context, ensuredCtx context.Context, conn *sql.Conn) error

// OLAPStore is implemented by drivers that are capable of storing, transforming and serving analytical queries.
// NOTE crud APIs are not safe to be called with `WithConnection`
type OLAPStore interface {
	Dialect() Dialect
	WithConnection(ctx context.Context, priority int, longRunning, tx bool, fn WithConnectionFunc) error
	Exec(ctx context.Context, stmt *Statement) error
	Execute(ctx context.Context, stmt *Statement) (*Result, error)
	InformationSchema() InformationSchema

	CreateTableAsSelect(ctx context.Context, name string, view bool, sql string, tableOpts map[string]any) error
	InsertTableAsSelect(ctx context.Context, name, sql string, byName, inPlace bool, strategy IncrementalStrategy, uniqueKey []string) error
	DropTable(ctx context.Context, name string, view bool) error
	RenameTable(ctx context.Context, name, newName string, view bool) error
	AddTableColumn(ctx context.Context, tableName, columnName string, typ string) error
	AlterTableColumn(ctx context.Context, tableName, columnName string, newType string) error

	MayBeScaledToZero(ctx context.Context) bool
}

// Statement wraps a query to execute against an OLAP driver.
type Statement struct {
	Query            string
	Args             []any
	DryRun           bool
	Priority         int
	LongRunning      bool
	ExecutionTimeout time.Duration
}

// Result wraps the results of query.
type Result struct {
	*sqlx.Rows
	Schema    *runtimev1.StructType
	cleanupFn func() error
	cap       int64
	rows      int64
}

// SetCleanupFunc sets a function, which will be called when the Result is closed.
func (r *Result) SetCleanupFunc(fn func() error) {
	if r.cleanupFn == nil {
		r.cleanupFn = fn
		return
	}

	prevFn := r.cleanupFn
	r.cleanupFn = func() error {
		err1 := prevFn()
		err2 := fn()
		return errors.Join(err1, err2)
	}
}

// SetCap caps the number of rows to return. If the number is exceeded, an error is returned.
func (r *Result) SetCap(n int64) {
	if r.cap > 0 {
		panic("cap already set")
	}
	r.cap = n
}

// Next wraps rows.Next and enforces the cap set by SetCap.
func (r *Result) Next() bool {
	res := r.Rows.Next()
	if !res {
		return false
	}

	r.rows++
	if r.cap > 0 && r.rows > r.cap {
		return false
	}

	return true
}

// Err returns the error of the underlying rows.
func (r *Result) Err() error {
	err := r.Rows.Err()
	if err != nil {
		return err
	}

	if r.cap > 0 && r.rows > r.cap {
		return fmt.Errorf("result cap exceeded: returned more than %d rows", r.cap)
	}

	return nil
}

// Close wraps rows.Close and calls the Result's cleanup function (if it is set).
// Close should be idempotent.
func (r *Result) Close() error {
	firstErr := r.Rows.Close()
	if r.cleanupFn != nil {
		err := r.cleanupFn()
		if firstErr == nil {
			firstErr = err
		}

		// Prevent cleanupFn from being called multiple times.
		// NOTE: Not idempotent for error returned from cleanupFn.
		r.cleanupFn = nil
	}
	return firstErr
}

// InformationSchema contains information about existing tables in an OLAP driver.
// Table lookups should be case insensitive.
type InformationSchema interface {
	All(ctx context.Context, like string) ([]*Table, error)
	Lookup(ctx context.Context, db, schema, name string) (*Table, error)
}

// Table represents a table in an information schema.
type Table struct {
	Database                string
	DatabaseSchema          string
	IsDefaultDatabase       bool
	IsDefaultDatabaseSchema bool
	Name                    string
	View                    bool
	Schema                  *runtimev1.StructType
	UnsupportedCols         map[string]string
}

// IngestionSummary is details about ingestion
type IngestionSummary struct {
	BytesIngested int64
}

// IncrementalStrategy is a strategy to use for incrementally inserting data into a SQL table.
type IncrementalStrategy string

const (
	IncrementalStrategyUnspecified IncrementalStrategy = ""
	IncrementalStrategyAppend      IncrementalStrategy = "append"
	IncrementalStrategyMerge       IncrementalStrategy = "merge"
)

// Dialect enumerates OLAP query languages.
type Dialect int

const (
	DialectUnspecified Dialect = iota
	DialectDuckDB
	DialectDruid
	DialectClickHouse
	DialectPinot
)

func (d Dialect) String() string {
	switch d {
	case DialectUnspecified:
		return ""
	case DialectDuckDB:
		return "duckdb"
	case DialectDruid:
		return "druid"
	case DialectClickHouse:
		return "clickhouse"
	case DialectPinot:
		return "pinot"
	default:
		panic("not implemented")
	}
}

func (d Dialect) ContextKeyArgPrefix() string {
	return fmt.Sprintf("%s_context__", d.String())
}

func (d Dialect) CanPivot() bool {
	return d == DialectDuckDB
}

// EscapeIdentifier returns an escaped SQL identifier in the dialect.
func (d Dialect) EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(ident, "\"", "\"\"")) // nolint:gocritic // Because SQL escaping is different
}

func (d Dialect) EscapeStringValue(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}

func (d Dialect) ConvertToDateTruncSpecifier(grain runtimev1.TimeGrain) string {
	var str string
	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		str = "MILLISECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		str = "SECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		str = "MINUTE"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		str = "HOUR"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		str = "DAY"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		str = "WEEK"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		str = "MONTH"
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		str = "QUARTER"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		str = "YEAR"
	}

	if d == DialectClickHouse {
		return strings.ToLower(str)
	}
	return str
}

func (d Dialect) SupportsILike() bool {
	return d != DialectDruid && d != DialectPinot
}

// RequiresCastForLike returns true if the dialect requires an expression used in a LIKE or ILIKE condition to explicitly be cast to type TEXT.
func (d Dialect) RequiresCastForLike() bool {
	return d == DialectClickHouse
}

// EscapeTable returns an esacped fully qualified table name
func (d Dialect) EscapeTable(db, schema, table string) string {
	var sb strings.Builder
	if db != "" {
		sb.WriteString(d.EscapeIdentifier(db))
		sb.WriteString(".")
	}
	if schema != "" {
		sb.WriteString(d.EscapeIdentifier(schema))
		sb.WriteString(".")
	}
	sb.WriteString(d.EscapeIdentifier(table))
	return sb.String()
}

func (d Dialect) DimensionSelect(db, dbSchema, table string, dim *runtimev1.MetricsViewSpec_DimensionV2) (dimSelect, unnestClause string) {
	colName := d.EscapeIdentifier(dim.Name)
	if !dim.Unnest || d == DialectDruid {
		return fmt.Sprintf(`(%s) as %s`, d.MetricsViewDimensionExpression(dim), colName), ""
	}
	if dim.Unnest && d == DialectClickHouse {
		return fmt.Sprintf(`arrayJoin(%s) as %s`, d.MetricsViewDimensionExpression(dim), colName), ""
	}

	unnestColName := d.EscapeIdentifier(tempName(fmt.Sprintf("%s_%s_", "unnested", dim.Name)))
	unnestTableName := tempName("tbl")
	sel := fmt.Sprintf(`%s as %s`, unnestColName, colName)
	if dim.Expression == "" {
		// select "unnested_colName" as "colName" ... FROM "mv_table", LATERAL UNNEST("mv_table"."colName") tbl_name("unnested_colName") ...
		return sel, fmt.Sprintf(`, LATERAL UNNEST(%s.%s) %s(%s)`, d.EscapeTable(db, dbSchema, table), colName, unnestTableName, unnestColName)
	}

	return sel, fmt.Sprintf(`, LATERAL UNNEST(%s) %s(%s)`, dim.Expression, unnestTableName, unnestColName)
}

func (d Dialect) DimensionSelectPair(db, dbSchema, table string, dim *runtimev1.MetricsViewSpec_DimensionV2) (expr, alias, unnestClause string) {
	colName := d.EscapeIdentifier(dim.Name)
	if !dim.Unnest || d == DialectDruid {
		return d.MetricsViewDimensionExpression(dim), colName, ""
	}

	unnestColName := d.EscapeIdentifier(tempName(fmt.Sprintf("%s_%s_", "unnested", dim.Name)))
	unnestTableName := tempName("tbl")
	if dim.Expression == "" {
		// select "unnested_colName" as "colName" ... FROM "mv_table", LATERAL UNNEST("mv_table"."colName") tbl_name("unnested_colName") ...
		return unnestColName, colName, fmt.Sprintf(`, LATERAL UNNEST(%s.%s) %s(%s)`, d.EscapeTable(db, dbSchema, table), colName, unnestTableName, unnestColName)
	}

	return unnestColName, colName, fmt.Sprintf(`, LATERAL UNNEST(%s) %s(%s)`, dim.Expression, unnestTableName, unnestColName)
}

func (d Dialect) LateralUnnest(expr, tableAlias, colName string) (tbl string, auto bool, err error) {
	if d == DialectDruid || d == DialectPinot || d == DialectClickHouse {
		return "", true, nil
	}

	return fmt.Sprintf(`LATERAL UNNEST(%s) %s(%s)`, expr, tableAlias, d.EscapeIdentifier(colName)), false, nil
}

func (d Dialect) AutoUnnest(expr string) string {
	if d == DialectClickHouse {
		return fmt.Sprintf("arrayJoin(%s)", expr)
	}
	return expr
}

func (d Dialect) MetricsViewDimensionExpression(dimension *runtimev1.MetricsViewSpec_DimensionV2) string {
	if dimension.Expression != "" {
		return dimension.Expression
	}
	if dimension.Column != "" {
		return d.EscapeIdentifier(dimension.Column)
	}
	// Backwards compatibility for older projects that have not run reconcile on this metrics view.
	// In that case `column` will not be present.
	return d.EscapeIdentifier(dimension.Name)
}

func (d Dialect) SafeDivideExpression(numExpr, denExpr string) string {
	switch d {
	case DialectDruid:
		return fmt.Sprintf("SAFE_DIVIDE(%s, CAST(%s AS DOUBLE))", numExpr, denExpr)
	default:
		return fmt.Sprintf("(%s)/CAST(%s AS DOUBLE)", numExpr, denExpr)
	}
}

func (d Dialect) OrderByExpression(name string, desc bool) string {
	res := d.EscapeIdentifier(name)
	if desc {
		res += " DESC"
	}
	if d == DialectDuckDB {
		res += " NULLS LAST"
	}
	return res
}

func (d Dialect) JoinOnExpression(lhs, rhs string) string {
	if d == DialectClickHouse {
		return fmt.Sprintf("isNotDistinctFrom(%s, %s)", lhs, rhs)
	}
	return fmt.Sprintf("%s IS NOT DISTINCT FROM %s", lhs, rhs)
}

func (d Dialect) DateTruncExpr(dim *runtimev1.MetricsViewSpec_DimensionV2, grain runtimev1.TimeGrain, tz string, firstDayOfWeek, firstMonthOfYear int) (string, error) {
	if tz == "UTC" || tz == "Etc/UTC" {
		tz = ""
	}

	if tz != "" {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return "", fmt.Errorf("invalid time zone %q: %w", tz, err)
		}
	}

	var specifier string
	if tz != "" && d == DialectDruid {
		specifier = druidTimeFloorSpecifier(grain)
	} else {
		specifier = d.ConvertToDateTruncSpecifier(grain)
	}

	var expr string
	if dim.Expression != "" {
		expr = fmt.Sprintf("(%s)", dim.Expression)
	} else {
		expr = d.EscapeIdentifier(dim.Column)
	}

	switch d {
	case DialectDuckDB:
		var shift string
		if grain == runtimev1.TimeGrain_TIME_GRAIN_WEEK && firstDayOfWeek > 1 {
			offset := 8 - firstDayOfWeek
			shift = fmt.Sprintf("%d DAY", offset)
		} else if grain == runtimev1.TimeGrain_TIME_GRAIN_YEAR && firstMonthOfYear > 1 {
			offset := 13 - firstMonthOfYear
			shift = fmt.Sprintf("%d MONTH", offset)
		}

		if tz == "" {
			if shift == "" {
				return fmt.Sprintf("date_trunc('%s', %s::TIMESTAMP)::TIMESTAMP", specifier, expr), nil
			}
			return fmt.Sprintf("date_trunc('%s', %s::TIMESTAMP + INTERVAL %s)::TIMESTAMP - INTERVAL %s", specifier, expr, shift, shift), nil
		}

		// Optimization: date_trunc is faster for day+ granularity
		switch grain {
		case runtimev1.TimeGrain_TIME_GRAIN_DAY, runtimev1.TimeGrain_TIME_GRAIN_WEEK, runtimev1.TimeGrain_TIME_GRAIN_MONTH, runtimev1.TimeGrain_TIME_GRAIN_QUARTER, runtimev1.TimeGrain_TIME_GRAIN_YEAR:
			if shift == "" {
				return fmt.Sprintf("timezone('%s', date_trunc('%s', timezone('%s', %s::TIMESTAMPTZ)))::TIMESTAMP", tz, specifier, tz, expr), nil
			}
			return fmt.Sprintf("timezone('%s', date_trunc('%s', timezone('%s', %s::TIMESTAMPTZ) + INTERVAL %s) - INTERVAL %s)::TIMESTAMP", tz, specifier, tz, expr, shift, shift), nil
		}

		if shift == "" {
			return fmt.Sprintf("time_bucket(INTERVAL '1 %s', %s::TIMESTAMPTZ, '%s')", specifier, expr, tz), nil
		}
		return fmt.Sprintf("time_bucket(INTERVAL '1 %s', %s::TIMESTAMPTZ + INTERVAL %s, '%s') - INTERVAL %s", specifier, expr, shift, tz, shift), nil
	case DialectDruid:
		var shift int
		var shiftPeriod string
		if grain == runtimev1.TimeGrain_TIME_GRAIN_WEEK && firstDayOfWeek > 1 {
			shift = 8 - firstDayOfWeek
			shiftPeriod = "P1D"
		} else if grain == runtimev1.TimeGrain_TIME_GRAIN_YEAR && firstMonthOfYear > 1 {
			shift = 13 - firstMonthOfYear
			shiftPeriod = "P1M"
		}

		if tz == "" {
			if shift == 0 {
				return fmt.Sprintf("date_trunc('%s', %s)", specifier, expr), nil
			}
			return fmt.Sprintf("time_shift(date_trunc('%s', time_shift(%s, '%s', %d)), '%s', -%d)", specifier, expr, shiftPeriod, shift, shiftPeriod, shift), nil
		}

		if shift == 0 {
			return fmt.Sprintf("time_floor(%s, '%s', null, '%s')", expr, specifier, tz), nil
		}
		return fmt.Sprintf("time_shift(time_floor(time_shift(%s, '%s', %d), '%s', null, '%s'), '%s', -%d)", expr, shiftPeriod, shift, specifier, tz, shiftPeriod, shift), nil
	case DialectClickHouse:
		var shift string
		if grain == runtimev1.TimeGrain_TIME_GRAIN_WEEK && firstDayOfWeek > 1 {
			offset := 8 - firstDayOfWeek
			shift = fmt.Sprintf("%d DAY", offset)
		} else if grain == runtimev1.TimeGrain_TIME_GRAIN_YEAR && firstMonthOfYear > 1 {
			offset := 13 - firstMonthOfYear
			shift = fmt.Sprintf("%d MONTH", offset)
		}

		if tz == "" {
			if shift == "" {
				return fmt.Sprintf("date_trunc('%s', %s)::DateTime64", specifier, expr), nil
			}
			return fmt.Sprintf("date_trunc('%s', %s + INTERVAL %s)::DateTime64 - INTERVAL %s", specifier, expr, shift, shift), nil
		}

		if shift == "" {
			return fmt.Sprintf("date_trunc('%s', %s::DateTime64(6, '%s'))::DateTime64(6, '%s')", specifier, expr, tz, tz), nil
		}
		return fmt.Sprintf("date_trunc('%s', %s::DateTime64(6, '%s') + INTERVAL %s)::DateTime64(6, '%s') - INTERVAL %s", specifier, expr, tz, shift, tz, shift), nil
	case DialectPinot:
		// TODO: Handle tz instead of ignoring it.
		// TODO: Handle firstDayOfWeek and firstMonthOfYear. NOTE: We currently error when configuring these for Pinot in runtime/validate.go.
		return fmt.Sprintf("ToDateTime(date_trunc('%s', %s, 'MILLISECONDS', '%s'), 'yyyy-MM-dd''T''HH:mm:ss''Z''')", specifier, expr, tz), nil
	default:
		return "", fmt.Errorf("unsupported dialect %q", d)
	}
}

func (d Dialect) DateDiff(grain runtimev1.TimeGrain, t1, t2 time.Time) (string, error) {
	unit := d.ConvertToDateTruncSpecifier(grain)
	switch d {
	case DialectClickHouse:
		return fmt.Sprintf("DATEDIFF('%s', parseDateTimeBestEffort('%s'), parseDateTimeBestEffort('%s'))", unit, t1.Format(time.RFC3339), t2.Format(time.RFC3339)), nil
	case DialectDruid:
		return fmt.Sprintf("TIMESTAMPDIFF(%q, TIME_PARSE('%s'), TIME_PARSE('%s'))", unit, t1.Format(time.RFC3339), t2.Format(time.RFC3339)), nil
	case DialectDuckDB:
		return fmt.Sprintf("DATEDIFF('%s', TIMESTAMP '%s', TIMESTAMP '%s')", unit, t1.Format(time.RFC3339), t2.Format(time.RFC3339)), nil
	default:
		return "", fmt.Errorf("unsupported dialect %q", d)
	}
}

func druidTimeFloorSpecifier(grain runtimev1.TimeGrain) string {
	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return "PT0.001S"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return "PT1S"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return "PT1M"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return "PT1H"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return "P1D"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return "P1W"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return "P1M"
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return "P3M"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return "P1Y"
	}
	panic(fmt.Errorf("invalid time grain enum value %d", int(grain)))
}

func tempName(prefix string) string {
	return prefix + strings.ReplaceAll(uuid.New().String(), "-", "")
}
