package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var minCliVersion = version.Must(version.NewVersion("0.20.0"))

type Options struct {
	HTTPPort               int
	GRPCPort               int
	ExternalURL            string
	FrontendURL            string
	SessionKeyPairs        [][]byte
	AllowedOrigins         []string
	ServePrometheus        bool
	AuthDomain             string
	AuthClientID           string
	AuthClientSecret       string
	GithubAppName          string
	GithubAppWebhookSecret string
	GithubClientID         string
	GithubClientSecret     string
}

type Server struct {
	adminv1.UnsafeAdminServiceServer
	logger        *zap.Logger
	admin         *admin.Service
	opts          *Options
	cookies       *sessions.CookieStore
	authenticator *auth.Authenticator
	issuer        *runtimeauth.Issuer
	urls          *externalURLs
}

var _ adminv1.AdminServiceServer = (*Server)(nil)

func New(opts *Options, logger *zap.Logger, adm *admin.Service, issuer *runtimeauth.Issuer) (*Server, error) {
	externalURL, err := url.Parse(opts.ExternalURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse external URL: %w", err)
	}

	if len(opts.SessionKeyPairs) == 0 {
		return nil, fmt.Errorf("provided SessionKeyPairs is empty")
	}

	cookies := sessions.NewCookieStore(opts.SessionKeyPairs...)
	cookies.MaxAge(60 * 60 * 24 * 365 * 10) // 10 years
	cookies.Options.Secure = externalURL.Scheme == "https"
	cookies.Options.HttpOnly = true

	authenticator, err := auth.NewAuthenticator(logger, adm, cookies, &auth.AuthenticatorOptions{
		AuthDomain:       opts.AuthDomain,
		AuthClientID:     opts.AuthClientID,
		AuthClientSecret: opts.AuthClientSecret,
		ExternalURL:      opts.ExternalURL,
		FrontendURL:      opts.FrontendURL,
	})
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:        logger,
		admin:         adm,
		opts:          opts,
		cookies:       cookies,
		authenticator: authenticator,
		issuer:        issuer,
		urls:          newURLRegistry(opts),
	}, nil
}

// ServeGRPC Starts the gRPC server.
func (s *Server) ServeGRPC(ctx context.Context) error {
	server := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			observability.TracingStreamServerInterceptor(),
			observability.LoggingStreamServerInterceptor(s.logger),
			observability.RecovererStreamServerInterceptor(),
			grpc_validator.StreamServerInterceptor(),
			s.authenticator.StreamServerInterceptor(),
			grpc_auth.StreamServerInterceptor(CheckUserAgent),
		),
		grpc.ChainUnaryInterceptor(
			observability.TracingUnaryServerInterceptor(),
			observability.LoggingUnaryServerInterceptor(s.logger),
			observability.RecovererUnaryServerInterceptor(),
			grpc_validator.UnaryServerInterceptor(),
			s.authenticator.UnaryServerInterceptor(),
			grpc_auth.UnaryServerInterceptor(CheckUserAgent),
		),
	)

	adminv1.RegisterAdminServiceServer(server, s)
	s.logger.Sugar().Infof("serving admin gRPC on port:%v", s.opts.GRPCPort)
	return graceful.ServeGRPC(ctx, server, s.opts.GRPCPort)
}

// Starts the HTTP server.
func (s *Server) ServeHTTP(ctx context.Context) error {
	handler, err := s.HTTPHandler(ctx)
	if err != nil {
		return err
	}

	server := &http.Server{Handler: handler}
	s.logger.Sugar().Infof("serving admin HTTP on port:%v", s.opts.HTTPPort)
	return graceful.ServeHTTP(ctx, server, s.opts.HTTPPort)
}

