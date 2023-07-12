package queries

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"reflect"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type ColumnTimeGrain struct {
	TableName  string
	ColumnName string
	Result     runtimev1.TimeGrain
}

var _ runtime.Query = &ColumnTimeGrain{}

func (q *ColumnTimeGrain) Key() string {
	return fmt.Sprintf("ColumnTimeGrain:%s:%s", q.TableName, q.ColumnName)
}

func (q *ColumnTimeGrain) Deps() []string {
	return []string{q.TableName}
}

func (q *ColumnTimeGrain) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: int64(reflect.TypeOf(q.Result).Size()),
	}
}

func (q *ColumnTimeGrain) UnmarshalResult(v any) error {
	res, ok := v.(runtimev1.TimeGrain)
	if !ok {
		return fmt.Errorf("ColumnTimeGrain: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnTimeGrain) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	sampleSize := int64(500000)
	cq := &TableCardinality{
		TableName: q.TableName,
	}
	err := rt.Query(ctx, instanceID, cq, priority)
	if err != nil {
		return err
	}
	var useSample string
	if sampleSize > cq.Result {
		useSample = ""
	} else {
		useSample = fmt.Sprintf("USING SAMPLE %d ROWS", sampleSize)
	}

	estimateSQL := fmt.Sprintf(`
      WITH cleaned_column AS (
          SELECT %s as cd
          from %s
          %s
      ),
      time_grains as (
      SELECT 
          approx_count_distinct(extract('years' from cd)) as year,
          approx_count_distinct(extract('months' from cd)) as month,
          approx_count_distinct(extract('dayofyear' from cd)) as dayofyear,
          approx_count_distinct(extract('dayofmonth' from cd)) as dayofmonth,
          min(cd = last_day(cd)) = TRUE as lastdayofmonth,
          approx_count_distinct(extract('weekofyear' from cd)) as weekofyear,
          approx_count_distinct(extract('dayofweek' from cd)) as dayofweek,
          approx_count_distinct(extract('hour' from cd)) as hour,
          approx_count_distinct(extract('minute' from cd)) as minute,
          approx_count_distinct(extract('second' from cd)) as second,
          approx_count_distinct(extract('millisecond' from cd) - extract('seconds' from cd) * 1000) as ms
      FROM cleaned_column
      )
      SELECT 
        COALESCE(
            case WHEN ms > 1 THEN 'MILLISECOND' else NULL END,
            CASE WHEN second > 1 THEN 'SECOND' else NULL END,
            CASE WHEN minute > 1 THEN 'MINUTE' else null END,
            CASE WHEN hour > 1 THEN 'HOUR' else null END,
            -- cases above, if equal to 1, then we have some candidates for
            -- bigger time grains. We need to reverse from here
            -- years, months, weeks, days.
            CASE WHEN dayofyear = 1 and year > 1 THEN 'YEAR' else null END,
            CASE WHEN (dayofmonth = 1 OR lastdayofmonth) and month > 1 THEN 'MONTH' else null END,
            CASE WHEN dayofweek = 1 and weekofyear > 1 THEN 'WEEK' else null END,
            CASE WHEN hour = 1 THEN 'DAY' else null END
        ) as estimatedSmallestTimeGrain
      FROM time_grains
      `,
		safeName(q.ColumnName),
		safeName(q.TableName),
		useSample,
	)

	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            estimateSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	var timeGrainString sql.NullString
	if rows.Next() {
		err := rows.Scan(&timeGrainString)
		if err != nil {
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	if !timeGrainString.Valid {
		q.Result = runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED // It's the default value. This is just to clarify intended behavior.
		return nil
	}

	q.Result = toTimeGrain(timeGrainString.String)
	return nil
}

func (q *ColumnTimeGrain) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
