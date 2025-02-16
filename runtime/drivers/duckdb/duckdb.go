package duckdb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/duckdb/extensions"
	"github.com/rilldata/rill/runtime/drivers/file"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/priorityqueue"
	"github.com/rilldata/rill/runtime/pkg/rduckdb"
	"github.com/rilldata/rill/runtime/storage"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"gocloud.dev/blob"
	"golang.org/x/sync/semaphore"
)

func init() {
	drivers.Register("duckdb", Driver{name: "duckdb"})
	drivers.Register("motherduck", Driver{name: "motherduck"})
	drivers.RegisterAsConnector("duckdb", Driver{name: "duckdb"})
	drivers.RegisterAsConnector("motherduck", Driver{name: "motherduck"})
}

var spec = drivers.Spec{
	DisplayName: "DuckDB",
	Description: "DuckDB SQL connector.",
	DocsURL:     "https://docs.rilldata.com/reference/connectors/motherduck",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "path",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Path",
			Description: "Path to external DuckDB database.",
			Placeholder: "/path/to/main.db",
		},
	},
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "db",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "DB",
			Description: "Path to DuckDB database",
			Placeholder: "/path/to/duckdb.db",
		},
		{
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
			Description: "Query to extract data from DuckDB.",
			Placeholder: "select * from table;",
		},
		{
			Key:         "name",
			Type:        drivers.StringPropertyType,
			DisplayName: "Source name",
			Description: "The name of the source",
			Placeholder: "my_new_source",
			Required:    true,
		},
	},
	ImplementsCatalog: true,
	ImplementsOLAP:    true,
}

var motherduckSpec = drivers.Spec{
	DisplayName: "MotherDuck",
	Description: "MotherDuck SQL connector.",
	DocsURL:     "https://docs.rilldata.com/reference/connectors/motherduck",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "token",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
	},
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "MotherDuck Connection String",
			Placeholder: "md:motherduck.db",
		},
		{
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
			Description: "Query to extract data from MotherDuck.",
			Placeholder: "select * from table;",
		},
		{
			Key:         "token",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Access token",
			Description: "MotherDuck access token",
			Placeholder: "your.access_token.here",
			Secret:      true,
		},
		{
			Key:         "name",
			Type:        drivers.StringPropertyType,
			DisplayName: "Source name",
			Description: "The name of the source",
			Placeholder: "my_new_source",
			Required:    true,
		},
	},
}

type Driver struct {
	name string
}

func (d Driver) Open(instanceID string, cfgMap map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("duckdb driver can't be shared")
	}

	err := extensions.InstallExtensionsOnce()
	if err != nil {
		logger.Warn("failed to install embedded DuckDB extensions, let DuckDB download them", zap.Error(err))
	}

	cfg, err := newConfig(cfgMap)
	if err != nil {
		return nil, err
	}

	// See note in connection struct
	olapSemSize := cfg.PoolSize - 1
	if olapSemSize < 1 {
		olapSemSize = 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &connection{
		instanceID:     instanceID,
		config:         cfg,
		logger:         logger,
		activity:       ac,
		storage:        st,
		metaSem:        semaphore.NewWeighted(1),
		olapSem:        priorityqueue.NewSemaphore(olapSemSize),
		longRunningSem: semaphore.NewWeighted(1), // Currently hard-coded to 1
		dbCond:         sync.NewCond(&sync.Mutex{}),
		driverConfig:   cfgMap,
		driverName:     d.name,
		connTimes:      make(map[int]time.Time),
		ctx:            ctx,
		cancel:         cancel,
	}
	remote, ok, err := st.OpenBucket(context.Background())
	if err != nil {
		return nil, err
	}
	if ok {
		c.remote = remote
	}

	// register a callback to add a gauge on number of connections in use per db
	attrs := []attribute.KeyValue{attribute.String("instance_id", instanceID)}
	c.registration = observability.Must(meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		observer.ObserveInt64(connectionsInUse, int64(c.dbConnCount), metric.WithAttributes(attrs...))
		return nil
	}, connectionsInUse))

	// Open the DB
	err = c.reopenDB(context.Background())
	if err != nil {
		if remote != nil {
			_ = remote.Close()
		}
		// Check for another process currently accessing the DB
		if strings.Contains(err.Error(), "Could not set lock on file") {
			return nil, fmt.Errorf("failed to open database (is Rill already running?): %w", err)
		}
		return nil, err
	}

	go c.periodicallyEmitStats(time.Minute)

	go c.periodicallyCheckConnDurations(time.Minute)

	return c, nil
}

