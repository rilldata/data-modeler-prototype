package local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/examples"
	"github.com/rilldata/rill/cli/pkg/update"
	"github.com/rilldata/rill/cli/pkg/variable"
	"github.com/rilldata/rill/cli/pkg/web"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	runtimeserver "github.com/rilldata/rill/runtime/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

type LogFormat string

// Default log formats for logger
const (
	LogFormatConsole = "console"
	LogFormatJSON    = "json"
)

// Default instance config on local.
const (
	DefaultInstanceID = "default"
	DefaultOLAPDriver = "duckdb"
	DefaultOLAPDSN    = "stage.db"
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
	Version               config.Version
	Verbose               bool
	ProjectPath           string
	observabilityShutdown observability.ShutdownFunc
}

func NewApp(ctx context.Context, ver config.Version, verbose bool, olapDriver, olapDSN, projectPath string, logFormat LogFormat, variables []string) (*App, error) {
	// Setup a friendly-looking colored/json logger
	var logger *zap.Logger
	var err error
	switch logFormat {
	case LogFormatJSON:
		cfg := zap.NewProductionConfig()
		cfg.DisableStacktrace = true
		cfg.Level.SetLevel(zapcore.DebugLevel)
		logger, err = cfg.Build()
	case LogFormatConsole:
		encCfg := zap.NewDevelopmentEncoderConfig()
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger = zap.New(zapcore.NewCore(
			zapcore.NewConsoleEncoder(encCfg),
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		))
	}

	if err != nil {
		return nil, err
	}

	// Set logging level
	lvl := zap.InfoLevel
	if verbose {
		lvl = zap.DebugLevel
	}
	logger = logger.WithOptions(zap.IncreaseLevel(lvl))

	// Init Prometheus telemetry
	shutdown, err := observability.Start(ctx, logger, &observability.Options{
		MetricsExporter: observability.PrometheusExporter,
		TracesExporter:  observability.NoopExporter,
		ServiceName:     "rill-local",
		ServiceVersion:  ver.String(),
	})
	if err != nil {
		return nil, err
	}

	// Create a local runtime with an in-memory metastore
	rtOpts := &runtime.Options{
		ConnectionCacheSize: 100,
		MetastoreDriver:     "sqlite",
		MetastoreDSN:        "file:rill?mode=memory&cache=shared",
		QueryCacheSize:      10000,
		AllowHostAccess:     true,
	}
	rt, err := runtime.New(rtOpts, logger)
	if err != nil {
		return nil, err
	}

	// Get full path to project
	projectPath, err = filepath.Abs(projectPath)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(projectPath, os.ModePerm) // Create project dir if it doesn't exist
	if err != nil {
		return nil, err
	}

	// If the OLAP is the default OLAP (DuckDB in stage.db), we make it relative to the project directory (not the working directory)
	if olapDriver == DefaultOLAPDriver && olapDSN == DefaultOLAPDSN {
		olapDSN = path.Join(projectPath, olapDSN)
	}

	parsedVariables, err := variable.Parse(variables)
	if err != nil {
		return nil, err
	}

	// Create instance with its repo set to the project directory
	inst := &drivers.Instance{
		ID:           DefaultInstanceID,
		OLAPDriver:   olapDriver,
		OLAPDSN:      olapDSN,
		RepoDriver:   "file",
		RepoDSN:      projectPath,
		EmbedCatalog: olapDriver == "duckdb",
		Variables:    parsedVariables,
	}
	err = rt.CreateInstance(ctx, inst)
	if err != nil {
		return nil, err
	}

	// Done
	app := &App{
		Context:               ctx,
		Runtime:               rt,
		Instance:              inst,
		Logger:                logger.Sugar(),
		BaseLogger:            logger,
		Version:               ver,
		Verbose:               verbose,
		ProjectPath:           projectPath,
		observabilityShutdown: shutdown,
	}
	return app, nil
}

func (a *App) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := a.observabilityShutdown(ctx)
	if err != nil {
		fmt.Printf("telemetry shutdown failed: %s\n", err.Error())
	}

	return a.Runtime.Close()
}

func (a *App) IsProjectInit() bool {
	repo, err := a.Runtime.Repo(a.Context, a.Instance.ID)
	if err != nil {
		panic(err) // checks in New should ensure it never happens
	}

	c := rillv1beta.New(repo, a.Instance.ID)
	return c.IsInit(a.Context)
}

