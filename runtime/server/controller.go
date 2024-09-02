package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// timeLayoutUnseparated formats an absolute timestamp as a string with millisecond precision without any separators.
// E.g. for "2006-01-02T15:04:05.999Z" it outputs "200601021504059999".
const timeLayoutUnseparated = "200601021504059999"

// ListResources implements runtimev1.RuntimeServiceServer
func (s *Server) ListResources(ctx context.Context, req *runtimev1.ListResourcesRequest) (*runtimev1.ListResourcesResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.kind", req.Kind),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rs, err := ctrl.List(ctx, req.Kind, req.Path, false)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	slices.SortFunc(rs, func(a, b *runtimev1.Resource) int {
		an := a.Meta.Name
		bn := b.Meta.Name
		if an.Kind < bn.Kind {
			return -1
		}
		if an.Kind > bn.Kind {
			return 1
		}
		return strings.Compare(an.Name, bn.Name)
	})

	i := 0
	for i < len(rs) {
		r := rs[i]
		r, access, err := s.applySecurityPolicy(ctx, req.InstanceId, r)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if !access {
			// Remove from the slice
			rs[i] = rs[len(rs)-1]
			rs[len(rs)-1] = nil
			rs = rs[:len(rs)-1]
			continue
		}
		rs[i] = r
		i++
	}

	return &runtimev1.ListResourcesResponse{Resources: rs}, nil
}

// WatchResources implements runtimev1.RuntimeServiceServer
func (s *Server) WatchResources(req *runtimev1.WatchResourcesRequest, ss runtimev1.RuntimeService_WatchResourcesServer) error {
	observability.AddRequestAttributes(ss.Context(),
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.kind", req.Kind),
	)

	if !auth.GetClaims(ss.Context()).CanInstance(req.InstanceId, auth.ReadObjects) {
		return ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ss.Context(), req.InstanceId)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if req.Replay {
		rs, err := ctrl.List(ss.Context(), req.Kind, "", false)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		for _, r := range rs {
			r, access, err := s.applySecurityPolicy(ss.Context(), req.InstanceId, r)
			if err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
			if !access {
				continue
			}

			err = ss.Send(&runtimev1.WatchResourcesResponse{
				Event:    runtimev1.ResourceEvent_RESOURCE_EVENT_WRITE,
				Resource: r,
			})
			if err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	return ctrl.Subscribe(ss.Context(), func(e runtimev1.ResourceEvent, n *runtimev1.ResourceName, r *runtimev1.Resource) {
		if r != nil { // r is nil for deletion events
			var access bool
			var err error
			r, access, err = s.applySecurityPolicy(ss.Context(), req.InstanceId, r)
			if err != nil {
				s.logger.Info("failed to apply security policy", zap.String("name", n.Name), zap.Error(err))
				return
			}
			if !access {
				return
			}
		}

		err = ss.Send(&runtimev1.WatchResourcesResponse{
			Event:    e,
			Name:     n,
			Resource: r,
		})
		if err != nil {
			s.logger.Info("failed to send resource event", zap.Error(err))
		}
	})
}

// GetResource implements runtimev1.RuntimeServiceServer
func (s *Server) GetResource(ctx context.Context, req *runtimev1.GetResourceRequest) (*runtimev1.GetResourceResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.name.kind", req.Name.Kind),
		attribute.String("args.name.name", req.Name.Name),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	r, err := ctrl.Get(ctx, req.Name, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	r, access, err := s.applySecurityPolicy(ctx, req.InstanceId, r)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !access {
		return nil, status.Error(codes.NotFound, "resource not found")
	}

	return &runtimev1.GetResourceResponse{Resource: r}, nil
}

// GetModelSplits implements runtimev1.RuntimeServiceServer
func (s *Server) GetModelSplits(ctx context.Context, req *runtimev1.GetModelSplitsRequest) (*runtimev1.GetModelSplitsResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.model", req.Model),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	n := &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: req.Model}
	r, err := ctrl.Get(ctx, n, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	r, access, err := s.applySecurityPolicy(ctx, req.InstanceId, r)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !access {
		return nil, status.Error(codes.NotFound, "resource not found")
	}

	splitsModelID := r.GetModel().State.SplitsModelId
	if splitsModelID == "" {
		return &runtimev1.GetModelSplitsResponse{}, nil
	}

	afterIdx := -1
	afterKey := ""
	if req.PageToken != "" {
		err := unmarshalPageToken(req.PageToken, &afterIdx, &afterKey)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse page token: %v", err)
		}
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer release()

	splits, err := catalog.FindModelSplits(ctx, splitsModelID, afterIdx, afterKey, validPageSize(req.PageSize))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var nextPageToken string
	if len(splits) == validPageSize(req.PageSize) {
		last := splits[len(splits)-1]
		nextPageToken = marshalPageToken(last.Index, last.Key)
	}

	return &runtimev1.GetModelSplitsResponse{
		Splits:        modelSplitsToPB(splits),
		NextPageToken: nextPageToken,
	}, nil
}

// CreateTrigger implements runtimev1.RuntimeServiceServer
func (s *Server) CreateTrigger(ctx context.Context, req *runtimev1.CreateTriggerRequest) (*runtimev1.CreateTriggerResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditTrigger) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Build refresh trigger spec
	spec := &runtimev1.RefreshTriggerSpec{
		Resources: req.Resources,
		Models:    req.Models,
	}

	// Handle the convenience flag for the project parser.
	if req.Parser {
		spec.Resources = append(spec.Resources, runtime.GlobalProjectParserName)
	}

	// Handle the convenience flags for all sources and models.
	if req.AllSourcesModels || req.AllSourcesModelsFull {
		// Add all sources.
		// Note: Don't need to handle "full" here since source refreshes are always full refreshes.
		rs, err := ctrl.List(ctx, runtime.ResourceKindSource, "", false)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to list sources: %w", err).Error())
		}
		for _, r := range rs {
			spec.Resources = append(spec.Resources, r.Meta.Name)
		}

		// Add all models.
		rs, err = ctrl.List(ctx, runtime.ResourceKindModel, "", false)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to list models: %w", err).Error())
		}
		for _, r := range rs {
			spec.Models = append(spec.Models, &runtimev1.RefreshModelTrigger{
				Model: r.Meta.Name.Name,
				Full:  req.AllSourcesModelsFull,
			})
		}
	}

	// Create the trigger resource
	name := fmt.Sprintf("trigger_adhoc_%s", time.Now().Format(timeLayoutUnseparated))
	n := &runtimev1.ResourceName{Kind: runtime.ResourceKindRefreshTrigger, Name: name}
	r := &runtimev1.Resource{Resource: &runtimev1.Resource_RefreshTrigger{RefreshTrigger: &runtimev1.RefreshTrigger{Spec: spec}}}
	err = ctrl.Create(ctx, n, nil, nil, nil, true, r)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to create trigger: %w", err).Error())
	}

	return &runtimev1.CreateTriggerResponse{}, nil
}