func (d Driver) Spec() drivers.Spec {
	if d.name == "motherduck" {
		return motherduckSpec
	}
	return spec
}

func (d Driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

func (d Driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	// The "sql" property of a DuckDB source can reference other connectors like S3.
	// We try to extract those and return them here.
	// We will in most error cases just return nil and let errors be handled during source ingestion.

	sql, ok := src["sql"].(string)
	if !ok {
		return nil, nil
	}

	ast, err := duckdbsql.Parse(sql)
	if err != nil {
		return nil, nil
	}

	res := make([]string, 0)

	refs := ast.GetTableRefs()
	for _, ref := range refs {
		if len(ref.Paths) == 0 {
			continue
		}

		uri, err := url.Parse(ref.Paths[0])
		if err != nil {
			return nil, err
		}

		switch uri.Scheme {
		case "s3", "azure":
			res = append(res, uri.Scheme)
		case "gs":
			res = append(res, "gcs")
		default:
			// Ignore
		}
	}

	return res, nil
}

type connection struct {
	instanceID string
	// do not use directly it can also be nil or closed
	// use acquireOLAPConn/acquireMetaConn for select and acquireDB for write queries
	db rduckdb.DB
	// driverConfig is input config passed during Open
	driverConfig map[string]any
	driverName   string
	// config is parsed configs
	config   *config
	logger   *zap.Logger
	activity *activity.Client
	storage  *storage.Client
	remote   *blob.Bucket
	// This driver may issue both OLAP and "meta" queries (like catalog info) against DuckDB.
	// Meta queries are usually fast, but OLAP queries may take a long time. To enable predictable parallel performance,
	// we gate queries with semaphores that limits the number of concurrent queries of each type.
	// The metaSem allows 1 query at a time and the olapSem allows cfg.PoolSize-1 queries at a time.
	// When cfg.PoolSize is 1, we set olapSem to still allow 1 query at a time.
	// This creates contention for the same connection in database/sql's pool, but its locks will handle that.
	metaSem *semaphore.Weighted
	olapSem *priorityqueue.Semaphore
	// The OLAP interface additionally provides an option to limit the number of long-running queries, as designated by the caller.
	// longRunningSem enforces this limitation.
	longRunningSem *semaphore.Weighted
	// If DuckDB encounters a fatal error, all queries will fail until the DB has been reopened.
	// When dbReopen is set to true, dbCond will be used to stop acquisition of new connections,
	// and then when dbConnCount becomes 0, the DB will be reopened and dbReopen set to false again.
	// If the reopen fails, dbErr will be set and all subsequent connection acquires will return it.
	dbConnCount int
	dbCond      *sync.Cond
	dbReopen    bool
	dbErr       error
	// State for maintaining connection acquire times, which enables periodically checking for hanging DuckDB queries (we have previously seen deadlocks in DuckDB).
	connTimesMu    sync.Mutex
	nextConnID     int
	connTimes      map[int]time.Time
	hangingConnErr error
	// Cancellable context to control internal processes like emitting the stats
	ctx    context.Context
	cancel context.CancelFunc
	// registration should be unregistered on close
	registration metric.Registration
}

var _ drivers.OLAPStore = &connection{}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	conn, rel, err := c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	err = conn.PingContext(ctx)
	_ = rel()
	c.connTimesMu.Lock()
	defer c.connTimesMu.Unlock()
	return errors.Join(err, c.hangingConnErr)
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return c.driverName
}

