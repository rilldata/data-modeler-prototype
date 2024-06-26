package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewToplist struct {
	MetricsViewName    string                                `json:"metrics_view_name,omitempty"`
	DimensionName      string                                `json:"dimension_name,omitempty"`
	MeasureNames       []string                              `json:"measure_names,omitempty"`
	TimeStart          *timestamppb.Timestamp                `json:"time_start,omitempty"`
	TimeEnd            *timestamppb.Timestamp                `json:"time_end,omitempty"`
	Limit              *int64                                `json:"limit,omitempty"`
	Offset             int64                                 `json:"offset,omitempty"`
	Sort               []*runtimev1.MetricsViewSort          `json:"sort,omitempty"`
	Where              *runtimev1.Expression                 `json:"where,omitempty"`
	Filter             *runtimev1.MetricsViewFilter          `json:"filter,omitempty"` // backwards compatibility
	Having             *runtimev1.Expression                 `json:"having,omitempty"`
	SecurityAttributes map[string]any                        `json:"security_attributes,omitempty"`
	SecurityPolicy     *runtimev1.MetricsViewSpec_SecurityV2 `json:"security_policy,omitempty"`

	Result *runtimev1.MetricsViewToplistResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewToplist{}

func (q *MetricsViewToplist) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewToplist:%s", r)
}

func (q *MetricsViewToplist) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewToplist) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewToplist) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewToplistResponse)
	if !ok {
		return fmt.Errorf("MetricsViewToplist: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewToplist) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	ds := []*runtimev1.MetricsViewAggregationDimension{{Name: q.DimensionName}}
	ms := make([]*runtimev1.MetricsViewAggregationMeasure, len(q.MeasureNames))
	for i, m := range q.MeasureNames {
		ms[i] = &runtimev1.MetricsViewAggregationMeasure{Name: m}
	}

	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityAttributes, q.SecurityPolicy, ds, ms)
	if err != nil {
		return err
	}

	qry, err := q.rewriteToMetricsViewQuery(false)
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}

	e, err := metricsview.NewExecutor(ctx, rt, instanceID, mv, security, priority)
	if err != nil {
		return err
	}
	defer e.Close()

	res, _, err := e.Query(ctx, qry, nil)
	if err != nil {
		return err
	}
	defer res.Close()

	data, err := rowsToData(res)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewToplistResponse{
		Meta: structTypeToMetricsViewColumn(res.Schema),
		Data: data,
	}
	return nil
}

func (q *MetricsViewToplist) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	ds := []*runtimev1.MetricsViewAggregationDimension{{Name: q.DimensionName}}
	ms := make([]*runtimev1.MetricsViewAggregationMeasure, len(q.MeasureNames))
	for i, m := range q.MeasureNames {
		ms[i] = &runtimev1.MetricsViewAggregationMeasure{Name: m}
	}

	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityAttributes, q.SecurityPolicy, ds, ms)
	if err != nil {
		return err
	}

	qry, err := q.rewriteToMetricsViewQuery(true)
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}

	e, err := metricsview.NewExecutor(ctx, rt, instanceID, mv, security, opts.Priority)
	if err != nil {
		return err
	}
	defer e.Close()

	var format string
	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		format = "csv"
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		format = "xlsx"
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		format = "parquet"
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format.String())
	}

	path, err := e.Export(ctx, qry, nil, format)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(path) }()

	filename := q.generateFilename()
	err = opts.PreWriteHook(filename)
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}

	return nil
}

func (q *MetricsViewToplist) rewriteToMetricsViewQuery(export bool) (*metricsview.Query, error) {
	qry := &metricsview.Query{MetricsView: q.MetricsViewName}

	qry.Dimensions = append(qry.Dimensions, metricsview.Dimension{Name: q.DimensionName})

	for _, m := range q.MeasureNames {
		qry.Measures = append(qry.Measures, metricsview.Measure{Name: m})
	}

	if q.TimeStart != nil || q.TimeEnd != nil {
		res := &metricsview.TimeRange{}
		if q.TimeStart != nil {
			res.Start = q.TimeStart.AsTime()
		}
		if q.TimeEnd != nil {
			res.End = q.TimeEnd.AsTime()
		}
		qry.TimeRange = res
	}

	if q.Limit != nil {
		qry.Limit = q.Limit
	}

	if q.Offset != 0 {
		qry.Offset = &q.Offset
	}

	for _, s := range q.Sort {
		qry.Sort = append(qry.Sort, metricsview.Sort{
			Name: s.Name,
			Desc: !s.Ascending,
		})
	}

	if q.Filter != nil { // backwards backwards compatibility
		if q.Where != nil {
			return nil, fmt.Errorf("both filter and where is provided")
		}
		q.Where = convertFilterToExpression(q.Filter)
	}

	if q.Where != nil {
		qry.Where = metricsview.NewExpressionFromProto(q.Where)
	}

	if q.Having != nil {
		qry.Having = metricsview.NewExpressionFromProto(q.Having)
	}

	qry.Label = export

	return qry, nil
}

func (q *MetricsViewToplist) generateFilename() string {
	filename := strings.ReplaceAll(q.MetricsViewName, `"`, `_`)
	filename += "_" + q.DimensionName
	if q.TimeStart != nil || q.TimeEnd != nil || q.Where != nil || q.Having != nil {
		filename += "_filtered"
	}
	return filename
}
