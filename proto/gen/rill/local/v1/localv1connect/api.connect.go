// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: rill/local/v1/api.proto

package localv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/rilldata/rill/proto/gen/rill/local/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// LocalServiceName is the fully-qualified name of the LocalService service.
	LocalServiceName = "rill.local.v1.LocalService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// LocalServicePingProcedure is the fully-qualified name of the LocalService's Ping RPC.
	LocalServicePingProcedure = "/rill.local.v1.LocalService/Ping"
	// LocalServiceGetMetadataProcedure is the fully-qualified name of the LocalService's GetMetadata
	// RPC.
	LocalServiceGetMetadataProcedure = "/rill.local.v1.LocalService/GetMetadata"
	// LocalServiceGetVersionProcedure is the fully-qualified name of the LocalService's GetVersion RPC.
	LocalServiceGetVersionProcedure = "/rill.local.v1.LocalService/GetVersion"
	// LocalServicePushToGithubProcedure is the fully-qualified name of the LocalService's PushToGithub
	// RPC.
	LocalServicePushToGithubProcedure = "/rill.local.v1.LocalService/PushToGithub"
	// LocalServiceDeployProjectProcedure is the fully-qualified name of the LocalService's
	// DeployProject RPC.
	LocalServiceDeployProjectProcedure = "/rill.local.v1.LocalService/DeployProject"
	// LocalServiceRedeployProjectProcedure is the fully-qualified name of the LocalService's
	// RedeployProject RPC.
	LocalServiceRedeployProjectProcedure = "/rill.local.v1.LocalService/RedeployProject"
	// LocalServiceGetCurrentUserProcedure is the fully-qualified name of the LocalService's
	// GetCurrentUser RPC.
	LocalServiceGetCurrentUserProcedure = "/rill.local.v1.LocalService/GetCurrentUser"
	// LocalServiceGetCurrentProjectProcedure is the fully-qualified name of the LocalService's
	// GetCurrentProject RPC.
	LocalServiceGetCurrentProjectProcedure = "/rill.local.v1.LocalService/GetCurrentProject"
	// LocalServiceGetUserOrgMetadataProcedure is the fully-qualified name of the LocalService's
	// GetUserOrgMetadata RPC.
	LocalServiceGetUserOrgMetadataProcedure = "/rill.local.v1.LocalService/GetUserOrgMetadata"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	localServiceServiceDescriptor                  = v1.File_rill_local_v1_api_proto.Services().ByName("LocalService")
	localServicePingMethodDescriptor               = localServiceServiceDescriptor.Methods().ByName("Ping")
	localServiceGetMetadataMethodDescriptor        = localServiceServiceDescriptor.Methods().ByName("GetMetadata")
	localServiceGetVersionMethodDescriptor         = localServiceServiceDescriptor.Methods().ByName("GetVersion")
	localServicePushToGithubMethodDescriptor       = localServiceServiceDescriptor.Methods().ByName("PushToGithub")
	localServiceDeployProjectMethodDescriptor      = localServiceServiceDescriptor.Methods().ByName("DeployProject")
	localServiceRedeployProjectMethodDescriptor    = localServiceServiceDescriptor.Methods().ByName("RedeployProject")
	localServiceGetCurrentUserMethodDescriptor     = localServiceServiceDescriptor.Methods().ByName("GetCurrentUser")
	localServiceGetCurrentProjectMethodDescriptor  = localServiceServiceDescriptor.Methods().ByName("GetCurrentProject")
	localServiceGetUserOrgMetadataMethodDescriptor = localServiceServiceDescriptor.Methods().ByName("GetUserOrgMetadata")
)

// LocalServiceClient is a client for the rill.local.v1.LocalService service.
type LocalServiceClient interface {
	// Ping returns the current time.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// GetMetadata returns information about the local Rill instance.
	GetMetadata(context.Context, *connect.Request[v1.GetMetadataRequest]) (*connect.Response[v1.GetMetadataResponse], error)
	// GetVersion returns details about the current and latest available Rill versions.
	GetVersion(context.Context, *connect.Request[v1.GetVersionRequest]) (*connect.Response[v1.GetVersionResponse], error)
	// PushToGithub create a Git repo from local project and pushed to users git account.
	PushToGithub(context.Context, *connect.Request[v1.PushToGithubRequest]) (*connect.Response[v1.PushToGithubResponse], error)
	// DeployProject deploys the local project to the Rill cloud.
	DeployProject(context.Context, *connect.Request[v1.DeployProjectRequest]) (*connect.Response[v1.DeployProjectResponse], error)
	// RedeployProject updates a deployed project.
	RedeployProject(context.Context, *connect.Request[v1.RedeployProjectRequest]) (*connect.Response[v1.RedeployProjectResponse], error)
	// GetCurrentUser returns the locally logged in user
	GetCurrentUser(context.Context, *connect.Request[v1.GetCurrentUserRequest]) (*connect.Response[v1.GetCurrentUserResponse], error)
	// GetCurrentProject returns the rill cloud project connected to the local project
	GetCurrentProject(context.Context, *connect.Request[v1.GetCurrentProjectRequest]) (*connect.Response[v1.GetCurrentProjectResponse], error)
	// GetUserOrgMetadata returns metadata about the current user's orgs.
	GetUserOrgMetadata(context.Context, *connect.Request[v1.GetUserOrgMetadataRequest]) (*connect.Response[v1.GetUserOrgMetadataResponse], error)
}