// Config used to open the Connection
func (c *connection) Config() map[string]any {
	return c.driverConfig
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	c.cancel()
	_ = c.registration.Unregister()
	if c.remote != nil {
		_ = c.remote.Close()
	}
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// AsRegistry Registry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore Catalog implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return c, true
}

// AsRepoStore Repo implements drivers.Connection.
func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsOLAP OLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	if opts.InputHandle == c && opts.OutputHandle == c {
		return &selfToSelfExecutor{c}, true
	}
	if opts.OutputHandle == c {
		if w, ok := opts.InputHandle.AsWarehouse(); ok {
			return &warehouseToSelfExecutor{c, w}, true
		}
		if f, ok := opts.InputHandle.AsFileStore(); ok && opts.InputConnector == "local_file" {
			return &localFileToSelfExecutor{c, f}, true
		}
		switch opts.InputHandle.Driver() {
		case "mysql", "postgres":
			return &sqlStoreToSelfExecutor{c}, true
		case "https":
			return &httpsToSelfExecutor{c}, true
		case "motherduck":
			return &mdToSelfExecutor{c}, true
		}
		if _, ok := opts.InputHandle.AsObjectStore(); ok {
			return &objectStoreToSelfExecutor{c}, true
		}
	}
	if opts.InputHandle == c {
		if opts.OutputHandle.Driver() == "file" {
			outputProps := &file.ModelOutputProperties{}
			if err := mapstructure.WeakDecode(opts.PreliminaryOutputProperties, outputProps); err != nil {
				return nil, false
			}
			if supportsExportFormat(outputProps.Format) {
				return &selfToFileExecutor{c}, true
			}
		}
	}
	return nil, false
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return c, true
}

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	olap, _ := to.(*connection)
	if c == to {
		if from == to {
			return newDuckDBToDuckDB(from, c, c.logger), true
		}
		switch from.Driver() {
		case "motherduck":
			return newMotherduckToDuckDB(from, c, c.logger), true
		case "postgres":
			return newDuckDBToDuckDB(from, c, c.logger), true
		case "mysql":
			return newDuckDBToDuckDB(from, c, c.logger), true
		}
		if store, ok := from.AsWarehouse(); ok {
			return NewWarehouseToDuckDB(store, olap, c.logger), true
		}
		if store, ok := from.AsObjectStore(); ok { // objectstore to duckdb transfer
			return NewObjectStoreToDuckDB(store, olap, c.logger), true
		}
		if store, ok := from.AsFileStore(); ok {
			return NewFileStoreToDuckDB(store, olap, c.logger), true
		}
	}
	return nil, false
}

func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// reopenDB opens the DuckDB handle anew. If c.db is already set, it closes the existing handle first.
func (c *connection) reopenDB(ctx context.Context) error {
	// If c.db is already open, close it first
	if c.db != nil {
		err := c.db.Close()
		if err != nil {
			return err
		}
		c.db = nil
	}

	// Queries to run when a new DuckDB connection is opened.
	var bootQueries []string

	// Add custom boot queries before any other (e.g. to override the extensions repository)
	if c.config.BootQueries != "" {
		bootQueries = append(bootQueries, c.config.BootQueries)
	}

	// Add required boot queries
	bootQueries = append(bootQueries,
		"INSTALL 'json'",
		"LOAD 'json'",
		"INSTALL 'icu'",
		"LOAD 'icu'",
		"INSTALL 'parquet'",
		"LOAD 'parquet'",
		"INSTALL 'httpfs'",
		"LOAD 'httpfs'",
		"INSTALL 'sqlite'",
		"LOAD 'sqlite'",
		"SET max_expression_depth TO 250",
		"SET timezone='UTC'",
		"SET old_implicit_casting = true",        // Implicit Cast to VARCHAR
		"SET allow_community_extensions = false", // This locks the configuration, so it can't later be enabled.
	)

	dataDir, err := c.storage.DataDir()
	if err != nil {
		return err
	}

	// We want to set preserve_insertion_order=false in hosted environments only (where source data is never viewed directly). Setting it reduces batch data ingestion time by ~40%.
	// Hack: Using AllowHostAccess as a proxy indicator for a hosted environment.
	if !c.config.AllowHostAccess {
		bootQueries = append(bootQueries, "SET preserve_insertion_order TO false")
	}

	// Add init SQL if provided
	if c.config.InitSQL != "" {
		bootQueries = append(bootQueries, c.config.InitSQL)
	}

	// Create new DB
	logger := slog.New(zapslog.NewHandler(c.logger.Core(), &zapslog.HandlerOptions{
		AddSource: true,
	}))
	c.db, err = rduckdb.NewDB(ctx, &rduckdb.DBOptions{
		LocalPath:      dataDir,
		Remote:         c.remote,
		CPU:            c.config.CPU,
		MemoryLimitGB:  c.config.MemoryLimitGB,
		ReadWriteRatio: c.config.ReadWriteRatio,
		ReadSettings:   c.config.readSettings(),
		WriteSettings:  c.config.writeSettings(),
		InitQueries:    bootQueries,
		Logger:         logger,
		OtelAttributes: []attribute.KeyValue{attribute.String("instance_id", c.instanceID)},
	})
	return err
}

