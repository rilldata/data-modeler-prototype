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
	EstimateSize() (int64, bool)

	CreateTableAsSelect(ctx context.Context, name string, view bool, sql string) error
	InsertTableAsSelect(ctx context.Context, name, sql string, byName, inPlace bool, strategy IncrementalStrategy, uniqueKey []string) error
	DropTable(ctx context.Context, name string, view bool) error
	RenameTable(ctx context.Context, name, newName string, view bool) error
	AddTableColumn(ctx context.Context, tableName, columnName string, typ string) error
	AlterTableColumn(ctx context.Context, tableName, columnName string, newType string) error
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
}

// SetCleanupFunc sets a function, which will be called when the Result is closed.
func (r *Result) SetCleanupFunc(fn func() error) {
	if r.cleanupFn != nil {
		panic("cleanup function already set")
	}
	r.cleanupFn = fn
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
	All(ctx context.Context) ([]*Table, error)
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

// EscapeIdentifier returns an escaped SQL identifier in the dialect.
func (d Dialect) EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(ident, "\"", "\"\""))
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

	unnestColName := d.EscapeIdentifier(tempName(fmt.Sprintf("%s_%s_", "unnested", dim.Name)))
	unnestTableName := tempName("tbl")
	sel := fmt.Sprintf(`%s as %s`, unnestColName, colName)
	if dim.Expression == "" {
		// select "unnested_colName" as "colName" ... FROM "mv_table", LATERAL UNNEST("mv_table"."colName") tbl_name("unnested_colName") ...
		return sel, fmt.Sprintf(`, LATERAL UNNEST(%s.%s) %s(%s)`, d.EscapeTable(db, dbSchema, table), colName, unnestTableName, unnestColName)
	}

	return sel, fmt.Sprintf(`, LATERAL UNNEST(%s) %s(%s)`, dim.Expression, unnestTableName, unnestColName)
}

func (d Dialect) MetricsViewDimensionExpression(dimension *runtimev1.MetricsViewSpec_DimensionV2) string {
	if dimension.Expression != "" {
		return dimension.Expression
	}
	if dimension.Column != "" {
		return d.EscapeIdentifier(dimension.Column)
	}
	// backwards compatibility for older projects that have not run reconcile on this dashboard
	// in that case `column` will not be present
	return d.EscapeIdentifier(dimension.Name)
}

func (d Dialect) SafeDivideExpression(numExpr, denExpr string) string {
	switch d {
	case DialectDruid:
		return fmt.Sprintf("SAFE_DIVIDE(%s, %s)", numExpr, denExpr)
	default:
		return fmt.Sprintf("CAST((%s) AS DOUBLE)/%s", numExpr, denExpr)
	}
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
				return fmt.Sprintf("date_trunc('%s', %s)", specifier, expr), nil
			}
			return fmt.Sprintf("date_trunc('%s', %s + INTERVAL %s) - INTERVAL %s", specifier, expr, shift, shift), nil
		}

		// TODO: Should this use date_trunc(grain, expr, tz) instead?
		if shift == "" {
			return fmt.Sprintf("toTimezone(date_trunc('%s', toTimezone(%s::TIMESTAMP, '%s')), '%s')", grain, expr, tz, tz), nil
		}
		return fmt.Sprintf("toTimezone(date_trunc('%s', toTimezone(%s::TIMESTAMP, '%s') + INTERVAL %s) - INTERVAL %s, '%s')", grain, expr, tz, shift, shift, tz), nil
	case DialectPinot:
		// TODO: Handle tz instead of ignoring it.
		// TODO: Handle firstDayOfWeek and firstMonthOfYear. NOTE: We currently error when configuring these for Pinot in runtime/validate.go.
		return fmt.Sprintf("ToDateTime(date_trunc('%s', %s, 'MILLISECONDS', '%s'), 'yyyy-MM-dd''T''HH:mm:ss''Z''')", specifier, expr, tz), nil
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
