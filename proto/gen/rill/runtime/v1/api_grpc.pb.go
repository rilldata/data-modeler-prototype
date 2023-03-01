// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: rill/runtime/v1/api.proto

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

// RuntimeServiceClient is the client API for RuntimeService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RuntimeServiceClient interface {
	// Ping returns information about the runtime
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	// ListInstances lists all the instances currently managed by the runtime
	ListInstances(ctx context.Context, in *ListInstancesRequest, opts ...grpc.CallOption) (*ListInstancesResponse, error)
	// GetInstance returns information about a specific instance
	GetInstance(ctx context.Context, in *GetInstanceRequest, opts ...grpc.CallOption) (*GetInstanceResponse, error)
	// CreateInstance creates a new instance
	CreateInstance(ctx context.Context, in *CreateInstanceRequest, opts ...grpc.CallOption) (*CreateInstanceResponse, error)
	// DeleteInstance deletes an instance
	DeleteInstance(ctx context.Context, in *DeleteInstanceRequest, opts ...grpc.CallOption) (*DeleteInstanceResponse, error)
	// ListFiles lists all the files matching a glob in a repo.
	// The files are sorted by their full path.
	ListFiles(ctx context.Context, in *ListFilesRequest, opts ...grpc.CallOption) (*ListFilesResponse, error)
	// GetFile returns the contents of a specific file in a repo.
	GetFile(ctx context.Context, in *GetFileRequest, opts ...grpc.CallOption) (*GetFileResponse, error)
	// PutFile creates or updates a file in a repo
	PutFile(ctx context.Context, in *PutFileRequest, opts ...grpc.CallOption) (*PutFileResponse, error)
	// DeleteFile deletes a file from a repo
	DeleteFile(ctx context.Context, in *DeleteFileRequest, opts ...grpc.CallOption) (*DeleteFileResponse, error)
	// RenameFile renames a file in a repo
	RenameFile(ctx context.Context, in *RenameFileRequest, opts ...grpc.CallOption) (*RenameFileResponse, error)
	// ListCatalogEntries lists all the entries registered in an instance's catalog (like tables, sources or metrics views)
	ListCatalogEntries(ctx context.Context, in *ListCatalogEntriesRequest, opts ...grpc.CallOption) (*ListCatalogEntriesResponse, error)
	// GetCatalogEntry returns information about a specific entry in the catalog
	GetCatalogEntry(ctx context.Context, in *GetCatalogEntryRequest, opts ...grpc.CallOption) (*GetCatalogEntryResponse, error)
	// TriggerRefresh triggers a refresh of a refreshable catalog object.
	// It currently only supports sources (which will be re-ingested), but will also support materialized models in the future.
	// It does not respond until the refresh has completed (will move to async jobs when the task scheduler is in place).
	TriggerRefresh(ctx context.Context, in *TriggerRefreshRequest, opts ...grpc.CallOption) (*TriggerRefreshResponse, error)
	// TriggerSync syncronizes the instance's catalog with the underlying OLAP's information schema.
	// If the instance has exposed=true, tables found in the information schema will be added to the catalog.
	TriggerSync(ctx context.Context, in *TriggerSyncRequest, opts ...grpc.CallOption) (*TriggerSyncResponse, error)
	// Reconcile applies a full set of artifacts from a repo to the catalog and infra.
	// It attempts to infer a minimal number of migrations to apply to reconcile the current state with
	// the desired state expressed in the artifacts. Any existing objects not described in the submitted
	// artifacts will be deleted.
	Reconcile(ctx context.Context, in *ReconcileRequest, opts ...grpc.CallOption) (*ReconcileResponse, error)
	// PutFileAndReconcile combines PutFile and Reconcile in a single endpoint to reduce latency.
	// It is equivalent to calling the two RPCs sequentially.
	PutFileAndReconcile(ctx context.Context, in *PutFileAndReconcileRequest, opts ...grpc.CallOption) (*PutFileAndReconcileResponse, error)
	// DeleteFileAndReconcile combines RenameFile and Reconcile in a single endpoint to reduce latency.
	DeleteFileAndReconcile(ctx context.Context, in *DeleteFileAndReconcileRequest, opts ...grpc.CallOption) (*DeleteFileAndReconcileResponse, error)
	// RenameFileAndReconcile combines RenameFile and Reconcile in a single endpoint to reduce latency.
	RenameFileAndReconcile(ctx context.Context, in *RenameFileAndReconcileRequest, opts ...grpc.CallOption) (*RenameFileAndReconcileResponse, error)
	RefreshAndReconcile(ctx context.Context, in *RefreshAndReconcileRequest, opts ...grpc.CallOption) (*RefreshAndReconcileResponse, error)
	// ListConnectors returns a description of all the connectors implemented in the runtime,
	// including their schema and validation rules
	ListConnectors(ctx context.Context, in *ListConnectorsRequest, opts ...grpc.CallOption) (*ListConnectorsResponse, error)
}

type runtimeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRuntimeServiceClient(cc grpc.ClientConnInterface) RuntimeServiceClient {
	return &runtimeServiceClient{cc}
}

func (c *runtimeServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) ListInstances(ctx context.Context, in *ListInstancesRequest, opts ...grpc.CallOption) (*ListInstancesResponse, error) {
	out := new(ListInstancesResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/ListInstances", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) GetInstance(ctx context.Context, in *GetInstanceRequest, opts ...grpc.CallOption) (*GetInstanceResponse, error) {
	out := new(GetInstanceResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/GetInstance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) CreateInstance(ctx context.Context, in *CreateInstanceRequest, opts ...grpc.CallOption) (*CreateInstanceResponse, error) {
	out := new(CreateInstanceResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/CreateInstance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) DeleteInstance(ctx context.Context, in *DeleteInstanceRequest, opts ...grpc.CallOption) (*DeleteInstanceResponse, error) {
	out := new(DeleteInstanceResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/DeleteInstance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) ListFiles(ctx context.Context, in *ListFilesRequest, opts ...grpc.CallOption) (*ListFilesResponse, error) {
	out := new(ListFilesResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/ListFiles", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) GetFile(ctx context.Context, in *GetFileRequest, opts ...grpc.CallOption) (*GetFileResponse, error) {
	out := new(GetFileResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/GetFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) PutFile(ctx context.Context, in *PutFileRequest, opts ...grpc.CallOption) (*PutFileResponse, error) {
	out := new(PutFileResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/PutFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) DeleteFile(ctx context.Context, in *DeleteFileRequest, opts ...grpc.CallOption) (*DeleteFileResponse, error) {
	out := new(DeleteFileResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/DeleteFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) RenameFile(ctx context.Context, in *RenameFileRequest, opts ...grpc.CallOption) (*RenameFileResponse, error) {
	out := new(RenameFileResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/RenameFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) ListCatalogEntries(ctx context.Context, in *ListCatalogEntriesRequest, opts ...grpc.CallOption) (*ListCatalogEntriesResponse, error) {
	out := new(ListCatalogEntriesResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/ListCatalogEntries", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) GetCatalogEntry(ctx context.Context, in *GetCatalogEntryRequest, opts ...grpc.CallOption) (*GetCatalogEntryResponse, error) {
	out := new(GetCatalogEntryResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/GetCatalogEntry", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) TriggerRefresh(ctx context.Context, in *TriggerRefreshRequest, opts ...grpc.CallOption) (*TriggerRefreshResponse, error) {
	out := new(TriggerRefreshResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/TriggerRefresh", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) TriggerSync(ctx context.Context, in *TriggerSyncRequest, opts ...grpc.CallOption) (*TriggerSyncResponse, error) {
	out := new(TriggerSyncResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/TriggerSync", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) Reconcile(ctx context.Context, in *ReconcileRequest, opts ...grpc.CallOption) (*ReconcileResponse, error) {
	out := new(ReconcileResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/Reconcile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) PutFileAndReconcile(ctx context.Context, in *PutFileAndReconcileRequest, opts ...grpc.CallOption) (*PutFileAndReconcileResponse, error) {
	out := new(PutFileAndReconcileResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/PutFileAndReconcile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) DeleteFileAndReconcile(ctx context.Context, in *DeleteFileAndReconcileRequest, opts ...grpc.CallOption) (*DeleteFileAndReconcileResponse, error) {
	out := new(DeleteFileAndReconcileResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/DeleteFileAndReconcile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) RenameFileAndReconcile(ctx context.Context, in *RenameFileAndReconcileRequest, opts ...grpc.CallOption) (*RenameFileAndReconcileResponse, error) {
	out := new(RenameFileAndReconcileResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/RenameFileAndReconcile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) RefreshAndReconcile(ctx context.Context, in *RefreshAndReconcileRequest, opts ...grpc.CallOption) (*RefreshAndReconcileResponse, error) {
	out := new(RefreshAndReconcileResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/RefreshAndReconcile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) ListConnectors(ctx context.Context, in *ListConnectorsRequest, opts ...grpc.CallOption) (*ListConnectorsResponse, error) {
	out := new(ListConnectorsResponse)
	err := c.cc.Invoke(ctx, "/rill.runtime.v1.RuntimeService/ListConnectors", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RuntimeServiceServer is the server API for RuntimeService service.
// All implementations must embed UnimplementedRuntimeServiceServer
// for forward compatibility
type RuntimeServiceServer interface {
	// Ping returns information about the runtime
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	// ListInstances lists all the instances currently managed by the runtime
	ListInstances(context.Context, *ListInstancesRequest) (*ListInstancesResponse, error)
	// GetInstance returns information about a specific instance
	GetInstance(context.Context, *GetInstanceRequest) (*GetInstanceResponse, error)
	// CreateInstance creates a new instance
	CreateInstance(context.Context, *CreateInstanceRequest) (*CreateInstanceResponse, error)
	// DeleteInstance deletes an instance
	DeleteInstance(context.Context, *DeleteInstanceRequest) (*DeleteInstanceResponse, error)
	// ListFiles lists all the files matching a glob in a repo.
	// The files are sorted by their full path.
	ListFiles(context.Context, *ListFilesRequest) (*ListFilesResponse, error)
	// GetFile returns the contents of a specific file in a repo.
	GetFile(context.Context, *GetFileRequest) (*GetFileResponse, error)
	// PutFile creates or updates a file in a repo
	PutFile(context.Context, *PutFileRequest) (*PutFileResponse, error)
	// DeleteFile deletes a file from a repo
	DeleteFile(context.Context, *DeleteFileRequest) (*DeleteFileResponse, error)
	// RenameFile renames a file in a repo
	RenameFile(context.Context, *RenameFileRequest) (*RenameFileResponse, error)
	// ListCatalogEntries lists all the entries registered in an instance's catalog (like tables, sources or metrics views)
	ListCatalogEntries(context.Context, *ListCatalogEntriesRequest) (*ListCatalogEntriesResponse, error)
	// GetCatalogEntry returns information about a specific entry in the catalog
	GetCatalogEntry(context.Context, *GetCatalogEntryRequest) (*GetCatalogEntryResponse, error)
	// TriggerRefresh triggers a refresh of a refreshable catalog object.
	// It currently only supports sources (which will be re-ingested), but will also support materialized models in the future.
	// It does not respond until the refresh has completed (will move to async jobs when the task scheduler is in place).
	TriggerRefresh(context.Context, *TriggerRefreshRequest) (*TriggerRefreshResponse, error)
	// TriggerSync syncronizes the instance's catalog with the underlying OLAP's information schema.
	// If the instance has exposed=true, tables found in the information schema will be added to the catalog.
	TriggerSync(context.Context, *TriggerSyncRequest) (*TriggerSyncResponse, error)
	// Reconcile applies a full set of artifacts from a repo to the catalog and infra.
	// It attempts to infer a minimal number of migrations to apply to reconcile the current state with
	// the desired state expressed in the artifacts. Any existing objects not described in the submitted
	// artifacts will be deleted.
	Reconcile(context.Context, *ReconcileRequest) (*ReconcileResponse, error)
	// PutFileAndReconcile combines PutFile and Reconcile in a single endpoint to reduce latency.
	// It is equivalent to calling the two RPCs sequentially.
	PutFileAndReconcile(context.Context, *PutFileAndReconcileRequest) (*PutFileAndReconcileResponse, error)
	// DeleteFileAndReconcile combines RenameFile and Reconcile in a single endpoint to reduce latency.
	DeleteFileAndReconcile(context.Context, *DeleteFileAndReconcileRequest) (*DeleteFileAndReconcileResponse, error)
	// RenameFileAndReconcile combines RenameFile and Reconcile in a single endpoint to reduce latency.
	RenameFileAndReconcile(context.Context, *RenameFileAndReconcileRequest) (*RenameFileAndReconcileResponse, error)
	RefreshAndReconcile(context.Context, *RefreshAndReconcileRequest) (*RefreshAndReconcileResponse, error)
	// ListConnectors returns a description of all the connectors implemented in the runtime,
	// including their schema and validation rules
	ListConnectors(context.Context, *ListConnectorsRequest) (*ListConnectorsResponse, error)
	mustEmbedUnimplementedRuntimeServiceServer()
}

// UnimplementedRuntimeServiceServer must be embedded to have forward compatible implementations.
type UnimplementedRuntimeServiceServer struct {
}

func (UnimplementedRuntimeServiceServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedRuntimeServiceServer) ListInstances(context.Context, *ListInstancesRequest) (*ListInstancesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListInstances not implemented")
}
func (UnimplementedRuntimeServiceServer) GetInstance(context.Context, *GetInstanceRequest) (*GetInstanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInstance not implemented")
}
func (UnimplementedRuntimeServiceServer) CreateInstance(context.Context, *CreateInstanceRequest) (*CreateInstanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateInstance not implemented")
}
func (UnimplementedRuntimeServiceServer) DeleteInstance(context.Context, *DeleteInstanceRequest) (*DeleteInstanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteInstance not implemented")
}
func (UnimplementedRuntimeServiceServer) ListFiles(context.Context, *ListFilesRequest) (*ListFilesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListFiles not implemented")
}
func (UnimplementedRuntimeServiceServer) GetFile(context.Context, *GetFileRequest) (*GetFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFile not implemented")
}
func (UnimplementedRuntimeServiceServer) PutFile(context.Context, *PutFileRequest) (*PutFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutFile not implemented")
}
func (UnimplementedRuntimeServiceServer) DeleteFile(context.Context, *DeleteFileRequest) (*DeleteFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteFile not implemented")
}
func (UnimplementedRuntimeServiceServer) RenameFile(context.Context, *RenameFileRequest) (*RenameFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RenameFile not implemented")
}
func (UnimplementedRuntimeServiceServer) ListCatalogEntries(context.Context, *ListCatalogEntriesRequest) (*ListCatalogEntriesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListCatalogEntries not implemented")
}
func (UnimplementedRuntimeServiceServer) GetCatalogEntry(context.Context, *GetCatalogEntryRequest) (*GetCatalogEntryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCatalogEntry not implemented")
}
func (UnimplementedRuntimeServiceServer) TriggerRefresh(context.Context, *TriggerRefreshRequest) (*TriggerRefreshResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TriggerRefresh not implemented")
}
func (UnimplementedRuntimeServiceServer) TriggerSync(context.Context, *TriggerSyncRequest) (*TriggerSyncResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TriggerSync not implemented")
}
func (UnimplementedRuntimeServiceServer) Reconcile(context.Context, *ReconcileRequest) (*ReconcileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Reconcile not implemented")
}
func (UnimplementedRuntimeServiceServer) PutFileAndReconcile(context.Context, *PutFileAndReconcileRequest) (*PutFileAndReconcileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutFileAndReconcile not implemented")
}
func (UnimplementedRuntimeServiceServer) DeleteFileAndReconcile(context.Context, *DeleteFileAndReconcileRequest) (*DeleteFileAndReconcileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteFileAndReconcile not implemented")
}
func (UnimplementedRuntimeServiceServer) RenameFileAndReconcile(context.Context, *RenameFileAndReconcileRequest) (*RenameFileAndReconcileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RenameFileAndReconcile not implemented")
}
func (UnimplementedRuntimeServiceServer) RefreshAndReconcile(context.Context, *RefreshAndReconcileRequest) (*RefreshAndReconcileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefreshAndReconcile not implemented")
}
func (UnimplementedRuntimeServiceServer) ListConnectors(context.Context, *ListConnectorsRequest) (*ListConnectorsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListConnectors not implemented")
}
func (UnimplementedRuntimeServiceServer) mustEmbedUnimplementedRuntimeServiceServer() {}

// UnsafeRuntimeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RuntimeServiceServer will
// result in compilation errors.
type UnsafeRuntimeServiceServer interface {
	mustEmbedUnimplementedRuntimeServiceServer()
}

func RegisterRuntimeServiceServer(s grpc.ServiceRegistrar, srv RuntimeServiceServer) {
	s.RegisterService(&RuntimeService_ServiceDesc, srv)
}

func _RuntimeService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_ListInstances_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListInstancesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).ListInstances(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/ListInstances",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).ListInstances(ctx, req.(*ListInstancesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_GetInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).GetInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/GetInstance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).GetInstance(ctx, req.(*GetInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_CreateInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).CreateInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/CreateInstance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).CreateInstance(ctx, req.(*CreateInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_DeleteInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).DeleteInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/DeleteInstance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).DeleteInstance(ctx, req.(*DeleteInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_ListFiles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListFilesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).ListFiles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/ListFiles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).ListFiles(ctx, req.(*ListFilesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_GetFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).GetFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/GetFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).GetFile(ctx, req.(*GetFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_PutFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).PutFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/PutFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).PutFile(ctx, req.(*PutFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_DeleteFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).DeleteFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/DeleteFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).DeleteFile(ctx, req.(*DeleteFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_RenameFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RenameFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).RenameFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/RenameFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).RenameFile(ctx, req.(*RenameFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_ListCatalogEntries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListCatalogEntriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).ListCatalogEntries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/ListCatalogEntries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).ListCatalogEntries(ctx, req.(*ListCatalogEntriesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_GetCatalogEntry_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCatalogEntryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).GetCatalogEntry(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/GetCatalogEntry",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).GetCatalogEntry(ctx, req.(*GetCatalogEntryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_TriggerRefresh_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TriggerRefreshRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).TriggerRefresh(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/TriggerRefresh",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).TriggerRefresh(ctx, req.(*TriggerRefreshRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_TriggerSync_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TriggerSyncRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).TriggerSync(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/TriggerSync",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).TriggerSync(ctx, req.(*TriggerSyncRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_Reconcile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReconcileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).Reconcile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/Reconcile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).Reconcile(ctx, req.(*ReconcileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_PutFileAndReconcile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutFileAndReconcileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).PutFileAndReconcile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/PutFileAndReconcile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).PutFileAndReconcile(ctx, req.(*PutFileAndReconcileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_DeleteFileAndReconcile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteFileAndReconcileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).DeleteFileAndReconcile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/DeleteFileAndReconcile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).DeleteFileAndReconcile(ctx, req.(*DeleteFileAndReconcileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_RenameFileAndReconcile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RenameFileAndReconcileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).RenameFileAndReconcile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/RenameFileAndReconcile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).RenameFileAndReconcile(ctx, req.(*RenameFileAndReconcileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_RefreshAndReconcile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefreshAndReconcileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).RefreshAndReconcile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/RefreshAndReconcile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).RefreshAndReconcile(ctx, req.(*RefreshAndReconcileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_ListConnectors_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListConnectorsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).ListConnectors(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.runtime.v1.RuntimeService/ListConnectors",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).ListConnectors(ctx, req.(*ListConnectorsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RuntimeService_ServiceDesc is the grpc.ServiceDesc for RuntimeService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RuntimeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rill.runtime.v1.RuntimeService",
	HandlerType: (*RuntimeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _RuntimeService_Ping_Handler,
		},
		{
			MethodName: "ListInstances",
			Handler:    _RuntimeService_ListInstances_Handler,
		},
		{
			MethodName: "GetInstance",
			Handler:    _RuntimeService_GetInstance_Handler,
		},
		{
			MethodName: "CreateInstance",
			Handler:    _RuntimeService_CreateInstance_Handler,
		},
		{
			MethodName: "DeleteInstance",
			Handler:    _RuntimeService_DeleteInstance_Handler,
		},
		{
			MethodName: "ListFiles",
			Handler:    _RuntimeService_ListFiles_Handler,
		},
		{
			MethodName: "GetFile",
			Handler:    _RuntimeService_GetFile_Handler,
		},
		{
			MethodName: "PutFile",
			Handler:    _RuntimeService_PutFile_Handler,
		},
		{
			MethodName: "DeleteFile",
			Handler:    _RuntimeService_DeleteFile_Handler,
		},
		{
			MethodName: "RenameFile",
			Handler:    _RuntimeService_RenameFile_Handler,
		},
		{
			MethodName: "ListCatalogEntries",
			Handler:    _RuntimeService_ListCatalogEntries_Handler,
		},
		{
			MethodName: "GetCatalogEntry",
			Handler:    _RuntimeService_GetCatalogEntry_Handler,
		},
		{
			MethodName: "TriggerRefresh",
			Handler:    _RuntimeService_TriggerRefresh_Handler,
		},
		{
			MethodName: "TriggerSync",
			Handler:    _RuntimeService_TriggerSync_Handler,
		},
		{
			MethodName: "Reconcile",
			Handler:    _RuntimeService_Reconcile_Handler,
		},
		{
			MethodName: "PutFileAndReconcile",
			Handler:    _RuntimeService_PutFileAndReconcile_Handler,
		},
		{
			MethodName: "DeleteFileAndReconcile",
			Handler:    _RuntimeService_DeleteFileAndReconcile_Handler,
		},
		{
			MethodName: "RenameFileAndReconcile",
			Handler:    _RuntimeService_RenameFileAndReconcile_Handler,
		},
		{
			MethodName: "RefreshAndReconcile",
			Handler:    _RuntimeService_RefreshAndReconcile_Handler,
		},
		{
			MethodName: "ListConnectors",
			Handler:    _RuntimeService_ListConnectors_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rill/runtime/v1/api.proto",
}