// acquireMetaConn gets a connection from the pool for "meta" queries like catalog and information schema (i.e. fast queries).
// It returns a function that puts the connection back in the pool (if applicable).
func (c *connection) acquireMetaConn(ctx context.Context) (*sqlx.Conn, func() error, error) {
	// Try to get conn from context (means the call is wrapped in WithConnection)
	conn := connFromContext(ctx)
	if conn != nil {
		return conn, func() error { return nil }, nil
	}

	// Acquire semaphore
	err := c.metaSem.Acquire(ctx, 1)
	if err != nil {
		return nil, nil, err
	}

	// Get new conn
	conn, releaseConn, err := c.acquireReadConnection(ctx)
	if err != nil {
		c.metaSem.Release(1)
		return nil, nil, err
	}

	// Build release func
	release := func() error {
		err := releaseConn()
		c.metaSem.Release(1)
		return err
	}

	return conn, release, nil
}

// acquireOLAPConn gets a connection from the pool for OLAP queries (i.e. slow queries).
// It returns a function that puts the connection back in the pool (if applicable).
func (c *connection) acquireOLAPConn(ctx context.Context, priority int, longRunning bool) (*sqlx.Conn, func() error, error) {
	// Try to get conn from context (means the call is wrapped in WithConnection)
	conn := connFromContext(ctx)
	if conn != nil {
		return conn, func() error { return nil }, nil
	}

	// Acquire long-running semaphore if applicable
	if longRunning {
		err := c.longRunningSem.Acquire(ctx, 1)
		if err != nil {
			return nil, nil, err
		}
	}

	// Acquire semaphore
	err := c.olapSem.Acquire(ctx, priority)
	if err != nil {
		if longRunning {
			c.longRunningSem.Release(1)
		}
		return nil, nil, err
	}

	// Get new conn
	conn, releaseConn, err := c.acquireReadConnection(ctx)
	if err != nil {
		c.olapSem.Release()
		if longRunning {
			c.longRunningSem.Release(1)
		}
		return nil, nil, err
	}

	// Build release func
	release := func() error {
		err := releaseConn()
		c.olapSem.Release()
		if longRunning {
			c.longRunningSem.Release(1)
		}
		return err
	}

	return conn, release, nil
}

// acquireReadConnection is a helper function to acquire a read connection from rduckdb.
// Do not use this function directly for OLAP queries. Use acquireOLAPConn, acquireMetaConn instead.
func (c *connection) acquireReadConnection(ctx context.Context) (*sqlx.Conn, func() error, error) {
	db, releaseDB, err := c.acquireDB()
	if err != nil {
		return nil, nil, err
	}

	conn, releaseConn, err := db.AcquireReadConnection(ctx)
	if err != nil {
		_ = releaseDB()
		return nil, nil, err
	}

	release := func() error {
		err := releaseConn()
		return errors.Join(err, releaseDB())
	}
	return conn, release, nil
}

