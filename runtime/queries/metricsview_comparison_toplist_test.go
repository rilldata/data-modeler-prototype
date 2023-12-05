package queries_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/timestamppb"

	// Register drivers
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
)

func TestMetricsViewsComparison_dim_order_comparison_toplist_vs_general_toplist(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView().Spec

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		MetricsView: mv,
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "dom",
				Type: runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_UNSPECIFIED,
				Desc: false,
			},
		},
		Limit: 10,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	dims := make([]string, 0, 10)
	previous := ""
	for i, r := range q.Result.Rows {
		if i == 10 {
			break
		}

		require.Equal(t, -1, strings.Compare(previous, r.DimensionValue.GetStringValue()))
		previous = r.DimensionValue.GetStringValue()
		dims = append(dims, r.DimensionValue.GetStringValue())
	}

	q = &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		MetricsView: mv,
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(maxTime),
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "dom",
				Type: runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE,
				Desc: false,
			},
		},
		Limit: 10,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	comparisonDims := make([]string, 0, 10)
	for i, r := range q.Result.Rows {
		if i == 10 {
			break
		}
		comparisonDims = append(comparisonDims, r.DimensionValue.GetStringValue())
	}
	require.Equal(t, dims, comparisonDims)
}

func TestMetricsViewsComparison_dim_order(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		MetricsView: mv.Spec,
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "dom",
				Type: runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_UNSPECIFIED,
				Desc: true,
			},
		},
		Limit: 250,
		MeasureFilter: &runtimev1.MeasureFilter{
			Entry: &runtimev1.MeasureFilter_MeasureFilterEntry{
				MeasureFilterEntry: &runtimev1.MeasureFilterEntry{
					Measure:    &runtimev1.MetricsViewAggregationMeasure{Name: "measure_1"},
					Expression: " > 3.25",
				},
			},
		},
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Len(t, q.Result.Rows, 3)
	require.NotEmpty(t, "sports.yahoo.com", q.Result.Rows[0].DimensionValue)
	require.NotEmpty(t, "news.google.com", q.Result.Rows[1].DimensionValue)
	require.NotEmpty(t, "instagram.com", q.Result.Rows[2].DimensionValue)
}

func TestMetricsViewsComparison_measure_order(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		MetricsView: mv.Spec,
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(maxTime),
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_1",
				Type: runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE,
				Desc: true,
			},
		},
		Limit: 250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.NotEmpty(t, "facebook.com", q.Result.Rows[0].DimensionValue)
	require.NotEmpty(t, "msn.com", q.Result.Rows[1].DimensionValue)
}

func TestMetricsViewsComparison_measure_filters(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		MetricsView: mv.Spec,
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(maxTime),
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_1",
				Type: runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE,
				Desc: true,
			},
		},
		Limit: 250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fmt.Println(q.Result.Rows)
	require.NotEmpty(t, "facebook.com", q.Result.Rows[0].DimensionValue)
	require.NotEmpty(t, "msn.com", q.Result.Rows[1].DimensionValue)
}

func TestMetricsViewsCompariso_export_xlsx(t *testing.T) {
	t.Parallel()
	rt, instanceId := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceId, 0)
	require.NoError(t, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	ctrl, err := rt.Controller(context.Background(), instanceId)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		MetricsView: mv.Spec,
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "domain",
				Type: runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_UNSPECIFIED,
				Desc: false,
			},
		},
		Limit: 10,
	}

	var buf bytes.Buffer

	err = q.Export(context.Background(), rt, instanceId, &buf, &runtime.ExportOptions{
		Format: runtimev1.ExportFormat_EXPORT_FORMAT_XLSX,
	})
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	rows, err := file.GetRows("Sheet1")
	require.NoError(t, err)

	require.Equal(t, 2, len(rows))
	require.Equal(t, "Domain Label", rows[0][0])
	require.Equal(t, "Total volume", rows[0][1])
}

func TestServer_MetricsViewTimeseries_export_csv(t *testing.T) {
	t.Parallel()
	rt, instanceId := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceId, 0)
	require.NoError(t, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	ctrl, err := rt.Controller(context.Background(), instanceId)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		MetricsView: mv.Spec,
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "domain",
				Type: runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_UNSPECIFIED,
				Desc: false,
			},
		},
		Limit: 10,
	}

	var buf bytes.Buffer

	err = q.Export(context.Background(), rt, instanceId, &buf, &runtime.ExportOptions{
		Format: runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
	})
	require.NoError(t, err)

	str := string(buf.Bytes())
	require.Equal(t, 2, strings.Count(str, "\n"))
	rowStrings := strings.Split(str, "\n")
	require.Equal(t, "Domain Label,Total volume", rowStrings[0])
}
