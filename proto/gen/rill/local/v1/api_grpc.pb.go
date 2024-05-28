// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: rill/local/v1/api.proto

package localv1

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
	LocalService_Ping_FullMethodName             = "/rill.local.v1.LocalService/Ping"
	LocalService_GetMetadata_FullMethodName      = "/rill.local.v1.LocalService/GetMetadata"
	LocalService_GetVersion_FullMethodName       = "/rill.local.v1.LocalService/GetVersion"
	LocalService_DeployValidation_FullMethodName = "/rill.local.v1.LocalService/DeployValidation"
	LocalService_PushToGithub_FullMethodName     = "/rill.local.v1.LocalService/PushToGithub"
	LocalService_Deploy_FullMethodName           = "/rill.local.v1.LocalService/Deploy"
)

// LocalServiceClient is the client API for LocalService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LocalServiceClient interface {
	// Ping returns the current time.
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	// GetMetadata returns information about the local Rill instance.
	GetMetadata(ctx context.Context, in *GetMetadataRequest, opts ...grpc.CallOption) (*GetMetadataResponse, error)
	// GetVersion returns details about the current and latest available Rill versions.
	GetVersion(ctx context.Context, in *GetVersionRequest, opts ...grpc.CallOption) (*GetVersionResponse, error)
	// DeployValidation validates a deploy request.
	DeployValidation(ctx context.Context, in *DeployValidationRequest, opts ...grpc.CallOption) (*DeployValidationResponse, error)
	// PushToGithub create a Git repo from local project and pushed to users git account.
	PushToGithub(ctx context.Context, in *PushToGithubRequest, opts ...grpc.CallOption) (*PushToGithubResponse, error)
	// Deploy deploys the local project to the Rill cloud.
	Deploy(ctx context.Context, in *DeployRequest, opts ...grpc.CallOption) (*DeployResponse, error)
}

type localServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewLocalServiceClient(cc grpc.ClientConnInterface) LocalServiceClient {
	return &localServiceClient{cc}
}

func (c *localServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, LocalService_Ping_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *localServiceClient) GetMetadata(ctx context.Context, in *GetMetadataRequest, opts ...grpc.CallOption) (*GetMetadataResponse, error) {
	out := new(GetMetadataResponse)
	err := c.cc.Invoke(ctx, LocalService_GetMetadata_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *localServiceClient) GetVersion(ctx context.Context, in *GetVersionRequest, opts ...grpc.CallOption) (*GetVersionResponse, error) {
	out := new(GetVersionResponse)
	err := c.cc.Invoke(ctx, LocalService_GetVersion_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *localServiceClient) DeployValidation(ctx context.Context, in *DeployValidationRequest, opts ...grpc.CallOption) (*DeployValidationResponse, error) {
	out := new(DeployValidationResponse)
	err := c.cc.Invoke(ctx, LocalService_DeployValidation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *localServiceClient) PushToGithub(ctx context.Context, in *PushToGithubRequest, opts ...grpc.CallOption) (*PushToGithubResponse, error) {
	out := new(PushToGithubResponse)
	err := c.cc.Invoke(ctx, LocalService_PushToGithub_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *localServiceClient) Deploy(ctx context.Context, in *DeployRequest, opts ...grpc.CallOption) (*DeployResponse, error) {
	out := new(DeployResponse)
	err := c.cc.Invoke(ctx, LocalService_Deploy_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LocalServiceServer is the server API for LocalService service.
// All implementations must embed UnimplementedLocalServiceServer
// for forward compatibility
type LocalServiceServer interface {
	// Ping returns the current time.
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	// GetMetadata returns information about the local Rill instance.
	GetMetadata(context.Context, *GetMetadataRequest) (*GetMetadataResponse, error)
	// GetVersion returns details about the current and latest available Rill versions.
	GetVersion(context.Context, *GetVersionRequest) (*GetVersionResponse, error)
	// DeployValidation validates a deploy request.
	DeployValidation(context.Context, *DeployValidationRequest) (*DeployValidationResponse, error)
	// PushToGithub create a Git repo from local project and pushed to users git account.
	PushToGithub(context.Context, *PushToGithubRequest) (*PushToGithubResponse, error)
	// Deploy deploys the local project to the Rill cloud.
	Deploy(context.Context, *DeployRequest) (*DeployResponse, error)
	mustEmbedUnimplementedLocalServiceServer()
}

// UnimplementedLocalServiceServer must be embedded to have forward compatible implementations.
type UnimplementedLocalServiceServer struct {
}

func (UnimplementedLocalServiceServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedLocalServiceServer) GetMetadata(context.Context, *GetMetadataRequest) (*GetMetadataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMetadata not implemented")
}
func (UnimplementedLocalServiceServer) GetVersion(context.Context, *GetVersionRequest) (*GetVersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVersion not implemented")
}
func (UnimplementedLocalServiceServer) DeployValidation(context.Context, *DeployValidationRequest) (*DeployValidationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeployValidation not implemented")
}
func (UnimplementedLocalServiceServer) PushToGithub(context.Context, *PushToGithubRequest) (*PushToGithubResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PushToGithub not implemented")
}
func (UnimplementedLocalServiceServer) Deploy(context.Context, *DeployRequest) (*DeployResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Deploy not implemented")
}
func (UnimplementedLocalServiceServer) mustEmbedUnimplementedLocalServiceServer() {}

// UnsafeLocalServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LocalServiceServer will
// result in compilation errors.
type UnsafeLocalServiceServer interface {
	mustEmbedUnimplementedLocalServiceServer()
}

func RegisterLocalServiceServer(s grpc.ServiceRegistrar, srv LocalServiceServer) {
	s.RegisterService(&LocalService_ServiceDesc, srv)
}

func _LocalService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocalServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LocalService_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocalServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocalService_GetMetadata_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMetadataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocalServiceServer).GetMetadata(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LocalService_GetMetadata_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocalServiceServer).GetMetadata(ctx, req.(*GetMetadataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocalService_GetVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocalServiceServer).GetVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LocalService_GetVersion_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocalServiceServer).GetVersion(ctx, req.(*GetVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocalService_DeployValidation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeployValidationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocalServiceServer).DeployValidation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LocalService_DeployValidation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocalServiceServer).DeployValidation(ctx, req.(*DeployValidationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocalService_PushToGithub_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushToGithubRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocalServiceServer).PushToGithub(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LocalService_PushToGithub_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocalServiceServer).PushToGithub(ctx, req.(*PushToGithubRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocalService_Deploy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeployRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocalServiceServer).Deploy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LocalService_Deploy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocalServiceServer).Deploy(ctx, req.(*DeployRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// LocalService_ServiceDesc is the grpc.ServiceDesc for LocalService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LocalService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rill.local.v1.LocalService",
	HandlerType: (*LocalServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _LocalService_Ping_Handler,
		},
		{
			MethodName: "GetMetadata",
			Handler:    _LocalService_GetMetadata_Handler,
		},
		{
			MethodName: "GetVersion",
			Handler:    _LocalService_GetVersion_Handler,
		},
		{
			MethodName: "DeployValidation",
			Handler:    _LocalService_DeployValidation_Handler,
		},
		{
			MethodName: "PushToGithub",
			Handler:    _LocalService_PushToGithub_Handler,
		},
		{
			MethodName: "Deploy",
			Handler:    _LocalService_Deploy_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rill/local/v1/api.proto",
}
