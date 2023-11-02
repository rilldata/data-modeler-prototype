package queries_test

import (
	"context"
	// "fmt"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestMetricsViewsTimeseries_month_grain(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_year"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"max_clicks"},
		MetricsViewName: "timeseries_year",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-01-01T00:00:00Z"),
		TimeEnd:         parseTime(t, "2024-01-01T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
		Limit:           250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	i := 0
	require.Equal(t, parseTime(t, "2023-01-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-02-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-04-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-05-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-06-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-07-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-08-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-09-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-10-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-12-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewsTimeseries_month_grain_IST(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_year"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"max_clicks"},
		MetricsViewName: "timeseries_year",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2022-12-31T18:30:00Z"),
		TimeEnd:         parseTime(t, "2024-01-31T18:30:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
		TimeZone:        "Asia/Kolkata",
		Limit:           250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	i := 0
	require.Equal(t, parseTime(t, "2022-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-01-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-02-28T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-04-30T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-05-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-06-30T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-07-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-08-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-09-30T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-10-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-30T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewsTimeseries_quarter_grain_IST(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_year"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"max_clicks"},
		MetricsViewName: "timeseries_year",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2022-12-31T18:30:00Z"),
		TimeEnd:         parseTime(t, "2024-01-31T18:30:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_QUARTER,
		TimeZone:        "Asia/Kolkata",
		Limit:           250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	i := 0
	require.Equal(t, parseTime(t, "2022-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-06-30T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-09-30T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewsTimeseries_year_grain_IST(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_year"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"max_clicks"},
		MetricsViewName: "timeseries_year",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2022-12-31T18:30:00Z"),
		TimeEnd:         parseTime(t, "2024-12-31T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		TimeZone:        "Asia/Kolkata",
		Limit:           250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	i := 0
	require.Equal(t, parseTime(t, "2022-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestTruncateTime(t *testing.T) {
	require.Equal(t, parseTestTime(t, "2019-01-07T04:20:07Z"), queries.TruncateTime(parseTestTime(t, "2019-01-07T04:20:07.29Z"), runtimev1.TimeGrain_TIME_GRAIN_SECOND, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-07T04:20:00Z"), queries.TruncateTime(parseTestTime(t, "2019-01-07T04:20:07Z"), runtimev1.TimeGrain_TIME_GRAIN_MINUTE, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-07T04:00:00Z"), queries.TruncateTime(parseTestTime(t, "2019-01-07T04:20:01Z"), runtimev1.TimeGrain_TIME_GRAIN_HOUR, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-07T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2019-01-07T04:20:01Z"), runtimev1.TimeGrain_TIME_GRAIN_DAY, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-10-09T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-10T04:20:01Z"), runtimev1.TimeGrain_TIME_GRAIN_WEEK, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-01T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2019-01-07T01:01:01Z"), runtimev1.TimeGrain_TIME_GRAIN_MONTH, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-04-01T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2019-05-07T01:01:01Z"), runtimev1.TimeGrain_TIME_GRAIN_QUARTER, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-01T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2019-02-07T01:01:01Z"), runtimev1.TimeGrain_TIME_GRAIN_YEAR, time.UTC, 1, 1))
}

func TestTruncateTime_Kathmandu(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Kathmandu")
	require.NoError(t, err)
	require.Equal(t, parseTestTime(t, "2019-01-07T04:20:07Z"), queries.TruncateTime(parseTestTime(t, "2019-01-07T04:20:07.29Z"), runtimev1.TimeGrain_TIME_GRAIN_SECOND, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-07T04:20:00Z"), queries.TruncateTime(parseTestTime(t, "2019-01-07T04:20:07Z"), runtimev1.TimeGrain_TIME_GRAIN_MINUTE, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-07T04:15:00Z"), queries.TruncateTime(parseTestTime(t, "2019-01-07T04:20:01Z"), runtimev1.TimeGrain_TIME_GRAIN_HOUR, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-06T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2019-01-07T04:20:01Z"), runtimev1.TimeGrain_TIME_GRAIN_DAY, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-10-08T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-10T04:20:01Z"), runtimev1.TimeGrain_TIME_GRAIN_WEEK, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-31T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2019-02-07T01:01:01Z"), runtimev1.TimeGrain_TIME_GRAIN_MONTH, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-03-31T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2019-05-07T01:01:01Z"), runtimev1.TimeGrain_TIME_GRAIN_QUARTER, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2018-12-31T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2019-02-07T01:01:01Z"), runtimev1.TimeGrain_TIME_GRAIN_YEAR, tz, 1, 1))
}

func TestTruncateTime_UTC_first_day(t *testing.T) {
	tz := time.UTC
	require.Equal(t, parseTestTime(t, "2023-10-08T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-10T04:20:01Z"), runtimev1.TimeGrain_TIME_GRAIN_WEEK, tz, 7, 1))
	require.Equal(t, parseTestTime(t, "2023-10-10T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-10T04:20:01Z"), runtimev1.TimeGrain_TIME_GRAIN_WEEK, tz, 2, 1))
	require.Equal(t, parseTestTime(t, "2023-10-10T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-11T04:20:01Z"), runtimev1.TimeGrain_TIME_GRAIN_WEEK, tz, 2, 1))
	require.Equal(t, parseTestTime(t, "2023-10-10T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-10T00:01:01Z"), runtimev1.TimeGrain_TIME_GRAIN_WEEK, tz, 2, 1))
}

func TestTruncateTime_Kathmandu_first_day(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Kathmandu")
	require.NoError(t, err)
	require.Equal(t, parseTestTime(t, "2023-10-07T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-10T04:20:01Z"), runtimev1.TimeGrain_TIME_GRAIN_WEEK, tz, 7, 1))
	require.Equal(t, parseTestTime(t, "2023-10-09T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-10T04:20:01Z"), runtimev1.TimeGrain_TIME_GRAIN_WEEK, tz, 2, 1))
	require.Equal(t, parseTestTime(t, "2023-10-09T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-11T04:20:01Z"), runtimev1.TimeGrain_TIME_GRAIN_WEEK, tz, 2, 1))
	require.Equal(t, parseTestTime(t, "2023-10-09T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-09T18:16:01Z"), runtimev1.TimeGrain_TIME_GRAIN_WEEK, tz, 2, 1))
}

func TestTruncateTime_UTC_first_month(t *testing.T) {
	tz := time.UTC
	require.Equal(t, parseTestTime(t, "2023-02-01T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-01T00:20:00Z"), runtimev1.TimeGrain_TIME_GRAIN_YEAR, tz, 2, 2))
	require.Equal(t, parseTestTime(t, "2023-03-01T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-01T00:20:00Z"), runtimev1.TimeGrain_TIME_GRAIN_YEAR, tz, 2, 3))
	require.Equal(t, parseTestTime(t, "2023-03-01T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2023-03-01T00:20:00Z"), runtimev1.TimeGrain_TIME_GRAIN_YEAR, tz, 2, 3))
	require.Equal(t, parseTestTime(t, "2022-12-01T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-01T00:20:00Z"), runtimev1.TimeGrain_TIME_GRAIN_YEAR, tz, 2, 12))
	require.Equal(t, parseTestTime(t, "2023-01-01T00:00:00Z"), queries.TruncateTime(parseTestTime(t, "2023-01-01T00:20:00Z"), runtimev1.TimeGrain_TIME_GRAIN_YEAR, tz, 2, 1))
}

func TestTruncateTime_Kathmandu_first_month(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Kathmandu")
	require.NoError(t, err)
	require.Equal(t, parseTestTime(t, "2023-01-31T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-02T00:20:00Z"), runtimev1.TimeGrain_TIME_GRAIN_YEAR, tz, 2, 2))
	require.Equal(t, parseTestTime(t, "2023-02-28T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-02T00:20:00Z"), runtimev1.TimeGrain_TIME_GRAIN_YEAR, tz, 2, 3))
	require.Equal(t, parseTestTime(t, "2023-02-28T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2023-03-02T00:20:00Z"), runtimev1.TimeGrain_TIME_GRAIN_YEAR, tz, 2, 3))
	require.Equal(t, parseTestTime(t, "2022-11-30T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2023-10-02T00:20:00Z"), runtimev1.TimeGrain_TIME_GRAIN_YEAR, tz, 2, 12))
	require.Equal(t, parseTestTime(t, "2022-12-31T18:15:00Z"), queries.TruncateTime(parseTestTime(t, "2023-01-02T00:20:00Z"), runtimev1.TimeGrain_TIME_GRAIN_YEAR, tz, 2, 1))
}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}
