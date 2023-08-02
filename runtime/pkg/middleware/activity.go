package middleware

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
)

func ActivityStreamServerInterceptor(activityClient activity.Client) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()
		subject := auth.GetClaims(ctx).Subject()

		ctx = activity.WithDims(ctx,
			attribute.String("user_id", subject),
			attribute.String("request_method", info.FullMethod),
		)
		wss := grpc_middleware.WrapServerStream(ss)
		wss.WrappedContext = ctx

		start := time.Now()
		defer func() {
			// Emit usage metric
			activityClient.Emit(ctx, "request_time_ms", float64(time.Since(start).Milliseconds()))
		}()

		return handler(srv, wss)
	}
}

func ActivityUnaryServerInterceptor(activityClient activity.Client) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		subject := auth.GetClaims(ctx).Subject()

		ctx = activity.WithDims(ctx,
			attribute.String("user_id", subject),
			attribute.String("request_method", info.FullMethod),
		)

		start := time.Now()
		defer func() {
			// Emit usage metric
			activityClient.Emit(ctx, "request_time_ms", float64(time.Since(start).Milliseconds()))
		}()

		return handler(ctx, req)
	}
}
