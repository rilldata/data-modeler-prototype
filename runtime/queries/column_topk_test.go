package queries

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestColumnTopK(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 'abc' AS col, 1 AS val, TIMESTAMP '2022-11-01 00:00:00' AS times 
		UNION ALL 
		SELECT 'def' AS col, 5 AS val, TIMESTAMP '2022-11-02 00:00:00' AS times
		UNION ALL 
		SELECT 'abc' AS col, 3 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times
		UNION ALL 
		SELECT null AS col, 1 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times
	`)

	q := &ColumnTopK{
		TableName:  "test",
		ColumnName: "col",
		Agg:        "count(*)",
		K:          50,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Equal(t, 3, len(q.Result.Entries))
	require.Equal(t, "abc", *q.Result.Entries[0].Value)
	require.Equal(t, 2, int(q.Result.Entries[0].Count))
	require.Nil(t, q.Result.Entries[1].Value)
	require.Equal(t, 1, int(q.Result.Entries[1].Count))
	require.Equal(t, "def", *q.Result.Entries[2].Value)
	require.Equal(t, 1, int(q.Result.Entries[2].Count))

	q = &ColumnTopK{
		TableName:  "test",
		ColumnName: "col",
		Agg:        "sum(val)",
		K:          50,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Equal(t, 3, len(q.Result.Entries))
	require.Equal(t, "def", *q.Result.Entries[0].Value)
	require.Equal(t, 5, int(q.Result.Entries[0].Count))
	require.Equal(t, "abc", *q.Result.Entries[1].Value)
	require.Equal(t, 4, int(q.Result.Entries[1].Count))
	require.Nil(t, q.Result.Entries[2].Value)
	require.Equal(t, 1, int(q.Result.Entries[2].Count))

	q = &ColumnTopK{
		TableName:  "test",
		ColumnName: "col",
		Agg:        "count(*)",
		K:          1,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Equal(t, 1, len(q.Result.Entries))
	require.Equal(t, "abc", *q.Result.Entries[0].Value)
	require.Equal(t, 2, int(q.Result.Entries[0].Count))
}

func BenchmarkColumnTopK(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &ColumnTopK{
			TableName:  "ad_bids",
			ColumnName: "domain",
			Agg:        "sum(bid_price)",
			K:          50,
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}
