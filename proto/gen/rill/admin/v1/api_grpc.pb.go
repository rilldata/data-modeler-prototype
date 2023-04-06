// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: rill/admin/v1/api.proto

package adminv1

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

// AdminServiceClient is the client API for AdminService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AdminServiceClient interface {
	// Ping returns information about the server
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	// ListOrganizations lists all the organizations currently managed by the admin
	ListOrganizations(ctx context.Context, in *ListOrganizationsRequest, opts ...grpc.CallOption) (*ListOrganizationsResponse, error)
	// GetOrganization returns information about a specific organization
	GetOrganization(ctx context.Context, in *GetOrganizationRequest, opts ...grpc.CallOption) (*GetOrganizationResponse, error)
	// CreateOrganization creates a new organization
	CreateOrganization(ctx context.Context, in *CreateOrganizationRequest, opts ...grpc.CallOption) (*CreateOrganizationResponse, error)
	// DeleteOrganization deletes an organizations
	DeleteOrganization(ctx context.Context, in *DeleteOrganizationRequest, opts ...grpc.CallOption) (*DeleteOrganizationResponse, error)
	// UpdateOrganization deletes an organizations
	UpdateOrganization(ctx context.Context, in *UpdateOrganizationRequest, opts ...grpc.CallOption) (*UpdateOrganizationResponse, error)
	// ListProjects lists all the projects currently available for given organizations
	ListProjects(ctx context.Context, in *ListProjectsRequest, opts ...grpc.CallOption) (*ListProjectsResponse, error)
	// GetProject returns information about a specific project
	GetProject(ctx context.Context, in *GetProjectRequest, opts ...grpc.CallOption) (*GetProjectResponse, error)
	// CreateProject creates a new project
	CreateProject(ctx context.Context, in *CreateProjectRequest, opts ...grpc.CallOption) (*CreateProjectResponse, error)
	// DeleteProject deletes an project
	DeleteProject(ctx context.Context, in *DeleteProjectRequest, opts ...grpc.CallOption) (*DeleteProjectResponse, error)
	// UpdateProject updates a project
	UpdateProject(ctx context.Context, in *UpdateProjectRequest, opts ...grpc.CallOption) (*UpdateProjectResponse, error)
	// CreateProject creates a new project
	TriggerReconcile(ctx context.Context, in *TriggerReconcileRequest, opts ...grpc.CallOption) (*TriggerReconcileResponse, error)
	// CreateProject creates a new project
	TriggerRefreshSource(ctx context.Context, in *TriggerRefreshSourceRequest, opts ...grpc.CallOption) (*TriggerRefreshSourceResponse, error)
	// CreateProject creates a new project
	TriggerRedeploy(ctx context.Context, in *TriggerRedeployRequest, opts ...grpc.CallOption) (*TriggerRedeployResponse, error)
	// GetCurrentUser returns the currently authenticated user (if any)
	GetCurrentUser(ctx context.Context, in *GetCurrentUserRequest, opts ...grpc.CallOption) (*GetCurrentUserResponse, error)
	// RevokeCurrentAuthToken revoke the current auth token
	RevokeCurrentAuthToken(ctx context.Context, in *RevokeCurrentAuthTokenRequest, opts ...grpc.CallOption) (*RevokeCurrentAuthTokenResponse, error)
	// GetGithubRepoRequest returns info about a Github repo based on the caller's installations.
	// If the caller has not granted access to the repository, instructions for granting access are returned.
	GetGithubRepoStatus(ctx context.Context, in *GetGithubRepoStatusRequest, opts ...grpc.CallOption) (*GetGithubRepoStatusResponse, error)
}

type adminServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAdminServiceClient(cc grpc.ClientConnInterface) AdminServiceClient {
	return &adminServiceClient{cc}
}

