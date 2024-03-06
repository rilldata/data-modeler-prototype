package druid

import (
	"context"
	"encoding/json"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestDruid_json_walker(t *testing.T) {
	j := `[["a", "b"], ["c", "d"], ["e", 1], ["k", 2]]`
	dec := json.NewDecoder(strings.NewReader(j))
	jw := JSONWalker{
		dec: dec,
	}
	require.True(t, jw.enterArray())
	require.True(t, jw.enterArray())
	a, err := jw.stringArrayValues()
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, a)

	require.True(t, jw.enterArray())
	a, err = jw.stringArrayValues()
	require.NoError(t, err)
	require.Equal(t, []string{"c", "d"}, a)

	require.True(t, jw.enterArray())
	k, err := jw.arrayValues()
	require.NoError(t, err)
	require.Equal(t, []any{"e", 1.0}, k)

	require.True(t, jw.enterArray())
	k, err = jw.arrayValues()
	require.NoError(t, err)
	require.Equal(t, []any{"k", 2.0}, k)

	require.NoError(t, jw.err)
}

func TestDriver_types(t *testing.T) {
	driver := &driversDriver{}
	handle, err := driver.Open(map[string]any{"pool_size": 2, "dsn": "http://localhost:8888/druid/v2/sql"}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
	require.True(t, ok)

	res, err := olap.Execute(context.Background(), &drivers.Statement{
		Query: `select 
		cast(1 as boolean) as bool1, 
		cast(1 as bigint) as bigint1, 
		timestamp '2021-01-01 00:00:00' as ts1,
		cast(1 as real) as double1,
		cast(1 as float) as float1,
		cast(1 as integer) as integer1,
		date '2023-01-01' as date1,
		`,
	})
	require.NoError(t, err)
	schema, err := rowsToSchema(res.Rows)
	require.NoError(t, err)
	require.True(t, len(schema.Fields) > 0)

	data, err := rowsToData(res)

	require.NoError(t, err)

	require.True(t, data[0].Fields["bool1"].GetBoolValue())
	require.Equal(t, 1.0, data[0].Fields["bigint1"].GetNumberValue())
	require.Equal(t, "2021-01-01T00:00:00Z", data[0].Fields["ts1"].GetStringValue())
	require.Equal(t, 1.0, data[0].Fields["double1"].GetNumberValue())
	require.Equal(t, 1.0, data[0].Fields["float1"].GetNumberValue())
	require.Equal(t, 1.0, data[0].Fields["integer1"].GetNumberValue())
	require.Equal(t, "2023-01-01T00:00:00.000Z", data[0].Fields["date1"].GetStringValue())
}

func TestDriver_array_type(t *testing.T) {
	driver := &driversDriver{}
	handle, err := driver.Open(map[string]any{"pool_size": 2, "dsn": "http://localhost:8888/druid/v2/sql"}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
	require.True(t, ok)

	res, err := olap.Execute(context.Background(), &drivers.Statement{
		Query: `select 
		array [1,2] as array1
		`,
	})
	require.NoError(t, err)
	schema, err := rowsToSchema(res.Rows)
	require.NoError(t, err)
	require.True(t, len(schema.Fields) > 0)

	data, err := rowsToData(res)

	require.NoError(t, err)

	require.Equal(t, 1.0, data[0].Fields["array1"].GetListValue().Values[0].GetNumberValue())
	require.Equal(t, 2.0, data[0].Fields["array1"].GetListValue().Values[1].GetNumberValue())
}

func TestDriver_json_type(t *testing.T) {
	driver := &driversDriver{}
	handle, err := driver.Open(map[string]any{"pool_size": 2, "dsn": "http://localhost:8888/druid/v2/sql"}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
	require.True(t, ok)

	res, err := olap.Execute(context.Background(), &drivers.Statement{
		Query: `select 
			json_object('a':'b') as json1 
		`,
	})
	require.NoError(t, err)
	schema, err := rowsToSchema(res.Rows)
	require.NoError(t, err)
	require.True(t, len(schema.Fields) > 0)

	data, err := rowsToData(res)

	require.NoError(t, err)

	require.Equal(t, "b", data[0].Fields["json1"].GetStructValue().Fields["a"].GetStringValue())
}

func TestDriver_error(t *testing.T) {
	driver := &driversDriver{}
	handle, err := driver.Open(map[string]any{"pool_size": 2, "dsn": "http://localhost:8888/druid/v2/sql"}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
	require.True(t, ok)

	_, err = olap.Execute(context.Background(), &drivers.Statement{
		Query: `select select`,
	})
	require.Error(t, err)
	require.True(t, strings.HasPrefix(err.Error(), `"error":"druidException"`))
}

func rowsToData(rows *drivers.Result) ([]*structpb.Struct, error) {
	var data []*structpb.Struct
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}

		rowStruct, err := pbutil.ToStruct(rowMap, rows.Schema)
		if err != nil {
			return nil, err
		}

		data = append(data, rowStruct)
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}
