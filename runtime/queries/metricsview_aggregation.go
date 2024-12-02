package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	metricssqlparser "github.com/rilldata/rill/runtime/pkg/metricssql"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

type MetricsViewAggregation struct {
	MetricsViewName     string                                         `json:"metrics_view,omitempty"`
	Dimensions          []*runtimev1.MetricsViewAggregationDimension   `json:"dimensions,omitempty"`
	Measures            []*runtimev1.MetricsViewAggregationMeasure     `json:"measures,omitempty"`
	Sort                []*runtimev1.MetricsViewAggregationSort        `json:"sort,omitempty"`
	TimeRange           *runtimev1.TimeRange                           `json:"time_range,omitempty"`
	ComparisonTimeRange *runtimev1.TimeRange                           `json:"comparison_time_range,omitempty"`
	Where               *runtimev1.Expression                          `json:"where,omitempty"`
	WhereSQL            string                                         `json:"where_sql,omitempty"`
	Having              *runtimev1.Expression                          `json:"having,omitempty"`
	HavingSQL           string                                         `json:"having_sql,omitempty"`
	Filter              *runtimev1.MetricsViewFilter                   `json:"filter,omitempty"` // Backwards compatibility
	Priority            int32                                          `json:"priority,omitempty"`
	Limit               *int64                                         `json:"limit,omitempty"`
	Offset              int64                                          `json:"offset,omitempty"`
	PivotOn             []string                                       `json:"pivot_on,omitempty"`
	SecurityClaims      *runtime.SecurityClaims                        `json:"security_claims,omitempty"`
	Aliases             []*runtimev1.MetricsViewComparisonMeasureAlias `json:"aliases,omitempty"`
	Exact               bool                                           `json:"exact,omitempty"`
	FillMissing         bool                                           `json:"fill_missing,omitempty"`

	Result    *runtimev1.MetricsViewAggregationResponse `json:"-"`
	Exporting bool                                      `json:"-"` // Deprecated: Remove when tests call Export directly
}

var _ runtime.Query = &MetricsViewAggregation{}

func (q *MetricsViewAggregation) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewAggregation:%s", string(r))
}

func (q *MetricsViewAggregation) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewAggregation) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewAggregation) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewAggregationResponse)
	if !ok {
		return fmt.Errorf("MetricsViewAggregation: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewAggregation) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityClaims)
	if err != nil {
		return err
	}

	qry, err := q.rewriteToMetricsViewQuery(q.Exporting)
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}

	e, err := metricsview.NewExecutor(ctx, rt, instanceID, mv, security, priority)
	if err != nil {
		return err
	}
	defer e.Close()

	res, err := e.Query(ctx, qry, nil)
	if err != nil {
		return err
	}
	defer res.Close()

	data, err := rowsToData(res)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewAggregationResponse{
		Schema: res.Schema,
		Data:   data,
	}
	return nil
}

func (q *MetricsViewAggregation) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	filename := strings.ReplaceAll(q.MetricsViewName, `"`, `_`)
	if !isTimeRangeNil(q.TimeRange) || q.Where != nil || q.Having != nil || q.WhereSQL != "" || q.HavingSQL != "" {
		filename += "_filtered"
	}

	// Resolve metrics view
	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityClaims)
	if err != nil {
		return err
	}

	// Route to metricsview executor
	qry, err := q.rewriteToMetricsViewQuery(true)
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}

	e, err := metricsview.NewExecutor(ctx, rt, instanceID, mv, security, opts.Priority)
	if err != nil {
		return err
	}
	defer e.Close()

	var format drivers.FileFormat
	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		format = drivers.FileFormatCSV
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		format = drivers.FileFormatXLSX
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		format = drivers.FileFormatParquet
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format.String())
	}

	path, err := e.Export(ctx, qry, nil, format)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(path) }()

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

