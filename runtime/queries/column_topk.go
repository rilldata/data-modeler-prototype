package queries

import (
	"context"
	"database/sql"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/structpb"
)

type ColumnTopK struct {
	TableName  string
	ColumnName string
	Agg        string
	K          int
	Result     *runtimev1.TopK
}

var _ runtime.Query = &ColumnTopK{}

func (q *ColumnTopK) Key() string {
	return fmt.Sprintf("ColumnTopK:%s:%s:%s:%d", q.TableName, q.ColumnName, q.Agg, q.K)
}

func (q *ColumnTopK) Deps() []string {
	return []string{q.TableName}
}

func (q *ColumnTopK) MarshalResult() any {
	return q.Result
}

func (q *ColumnTopK) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.TopK)
	if !ok {
		return fmt.Errorf("topk: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnTopK) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	// Get OLAP connection
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	// Check dialect
	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	// Build SQL
	qry := fmt.Sprintf("SELECT %s AS value, %s AS count FROM %s GROUP BY %s ORDER BY count DESC, value ASC LIMIT %d",
		quoteName(q.ColumnName),
		q.Agg,
		quoteName(q.TableName),
		quoteName(q.ColumnName),
		q.K,
	)

	// Run query
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    qry,
		Priority: priority,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	// Scan result
	res := &runtimev1.TopK{}
	for rows.Next() {
		entry := &runtimev1.TopK_Entry{}
		var val sql.NullString
		err = rows.Scan(&val, &entry.Count)
		if err != nil {
			return err
		}
		if val.Valid {
			entry.Value = structpb.NewStringValue(val.String)
		} else {
			entry.Value = structpb.NewNullValue()
		}
		res.Entries = append(res.Entries, entry)
	}

	// Save result
	q.Result = res
	return nil
}
