package server

import (
	"context"
	"testing"
	"time"

	"github.com/marcboeker/go-duckdb"
	"github.com/stretchr/testify/require"
	structpb "google.golang.org/protobuf/types/known/structpb"

	"fmt"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
)

func CreateSimpleTimeseriesTable(server *Server, instanceId string, t *testing.T, tableName string) *drivers.Result {
	result, err := server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "create table " + quoteName(tableName) + " (clicks double, time timestamp, device varchar)",
	})
	require.NoError(t, err)
	result.Close()
	result, _ = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "insert into " + quoteName(tableName) + " values (1.0, '2019-01-01 00:00:00', 'android'), (1.0, '2019-01-02 00:00:00', 'iphone')",
	})
	require.NoError(t, err)
	result.Close()
	result, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select count(*) from " + quoteName(tableName),
	})
	require.NoError(t, err)
	return result
}

func TestServer_Timeseries(t *testing.T) {
	server, instanceId, err := getTestServer(t)
	require.NoError(t, err)
	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))

	mx := "max"
	response, err := server.GenerateTimeSeries(context.Background(), &api.GenerateTimeSeriesRequest{
		InstanceId: instanceId,
		TableName:  "timeseries",
		Measures: &api.GenerateTimeSeriesRequest_BasicMeasures{
			BasicMeasures: []*api.BasicMeasureDefinition{
				{
					Expression: "max(clicks)",
					SqlName:    &mx,
				},
			},
		},
		TimestampColumnName: "time",
		TimeRange: &api.TimeSeriesTimeRange{
			Start:    "2019-01-01",
			End:      "2019-12-01",
			Interval: api.TimeGrain_YEAR,
		},
		Filters: &api.MetricsViewRequestFilter{
			Include: []*api.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	// printResults(results)
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["max"])
}

func TestServer_Timeseries_2measures(t *testing.T) {
	server, instanceId, err := getTestServer(t)
	require.NoError(t, err)
	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))

	mx := "max"
	sm := "sum"
	response, err := server.GenerateTimeSeries(context.Background(), &api.GenerateTimeSeriesRequest{
		InstanceId: instanceId,
		TableName:  "timeseries",
		Measures: &api.GenerateTimeSeriesRequest_BasicMeasures{
			BasicMeasures: []*api.BasicMeasureDefinition{
				{
					Expression: "max(clicks)",
					SqlName:    &mx,
				},
				{
					Expression: "sum(clicks)",
					SqlName:    &sm,
				},
			},
		},
		TimestampColumnName: "time",
		TimeRange: &api.TimeSeriesTimeRange{
			Start:    "2019-01-01",
			End:      "2019-12-01",
			Interval: api.TimeGrain_YEAR,
		},
		Filters: &api.MetricsViewRequestFilter{
			Include: []*api.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	// printResults(results)
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["max"])
	require.Equal(t, 2.0, results[0].Records["sum"])
}

func TestServer_Timeseries_1dim(t *testing.T) {
	server, instanceId, err := getTestServer(t)
	require.NoError(t, err)
	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))

	sm := "sum"
	response, err := server.GenerateTimeSeries(context.Background(), &api.GenerateTimeSeriesRequest{
		InstanceId: instanceId,
		TableName:  "timeseries",
		Measures: &api.GenerateTimeSeriesRequest_BasicMeasures{
			BasicMeasures: []*api.BasicMeasureDefinition{
				{
					Expression: "sum(clicks)",
					SqlName:    &sm,
				},
			},
		},
		TimestampColumnName: "time",
		TimeRange: &api.TimeSeriesTimeRange{
			Start:    "2019-01-01",
			End:      "2019-12-01",
			Interval: api.TimeGrain_YEAR,
		},
		Filters: &api.MetricsViewRequestFilter{
			Include: []*api.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	// printResults(results)
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["sum"])
}

func printResults(results []*api.TimeSeriesValue) {
	for _, result := range results {
		fmt.Printf("%v ", result.Ts)
		for k, value := range result.Records {
			fmt.Printf("%v:%v ", k, value)
		}
		fmt.Println()
	}
}

func TestServer_Timeseries_1day(t *testing.T) {
	server, instanceId, err := getTestServer(t)
	require.NoError(t, err)
	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))

	mx := "max"
	response, err := server.GenerateTimeSeries(context.Background(), &api.GenerateTimeSeriesRequest{
		InstanceId: instanceId,
		TableName:  "timeseries",
		Measures: &api.GenerateTimeSeriesRequest_BasicMeasures{
			BasicMeasures: []*api.BasicMeasureDefinition{
				{
					Expression: "max(clicks)",
					SqlName:    &mx,
				},
			},
		},
		TimestampColumnName: "time",
		TimeRange: &api.TimeSeriesTimeRange{
			Start:    "2019-01-01",
			End:      "2019-01-02",
			Interval: api.TimeGrain_DAY,
		},
		Filters: &api.MetricsViewRequestFilter{
			Include: []*api.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 2, len(results))
}

func TestServer_RangeSanity(t *testing.T) {
	server, instanceId, err := getTestServer(t)
	require.NoError(t, err)
	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	result.Close()
	result, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select min(time) min, max(time) max, max(time)-min(time) as r from timeseries",
	})
	require.NoError(t, err)
	var min, max time.Time
	var r duckdb.Interval
	result.Next()
	err = result.Scan(&min, &max, &r)
	require.NoError(t, err)
	require.Equal(t, time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC), min)
	require.Equal(t, int32(1), r.Days)
}

func TestServer_normaliseRanger(t *testing.T) {
	server, instanceId, err := getTestServer(t)
	require.NoError(t, err)
	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))
	r := &api.TimeSeriesTimeRange{
		Interval: api.TimeGrain_UNSPECIFIED,
	}
	r, err = server.normaliseTimeRange(context.Background(), &api.GenerateTimeSeriesRequest{
		InstanceId:          instanceId,
		TimeRange:           r,
		TableName:           "timeseries",
		TimestampColumnName: "time",
	})
	require.NoError(t, err)
	require.Equal(t, "2019-01-01 00:00:00", r.Start)
	require.Equal(t, "2019-01-02 00:00:00", r.End)
	require.Equal(t, api.TimeGrain_HOUR, r.Interval)
}