func (q *MetricsViewAggregation) rewriteToMetricsViewQuery(export bool) (*metricsview.Query, error) {
	qry := &metricsview.Query{MetricsView: q.MetricsViewName}
	var computedTimeDims []*metricsview.Dimension
	for _, d := range q.Dimensions {
		res := metricsview.Dimension{Name: d.Name}
		if d.Alias != "" {
			res.Name = d.Alias
		}
		if d.TimeZone != "" {
			qry.TimeZone = d.TimeZone
		}
		if d.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			res.Compute = &metricsview.DimensionCompute{
				TimeFloor: &metricsview.DimensionComputeTimeFloor{
					Dimension: d.Name,
					Grain:     metricsview.TimeGrainFromProto(d.TimeGrain),
				},
			}
			computedTimeDims = append(computedTimeDims, &res)
		}
		qry.Dimensions = append(qry.Dimensions, res)
	}

	var measureFilter *runtimev1.Expression

	for _, m := range q.Measures {
		res := metricsview.Measure{Name: m.Name}
		switch m.BuiltinMeasure {
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT:
			res.Compute = &metricsview.MeasureCompute{Count: true}
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT_DISTINCT:
			res.Compute = &metricsview.MeasureCompute{CountDistinct: &metricsview.MeasureComputeCountDistinct{
				Dimension: m.BuiltinMeasureArgs[0].GetStringValue(),
			}}
		}

		if m.Filter != nil {
			if len(q.Measures) > 1 {
				return nil, fmt.Errorf("measure-level filter is not supported when multiple measures are present")
			}
			measureFilter = m.Filter
		}

		if m.Compute != nil {
			switch c := m.Compute.(type) {
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue:
				res.Compute = &metricsview.MeasureCompute{ComparisonValue: &metricsview.MeasureComputeComparisonValue{
					Measure: c.ComparisonValue.Measure,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta:
				res.Compute = &metricsview.MeasureCompute{ComparisonDelta: &metricsview.MeasureComputeComparisonDelta{
					Measure: c.ComparisonDelta.Measure,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
				res.Compute = &metricsview.MeasureCompute{ComparisonRatio: &metricsview.MeasureComputeComparisonRatio{
					Measure: c.ComparisonRatio.Measure,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_PercentOfTotal:
				res.Compute = &metricsview.MeasureCompute{PercentOfTotal: &metricsview.MeasureComputePercentOfTotal{
					Measure: c.PercentOfTotal.Measure,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_Uri:
				res.Compute = &metricsview.MeasureCompute{URI: &metricsview.MeasureComputeURI{
					Dimension: c.Uri.Dimension,
				}}
			}
		}

		qry.Measures = append(qry.Measures, res)
	}

	qry.PivotOn = q.PivotOn

	for _, s := range q.Sort {
		qry.Sort = append(qry.Sort, metricsview.Sort{
			Name: s.Name,
			Desc: s.Desc,
		})
	}

	if q.TimeRange != nil {
		res := &metricsview.TimeRange{}
		if q.TimeRange.Start != nil {
			res.Start = q.TimeRange.Start.AsTime()
		}
		if q.TimeRange.End != nil {
			res.End = q.TimeRange.End.AsTime()
		}
		res.IsoDuration = q.TimeRange.IsoDuration
		res.IsoOffset = q.TimeRange.IsoOffset
		res.RoundToGrain = metricsview.TimeGrainFromProto(q.TimeRange.RoundToGrain)
		qry.TimeRange = res
	}

	if q.ComparisonTimeRange != nil {
		res := &metricsview.TimeRange{}
		if q.ComparisonTimeRange.Start != nil {
			res.Start = q.ComparisonTimeRange.Start.AsTime()
		}
		if q.ComparisonTimeRange.End != nil {
			res.End = q.ComparisonTimeRange.End.AsTime()
		}
		res.IsoDuration = q.ComparisonTimeRange.IsoDuration
		res.IsoOffset = q.ComparisonTimeRange.IsoOffset
		res.RoundToGrain = metricsview.TimeGrainFromProto(q.ComparisonTimeRange.RoundToGrain)
		qry.ComparisonTimeRange = res
	}

	if q.Filter != nil { // Backwards compatibility
		if q.Where != nil {
			return nil, fmt.Errorf("both filter and where is provided")
		}
		q.Where = convertFilterToExpression(q.Filter)
	}

	var err error
	qry.Where, err = metricViewExpression(q.Where, q.WhereSQL)
	if err != nil {
		return nil, err
	}

	// If a measure-level filter is present, we set qry.Where as the spine, and use (qry.Where AND measuresFilter) as the new where clause
	if measureFilter != nil {
		qry.Spine = &metricsview.Spine{Where: &metricsview.WhereSpine{Expression: qry.Where}}

		if qry.Where == nil {
			qry.Where = metricsview.NewExpressionFromProto(measureFilter)
		} else {
			qry.Where = &metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator: metricsview.OperatorAnd,
					Expressions: []*metricsview.Expression{
						qry.Where,
						metricsview.NewExpressionFromProto(measureFilter),
					},
				},
			}
		}
	}

	// If there is only one time dimension and null fill is enabled, we set the spine to the time range
	if q.FillMissing && len(computedTimeDims) > 0 {
		if qry.Spine != nil {
			// should we silently ignore instead of error ?
			return nil, fmt.Errorf("cannot have both where and time spine")
		}
		if len(computedTimeDims) > 1 {
			return nil, fmt.Errorf("cannot have multiple time spines")
		}
		if q.TimeRange == nil || q.TimeRange.Start == nil || q.TimeRange.End == nil {
			return nil, fmt.Errorf("time range is required for null fill")
		}

		timeDim := computedTimeDims[0]
		if timeDim.Compute.TimeFloor.Grain == metricsview.TimeGrainUnspecified {
			return nil, fmt.Errorf("time grain is required for null fill")
		}

		tz, err := time.LoadLocation(qry.TimeZone)
		if err != nil {
			return nil, fmt.Errorf("invalid time zone %q: %w", qry.TimeZone, err)
		}

		qry.Spine = &metricsview.Spine{}
		s := timeutil.TruncateTime(q.TimeRange.Start.AsTime(), timeDim.Compute.TimeFloor.Grain.ToTimeutil(), tz, 1, 1)
		e := q.TimeRange.End.AsTime()
		qry.Spine.TimeRange = &metricsview.TimeSpine{
			Start: s,
			End:   e,
			Grain: timeDim.Compute.TimeFloor.Grain,
			Alias: timeDim.Name,
		}
	}

	qry.Having, err = metricViewExpression(q.Having, q.HavingSQL)
	if err != nil {
		return nil, err
	}

	if q.Limit != nil {
		if *q.Limit == 0 {
			tmp := int64(100)
			q.Limit = &tmp
		}
		qry.Limit = q.Limit
	}

	if q.Offset != 0 {
		qry.Offset = &q.Offset
	}

	if len(q.PivotOn) > 0 {
		qry.PivotOn = q.PivotOn
	}

	qry.UseDisplayNames = export

	return qry, nil
}

func metricViewExpression(expr *runtimev1.Expression, sql string) (*metricsview.Expression, error) {
	if expr != nil && sql != "" {
		sqlExpr, err := metricssqlparser.ParseSQLFilter(sql)
		if err != nil {
			return nil, err
		}
		return &metricsview.Expression{
			Condition: &metricsview.Condition{
				Operator: metricsview.OperatorAnd,
				Expressions: []*metricsview.Expression{
					metricsview.NewExpressionFromProto(expr),
					sqlExpr,
				},
			},
		}, nil
	}
	if expr != nil {
		return metricsview.NewExpressionFromProto(expr), nil
	}
	if sql != "" {
		return metricssqlparser.ParseSQLFilter(sql)
	}
	return nil, nil
}
