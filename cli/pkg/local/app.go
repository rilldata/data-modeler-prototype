package local

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/pkce"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/debugserver"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	runtimeserver "github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/storage"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Default instance config on local.
const (
	DefaultInstanceID   = "default"
	DefaultOLAPDriver   = "duckdb"
	DefaultOLAPDSN      = "main.db"
	DefaultCatalogStore = "meta.db"
	DefaultDBDir        = "tmp"
)

// App encapsulates the logic associated with configuring and running the UI and the runtime in a local environment.
// Here, a local environment means a non-authenticated, single-instance and single-project setup on localhost.
// App encapsulates logic shared between different CLI commands, like start, init, build and source.
type App struct {
	Context               context.Context
	Runtime               *runtime.Runtime
	Instance              *drivers.Instance
	Logger                *zap.SugaredLogger
	BaseLogger            *zap.Logger
	Verbose               bool
	Debug                 bool
	ProjectPath           string
	ch                    *cmdutil.Helper
	observabilityShutdown observability.ShutdownFunc
	loggerCleanUp         func()
	pkceAuthenticators    map[string]*pkce.Authenticator // map of state to pkce authenticators
	localURL              string
	allowedOrigins        []string
}

type AppOptions struct {
	Ch             *cmdutil.Helper
	Verbose        bool
	Debug          bool
	Reset          bool
	Environment    string
	OlapDriver     string
	OlapDSN        string
	ProjectPath    string
	LogFormat      LogFormat
	Variables      map[string]string
	LocalURL       string
	AllowedOrigins []string
}

