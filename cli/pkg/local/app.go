package local

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bmatcuk/doublestar/v4"
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
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogFormat string

// Default log formats for logger
const (
	LogFormatConsole = "console"
	LogFormatJSON    = "json"
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
	Version               cmdutil.Version
	Verbose               bool
	Debug                 bool
	ProjectPath           string
	observabilityShutdown observability.ShutdownFunc
	loggerCleanUp         func()
	activity              *activity.Client
	adminURL              string
	pkceAuthenticators    map[string]*pkce.Authenticator // map of state to pkce authenticators
	ch                    *cmdutil.Helper
	localURL              string
}

type AppOptions struct {
	Version     cmdutil.Version
	Verbose     bool
	Debug       bool
	Reset       bool
	Environment string
	OlapDriver  string
	OlapDSN     string
	ProjectPath string
	LogFormat   LogFormat
	Variables   map[string]string
	Activity    *activity.Client
	AdminURL    string
	AdminToken  string
	CMDHelper   *cmdutil.Helper
	LocalURL    string
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
		ServiceVersion:  opts.Version.String(),
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
		DataDir:                      dbDirPath,
		SystemConnectors:             systemConnectors,
		SecurityEngineCacheSize:      1000,
		ControllerLogBufferCapacity:  10000,
		ControllerLogBufferSizeBytes: int64(datasize.MB * 16),
	}
	rt, err := runtime.New(ctx, rtOpts, logger, opts.Activity, email.New(sender))
	if err != nil {
		return nil, err
	}

	// Merge opts.Variables with some local overrides of the defaults in runtime/drivers.InstanceConfig.
	vars := map[string]string{
		"rill.download_row_limit": "0", // 0 means unlimited
		"rill.stage_changes":      "false",
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

	// If the OLAP is the default OLAP (DuckDB in stage.db), we make it relative to the project directory (not the working directory)
	defaultOLAP := false
	olapDSN := opts.OlapDSN
	olapCfg := make(map[string]string)
	if opts.OlapDriver == DefaultOLAPDriver && olapDSN == DefaultOLAPDSN {
		defaultOLAP = true
		olapDSN = path.Join(dbDirPath, olapDSN)
		// Set path which overrides the duckdb's default behaviour to store duckdb data in data_dir/<instance_id>/<connector> directory which is not backward compatible
		olapCfg["path"] = olapDSN
		val, err := isExternalStorageEnabled(dbDirPath, vars)
		if err != nil {
			return nil, err
		}

		olapCfg["external_table_storage"] = strconv.FormatBool(val)
	}

	// Set default DuckDB pool size to 4
	olapCfg["dsn"] = olapDSN
	if opts.OlapDriver == "duckdb" {
		olapCfg["pool_size"] = "4"
		if !defaultOLAP {
			olapCfg["error_on_incompatible_version"] = "true"
		}
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
			"admin_url":    opts.AdminURL,
			"access_token": opts.AdminToken,
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
		Version:               opts.Version,
		Verbose:               opts.Verbose,
		Debug:                 opts.Debug,
		ProjectPath:           projectPath,
		observabilityShutdown: shutdown,
		loggerCleanUp:         cleanupFn,
		activity:              opts.Activity,
		adminURL:              opts.AdminURL,
		pkceAuthenticators:    make(map[string]*pkce.Authenticator),
		ch:                    opts.CMDHelper,
		localURL:              opts.LocalURL,
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
		Version:          a.Version.Number,
		BuildCommit:      a.Version.Commit,
		BuildTime:        a.Version.Timestamp,
		IsDev:            a.Version.IsDev(),
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
		AllowedOrigins:  []string{"*"},
		ServePrometheus: true,
	}
	runtimeServer, err := runtimeserver.NewServer(ctx, opts, a.Runtime, runtimeServerLogger, ratelimit.NewNoop(), a.activity)
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
	go a.pollServer(ctx, httpPort, enableUI && openBrowser, secure)

	// Run the server
	err = group.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("server crashed: %w", err)
	}

	return nil
}

