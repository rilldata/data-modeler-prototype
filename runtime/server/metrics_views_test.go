package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

func createServerWithMetricsView(t *testing.T) (*Server, string) {
	metastore, err := drivers.Open("sqlite", "file:rill?mode=memory&cache=shared")
	require.NoError(t, err)

	err = metastore.Migrate(context.Background())
	require.NoError(t, err)

	server, err := NewServer(&ServerOptions{
		ConnectionCacheSize: 100,
	}, metastore, nil)
	require.NoError(t, err)

	resp, err := server.CreateInstance(context.Background(), &runtimev1.CreateInstanceRequest{
		OlapDriver:   "duckdb",
		OlapDsn:      "",
		RepoDriver:   "file",
		RepoDsn:      "../testproject/ad_bids",
		EmbedCatalog: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Instance.InstanceId)

	rr, err := server.Reconcile(context.Background(), &runtimev1.ReconcileRequest{
		InstanceId: resp.Instance.InstanceId,
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(rr.Errors))

	return server, resp.Instance.InstanceId
}

func TestServer_LookupMetricsView(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)

	mv, err := server.lookupMetricsView(context.Background(), instanceId, "ad_bids_metrics")
	require.NoError(t, err)
	require.Equal(t, 2, len(mv.Measures))
	require.Equal(t, 2, len(mv.Dimensions))
}

func TestServer_MetricsViewTotals(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_2measures(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)
	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0", "measure_1"},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
	require.Equal(t, 8.0, tr.Data.Fields["measure_1"].GetNumberValue())
}

func TestServer_MetricsViewTotals_TimeStart(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		TimeStart:       parseTime(t, "2022-01-02T00:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_TimeEnd(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		TimeEnd:         parseTime(t, "2022-01-02T00:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("msn.com"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_2In(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("msn.com"),
						structpb.NewStringValue("yahoo.com"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_2dim(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("yahoo.com"),
					},
				},
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewStringValue("Yahoo"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_like(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					Like: []*structpb.Value{
						structpb.NewStringValue("%com"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_in_and_like(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("yahoo"),
					},
					Like: []*structpb.Value{
						structpb.NewStringValue("%com"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_include_and_exclude(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					Like: []*structpb.Value{
						structpb.NewStringValue("%com"),
					},
				},
			},
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("yahoo.com"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_null(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewNullValue(),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_include_and_exclude_in_and_like(t *testing.T) {
	server, instanceId := createServerWithMetricsView(t)

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("msn.com"),
					},
					Like: []*structpb.Value{
						structpb.NewStringValue("%yahoo%"),
					},
				},
			},
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewNullValue(),
					},
					Like: []*structpb.Value{
						structpb.NewStringValue("Y%"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
}