// NewLocalServiceClient constructs a client for the rill.local.v1.LocalService service. By default,
// it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and
// sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC()
// or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewLocalServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) LocalServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &localServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+LocalServicePingProcedure,
			connect.WithSchema(localServicePingMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getMetadata: connect.NewClient[v1.GetMetadataRequest, v1.GetMetadataResponse](
			httpClient,
			baseURL+LocalServiceGetMetadataProcedure,
			connect.WithSchema(localServiceGetMetadataMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getVersion: connect.NewClient[v1.GetVersionRequest, v1.GetVersionResponse](
			httpClient,
			baseURL+LocalServiceGetVersionProcedure,
			connect.WithSchema(localServiceGetVersionMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		pushToGithub: connect.NewClient[v1.PushToGithubRequest, v1.PushToGithubResponse](
			httpClient,
			baseURL+LocalServicePushToGithubProcedure,
			connect.WithSchema(localServicePushToGithubMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		deployProject: connect.NewClient[v1.DeployProjectRequest, v1.DeployProjectResponse](
			httpClient,
			baseURL+LocalServiceDeployProjectProcedure,
			connect.WithSchema(localServiceDeployProjectMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		redeployProject: connect.NewClient[v1.RedeployProjectRequest, v1.RedeployProjectResponse](
			httpClient,
			baseURL+LocalServiceRedeployProjectProcedure,
			connect.WithSchema(localServiceRedeployProjectMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getCurrentUser: connect.NewClient[v1.GetCurrentUserRequest, v1.GetCurrentUserResponse](
			httpClient,
			baseURL+LocalServiceGetCurrentUserProcedure,
			connect.WithSchema(localServiceGetCurrentUserMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getCurrentProject: connect.NewClient[v1.GetCurrentProjectRequest, v1.GetCurrentProjectResponse](
			httpClient,
			baseURL+LocalServiceGetCurrentProjectProcedure,
			connect.WithSchema(localServiceGetCurrentProjectMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getUserOrgMetadata: connect.NewClient[v1.GetUserOrgMetadataRequest, v1.GetUserOrgMetadataResponse](
			httpClient,
			baseURL+LocalServiceGetUserOrgMetadataProcedure,
			connect.WithSchema(localServiceGetUserOrgMetadataMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// localServiceClient implements LocalServiceClient.
type localServiceClient struct {
	ping               *connect.Client[v1.PingRequest, v1.PingResponse]
	getMetadata        *connect.Client[v1.GetMetadataRequest, v1.GetMetadataResponse]
	getVersion         *connect.Client[v1.GetVersionRequest, v1.GetVersionResponse]
	pushToGithub       *connect.Client[v1.PushToGithubRequest, v1.PushToGithubResponse]
	deployProject      *connect.Client[v1.DeployProjectRequest, v1.DeployProjectResponse]
	redeployProject    *connect.Client[v1.RedeployProjectRequest, v1.RedeployProjectResponse]
	getCurrentUser     *connect.Client[v1.GetCurrentUserRequest, v1.GetCurrentUserResponse]
	getCurrentProject  *connect.Client[v1.GetCurrentProjectRequest, v1.GetCurrentProjectResponse]
	getUserOrgMetadata *connect.Client[v1.GetUserOrgMetadataRequest, v1.GetUserOrgMetadataResponse]
}

// Ping calls rill.local.v1.LocalService.Ping.
func (c *localServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// GetMetadata calls rill.local.v1.LocalService.GetMetadata.
func (c *localServiceClient) GetMetadata(ctx context.Context, req *connect.Request[v1.GetMetadataRequest]) (*connect.Response[v1.GetMetadataResponse], error) {
	return c.getMetadata.CallUnary(ctx, req)
}

// GetVersion calls rill.local.v1.LocalService.GetVersion.
func (c *localServiceClient) GetVersion(ctx context.Context, req *connect.Request[v1.GetVersionRequest]) (*connect.Response[v1.GetVersionResponse], error) {
	return c.getVersion.CallUnary(ctx, req)
}

// PushToGithub calls rill.local.v1.LocalService.PushToGithub.
func (c *localServiceClient) PushToGithub(ctx context.Context, req *connect.Request[v1.PushToGithubRequest]) (*connect.Response[v1.PushToGithubResponse], error) {
	return c.pushToGithub.CallUnary(ctx, req)
}

// DeployProject calls rill.local.v1.LocalService.DeployProject.
func (c *localServiceClient) DeployProject(ctx context.Context, req *connect.Request[v1.DeployProjectRequest]) (*connect.Response[v1.DeployProjectResponse], error) {
	return c.deployProject.CallUnary(ctx, req)
}

// RedeployProject calls rill.local.v1.LocalService.RedeployProject.
func (c *localServiceClient) RedeployProject(ctx context.Context, req *connect.Request[v1.RedeployProjectRequest]) (*connect.Response[v1.RedeployProjectResponse], error) {
	return c.redeployProject.CallUnary(ctx, req)
}

// GetCurrentUser calls rill.local.v1.LocalService.GetCurrentUser.
func (c *localServiceClient) GetCurrentUser(ctx context.Context, req *connect.Request[v1.GetCurrentUserRequest]) (*connect.Response[v1.GetCurrentUserResponse], error) {
	return c.getCurrentUser.CallUnary(ctx, req)
}

// GetCurrentProject calls rill.local.v1.LocalService.GetCurrentProject.
func (c *localServiceClient) GetCurrentProject(ctx context.Context, req *connect.Request[v1.GetCurrentProjectRequest]) (*connect.Response[v1.GetCurrentProjectResponse], error) {
	return c.getCurrentProject.CallUnary(ctx, req)
}

// GetUserOrgMetadata calls rill.local.v1.LocalService.GetUserOrgMetadata.
func (c *localServiceClient) GetUserOrgMetadata(ctx context.Context, req *connect.Request[v1.GetUserOrgMetadataRequest]) (*connect.Response[v1.GetUserOrgMetadataResponse], error) {
	return c.getUserOrgMetadata.CallUnary(ctx, req)
}

// LocalServiceHandler is an implementation of the rill.local.v1.LocalService service.
type LocalServiceHandler interface {
	// Ping returns the current time.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// GetMetadata returns information about the local Rill instance.
	GetMetadata(context.Context, *connect.Request[v1.GetMetadataRequest]) (*connect.Response[v1.GetMetadataResponse], error)
	// GetVersion returns details about the current and latest available Rill versions.
	GetVersion(context.Context, *connect.Request[v1.GetVersionRequest]) (*connect.Response[v1.GetVersionResponse], error)
	// PushToGithub create a Git repo from local project and pushed to users git account.
	PushToGithub(context.Context, *connect.Request[v1.PushToGithubRequest]) (*connect.Response[v1.PushToGithubResponse], error)
	// DeployProject deploys the local project to the Rill cloud.
	DeployProject(context.Context, *connect.Request[v1.DeployProjectRequest]) (*connect.Response[v1.DeployProjectResponse], error)
	// RedeployProject updates a deployed project.
	RedeployProject(context.Context, *connect.Request[v1.RedeployProjectRequest]) (*connect.Response[v1.RedeployProjectResponse], error)
	// GetCurrentUser returns the locally logged in user
	GetCurrentUser(context.Context, *connect.Request[v1.GetCurrentUserRequest]) (*connect.Response[v1.GetCurrentUserResponse], error)
	// GetCurrentProject returns the rill cloud project connected to the local project
	GetCurrentProject(context.Context, *connect.Request[v1.GetCurrentProjectRequest]) (*connect.Response[v1.GetCurrentProjectResponse], error)
	// GetUserOrgMetadata returns metadata about the current user's orgs.
	GetUserOrgMetadata(context.Context, *connect.Request[v1.GetUserOrgMetadataRequest]) (*connect.Response[v1.GetUserOrgMetadataResponse], error)
}

// NewLocalServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewLocalServiceHandler(svc LocalServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	localServicePingHandler := connect.NewUnaryHandler(
		LocalServicePingProcedure,
		svc.Ping,
		connect.WithSchema(localServicePingMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	localServiceGetMetadataHandler := connect.NewUnaryHandler(
		LocalServiceGetMetadataProcedure,
		svc.GetMetadata,
		connect.WithSchema(localServiceGetMetadataMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	localServiceGetVersionHandler := connect.NewUnaryHandler(
		LocalServiceGetVersionProcedure,
		svc.GetVersion,
		connect.WithSchema(localServiceGetVersionMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	localServicePushToGithubHandler := connect.NewUnaryHandler(
		LocalServicePushToGithubProcedure,
		svc.PushToGithub,
		connect.WithSchema(localServicePushToGithubMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	localServiceDeployProjectHandler := connect.NewUnaryHandler(
		LocalServiceDeployProjectProcedure,
		svc.DeployProject,
		connect.WithSchema(localServiceDeployProjectMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	localServiceRedeployProjectHandler := connect.NewUnaryHandler(
		LocalServiceRedeployProjectProcedure,
		svc.RedeployProject,
		connect.WithSchema(localServiceRedeployProjectMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	localServiceGetCurrentUserHandler := connect.NewUnaryHandler(
		LocalServiceGetCurrentUserProcedure,
		svc.GetCurrentUser,
		connect.WithSchema(localServiceGetCurrentUserMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	localServiceGetCurrentProjectHandler := connect.NewUnaryHandler(
		LocalServiceGetCurrentProjectProcedure,
		svc.GetCurrentProject,
		connect.WithSchema(localServiceGetCurrentProjectMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	localServiceGetUserOrgMetadataHandler := connect.NewUnaryHandler(
		LocalServiceGetUserOrgMetadataProcedure,
		svc.GetUserOrgMetadata,
		connect.WithSchema(localServiceGetUserOrgMetadataMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/rill.local.v1.LocalService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case LocalServicePingProcedure:
			localServicePingHandler.ServeHTTP(w, r)
		case LocalServiceGetMetadataProcedure:
			localServiceGetMetadataHandler.ServeHTTP(w, r)
		case LocalServiceGetVersionProcedure:
			localServiceGetVersionHandler.ServeHTTP(w, r)
		case LocalServicePushToGithubProcedure:
			localServicePushToGithubHandler.ServeHTTP(w, r)
		case LocalServiceDeployProjectProcedure:
			localServiceDeployProjectHandler.ServeHTTP(w, r)
		case LocalServiceRedeployProjectProcedure:
			localServiceRedeployProjectHandler.ServeHTTP(w, r)
		case LocalServiceGetCurrentUserProcedure:
			localServiceGetCurrentUserHandler.ServeHTTP(w, r)
		case LocalServiceGetCurrentProjectProcedure:
			localServiceGetCurrentProjectHandler.ServeHTTP(w, r)
		case LocalServiceGetUserOrgMetadataProcedure:
			localServiceGetUserOrgMetadataHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedLocalServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedLocalServiceHandler struct{}

func (UnimplementedLocalServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("rill.local.v1.LocalService.Ping is not implemented"))
}

func (UnimplementedLocalServiceHandler) GetMetadata(context.Context, *connect.Request[v1.GetMetadataRequest]) (*connect.Response[v1.GetMetadataResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("rill.local.v1.LocalService.GetMetadata is not implemented"))
}

func (UnimplementedLocalServiceHandler) GetVersion(context.Context, *connect.Request[v1.GetVersionRequest]) (*connect.Response[v1.GetVersionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("rill.local.v1.LocalService.GetVersion is not implemented"))
}

func (UnimplementedLocalServiceHandler) PushToGithub(context.Context, *connect.Request[v1.PushToGithubRequest]) (*connect.Response[v1.PushToGithubResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("rill.local.v1.LocalService.PushToGithub is not implemented"))
}

func (UnimplementedLocalServiceHandler) DeployProject(context.Context, *connect.Request[v1.DeployProjectRequest]) (*connect.Response[v1.DeployProjectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("rill.local.v1.LocalService.DeployProject is not implemented"))
}

func (UnimplementedLocalServiceHandler) RedeployProject(context.Context, *connect.Request[v1.RedeployProjectRequest]) (*connect.Response[v1.RedeployProjectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("rill.local.v1.LocalService.RedeployProject is not implemented"))
}

func (UnimplementedLocalServiceHandler) GetCurrentUser(context.Context, *connect.Request[v1.GetCurrentUserRequest]) (*connect.Response[v1.GetCurrentUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("rill.local.v1.LocalService.GetCurrentUser is not implemented"))
}

func (UnimplementedLocalServiceHandler) GetCurrentProject(context.Context, *connect.Request[v1.GetCurrentProjectRequest]) (*connect.Response[v1.GetCurrentProjectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("rill.local.v1.LocalService.GetCurrentProject is not implemented"))
}

func (UnimplementedLocalServiceHandler) GetUserOrgMetadata(context.Context, *connect.Request[v1.GetUserOrgMetadataRequest]) (*connect.Response[v1.GetUserOrgMetadataResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("rill.local.v1.LocalService.GetUserOrgMetadata is not implemented"))
}