func (a *App) InitProject(exampleName string) error {
	repo, err := a.Runtime.Repo(a.Context, a.Instance.ID)
	if err != nil {
		panic(err) // checks in New should ensure it never happens
	}

	c := rillv1beta.New(repo, a.Instance.ID)
	if c.IsInit(a.Context) {
		return fmt.Errorf("a Rill project already exists")
	}

	// Check if project path is pwd for nicer log messages
	pwd, _ := os.Getwd()
	isPwd := a.ProjectPath == pwd

	// If no example is provided, init an empty project
	if exampleName == "" {
		// Infer a default project name from its path
		defaultName := filepath.Base(a.ProjectPath)
		if defaultName == "" || defaultName == "." || defaultName == ".." {
			defaultName = "untitled"
		}

		// Init empty project
		err := c.InitEmpty(a.Context, defaultName, a.Version.Number)
		if err != nil {
			if isPwd {
				return fmt.Errorf("failed to initialize project in the current directory (detailed error: %w)", err)
			}
			return fmt.Errorf("failed to initialize project in '%s' (detailed error: %w)", a.ProjectPath, err)
		}

		// Log success
		if isPwd {
			a.Logger.Infof("Initialized empty project in the current directory")
		} else {
			a.Logger.Infof("Initialized empty project at '%s'", a.ProjectPath)
		}

		return nil
	}

	// It's an example project. We currently only support examples through direct file unpacking.
	// TODO: Support unpacking examples through rillv1beta, instead of unpacking files.

	err = examples.Init(exampleName, a.ProjectPath)
	if err != nil {
		if errors.Is(err, examples.ErrExampleNotFound) {
			return fmt.Errorf("example project '%s' not found", exampleName)
		}
		return fmt.Errorf("failed to initialize project (detailed error: %w)", err)
	}

	if isPwd {
		a.Logger.Infof("Initialized example project '%s' in the current directory", exampleName)
	} else {
		a.Logger.Infof("Initialized example project '%s' in directory '%s'", exampleName, a.ProjectPath)
	}

	return nil
}

func (a *App) Reconcile(strict bool) error {
	a.Logger.Infof("Hydrating project '%s'", a.ProjectPath)
	res, err := a.Runtime.Reconcile(a.Context, a.Instance.ID, nil, nil, false, false)
	if err != nil {
		return err
	}
	if a.Context.Err() != nil {
		a.Logger.Errorf("Hydration canceled")
	}
	for _, path := range res.AffectedPaths {
		a.Logger.Infof("Reconciled: %s", path)
	}
	for _, merr := range res.Errors {
		a.Logger.Errorf("%s: %s", merr.FilePath, merr.Message)
	}
	if len(res.Errors) == 0 {
		a.Logger.Infof("Hydration completed!")
	} else if strict {
		a.Logger.Fatalf("Hydration failed")
	} else {
		a.Logger.Infof("Hydration failed")
	}
	return nil
}

func (a *App) ReconcileSource(sourcePath string) error {
	a.Logger.Infof("Reconciling source and impacted models in project '%s'", a.ProjectPath)
	paths := []string{sourcePath}
	res, err := a.Runtime.Reconcile(a.Context, a.Instance.ID, paths, paths, false, false)
	if err != nil {
		return err
	}
	if a.Context.Err() != nil {
		a.Logger.Errorf("Hydration canceled")
		return nil
	}
	for _, path := range res.AffectedPaths {
		a.Logger.Infof("Reconciled: %s", path)
	}
	for _, merr := range res.Errors {
		a.Logger.Errorf("%s: %s", merr.FilePath, merr.Message)
	}
	if len(res.Errors) == 0 {
		a.Logger.Infof("Hydration completed!")
	} else {
		a.Logger.Infof("Hydration failed")
	}
	return nil
}