func (a *App) pollServer(ctx context.Context, httpPort int, openOnHealthy, secure bool) {
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

	a.activity.RecordBehavioralLegacy(activity.BehavioralEventAppStart, attribute.StringSlice("connectors", connectorNames), attribute.String("olap_connector", a.Instance.OLAPConnector))

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

func initLogger(isVerbose bool, logFormat LogFormat) (logger *zap.Logger, cleanupFn func()) {
	logLevel := zapcore.InfoLevel
	if isVerbose {
		logLevel = zapcore.DebugLevel
	}

	logPath, err := dotrill.ResolveFilename("rill.log", true)
	if err != nil {
		panic(err)
	}
	// lumberjack.Logger is already safe for concurrent use, so we don't need to
	// lock it.
	luLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     30, // days
		Compress:   true,
	}
	cfg := zap.NewProductionEncoderConfig()
	// hide logger name like `console`
	cfg.NameKey = zapcore.OmitKey
	fileCore := zapcore.NewCore(zapcore.NewJSONEncoder(cfg), zapcore.AddSync(luLogger), logLevel)

	var consoleEncoder zapcore.Encoder
	opts := make([]zap.Option, 0)
	switch logFormat {
	case LogFormatJSON:
		cfg := zap.NewProductionEncoderConfig()
		cfg.NameKey = zapcore.OmitKey
		// never
		opts = append(opts, zap.AddStacktrace(zapcore.InvalidLevel))
		consoleEncoder = zapcore.NewJSONEncoder(cfg)
	case LogFormatConsole:
		encCfg := zap.NewDevelopmentEncoderConfig()
		encCfg.NameKey = zapcore.OmitKey
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000")
		consoleEncoder = zapcore.NewConsoleEncoder(encCfg)
	}

	// if it's not verbose, skip instance_id field
	if !isVerbose {
		consoleEncoder = skipFieldZapEncoder{
			Encoder: consoleEncoder,
			fields:  []string{"instance_id"},
		}
	}

	core := zapcore.NewTee(
		fileCore,
		zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), logLevel),
	)

	return zap.New(core, opts...), func() {
		_ = logger.Sync()
		luLogger.Close()
	}
}

// skipFieldZapEncoder skips fields with the given keys. only string fields are supported.
type skipFieldZapEncoder struct {
	zapcore.Encoder
	fields []string
}

func (s skipFieldZapEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	res := make([]zapcore.Field, 0, len(fields))
	for _, field := range fields {
		skip := false
		for _, skipField := range s.fields {
			if field.Key == skipField {
				skip = true
				break
			}
		}
		if !skip {
			res = append(res, field)
		}
	}
	return s.Encoder.EncodeEntry(entry, res)
}

func (s skipFieldZapEncoder) Clone() zapcore.Encoder {
	return skipFieldZapEncoder{
		Encoder: s.Encoder.Clone(),
		fields:  s.fields,
	}
}

func (s skipFieldZapEncoder) AddString(key, val string) {
	skip := false
	for _, skipField := range s.fields {
		if key == skipField {
			skip = true
			break
		}
	}
	if !skip {
		s.Encoder.AddString(key, val)
	}
}

// isExternalStorageEnabled determines if external storage can be enabled.
// we can't always enable `external_table_storage` if the project dir already has a db file
// it could have been created with older logic where every source was a table in the main db
func isExternalStorageEnabled(dbPath string, variables map[string]string) (bool, error) {
	_, err := os.Stat(filepath.Join(dbPath, DefaultOLAPDSN))
	if err != nil {
		// fresh project
		// check if flag explicitly passed
		val, ok := variables["connector.duckdb.external_table_storage"]
		if !ok {
			// mark enabled by default
			return true, nil
		}
		return strconv.ParseBool(val)
	}

	fsRoot := os.DirFS(dbPath)
	glob := path.Clean(path.Join("./", filepath.Join("*", "version.txt")))

	matches, err := doublestar.Glob(fsRoot, glob)
	if err != nil {
		return false, err
	}
	return len(matches) > 0, nil
}
