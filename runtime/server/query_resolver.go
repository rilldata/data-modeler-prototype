package server

import (
	"context"
	"errors"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// QueryResolver enables superusers and project admins to query a resolver within a project
func (s *Server) QueryResolver(ctx context.Context, req *runtimev1.QueryResolverRequest) (*runtimev1.QueryResolverResponse, error) {
	// Validate the caller has the ReadResolvers permission
	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.ReadResolvers) {
		return nil, status.Error(codes.PermissionDenied, "only superusers and project admins can query resolvers")
	}

	// Resolver should exist
	initializer, ok := runtime.ResolverInitializers[req.Resolver]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no resolver found of type %q", req.Resolver)
	}

	// Initialize the resolver
	resolver, err := initializer(ctx, &runtime.ResolverOptions{
		Runtime:    s.runtime,
		InstanceID: req.InstanceId,
		Properties: req.ResolverProperties.AsMap(),
		Args:       req.ResolverArgs.AsMap(),
		Claims:     claims.SecurityClaims(),
		ForExport:  false,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer resolver.Close()

	// Query the resolver
	res, err := resolver.ResolveInteractive(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer res.Close()

	data := make([]*structpb.Struct, 0)
	for {
		// Next returns the next row of data. It returns io.EOF when there are no more rows.
		row, err := res.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		rowStruct, err := pbutil.ToStruct(row, res.Schema())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		data = append(data, rowStruct)
	}

	// Return the response
	return &runtimev1.QueryResolverResponse{
		Schema: res.Schema(),
		Data:   data,
	}, nil
}