// HTTPHandler HTTP handler serving REST gateway.
func (s *Server) HTTPHandler(ctx context.Context) (http.Handler, error) {
	// Create REST gateway
	gwMux := gateway.NewServeMux(
		gateway.WithErrorHandler(HTTPErrorHandler),
		gateway.WithMetadata(s.authenticator.Annotator),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcAddress := fmt.Sprintf(":%d", s.opts.GRPCPort)
	err := adminv1.RegisterAdminServiceHandlerFromEndpoint(ctx, gwMux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}

	// Create regular http mux and mount gwMux on it
	mux := http.NewServeMux()
	mux.Handle("/v1/", gwMux)

	// Add Prometheus
	if s.opts.ServePrometheus {
		mux.Handle("/metrics", promhttp.Handler())
	}

	// Server public JWKS for runtime JWT verification
	mux.Handle("/.well-known/jwks.json", s.issuer.WellKnownHandler())

	// Add auth endpoints (not gRPC handlers, just regular endpoints on /auth/*)
	s.authenticator.RegisterEndpoints(mux)

	// Add Github-related endpoints (not gRPC handlers, just regular endpoints on /github/*)
	s.registerGithubEndpoints(mux)

	// Build CORS options for admin server

	// If the AllowedOrigins contains a "*" we want to return the requester's origin instead of "*" in the "Access-Control-Allow-Origin" header.
	// This is useful in development. In production, we set AllowedOrigins to non-wildcard values, so this does not have security implications.
	// Details: https://github.com/rs/cors#allow--with-credentials-security-protection
	var allowedOriginFunc func(string) bool
	allowedOrigins := s.opts.AllowedOrigins
	for _, origin := range s.opts.AllowedOrigins {
		if origin == "*" {
			allowedOriginFunc = func(origin string) bool { return true }
			allowedOrigins = nil
			break
		}
	}

	corsOpts := cors.Options{
		AllowedOrigins:  allowedOrigins,
		AllowOriginFunc: allowedOriginFunc,
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders: []string{"*"},
		// We use cookies for browser sessions, so this is required to allow ui.rilldata.com to make authenticated requests to admin.rilldata.com
		AllowCredentials: true,
		// Set max age to 1 hour (default if not set is 5 seconds)
		MaxAge: 60 * 60,
	}

	// Wrap mux with CORS middleware
	handler := cors.New(corsOpts).Handler(mux)

	return handler, nil
}

// HTTPErrorHandler wraps gateway.DefaultHTTPErrorHandler to map gRPC unknown errors (i.e. errors without an explicit
// code) to HTTP status code 400 instead of 500.
func HTTPErrorHandler(ctx context.Context, mux *gateway.ServeMux, marshaler gateway.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	s := status.Convert(err)
	if s.Code() == codes.Unknown {
		err = &gateway.HTTPStatusError{HTTPStatus: http.StatusBadRequest, Err: err}
	}
	gateway.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
}

// Ping implements AdminService
func (s *Server) Ping(ctx context.Context, req *adminv1.PingRequest) (*adminv1.PingResponse, error) {
	resp := &adminv1.PingResponse{
		Version: "", // TODO: Return version
		Time:    timestamppb.New(time.Now()),
	}
	return resp, nil
}

func CheckUserAgent(ctx context.Context) (context.Context, error) {
	userAgent := strings.Split(metautils.ExtractIncoming(ctx).Get("user-agent"), " ")
	var ver string
	for _, s := range userAgent {
		if strings.HasPrefix(s, "rill-cli/") {
			ver = strings.TrimPrefix(s, "rill-cli/")
		}
	}

	// Check if build from source
	if ver == "unknown" || ver == "" {
		return ctx, nil
	}

	v, err := version.NewVersion(ver)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("could not parse rill-cli version: %s", err.Error()))
	}

	if v.LessThan(minCliVersion) {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("Rill %s is no longer supported, please upgrade to the latest version", v))
	}

	return ctx, nil
}

type externalURLs struct {
	githubConnectUI       string
	githubConnect         string
	githubConnectRetry    string
	githubConnectRequest  string
	githubConnectSuccess  string
	githubAppInstallation string
	githubAuth            string
	githubAuthCallback    string
	githubAuthRetry       string
	authLogin             string
}

func newURLRegistry(opts *Options) *externalURLs {
	return &externalURLs{
		githubConnectUI:       urlutil.MustJoinURL(opts.FrontendURL, "/-/github/connect"),
		githubConnect:         urlutil.MustJoinURL(opts.ExternalURL, "/github/connect"),
		githubConnectRetry:    urlutil.MustJoinURL(opts.FrontendURL, "/-/github/connect/retry-install"),
		githubConnectRequest:  urlutil.MustJoinURL(opts.FrontendURL, "/-/github/connect/request"),
		githubConnectSuccess:  urlutil.MustJoinURL(opts.FrontendURL, "/-/github/connect/success"),
		githubAppInstallation: fmt.Sprintf("https://github.com/apps/%s/installations/new", opts.GithubAppName),
		githubAuth:            urlutil.MustJoinURL(opts.ExternalURL, "/github/auth/login"),
		githubAuthCallback:    urlutil.MustJoinURL(opts.ExternalURL, "/github/auth/callback"),
		githubAuthRetry:       urlutil.MustJoinURL(opts.FrontendURL, "/-/github/connect/retry-auth"),
		authLogin:             urlutil.MustJoinURL(opts.ExternalURL, "/auth/login"),
	}
}