func NewApp(ctx context.Context, opts *AppOptions) (*App, error) {
	// Setup logger
	logger, cleanupFn := initLogger(opts.Verbose, opts.LogFormat)
	sugarLogger := logger.Sugar()

	// Init Prometheus telemetry
	shutdown, err := observability.Start(ctx, logger, &observability.Options{
		MetricsExporter: observability.PrometheusExporter,
		TracesExporter:  observability.NoopExporter,
		ServiceName:     "rill-local",
		ServiceVersion:  opts.Ch.Version.String(),
	})
	if err != nil {
		return nil, err
	}

	// Get full path to project
	projectPath, err := filepath.Abs(opts.ProjectPath)
	if err != nil {
		return nil, err
	}
	dbDirPath := filepath.Join(projectPath, DefaultDBDir)
	err = os.MkdirAll(dbDirPath, os.ModePerm) // Create project dir and db dir if it doesn't exist
	if err != nil {
		return nil, err
	}

	// old behaviour when data was stored in a stage.db file in the project directory.
	// drop old file, remove this code after some time
	_, err = os.Stat(filepath.Join(projectPath, "stage.db"))
	if err == nil { // a old stage.db file exists
		_ = os.Remove(filepath.Join(projectPath, "stage.db"))
		_ = os.Remove(filepath.Join(projectPath, "stage.db.wal"))
		logger.Info("Dropping old stage.db file and rebuilding project")
	}

	// Create a local runtime with an in-memory metastore
	systemConnectors := []*runtimev1.Connector{
		{
			Type:   "sqlite",
			Name:   "metastore",
			Config: map[string]string{"dsn": "file:rill?mode=memory&cache=shared"},
		},
	}

	// Sender for sending transactional emails.
	// We use a noop sender by default, but you can uncomment the SMTP sender to send emails from localhost for testing.
	sender := email.NewNoopSender()
	// Uncomment to send emails for testing:
	// err = godotenv.Load()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to load .env file: %w", err)
	// }
	// smtpPort, err := strconv.Atoi(os.Getenv("RILL_RUNTIME_EMAIL_SMTP_PORT"))
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get SMTP port: %w", err)
	// }
	// sender, err := email.NewSMTPSender(&email.SMTPOptions{
	// 	SMTPHost:     os.Getenv("RILL_RUNTIME_EMAIL_SMTP_HOST"),
	// 	SMTPPort:     smtpPort,
	// 	SMTPUsername: os.Getenv("RILL_RUNTIME_EMAIL_SMTP_USERNAME"),
	// 	SMTPPassword: os.Getenv("RILL_RUNTIME_EMAIL_SMTP_PASSWORD"),
	// 	FromEmail:    os.Getenv("RILL_RUNTIME_EMAIL_SENDER_EMAIL"),
	// 	FromName:     os.Getenv("RILL_RUNTIME_EMAIL_SENDER_NAME"),
	// 	BCC:          os.Getenv("RILL_RUNTIME_EMAIL_BCC"),
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create email sender: %w", err)
	// }
	rtOpts := &runtime.Options{
		ConnectionCacheSize:          100,
		MetastoreConnector:           "metastore",
		QueryCacheSizeBytes:          int64(datasize.MB * 100),
		AllowHostAccess:              true,
		SystemConnectors:             systemConnectors,
		SecurityEngineCacheSize:      1000,
		ControllerLogBufferCapacity:  10000,
		ControllerLogBufferSizeBytes: int64(datasize.MB * 16),
	}
	st, err := storage.New(dbDirPath, nil)
	if err != nil {
		return nil, err
	}
	rt, err := runtime.New(ctx, rtOpts, logger, st, opts.Ch.Telemetry(ctx), email.New(sender))
	if err != nil {
		return nil, err
	}

	// Merge opts.Variables with some local overrides of the defaults in runtime/drivers.InstanceConfig.
	vars := map[string]string{
		"rill.download_limit_bytes": "0", // 0 means unlimited
		"rill.stage_changes":        "false",
	}
	for k, v := range opts.Variables {
		vars[k] = v
	}

	// Prepare connectors for the instance
	var connectors []*runtimev1.Connector

	// Reset tmp dir
	if opts.Reset {
		_ = os.RemoveAll(dbDirPath)
		err = os.MkdirAll(dbDirPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	olapCfg := make(map[string]string)
	if opts.OlapDriver == "duckdb" {
		if opts.OlapDSN != DefaultOLAPDSN {
			return nil, fmt.Errorf("setting DSN for DuckDB is not supported")
		}
		// Set default DuckDB pool size to 4
		olapCfg["pool_size"] = "4"
	}

	// Add OLAP connector
	olapConnector := &runtimev1.Connector{
		Type:   opts.OlapDriver,
		Name:   opts.OlapDriver,
		Config: olapCfg,
	}
	connectors = append(connectors, olapConnector)

	// The repo connector is the local project directory
	repoConnector := &runtimev1.Connector{
		Type:   "file",
		Name:   "repo",
		Config: map[string]string{"dsn": projectPath},
	}
	connectors = append(connectors, repoConnector)

	// The catalog connector is a SQLite database in the project directory's tmp folder
	catalogConnector := &runtimev1.Connector{
		Type:   "sqlite",
		Name:   "catalog",
		Config: map[string]string{"dsn": fmt.Sprintf("file:%s?cache=shared", filepath.Join(dbDirPath, DefaultCatalogStore))},
	}
	connectors = append(connectors, catalogConnector)

	// Use the admin service for AI
	aiConnector := &runtimev1.Connector{
		Name: "admin",
		Type: "admin",
		Config: map[string]string{
			"admin_url":    opts.Ch.AdminURL(),
			"access_token": opts.Ch.AdminToken(),
		},
	}
	connectors = append(connectors, aiConnector)

	// Print start status – need to do it before creating the instance, since doing so immediately starts the controller
	isInit := IsProjectInit(projectPath)
	if isInit {
		sugarLogger.Infof("Hydrating project '%s'", projectPath)
	}

	// Create instance with its repo set to the project directory
	inst := &drivers.Instance{
		ID:               DefaultInstanceID,
		Environment:      opts.Environment,
		OLAPConnector:    olapConnector.Name,
		RepoConnector:    repoConnector.Name,
		AIConnector:      aiConnector.Name,
		CatalogConnector: catalogConnector.Name,
		Connectors:       connectors,
		Variables:        vars,
		Annotations:      map[string]string{},
		WatchRepo:        true,
		// ModelMaterializeDelaySeconds:     30, // TODO: Enable when we support skipping it for the initial load
		IgnoreInitialInvalidProjectError: !isInit, // See ProjectParser reconciler for details
	}
	err = rt.CreateInstance(ctx, inst)
	if err != nil {
		return nil, err
	}

	// Create app
	app := &App{
		Context:               ctx,
		Runtime:               rt,
		Instance:              inst,
		Logger:                sugarLogger,
		BaseLogger:            logger,
		Verbose:               opts.Verbose,
		Debug:                 opts.Debug,
		ProjectPath:           projectPath,
		ch:                    opts.Ch,
		observabilityShutdown: shutdown,
		loggerCleanUp:         cleanupFn,
		pkceAuthenticators:    make(map[string]*pkce.Authenticator),
		localURL:              opts.LocalURL,
		allowedOrigins:        opts.AllowedOrigins,
	}

	// Collect and emit information about connectors at start time
	err = app.emitStartEvent(ctx)
	if err != nil {
		logger.Debug("failed to emit start event", zap.Error(err))
	}

	return app, nil
}

func (a *App) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := a.observabilityShutdown(ctx)
	if err != nil {
		a.Logger.Error("Observability shutdown failed", zap.Error(err))
	}

	err = a.Runtime.Close()
	if err != nil {
		a.Logger.Error("Graceful shutdown failed", zap.Error(err))
	} else {
		a.Logger.Info("Rill shutdown gracefully")
	}

	a.loggerCleanUp()
	return nil
}

func (a *App) Serve(httpPort, grpcPort int, enableUI, openBrowser, readonly bool, userID, tlsCertPath, tlsKeyPath string) error {
	// Get analytics info
	installID, enabled, err := dotrill.AnalyticsInfo()
	if err != nil {
		a.Logger.Warnf("error finding install ID: %v", err)
	}

	// Build local metadata
	metadata := &localMetadata{
		InstanceID:       a.Instance.ID,
		GRPCPort:         grpcPort,
		InstallID:        installID,
		ProjectPath:      a.ProjectPath,
		UserID:           userID,
		Version:          a.ch.Version.Number,
		BuildCommit:      a.ch.Version.Commit,
		BuildTime:        a.ch.Version.Timestamp,
		IsDev:            a.ch.Version.IsDev(),
		AnalyticsEnabled: enabled,
		Readonly:         readonly,
	}

	// Create the local server handler
	localServer := &Server{
		logger:   a.BaseLogger,
		app:      a,
		metadata: metadata,
	}

	// Prepare errgroup and context with graceful shutdown
	gctx := graceful.WithCancelOnTerminate(a.Context)
	group, ctx := errgroup.WithContext(gctx)

	// Create server logger for the runtime
	runtimeServerLogger := a.BaseLogger
	if !a.Verbose {
		// It only logs error messages when !verbose to prevent lots of req/res info messages.
		runtimeServerLogger = a.BaseLogger.WithOptions(zap.IncreaseLevel(zap.ErrorLevel))
	}

	// Create a runtime server
	opts := &runtimeserver.Options{
		HTTPPort:        httpPort,
		GRPCPort:        grpcPort,
		TLSCertPath:     tlsCertPath,
		TLSKeyPath:      tlsKeyPath,
		AllowedOrigins:  a.allowedOrigins,
		ServePrometheus: true,
	}
	runtimeServer, err := runtimeserver.NewServer(ctx, opts, a.Runtime, runtimeServerLogger, ratelimit.NewNoop(), a.ch.Telemetry(ctx))
	if err != nil {
		return err
	}

	// Start the gRPC server
	group.Go(func() error {
		return runtimeServer.ServeGRPC(ctx)
	})

	// if keypath and certpath are provided
	secure := tlsCertPath != "" && tlsKeyPath != ""

	// Start the local HTTP server
	group.Go(func() error {
		return runtimeServer.ServeHTTP(ctx, func(mux *http.ServeMux) {
			// Inject local-only endpoints on the runtime server
			localServer.RegisterHandlers(mux, httpPort, secure, enableUI)
		})
	})

	// Start debug server on port 6060
	if a.Debug {
		group.Go(func() error { return debugserver.ServeHTTP(ctx, 6060) })
	}

	// Open the browser when health check succeeds
	go a.PollServer(ctx, httpPort, enableUI && openBrowser, secure)

	// Run the server
	err = group.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("server crashed: %w", err)
	}

	return nil
}