// acquireDB returns rduckDB handle.
// acquireDB implements the connection tracking and DB reopening logic described in the struct definition for connection.
// It should not be used directly for select queries. For select queries use acquireOLAPConn and acquireMetaConn.
// It should only be used for write queries.
func (c *connection) acquireDB() (rduckdb.DB, func() error, error) {
	c.dbCond.L.Lock()
	for {
		if c.dbErr != nil {
			c.dbCond.L.Unlock()
			return nil, nil, c.dbErr
		}
		if !c.dbReopen {
			break
		}
		c.dbCond.Wait()
	}

	c.dbConnCount++
	c.dbCond.L.Unlock()

	c.connTimesMu.Lock()
	connID := c.nextConnID
	c.nextConnID++
	c.connTimes[connID] = time.Now()
	c.connTimesMu.Unlock()

	release := func() error {
		c.connTimesMu.Lock()
		delete(c.connTimes, connID)
		c.connTimesMu.Unlock()
		c.dbCond.L.Lock()
		c.dbConnCount--
		if c.dbConnCount == 0 && c.dbReopen {
			c.triggerReopen()
		}
		c.dbCond.L.Unlock()
		return nil
	}
	return c.db, release, nil
}

func (c *connection) triggerReopen() {
	go func() {
		c.dbCond.L.Lock()
		defer c.dbCond.L.Unlock()
		if !c.dbReopen || c.dbConnCount == 0 {
			c.logger.Error("triggerReopen called but should not reopen", zap.Bool("dbReopen", c.dbReopen), zap.Int("dbConnCount", c.dbConnCount))
			return
		}
		c.dbReopen = false
		err := c.reopenDB(c.ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				c.logger.Error("reopen of DuckDB failed - the handle is now permanently locked", zap.Error(err))
			}
		}
		c.dbErr = err
		c.dbCond.Broadcast()
	}()
}

// checkErr marks the DB for reopening if the error is an internal DuckDB error.
// In all other cases, it just proxies the err.
// It should be wrapped around errors returned from DuckDB queries. **It must be called while still holding an acquired DuckDB connection.**
func (c *connection) checkErr(err error) error {
	if err != nil {
		if strings.HasPrefix(err.Error(), "INTERNAL Error:") || strings.HasPrefix(err.Error(), "FATAL Error") {
			c.dbCond.L.Lock()
			defer c.dbCond.L.Unlock()
			c.dbReopen = true
			c.logger.Error("encountered internal DuckDB error - scheduling reopen of DuckDB", zap.Error(err))
		}
	}
	return err
}

// Periodically collects stats using pragma_database_size() and emits as activity events
// nolint
func (c *connection) periodicallyEmitStats(d time.Duration) {
	if c.activity == nil {
		// Activity client isn't set, there is no need to report stats
		return
	}

	statTicker := time.NewTicker(d)
	for {
		select {
		case <-statTicker.C:
			estimatedDBSize := c.estimateSize()
			c.activity.RecordMetric(c.ctx, "duckdb_estimated_size_bytes", float64(estimatedDBSize))
		case <-c.ctx.Done():
			statTicker.Stop()
			return
		}
	}
}

// maxAcquiredConnDuration is the maximum duration a connection can be held for before we consider it potentially hanging/deadlocked.
const maxAcquiredConnDuration = 3 * time.Hour

// periodicallyCheckConnDurations periodically checks the durations of all acquired connections and logs a warning if any have been held for longer than maxAcquiredConnDuration.
func (c *connection) periodicallyCheckConnDurations(d time.Duration) {
	connDurationTicker := time.NewTicker(d)
	defer connDurationTicker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-connDurationTicker.C:
			c.connTimesMu.Lock()
			var connErr error
			for connID, connTime := range c.connTimes {
				if time.Since(connTime) > maxAcquiredConnDuration {
					connErr = fmt.Errorf("duckdb: a connection has been held for longer than the maximum allowed duration")
					c.logger.Error("duckdb: a connection has been held for longer than the maximum allowed duration", zap.Int("conn_id", connID), zap.Duration("duration", time.Since(connTime)))
				}
			}
			c.hangingConnErr = connErr
			c.connTimesMu.Unlock()
		}
	}
}
