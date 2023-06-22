package queries

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type ColumnCardinality struct {
	TableName  string
	ColumnName string
	Result     float64
}

var _ runtime.Query = &ColumnCardinality{}

func (q *ColumnCardinality) Key() string {
	return fmt.Sprintf("ColumnCardinality:%s:%s", q.TableName, q.ColumnName)
}

func (q *ColumnCardinality) Deps() []string {
	return []string{q.TableName}
}

func (q *ColumnCardinality) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: int64(reflect.TypeOf(q.Result).Size()),
	}
}

func (q *ColumnCardinality) UnmarshalResult(v any) error {
	res, ok := v.(float64)
	if !ok {
		return fmt.Errorf("ColumnCardinality: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnCardinality) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	requestSQL := fmt.Sprintf("SELECT approx_count_distinct(%s) as count from %s", safeName(q.ColumnName), safeName(q.TableName))

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            requestSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	var count float64
	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return err
		}
		q.Result = count
		return nil
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return errors.New("no rows returned")
}

func (q *ColumnCardinality) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int, format runtimev1.ExportFormat, w io.Writer) error {
	return ErrExportNotSupported
}