func (a *App) PollServer(ctx context.Context, httpPort int, openOnHealthy, secure bool) {
	client := &http.Client{Timeout: time.Second}

	scheme := "http"
	if secure {
		scheme = "https"
		client.Transport = &http.Transport{
			// nolint:gosec // this is a health check against localhost, so it's safe to ignore the cert
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	uri := fmt.Sprintf("%s://localhost:%d", scheme, httpPort)

	for {
		// Wait a bit before (re)trying.
		//
		// We sleep before the first health check as a slightly hacky way to protect against the situation where
		// another Rill server is already running, which will pass the health check as a false positive.
		// By sleeping first, the ctx is in practice sure to have been cancelled with a "port taken" error at that point.
		time.Sleep(250 * time.Millisecond)

		// Check for cancellation
		if ctx.Err() != nil {
			return
		}

		// Check if server is up
		resp, err := client.Get(uri + "/v1/ping")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < http.StatusInternalServerError {
				break
			}
		}
	}

	// Health check succeeded
	a.Logger.Infof("Serving Rill on: %s", uri)
	if openOnHealthy {
		// Check for cancellation again to be safe
		if ctx.Err() != nil {
			return
		}

		// Open the browser
		err := browser.Open(uri)
		if err != nil {
			a.Logger.Debugf("could not open browser: %v", err)
		}
	}
}

// emitStartEvent sends a telemetry event with information about the project' state.
// It is not a blocking operation (events are flushed in the background).
func (a *App) emitStartEvent(ctx context.Context) error {
	repo, instanceID, err := cmdutil.RepoForProjectPath(a.ProjectPath)
	if err != nil {
		return err
	}

	parser, err := rillv1.Parse(ctx, repo, instanceID, a.Instance.Environment, a.Instance.OLAPConnector)
	if err != nil {
		return err
	}

	connectors := parser.AnalyzeConnectors(ctx)
	for _, c := range connectors {
		if c.Err != nil {
			return err
		}
	}

	var connectorNames []string
	for _, connector := range connectors {
		connectorNames = append(connectorNames, connector.Name)
	}

	a.ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventAppStart, attribute.StringSlice("connectors", connectorNames), attribute.String("olap_connector", a.Instance.OLAPConnector))

	return nil
}

// IsProjectInit checks if the project is initialized by checking if rill.yaml exists in the project directory.
// It doesn't use any runtime functions since we need the ability to check this before creating the instance.
func IsProjectInit(projectPath string) bool {
	rillYAML := filepath.Join(projectPath, "rill.yaml")
	if _, err := os.Stat(rillYAML); err != nil {
		return false
	}
	return true
}
