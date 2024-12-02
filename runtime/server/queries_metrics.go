package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricsViewAggregation implements QueryService.
func (s *Server) MetricsViewAggregation(ctx context.Context, req *runtimev1.MetricsViewAggregationRequest) (*runtimev1.MetricsViewAggregationResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsView),
		attribute.StringSlice("args.dimensions.names", marshalMetricsViewAggregationDimension(req.Dimensions)),
		attribute.StringSlice("args.measures.names", marshalMetricsViewAggregationMeasures(req.Measures)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewAggregationSort(req.Sort)),
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.Int64("args.limit", req.Limit),
		attribute.Int64("args.offset", req.Offset),
		attribute.Int("args.priority", int(req.Priority)),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	tr := req.TimeRange
	if req.TimeStart != nil || req.TimeEnd != nil {
		tr = &runtimev1.TimeRange{
			Start: req.TimeStart,
			End:   req.TimeEnd,
		}
	}

	q := &queries.MetricsViewAggregation{
		MetricsViewName:     req.MetricsView,
		Dimensions:          req.Dimensions,
		Measures:            req.Measures,
		Sort:                req.Sort,
		TimeRange:           tr,
		ComparisonTimeRange: req.ComparisonTimeRange,
		Where:               req.Where,
		WhereSQL:            req.WhereSql,
		Having:              req.Having,
		HavingSQL:           req.HavingSql,
		Filter:              req.Filter,
		Limit:               &req.Limit,
		Offset:              req.Offset,
		PivotOn:             req.PivotOn,
		SecurityClaims:      auth.GetClaims(ctx).SecurityClaims(),
		Exact:               req.Exact,
		Aliases:             req.Aliases,
		FillMissing:         req.FillMissing,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// MetricsViewToplist implements QueryService.
func (s *Server) MetricsViewToplist(ctx context.Context, req *runtimev1.MetricsViewToplistRequest) (*runtimev1.MetricsViewToplistResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.String("args.dimension", req.DimensionName),
		attribute.StringSlice("args.measures", req.MeasureNames),
		attribute.Int64("args.limit", req.Limit),
		attribute.Int64("args.offset", req.Offset),
		attribute.Int("args.priority", int(req.Priority)),
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewSort(req.Sort)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	q := &queries.MetricsViewToplist{
		MetricsViewName: req.MetricsViewName,
		DimensionName:   req.DimensionName,
		MeasureNames:    req.MeasureNames,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Limit:           &req.Limit,
		Offset:          req.Offset,
		Sort:            req.Sort,
		Where:           req.Where,
		WhereSQL:        req.WhereSql,
		Having:          req.Having,
		HavingSQL:       req.HavingSql,
		Filter:          req.Filter,
		SecurityClaims:  auth.GetClaims(ctx).SecurityClaims(),
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// MetricsViewComparison implements QueryService.
func (s *Server) MetricsViewComparison(ctx context.Context, req *runtimev1.MetricsViewComparisonRequest) (*runtimev1.MetricsViewComparisonResponse, error) {
	measureNames := make([]string, 0, len(req.Measures))
	for _, m := range req.Measures {
		measureNames = append(measureNames, m.Name)
	}
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.String("args.dimension", req.Dimension.Name),
		attribute.StringSlice("args.measures", measureNames),
		attribute.StringSlice("args.comparison_measures", req.ComparisonMeasures),
		attribute.StringSlice("args.sort.names", marshalMetricsViewComparisonSort(req.Sort)),
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.Int64("args.limit", req.Limit),
		attribute.Int64("args.offset", req.Offset),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if req.TimeRange != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.base_time_range.start", safeTimeStr(req.TimeRange.Start)))
		observability.AddRequestAttributes(ctx, attribute.String("args.base_time_range.end", safeTimeStr(req.TimeRange.End)))
	}
	if req.ComparisonTimeRange != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.comparison_time_range.start", safeTimeStr(req.ComparisonTimeRange.Start)))
		observability.AddRequestAttributes(ctx, attribute.String("args.comparison_time_range.end", safeTimeStr(req.ComparisonTimeRange.End)))
	}

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	q := &queries.MetricsViewComparison{
		MetricsViewName:     req.MetricsViewName,
		DimensionName:       req.Dimension.Name,
		Measures:            req.Measures,
		ComparisonMeasures:  req.ComparisonMeasures,
		TimeRange:           req.TimeRange,
		ComparisonTimeRange: req.ComparisonTimeRange,
		Limit:               req.Limit,
		Offset:              req.Offset,
		Sort:                req.Sort,
		Where:               req.Where,
		WhereSQL:            req.WhereSql,
		Having:              req.Having,
		HavingSQL:           req.HavingSql,
		Exact:               req.Exact,
		Filter:              req.Filter,
		SecurityClaims:      auth.GetClaims(ctx).SecurityClaims(),
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// MetricsViewTimeSeries implements QueryService.
func (s *Server) MetricsViewTimeSeries(ctx context.Context, req *runtimev1.MetricsViewTimeSeriesRequest) (*runtimev1.MetricsViewTimeSeriesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.StringSlice("args.measures", req.MeasureNames),
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.String("args.time_granularity", req.TimeGranularity.String()),
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	q := &queries.MetricsViewTimeSeries{
		MetricsViewName: req.MetricsViewName,
		MeasureNames:    req.MeasureNames,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		TimeGranularity: req.TimeGranularity,
		Where:           req.Where,
		WhereSQL:        req.WhereSql,
		Having:          req.Having,
		HavingSQL:       req.HavingSql,
		TimeZone:        req.TimeZone,
		Filter:          req.Filter,
		SecurityClaims:  auth.GetClaims(ctx).SecurityClaims(),
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return q.Result, nil
}

// MetricsViewTotals implements QueryService.
func (s *Server) MetricsViewTotals(ctx context.Context, req *runtimev1.MetricsViewTotalsRequest) (*runtimev1.MetricsViewTotalsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.StringSlice("args.measures", req.MeasureNames),
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	q := &queries.MetricsViewTotals{
		MetricsViewName: req.MetricsViewName,
		MeasureNames:    req.MeasureNames,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Where:           req.Where,
		WhereSQL:        req.WhereSql,
		Filter:          req.Filter,
		SecurityClaims:  auth.GetClaims(ctx).SecurityClaims(),
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return q.Result, nil
}

// MetricsViewRows implements QueryService.
func (s *Server) MetricsViewRows(ctx context.Context, req *runtimev1.MetricsViewRowsRequest) (*runtimev1.MetricsViewRowsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.String("args.time_granularity", req.TimeGranularity.String()),
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewSort(req.Sort)),
		attribute.Int("args.limit", int(req.Limit)),
		attribute.Int64("args.offset", req.Offset),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	mv, security, err := resolveMVAndSecurity(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	limit := int64(req.Limit)

	q := &queries.MetricsViewRows{
		MetricsViewName:    req.MetricsViewName,
		TimeStart:          req.TimeStart,
		TimeEnd:            req.TimeEnd,
		TimeGranularity:    req.TimeGranularity,
		Where:              req.Where,
		Sort:               req.Sort,
		Limit:              &limit,
		Offset:             req.Offset,
		TimeZone:           req.TimeZone,
		MetricsView:        mv,
		ResolvedMVSecurity: security,
		Filter:             req.Filter,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// MetricsViewTimeRange implements QueryService.
func (s *Server) MetricsViewTimeRange(ctx context.Context, req *runtimev1.MetricsViewTimeRangeRequest) (*runtimev1.MetricsViewTimeRangeResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	mv, security, err := resolveMVAndSecurity(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTimeRange{
		MetricsViewName:    req.MetricsViewName,
		MetricsView:        mv,
		ResolvedMVSecurity: security,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

func (s *Server) MetricsViewSchema(ctx context.Context, req *runtimev1.MetricsViewSchemaRequest) (*runtimev1.MetricsViewSchemaResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	q := &queries.MetricsViewSchema{
		MetricsViewName: req.MetricsViewName,
		SecurityClaims:  auth.GetClaims(ctx).SecurityClaims(),
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

func (s *Server) MetricsViewSearch(ctx context.Context, req *runtimev1.MetricsViewSearchRequest) (*runtimev1.MetricsViewSearchResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.StringSlice("args.dimensions.names", req.Dimensions),
		attribute.String("args.search", req.Search),
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	limit := int64(req.Limit)
	q := &queries.MetricsViewSearch{
		MetricsViewName: req.MetricsViewName,
		Dimensions:      req.Dimensions,
		Search:          req.Search,
		TimeRange:       req.TimeRange,
		Where:           req.Where,
		Having:          req.Having,
		Priority:        req.Priority,
		Limit:           &limit,
		SecurityClaims:  auth.GetClaims(ctx).SecurityClaims(),
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

func resolveMVAndSecurity(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName string) (*runtimev1.MetricsViewSpec, *runtime.ResolvedSecurity, error) {
	res, mv, err := lookupMetricsView(ctx, rt, instanceID, metricsViewName)
	if err != nil {
		return nil, nil, err
	}

	resolvedSecurity, err := rt.ResolveSecurity(instanceID, auth.GetClaims(ctx).SecurityClaims(), res)
	if err != nil {
		return nil, nil, err
	}
	if !resolvedSecurity.CanAccess() {
		return nil, nil, ErrForbidden
	}

	return mv, resolvedSecurity, nil
}

func resolveMVAndSecurityFromAttributes(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName string, claims *runtime.SecurityClaims) (*runtimev1.MetricsViewSpec, *runtime.ResolvedSecurity, error) {
	res, mv, err := lookupMetricsView(ctx, rt, instanceID, metricsViewName)
	if err != nil {
		return nil, nil, err
	}

	resolvedSecurity, err := rt.ResolveSecurity(instanceID, claims, res)
	if err != nil {
		return nil, nil, err
	}

	if !resolvedSecurity.CanAccess() {
		return nil, nil, ErrForbidden
	}

	return mv, resolvedSecurity, nil
}

// returns the metrics view and the time the catalog was last updated
func lookupMetricsView(ctx context.Context, rt *runtime.Runtime, instanceID, name string) (*runtimev1.Resource, *runtimev1.MetricsViewSpec, error) {
	ctrl, err := rt.Controller(ctx, instanceID)
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name}, false)
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mv := res.GetMetricsView()
	spec := mv.State.ValidSpec
	if spec == nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "metrics view %q is invalid", name)
	}

	return res, spec, nil
}