func (a *App) Serve(httpPort, grpcPort int, enableUI, openBrowser, readonly bool, userID string) error {
	// Get analytics info
	installID, enabled, err := dotrill.AnalyticsInfo()
	if err != nil {
		a.Logger.Warnf("error finding install ID: %v", err)
	}

	// Build local info for frontend
	inf := &localInfo{
		InstanceID:       a.Instance.ID,
		GRPCPort:         grpcPort,
		InstallID:        installID,
		ProjectPath:      a.ProjectPath,
		UserID:           userID,
		Version:          a.Version.Number,
		BuildCommit:      a.Version.Commit,
		BuildTime:        a.Version.Timestamp,
		IsDev:            a.Version.IsDev(),
		AnalyticsEnabled: enabled,
		Readonly:         readonly,
	}

	// Create server logger.
	// It only logs error messages when !verbose to prevent lots of req/res info messages.
	lvl := zap.ErrorLevel
	if a.Verbose {
		lvl = zap.DebugLevel
	}
	serverLogger := a.BaseLogger.WithOptions(zap.IncreaseLevel(lvl))

	// Prepare errgroup and context with graceful shutdown
	gctx := graceful.WithCancelOnTerminate(a.Context)
	group, ctx := errgroup.WithContext(gctx)

	// Create a runtime server
	opts := &runtimeserver.Options{
		HTTPPort:        httpPort,
		GRPCPort:        grpcPort,
		AllowedOrigins:  []string{"*"},
		ServePrometheus: true,
	}
	runtimeServer, err := runtimeserver.NewServer(opts, a.Runtime, serverLogger)
	if err != nil {
		return err
	}

	// Start the gRPC server
	group.Go(func() error {
		return runtimeServer.ServeGRPC(ctx)
	})

	// Start the local HTTP server
	group.Go(func() error {
		return runtimeServer.ServeHTTP(ctx, func(mux *http.ServeMux) {
			// Inject local-only endpoints on the server for the local UI and local backend endpoints
			if enableUI {
				mux.Handle("/", web.StaticHandler())
			}
			mux.Handle("/local/config", a.infoHandler(inf))
			mux.Handle("/local/version", a.versionHandler())
			mux.Handle("/local/track", a.trackingHandler(inf))
		})
	})

	// Open the browser when health check succeeds
	go a.pollServer(ctx, httpPort, enableUI && openBrowser)

	// Run the server
	err = group.Wait()
	if err != nil {
		return fmt.Errorf("server crashed: %w", err)
	}
	a.Logger.Info("Rill shutdown gracefully")
	return nil
}

func (a *App) pollServer(ctx context.Context, httpPort int, openOnHealthy bool) {
	// Basic health check
	uri := fmt.Sprintf("http://localhost:%d", httpPort)
	client := http.Client{Timeout: time.Second}
	for {
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

		// Wait a bit and retry
		time.Sleep(250 * time.Millisecond)
	}

	// Health check succeeded
	a.Logger.Infof("Serving Rill on: %s", uri)
	if openOnHealthy {
		err := browser.Open(uri)
		if err != nil {
			a.Logger.Debugf("could not open browser: %v", err)
		}
	}
}

type localInfo struct {
	InstanceID       string `json:"instance_id"`
	GRPCPort         int    `json:"grpc_port"`
	InstallID        string `json:"install_id"`
	UserID           string `json:"user_id"`
	ProjectPath      string `json:"project_path"`
	Version          string `json:"version"`
	BuildCommit      string `json:"build_commit"`
	BuildTime        string `json:"build_time"`
	IsDev            bool   `json:"is_dev"`
	AnalyticsEnabled bool   `json:"analytics_enabled"`
	Readonly         bool   `json:"readonly"`
}

// infoHandler servers the local info struct.
func (a *App) infoHandler(info *localInfo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(info)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		_, err = w.Write(data)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to write response data: %s", err), http.StatusInternalServerError)
			return
		}
	})
}

// versionHandler servers the version struct.
func (a *App) versionHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the latest version available
		latestVersion, err := update.LatestVersion(r.Context())
		if err != nil {
			a.Logger.Warnf("error finding latest version: %v", err)
		}

		inf := &versionInfo{
			CurrentVersion: a.Version.Number,
			LatestVersion:  latestVersion,
		}

		data, err := json.Marshal(inf)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		_, err = w.Write(data)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to write response data: %s", err), http.StatusInternalServerError)
			return
		}
	})
}

type versionInfo struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
}

// trackingHandler proxies events to intake.rilldata.io.
func (a *App) trackingHandler(info *localInfo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !info.AnalyticsEnabled {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Proxy request to rill intake
		proxyReq, err := http.NewRequest(r.Method, "https://intake.rilldata.io/events/data-modeler-metrics", r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		// Copy the auth header
		proxyReq.Header = http.Header{
			"Authorization": r.Header["Authorization"],
		}

		// Send proxied request
		resp, err := http.DefaultClient.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Done
		w.WriteHeader(http.StatusOK)
	})
}

func ParseLogFormat(format string) (LogFormat, bool) {
	switch format {
	case "json":
		return LogFormatJSON, true
	case "console":
		return LogFormatConsole, true
	default:
		return "", false
	}
}