func (c *adminServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) ListOrganizations(ctx context.Context, in *ListOrganizationsRequest, opts ...grpc.CallOption) (*ListOrganizationsResponse, error) {
	out := new(ListOrganizationsResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/ListOrganizations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) GetOrganization(ctx context.Context, in *GetOrganizationRequest, opts ...grpc.CallOption) (*GetOrganizationResponse, error) {
	out := new(GetOrganizationResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/GetOrganization", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) CreateOrganization(ctx context.Context, in *CreateOrganizationRequest, opts ...grpc.CallOption) (*CreateOrganizationResponse, error) {
	out := new(CreateOrganizationResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/CreateOrganization", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) DeleteOrganization(ctx context.Context, in *DeleteOrganizationRequest, opts ...grpc.CallOption) (*DeleteOrganizationResponse, error) {
	out := new(DeleteOrganizationResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/DeleteOrganization", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) UpdateOrganization(ctx context.Context, in *UpdateOrganizationRequest, opts ...grpc.CallOption) (*UpdateOrganizationResponse, error) {
	out := new(UpdateOrganizationResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/UpdateOrganization", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) ListProjects(ctx context.Context, in *ListProjectsRequest, opts ...grpc.CallOption) (*ListProjectsResponse, error) {
	out := new(ListProjectsResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/ListProjects", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) GetProject(ctx context.Context, in *GetProjectRequest, opts ...grpc.CallOption) (*GetProjectResponse, error) {
	out := new(GetProjectResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/GetProject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) CreateProject(ctx context.Context, in *CreateProjectRequest, opts ...grpc.CallOption) (*CreateProjectResponse, error) {
	out := new(CreateProjectResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/CreateProject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) DeleteProject(ctx context.Context, in *DeleteProjectRequest, opts ...grpc.CallOption) (*DeleteProjectResponse, error) {
	out := new(DeleteProjectResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/DeleteProject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) UpdateProject(ctx context.Context, in *UpdateProjectRequest, opts ...grpc.CallOption) (*UpdateProjectResponse, error) {
	out := new(UpdateProjectResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/UpdateProject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) TriggerReconcile(ctx context.Context, in *TriggerReconcileRequest, opts ...grpc.CallOption) (*TriggerReconcileResponse, error) {
	out := new(TriggerReconcileResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/TriggerReconcile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) TriggerRefreshSource(ctx context.Context, in *TriggerRefreshSourceRequest, opts ...grpc.CallOption) (*TriggerRefreshSourceResponse, error) {
	out := new(TriggerRefreshSourceResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/TriggerRefreshSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) TriggerRedeploy(ctx context.Context, in *TriggerRedeployRequest, opts ...grpc.CallOption) (*TriggerRedeployResponse, error) {
	out := new(TriggerRedeployResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/TriggerRedeploy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) GetCurrentUser(ctx context.Context, in *GetCurrentUserRequest, opts ...grpc.CallOption) (*GetCurrentUserResponse, error) {
	out := new(GetCurrentUserResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/GetCurrentUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) RevokeCurrentAuthToken(ctx context.Context, in *RevokeCurrentAuthTokenRequest, opts ...grpc.CallOption) (*RevokeCurrentAuthTokenResponse, error) {
	out := new(RevokeCurrentAuthTokenResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/RevokeCurrentAuthToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) GetGithubRepoStatus(ctx context.Context, in *GetGithubRepoStatusRequest, opts ...grpc.CallOption) (*GetGithubRepoStatusResponse, error) {
	out := new(GetGithubRepoStatusResponse)
	err := c.cc.Invoke(ctx, "/rill.admin.v1.AdminService/GetGithubRepoStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AdminServiceServer is the server API for AdminService service.
// All implementations must embed UnimplementedAdminServiceServer
// for forward compatibility
type AdminServiceServer interface {
	// Ping returns information about the server
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	// ListOrganizations lists all the organizations currently managed by the admin
	ListOrganizations(context.Context, *ListOrganizationsRequest) (*ListOrganizationsResponse, error)
	// GetOrganization returns information about a specific organization
	GetOrganization(context.Context, *GetOrganizationRequest) (*GetOrganizationResponse, error)
	// CreateOrganization creates a new organization
	CreateOrganization(context.Context, *CreateOrganizationRequest) (*CreateOrganizationResponse, error)
	// DeleteOrganization deletes an organizations
	DeleteOrganization(context.Context, *DeleteOrganizationRequest) (*DeleteOrganizationResponse, error)
	// UpdateOrganization deletes an organizations
	UpdateOrganization(context.Context, *UpdateOrganizationRequest) (*UpdateOrganizationResponse, error)
	// ListProjects lists all the projects currently available for given organizations
	ListProjects(context.Context, *ListProjectsRequest) (*ListProjectsResponse, error)
	// GetProject returns information about a specific project
	GetProject(context.Context, *GetProjectRequest) (*GetProjectResponse, error)
	// CreateProject creates a new project
	CreateProject(context.Context, *CreateProjectRequest) (*CreateProjectResponse, error)
	// DeleteProject deletes an project
	DeleteProject(context.Context, *DeleteProjectRequest) (*DeleteProjectResponse, error)
	// UpdateProject updates a project
	UpdateProject(context.Context, *UpdateProjectRequest) (*UpdateProjectResponse, error)
	// CreateProject creates a new project
	TriggerReconcile(context.Context, *TriggerReconcileRequest) (*TriggerReconcileResponse, error)
	// CreateProject creates a new project
	TriggerRefreshSource(context.Context, *TriggerRefreshSourceRequest) (*TriggerRefreshSourceResponse, error)
	// CreateProject creates a new project
	TriggerRedeploy(context.Context, *TriggerRedeployRequest) (*TriggerRedeployResponse, error)
	// GetCurrentUser returns the currently authenticated user (if any)
	GetCurrentUser(context.Context, *GetCurrentUserRequest) (*GetCurrentUserResponse, error)
	// RevokeCurrentAuthToken revoke the current auth token
	RevokeCurrentAuthToken(context.Context, *RevokeCurrentAuthTokenRequest) (*RevokeCurrentAuthTokenResponse, error)
	// GetGithubRepoRequest returns info about a Github repo based on the caller's installations.
	// If the caller has not granted access to the repository, instructions for granting access are returned.
	GetGithubRepoStatus(context.Context, *GetGithubRepoStatusRequest) (*GetGithubRepoStatusResponse, error)
	mustEmbedUnimplementedAdminServiceServer()
}

// UnimplementedAdminServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAdminServiceServer struct {
}

func (UnimplementedAdminServiceServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedAdminServiceServer) ListOrganizations(context.Context, *ListOrganizationsRequest) (*ListOrganizationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListOrganizations not implemented")
}
func (UnimplementedAdminServiceServer) GetOrganization(context.Context, *GetOrganizationRequest) (*GetOrganizationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrganization not implemented")
}
func (UnimplementedAdminServiceServer) CreateOrganization(context.Context, *CreateOrganizationRequest) (*CreateOrganizationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateOrganization not implemented")
}
func (UnimplementedAdminServiceServer) DeleteOrganization(context.Context, *DeleteOrganizationRequest) (*DeleteOrganizationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteOrganization not implemented")
}
func (UnimplementedAdminServiceServer) UpdateOrganization(context.Context, *UpdateOrganizationRequest) (*UpdateOrganizationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOrganization not implemented")
}
func (UnimplementedAdminServiceServer) ListProjects(context.Context, *ListProjectsRequest) (*ListProjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListProjects not implemented")
}
func (UnimplementedAdminServiceServer) GetProject(context.Context, *GetProjectRequest) (*GetProjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProject not implemented")
}
func (UnimplementedAdminServiceServer) CreateProject(context.Context, *CreateProjectRequest) (*CreateProjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateProject not implemented")
}
func (UnimplementedAdminServiceServer) DeleteProject(context.Context, *DeleteProjectRequest) (*DeleteProjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteProject not implemented")
}
func (UnimplementedAdminServiceServer) UpdateProject(context.Context, *UpdateProjectRequest) (*UpdateProjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProject not implemented")
}
func (UnimplementedAdminServiceServer) TriggerReconcile(context.Context, *TriggerReconcileRequest) (*TriggerReconcileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TriggerReconcile not implemented")
}
func (UnimplementedAdminServiceServer) TriggerRefreshSource(context.Context, *TriggerRefreshSourceRequest) (*TriggerRefreshSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TriggerRefreshSource not implemented")
}
func (UnimplementedAdminServiceServer) TriggerRedeploy(context.Context, *TriggerRedeployRequest) (*TriggerRedeployResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TriggerRedeploy not implemented")
}
func (UnimplementedAdminServiceServer) GetCurrentUser(context.Context, *GetCurrentUserRequest) (*GetCurrentUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCurrentUser not implemented")
}
func (UnimplementedAdminServiceServer) RevokeCurrentAuthToken(context.Context, *RevokeCurrentAuthTokenRequest) (*RevokeCurrentAuthTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RevokeCurrentAuthToken not implemented")
}
func (UnimplementedAdminServiceServer) GetGithubRepoStatus(context.Context, *GetGithubRepoStatusRequest) (*GetGithubRepoStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGithubRepoStatus not implemented")
}
func (UnimplementedAdminServiceServer) mustEmbedUnimplementedAdminServiceServer() {}

// UnsafeAdminServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AdminServiceServer will
// result in compilation errors.
type UnsafeAdminServiceServer interface {
	mustEmbedUnimplementedAdminServiceServer()
}

func RegisterAdminServiceServer(s grpc.ServiceRegistrar, srv AdminServiceServer) {
	s.RegisterService(&AdminService_ServiceDesc, srv)
}

func _AdminService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_ListOrganizations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListOrganizationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).ListOrganizations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/ListOrganizations",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).ListOrganizations(ctx, req.(*ListOrganizationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_GetOrganization_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOrganizationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).GetOrganization(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/GetOrganization",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).GetOrganization(ctx, req.(*GetOrganizationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_CreateOrganization_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateOrganizationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).CreateOrganization(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/CreateOrganization",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).CreateOrganization(ctx, req.(*CreateOrganizationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_DeleteOrganization_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteOrganizationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).DeleteOrganization(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/DeleteOrganization",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).DeleteOrganization(ctx, req.(*DeleteOrganizationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_UpdateOrganization_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateOrganizationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).UpdateOrganization(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/UpdateOrganization",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).UpdateOrganization(ctx, req.(*UpdateOrganizationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_ListProjects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListProjectsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).ListProjects(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/ListProjects",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).ListProjects(ctx, req.(*ListProjectsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_GetProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).GetProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/GetProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).GetProject(ctx, req.(*GetProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_CreateProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).CreateProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/CreateProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).CreateProject(ctx, req.(*CreateProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_DeleteProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).DeleteProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/DeleteProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).DeleteProject(ctx, req.(*DeleteProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_UpdateProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).UpdateProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/UpdateProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).UpdateProject(ctx, req.(*UpdateProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_TriggerReconcile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TriggerReconcileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).TriggerReconcile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/TriggerReconcile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).TriggerReconcile(ctx, req.(*TriggerReconcileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_TriggerRefreshSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TriggerRefreshSourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).TriggerRefreshSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/TriggerRefreshSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).TriggerRefreshSource(ctx, req.(*TriggerRefreshSourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_TriggerRedeploy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TriggerRedeployRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).TriggerRedeploy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/TriggerRedeploy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).TriggerRedeploy(ctx, req.(*TriggerRedeployRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_GetCurrentUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCurrentUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).GetCurrentUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/GetCurrentUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).GetCurrentUser(ctx, req.(*GetCurrentUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_RevokeCurrentAuthToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RevokeCurrentAuthTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).RevokeCurrentAuthToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/RevokeCurrentAuthToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).RevokeCurrentAuthToken(ctx, req.(*RevokeCurrentAuthTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_GetGithubRepoStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGithubRepoStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).GetGithubRepoStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rill.admin.v1.AdminService/GetGithubRepoStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).GetGithubRepoStatus(ctx, req.(*GetGithubRepoStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AdminService_ServiceDesc is the grpc.ServiceDesc for AdminService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AdminService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rill.admin.v1.AdminService",
	HandlerType: (*AdminServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _AdminService_Ping_Handler,
		},
		{
			MethodName: "ListOrganizations",
			Handler:    _AdminService_ListOrganizations_Handler,
		},
		{
			MethodName: "GetOrganization",
			Handler:    _AdminService_GetOrganization_Handler,
		},
		{
			MethodName: "CreateOrganization",
			Handler:    _AdminService_CreateOrganization_Handler,
		},
		{
			MethodName: "DeleteOrganization",
			Handler:    _AdminService_DeleteOrganization_Handler,
		},
		{
			MethodName: "UpdateOrganization",
			Handler:    _AdminService_UpdateOrganization_Handler,
		},
		{
			MethodName: "ListProjects",
			Handler:    _AdminService_ListProjects_Handler,
		},
		{
			MethodName: "GetProject",
			Handler:    _AdminService_GetProject_Handler,
		},
		{
			MethodName: "CreateProject",
			Handler:    _AdminService_CreateProject_Handler,
		},
		{
			MethodName: "DeleteProject",
			Handler:    _AdminService_DeleteProject_Handler,
		},
		{
			MethodName: "UpdateProject",
			Handler:    _AdminService_UpdateProject_Handler,
		},
		{
			MethodName: "TriggerReconcile",
			Handler:    _AdminService_TriggerReconcile_Handler,
		},
		{
			MethodName: "TriggerRefreshSource",
			Handler:    _AdminService_TriggerRefreshSource_Handler,
		},
		{
			MethodName: "TriggerRedeploy",
			Handler:    _AdminService_TriggerRedeploy_Handler,
		},
		{
			MethodName: "GetCurrentUser",
			Handler:    _AdminService_GetCurrentUser_Handler,
		},
		{
			MethodName: "RevokeCurrentAuthToken",
			Handler:    _AdminService_RevokeCurrentAuthToken_Handler,
		},
		{
			MethodName: "GetGithubRepoStatus",
			Handler:    _AdminService_GetGithubRepoStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rill/admin/v1/api.proto",
}
