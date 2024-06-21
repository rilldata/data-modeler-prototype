package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
)

type MetricsViewSchema struct {
	MetricsViewName    string                                `json:"metrics_view_name,omitempty"`
	SecurityAttributes map[string]any                        `json:"security_attributes,omitempty"`
	SecurityPolicy     *runtimev1.MetricsViewSpec_SecurityV2 `json:"security_policy,omitempty"`

	Result *runtimev1.MetricsViewSchemaResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewSchema{}

func (q *MetricsViewSchema) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewSchema:%s", string(r))
}

func (q *MetricsViewSchema) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewSchema) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewSchema) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewSchemaResponse)
	if !ok {
		return fmt.Errorf("MetricsViewSchema: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewSchema) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	// Resolve metrics view
	mv, _, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityAttributes, q.SecurityPolicy, nil, nil)
	if err != nil {
		return err
	}

	e, err := metricsview.NewExecutor(ctx, rt, instanceID, mv, nil, priority)
	if err != nil {
		return err
	}
	defer e.Close()

	schema, err := e.Schema(ctx)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewSchemaResponse{
		Schema: schema,
	}

	return nil
}

func (q *MetricsViewSchema) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return nil
}
