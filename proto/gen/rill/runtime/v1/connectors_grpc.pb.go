// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: rill/runtime/v1/connectors.proto

package runtimev1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	ConnectorService_ListS3Buckets_FullMethodName        = "/rill.runtime.v1.ConnectorService/ListS3Buckets"
	ConnectorService_ListS3BucketObjects_FullMethodName  = "/rill.runtime.v1.ConnectorService/ListS3BucketObjects"
	ConnectorService_ListGCSBuckets_FullMethodName       = "/rill.runtime.v1.ConnectorService/ListGCSBuckets"
	ConnectorService_ListGCSBucketObjects_FullMethodName = "/rill.runtime.v1.ConnectorService/ListGCSBucketObjects"
)

// ConnectorServiceClient is the client API for ConnectorService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ConnectorServiceClient interface {
	ListS3Buckets(ctx context.Context, in *ListS3BucketsRequest, opts ...grpc.CallOption) (*ListS3BucketsResponse, error)
	ListS3BucketObjects(ctx context.Context, in *ListS3BucketObjectsRequest, opts ...grpc.CallOption) (*ListS3BucketObjectsResponse, error)
	ListGCSBuckets(ctx context.Context, in *ListGCSBucketsRequest, opts ...grpc.CallOption) (*ListGCSBucketsResponse, error)
	ListGCSBucketObjects(ctx context.Context, in *ListGCSBucketObjectsRequest, opts ...grpc.CallOption) (*ListGCSBucketObjectsResponse, error)
}

type connectorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewConnectorServiceClient(cc grpc.ClientConnInterface) ConnectorServiceClient {
	return &connectorServiceClient{cc}
}

func (c *connectorServiceClient) ListS3Buckets(ctx context.Context, in *ListS3BucketsRequest, opts ...grpc.CallOption) (*ListS3BucketsResponse, error) {
	out := new(ListS3BucketsResponse)
	err := c.cc.Invoke(ctx, ConnectorService_ListS3Buckets_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *connectorServiceClient) ListS3BucketObjects(ctx context.Context, in *ListS3BucketObjectsRequest, opts ...grpc.CallOption) (*ListS3BucketObjectsResponse, error) {
	out := new(ListS3BucketObjectsResponse)
	err := c.cc.Invoke(ctx, ConnectorService_ListS3BucketObjects_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *connectorServiceClient) ListGCSBuckets(ctx context.Context, in *ListGCSBucketsRequest, opts ...grpc.CallOption) (*ListGCSBucketsResponse, error) {
	out := new(ListGCSBucketsResponse)
	err := c.cc.Invoke(ctx, ConnectorService_ListGCSBuckets_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *connectorServiceClient) ListGCSBucketObjects(ctx context.Context, in *ListGCSBucketObjectsRequest, opts ...grpc.CallOption) (*ListGCSBucketObjectsResponse, error) {
	out := new(ListGCSBucketObjectsResponse)
	err := c.cc.Invoke(ctx, ConnectorService_ListGCSBucketObjects_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ConnectorServiceServer is the server API for ConnectorService service.
// All implementations must embed UnimplementedConnectorServiceServer
// for forward compatibility
type ConnectorServiceServer interface {
	ListS3Buckets(context.Context, *ListS3BucketsRequest) (*ListS3BucketsResponse, error)
	ListS3BucketObjects(context.Context, *ListS3BucketObjectsRequest) (*ListS3BucketObjectsResponse, error)
	ListGCSBuckets(context.Context, *ListGCSBucketsRequest) (*ListGCSBucketsResponse, error)
	ListGCSBucketObjects(context.Context, *ListGCSBucketObjectsRequest) (*ListGCSBucketObjectsResponse, error)
	mustEmbedUnimplementedConnectorServiceServer()
}

// UnimplementedConnectorServiceServer must be embedded to have forward compatible implementations.
type UnimplementedConnectorServiceServer struct {
}

func (UnimplementedConnectorServiceServer) ListS3Buckets(context.Context, *ListS3BucketsRequest) (*ListS3BucketsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListS3Buckets not implemented")
}
func (UnimplementedConnectorServiceServer) ListS3BucketObjects(context.Context, *ListS3BucketObjectsRequest) (*ListS3BucketObjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListS3BucketObjects not implemented")
}
func (UnimplementedConnectorServiceServer) ListGCSBuckets(context.Context, *ListGCSBucketsRequest) (*ListGCSBucketsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListGCSBuckets not implemented")
}
func (UnimplementedConnectorServiceServer) ListGCSBucketObjects(context.Context, *ListGCSBucketObjectsRequest) (*ListGCSBucketObjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListGCSBucketObjects not implemented")
}
func (UnimplementedConnectorServiceServer) mustEmbedUnimplementedConnectorServiceServer() {}

// UnsafeConnectorServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ConnectorServiceServer will
// result in compilation errors.
type UnsafeConnectorServiceServer interface {
	mustEmbedUnimplementedConnectorServiceServer()
}

func RegisterConnectorServiceServer(s grpc.ServiceRegistrar, srv ConnectorServiceServer) {
	s.RegisterService(&ConnectorService_ServiceDesc, srv)
}

func _ConnectorService_ListS3Buckets_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListS3BucketsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConnectorServiceServer).ListS3Buckets(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ConnectorService_ListS3Buckets_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConnectorServiceServer).ListS3Buckets(ctx, req.(*ListS3BucketsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ConnectorService_ListS3BucketObjects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListS3BucketObjectsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConnectorServiceServer).ListS3BucketObjects(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ConnectorService_ListS3BucketObjects_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConnectorServiceServer).ListS3BucketObjects(ctx, req.(*ListS3BucketObjectsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ConnectorService_ListGCSBuckets_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListGCSBucketsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConnectorServiceServer).ListGCSBuckets(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ConnectorService_ListGCSBuckets_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConnectorServiceServer).ListGCSBuckets(ctx, req.(*ListGCSBucketsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ConnectorService_ListGCSBucketObjects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListGCSBucketObjectsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConnectorServiceServer).ListGCSBucketObjects(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ConnectorService_ListGCSBucketObjects_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConnectorServiceServer).ListGCSBucketObjects(ctx, req.(*ListGCSBucketObjectsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ConnectorService_ServiceDesc is the grpc.ServiceDesc for ConnectorService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ConnectorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rill.runtime.v1.ConnectorService",
	HandlerType: (*ConnectorServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListS3Buckets",
			Handler:    _ConnectorService_ListS3Buckets_Handler,
		},
		{
			MethodName: "ListS3BucketObjects",
			Handler:    _ConnectorService_ListS3BucketObjects_Handler,
		},
		{
			MethodName: "ListGCSBuckets",
			Handler:    _ConnectorService_ListGCSBuckets_Handler,
		},
		{
			MethodName: "ListGCSBucketObjects",
			Handler:    _ConnectorService_ListGCSBucketObjects_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rill/runtime/v1/connectors.proto",
}