// applySecurityPolicy applies relevant security policies to the resource.
// The input resource will not be modified in-place (so no need to set clone=true when obtaining it from the catalog).
func (s *Server) applySecurityPolicy(ctx context.Context, instID string, r *runtimev1.Resource) (*runtimev1.Resource, bool, error) {
	security, err := s.runtime.ResolveSecurity(instID, auth.GetClaims(ctx).SecurityClaims(), r)
	if err != nil {
		return nil, false, err
	}

	if security == nil {
		return r, true, nil
	}

	if !security.CanAccess() {
		return nil, false, nil
	}

	// Some resources may need deeper checks than just access.
	switch r.Resource.(type) {
	case *runtimev1.Resource_MetricsView:
		// For metrics views, we need to remove fields excluded by the field access rules.
		return s.applyMetricsViewSecurity(r, security), true, nil
	default:
		// The resource can be returned as is.
		return r, true, nil
	}
}

// applyMetricsViewSecurity rewrites a metrics view based on the field access conditions of a security policy.
func (s *Server) applyMetricsViewSecurity(r *runtimev1.Resource, security *runtime.ResolvedSecurity) *runtimev1.Resource {
	if security.CanAccessAllFields() {
		return r
	}

	mv := r.GetMetricsView()
	specDims, specMeasures, specChanged := s.applyMetricsViewSpecSecurity(mv.Spec, security)
	validSpecDims, validSpecMeasures, validSpecChanged := s.applyMetricsViewSpecSecurity(mv.State.ValidSpec, security)

	if !specChanged && !validSpecChanged {
		return r
	}

	mv = proto.Clone(mv).(*runtimev1.MetricsViewV2)

	if specChanged {
		mv.Spec.Dimensions = specDims
		mv.Spec.Measures = specMeasures
	}

	if validSpecChanged {
		mv.State.ValidSpec.Dimensions = validSpecDims
		mv.State.ValidSpec.Measures = validSpecMeasures
	}

	// We mustn't modify the resource in-place
	return &runtimev1.Resource{
		Meta:     r.Meta,
		Resource: &runtimev1.Resource_MetricsView{MetricsView: mv},
	}
}

// applyMetricsViewSpecSecurity rewrites a metrics view spec based on the field access conditions of a security policy.
func (s *Server) applyMetricsViewSpecSecurity(spec *runtimev1.MetricsViewSpec, policy *runtime.ResolvedSecurity) ([]*runtimev1.MetricsViewSpec_DimensionV2, []*runtimev1.MetricsViewSpec_MeasureV2, bool) {
	if spec == nil {
		return nil, nil, false
	}

	var dims []*runtimev1.MetricsViewSpec_DimensionV2
	for _, dim := range spec.Dimensions {
		if policy.CanAccessField(dim.Name) {
			dims = append(dims, dim)
		}
	}

	var ms []*runtimev1.MetricsViewSpec_MeasureV2
	for _, m := range spec.Measures {
		if policy.CanAccessField(m.Name) {
			ms = append(ms, m)
		}
	}

	if len(dims) == len(spec.Dimensions) && len(ms) == len(spec.Measures) {
		return nil, nil, false
	}

	return dims, ms, true
}

// modelSplitsToPB converts a slice of drivers.ModelSplit to a slice of runtimev1.ModelSplit.
func modelSplitsToPB(splits []drivers.ModelSplit) []*runtimev1.ModelSplit {
	pbs := make([]*runtimev1.ModelSplit, len(splits))
	for i, split := range splits {
		pbs[i] = modelSplitToPB(split)
	}
	return pbs
}

// modelSplitToPB converts a drivers.ModelSplit to a runtimev1.ModelSplit.
func modelSplitToPB(split drivers.ModelSplit) *runtimev1.ModelSplit {
	var data map[string]interface{}
	if err := json.Unmarshal(split.DataJSON, &data); err != nil {
		panic(err)
	}

	var watermark, executedOn *timestamppb.Timestamp
	if split.Watermark != nil {
		watermark = timestamppb.New(*split.Watermark)
	}
	if split.ExecutedOn != nil {
		executedOn = timestamppb.New(*split.ExecutedOn)
	}

	return &runtimev1.ModelSplit{
		Key:        split.Key,
		Data:       must(structpb.NewStruct(data)),
		Watermark:  watermark,
		ExecutedOn: executedOn,
		Error:      split.Error,
		ElapsedMs:  uint32(split.Elapsed.Milliseconds()),
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
