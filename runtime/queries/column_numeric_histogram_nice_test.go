package queries_test

import (
	"context"
	"math"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func nice(start, stop, count float64) []float64 {
	return niceAndStepArray(start, stop, count)[:2]
}

func niceAndStepArray(start, stop, count float64) []float64 {
	x, x1, x2 := queries.NiceAndStep(start, stop, count)
	return []float64{x, x1, x2}
}

func TestTimeseries_normaliseTimeRange_Specified1(t *testing.T) {
	rt, instanceID := instanceWith2RowsModel(t)

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			Start:    parseTime(t, "2018-01-01T00:00:00Z"),
		},
	}

	r, err := q.ResolveNormaliseTimeRange(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2018-01-01T00:00:00Z").AsTime(), r.Start.AsTime())
	require.Equal(t, parseTime(t, "2020-01-01T00:00:00.000Z").AsTime(), r.End.AsTime())
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_YEAR, r.Interval)
}

func TestNice_AnyNaN(t *testing.T) {
	Equal(t, []float64{math.NaN(), 1}, nice(math.NaN(), 1, 1))
	Equal(t, []float64{0, math.NaN()}, nice(0, math.NaN(), 1))
	Equal(t, []float64{0, 1}, nice(0, 1, math.NaN()))
	Equal(t, []float64{math.NaN(), math.NaN()}, nice(math.NaN(), math.NaN(), 1))
	Equal(t, []float64{0, math.NaN()}, nice(0, math.NaN(), math.NaN()))
	Equal(t, []float64{math.NaN(), 1}, nice(math.NaN(), 1, math.NaN()))
	Equal(t, []float64{math.NaN(), math.NaN()}, nice(math.NaN(), math.NaN(), math.NaN()))
}

func TestNice_StartStopEqual(t *testing.T) {
	Equal(t, []float64{1, 1}, nice(1, 1, -1))
	Equal(t, []float64{1, 1}, nice(1, 1, 0))
	Equal(t, []float64{1, 1}, nice(1, 1, math.NaN()))
	Equal(t, []float64{1, 1}, nice(1, 1, 1))
	Equal(t, []float64{1, 1}, nice(1, 1, 10))
}

func TestNice_NotPositiveCount(t *testing.T) {
	Equal(t, []float64{0, 1}, nice(0, 1, -1))
	Equal(t, []float64{0, 1}, nice(0, 1, 0))
}

func TestNice_InfinityCount(t *testing.T) {
	Equal(t, []float64{0, 1}, nice(0, 1, math.Inf(1)))
	Equal(t, []float64{0, 1}, nice(0, 1, math.Inf(-1)))
}

func TestNice_ExpectedValues(t *testing.T) {
	Equal(t, []float64{0.132, 0.876, 1 / 1000.0}, niceAndStepArray(0.132, 0.876, 1000))
	Equal(t, []float64{0.13, 0.88, 1 / 100.0}, niceAndStepArray(0.132, 0.876, 100))
	Equal(t, []float64{0.12, 0.88, 1 / 50.0}, niceAndStepArray(0.132, 0.876, 30))
	Equal(t, []float64{0.1, 0.9, 1 / 10.0}, niceAndStepArray(0.132, 0.876, 10))
	Equal(t, []float64{0.1, 0.9, 1 / 10.0}, niceAndStepArray(0.132, 0.876, 6))
	Equal(t, []float64{0, 1, 1 / 5.0}, niceAndStepArray(0.132, 0.876, 5))
	Equal(t, []float64{0, 1, 1 / 5.0}, niceAndStepArray(0.132, 0.876, 4))
	Equal(t, []float64{0, 1, 1 / 2.0}, niceAndStepArray(0.132, 0.876, 3))
	Equal(t, []float64{0, 1, 1 / 2.0}, niceAndStepArray(0.132, 0.876, 2))
	Equal(t, []float64{0, 1, 1}, niceAndStepArray(0.132, 0.876, 1))
	Equal(t, []float64{0.132, 0.876, 0}, niceAndStepArray(0.132, 0.876, 0))
	Equal(t, []float64{0.132, 0.876, 0}, niceAndStepArray(0.132, 0.876, -1))

	Equal(t, []float64{132, 876, 1}, niceAndStepArray(132, 876, 1000))
	Equal(t, []float64{130, 880, 10}, niceAndStepArray(132, 876, 100))
	Equal(t, []float64{120, 880, 20}, niceAndStepArray(132, 876, 30))
	Equal(t, []float64{100, 900, 100}, niceAndStepArray(132, 876, 10))
	Equal(t, []float64{100, 900, 100}, niceAndStepArray(132, 876, 6))
	Equal(t, []float64{0, 1000, 200}, niceAndStepArray(132, 876, 5))
	Equal(t, []float64{0, 1000, 200}, niceAndStepArray(132, 876, 4))
	Equal(t, []float64{0, 1000, 500}, niceAndStepArray(132, 876, 3))
	Equal(t, []float64{0, 1000, 500}, niceAndStepArray(132, 876, 2))
	Equal(t, []float64{0, 1000, 1000}, niceAndStepArray(132, 876, 1))
	Equal(t, []float64{132, 876, 0}, niceAndStepArray(132, 876, 0))
	Equal(t, []float64{132, 876, 0}, niceAndStepArray(132, 876, -1))

	Equal(t, []float64{0.132, 0.876, 0}, niceAndStepArray(0.132, 0.876, math.NaN()))
	Equal(t, []float64{0.132, 0.876, 0}, niceAndStepArray(0.132, 0.876, math.Inf(1)))
	Equal(t, []float64{0.132, 0.876, 0}, niceAndStepArray(0.132, 0.876, math.Inf(-1)))
}

func Equal(t *testing.T, expected []float64, actual []float64) {
	if len(expected) != len(actual) {
		t.Errorf("\n%s\nExpected:\n %v but got:\n %v", strings.Join(assert.CallerInfo()[1:], "\n\t\t\t"), expected, actual)
		t.FailNow()
	}

	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] && (!math.IsNaN(expected[i]) || !math.IsNaN(actual[i])) {
			t.Errorf("\n%s\nExpected:\n %v but got:\n %v", strings.Join(assert.CallerInfo()[1:], "\n\t\t\t"), expected, actual)

			t.FailNow()
		}
	}
}
