package rduckdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"
	"go.opentelemetry.io/otel/attribute"
	"gocloud.dev/blob"
	"golang.org/x/sync/semaphore"
)

var errNotFound = errors.New("rduckdb: not found")

type DB interface {
	// Close closes the database.
	Close() error

	// AcquireReadConnection returns a connection to the database for reading.
	// Once done the connection should be released by calling the release function.
	// This connection must only be used for select queries or for creating and working with temporary tables.
	AcquireReadConnection(ctx context.Context) (conn *sqlx.Conn, release func() error, err error)

	// Size returns the size of the database in bytes.
	// It is currently implemented as sum of the size of all serving `.db` files.
	Size() int64

	// CRUD APIs

	// CreateTableAsSelect creates a new table by name from the results of the given SQL query.
	CreateTableAsSelect(ctx context.Context, name string, sql string, opts *CreateTableOptions) error

	// MutateTable allows mutating a table in the database by calling the mutateFn.
	MutateTable(ctx context.Context, name string, mutateFn func(ctx context.Context, conn *sqlx.Conn) error) error

	// DropTable removes a table from the database.
	DropTable(ctx context.Context, name string) error

	// RenameTable renames a table in the database.
	RenameTable(ctx context.Context, oldName, newName string) error

	// Meta APIs

	// Schema returns the schema of the database.
	Schema(ctx context.Context, like string, matchCase bool) ([]*Table, error)
}

type DBOptions struct {
	// LocalPath is the path where local db files will be stored. Should be unique for each database.
	LocalPath string
	// Remote is the blob storage bucket where the database files will be stored. This is the source of truth.
	// The local db will be eventually synced with the remote.
	Remote *blob.Bucket
	// CPU cores available for the DB. If no ratio is set then this is split evenly between read and write.
	CPU int `mapstructure:"cpu"`
	// MemoryLimitGB is the amount of memory available for the DB. If no ratio is set then this is split evenly between read and write.
	MemoryLimitGB int `mapstructure:"memory_limit_gb"`
	// ReadWriteRatio is the ratio of resources to allocate to the read DB. If set, CPU and MemoryLimitGB are distributed based on this ratio.
	ReadWriteRatio float64 `mapstructure:"read_write_ratio"`
	// ReadSettings are settings applied the read duckDB handle.
	ReadSettings map[string]string
	// WriteSettings are settings applied the write duckDB handle.
	WriteSettings map[string]string
	// InitQueries are the queries to run when the database is first created.
	InitQueries []string

	Logger         *slog.Logger
	OtelAttributes []attribute.KeyValue
}

func (d *DBOptions) ValidateSettings() error {
	if d.ReadWriteRatio < 0 || d.ReadWriteRatio > 1 {
		return fmt.Errorf("read_write_ratio should be between 0 and 1")
	}
	if d.ReadSettings == nil {
		d.ReadSettings = make(map[string]string)
	}
	if d.WriteSettings == nil {
		d.WriteSettings = make(map[string]string)
	}
	memoryLimitBytes := int64(d.MemoryLimitGB * 1000 * 1000 * 1000)
	if memoryLimitBytes == 0 {
		db, err := sql.Open("duckdb", "")
		if err != nil {
			return err
		}
		defer db.Close()

		row := db.QueryRow("SELECT value FROM duckdb_settings() WHERE name = 'max_memory'")
		var maxMemory string
		err = row.Scan(&maxMemory)
		if err != nil {
			return fmt.Errorf("unable to get max_memory: %w", err)
		}

		bytes, err := humanReadableSizeToBytes(maxMemory)
		if err != nil {
			return fmt.Errorf("unable to parse max_memory: %w", err)
		}

		memoryLimitBytes = int64(bytes)
	}

	threads := d.CPU
	if threads == 0 {
		db, err := sql.Open("duckdb", "")
		if err != nil {
			return err
		}
		defer db.Close()

		row := db.QueryRow("SELECT value FROM duckdb_settings() WHERE name = 'threads'")
		err = row.Scan(&threads)
		if err != nil {
			return fmt.Errorf("unable to get threads: %w", err)
		}
	}

	d.ReadSettings["memory_limit"] = fmt.Sprintf("%d bytes", int64(float64(memoryLimitBytes)*d.ReadWriteRatio))
	d.WriteSettings["memory_limit"] = fmt.Sprintf("%d bytes", int64(float64(memoryLimitBytes)*(1-d.ReadWriteRatio)))

	readThreads := math.Floor(float64(threads) * d.ReadWriteRatio)
	if readThreads <= 1 {
		d.ReadSettings["threads"] = "1"
	} else {
		d.ReadSettings["threads"] = strconv.Itoa(int(readThreads))
	}
	writeThreads := threads - int(readThreads)
	if writeThreads <= 1 {
		d.WriteSettings["threads"] = "1"
	} else {
		d.WriteSettings["threads"] = strconv.Itoa(writeThreads)
	}
	return nil
}

type CreateTableOptions struct {
	// View specifies whether the created table is a view.
	View bool
	// If BeforeCreateFn is set, it will be executed before the create query is executed.
	BeforeCreateFn func(ctx context.Context, conn *sqlx.Conn) error
	// If AfterCreateFn is set, it will be executed after the create query is executed.
	// This will execute even if the create query fails.
	AfterCreateFn func(ctx context.Context, conn *sqlx.Conn) error
}

// NewDB creates a new DB instance.
// dbIdentifier is a unique identifier for the database reported in metrics.
func NewDB(ctx context.Context, opts *DBOptions) (DB, error) {
	err := opts.ValidateSettings()
	if err != nil {
		return nil, err
	}

	bgctx, cancel := context.WithCancel(context.Background())
	db := &db{
		opts:       opts,
		localPath:  opts.LocalPath,
		remote:     opts.Remote,
		writeSem:   semaphore.NewWeighted(1),
		metaSem:    semaphore.NewWeighted(1),
		localDirty: true,
		logger:     opts.Logger,
		ctx:        bgctx,
		cancel:     cancel,
	}
	// create local path
	err = os.MkdirAll(db.localPath, fs.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("unable to create local path: %w", err)
	}

	// migrate db from old storage structure to new
	err = db.migrateDB()
	if err != nil && !errors.Is(err, context.Canceled) {
		// do not return error just truncate the directory and start fresh
		db.logger.Error("failed to migrate db", slog.String("error", err.Error()))
		err = os.RemoveAll(db.localPath)
		if err != nil {
			return nil, err
		}
		err = os.MkdirAll(db.localPath, fs.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	// By adding a _duckdb_on_gcs_.txt we can check if the db files are synced to cloud storage.
	// If db files are present on cloud storage, the source of truth is cloud storage else local storage.
	// The file _duckdb_on_gcs_.txt with true/false as content
	// This is a temporary solution and will be removed in future when we enable cloud storage completely
	duckdbONGCS, _ := db.duckdbOnGCS()
	if !duckdbONGCS && db.remote != nil {
		// switched on remote storage
		// push local data to remote
		err := db.iterateLocalTables(false, func(name string, meta *tableMeta) error {
			return db.pushToRemote(ctx, name, nil, meta)
		})
		if err != nil {
			return nil, fmt.Errorf("unable to write local data to remote: %w", err)
		}
	}
	err = os.WriteFile(filepath.Join(db.localPath, "_duckdb_on_gcs_.txt"), []byte(strconv.FormatBool(db.remote != nil)), fs.ModePerm)
	if err != nil {
		return nil, err
	}

	// sync local data
	err = db.pullFromRemote(ctx, false)
	if err != nil {
		return nil, err
	}

	// collect all tables
	var tables []*tableMeta
	_ = db.iterateLocalTables(false, func(name string, meta *tableMeta) error {
		tables = append(tables, meta)
		return nil
	})

	// catalog
	db.catalog = newCatalog(
		func(name, version string) {
			go func() {
				err := db.removeTableVersion(bgctx, name, version)
				if err != nil && !errors.Is(err, context.Canceled) {
					db.logger.Error("error in removing table version", slog.String("name", name), slog.String("version", version), slog.String("error", err.Error()))
				}
			}()
		},
		func(i int) {
			go func() {
				err := db.removeSnapshot(bgctx, i)
				if err != nil && !errors.Is(err, context.Canceled) {
					db.logger.Error("error in removing snapshot", slog.Int("id", i), slog.String("error", err.Error()))
				}
			}()
		},
		tables,
		opts.Logger,
	)

	db.dbHandle, err = db.openDBAndAttach(ctx, filepath.Join(db.localPath, "main.db"), "", true)
	if err != nil {
		if strings.Contains(err.Error(), "Symbol not found") {
			fmt.Printf("Your version of macOS is not supported. Please upgrade to the latest major release of macOS. See this link for details: https://support.apple.com/en-in/macos/upgrade")
			os.Exit(1)
		}
		return nil, err
	}

	go db.localDBMonitor()
	return db, nil
}

type db struct {
	opts *DBOptions

	localPath string
	remote    *blob.Bucket

	// dbHandle serves executes meta queries and serves read queries
	dbHandle *sqlx.DB
	// writeSem ensures only one write operation is allowed at a time
	writeSem *semaphore.Weighted
	// metaSem enures only one meta operation can run on a duckb handle.
	// Meta operations are attach, detach, create view queries done on the db handle
	metaSem *semaphore.Weighted
	// localDirty is set to true when a change is committed to the remote but not yet reflected in the local db
	localDirty bool
	catalog    *catalog

	logger *slog.Logger

	// ctx and cancel to cancel background operations
	ctx    context.Context
	cancel context.CancelFunc
}

var _ DB = &db{}

func (d *db) Close() error {
	// close background operations
	d.cancel()
	return d.dbHandle.Close()
}

func (d *db) AcquireReadConnection(ctx context.Context) (*sqlx.Conn, func() error, error) {
	snapshot := d.catalog.acquireSnapshot()

	conn, err := d.dbHandle.Connx(ctx)
	if err != nil {
		d.catalog.releaseSnapshot(snapshot)
		return nil, nil, err
	}

	err = d.prepareSnapshot(ctx, conn, snapshot)
	if err != nil {
		d.catalog.releaseSnapshot(snapshot)
		_ = conn.Close()
		return nil, nil, err
	}

	release := func() error {
		d.catalog.releaseSnapshot(snapshot)
		return conn.Close()
	}
	return conn, release, nil
}

func (d *db) CreateTableAsSelect(ctx context.Context, name, query string, opts *CreateTableOptions) (createErr error) {
	d.logger.Debug("create: create table", slog.String("name", name), slog.Bool("view", opts.View))
	err := d.writeSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.writeSem.Release(1)

	// pull latest changes from remote
	err = d.pullFromRemote(ctx, true)
	if err != nil {
		return err
	}

	// check if some older version exists
	oldMeta, _ := d.catalog.tableMeta(name)
	if oldMeta != nil {
		d.logger.Debug("create: old version", slog.String("table", name), slog.String("version", oldMeta.Version))
	}

	// create new version directory
	newVersion := newVersion()
	newMeta := &tableMeta{
		Name:           name,
		Version:        newVersion,
		CreatedVersion: newVersion,
	}
	var dsn string
	if opts.View {
		dsn = ""
		newMeta.SQL = query
		err = d.initLocalTable(name, "")
		if err != nil {
			return fmt.Errorf("create: unable to create dir %q: %w", name, err)
		}
	} else {
		err = d.initLocalTable(name, newVersion)
		if err != nil {
			return fmt.Errorf("create: unable to create dir %q: %w", name, err)
		}
		dsn = d.localDBPath(name, newVersion)
		defer func() {
			if createErr != nil {
				_ = d.deleteLocalTableFiles(name, newVersion)
			}
		}()
	}

	// need to attach existing table so that any views dependent on this table are correctly attached
	conn, release, err := d.acquireWriteConn(ctx, dsn, name, true)
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()

	safeName := safeSQLName(name)
	var typ string
	if opts.View {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}
	newMeta.Type = typ
	if opts.BeforeCreateFn != nil {
		err = opts.BeforeCreateFn(ctx, conn)
		if err != nil {
			return fmt.Errorf("create: BeforeCreateFn returned error: %w", err)
		}
	}
	execAfterCreate := func() error {
		if opts.AfterCreateFn == nil {
			return nil
		}
		err = opts.AfterCreateFn(ctx, conn)
		if err != nil {
			return fmt.Errorf("create: AfterCreateFn returned error: %w", err)
		}
		return nil
	}
	// ingest data
	_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE %s %s AS (%s\n)", typ, safeName, query), nil)
	if err != nil {
		return errors.Join(fmt.Errorf("create: create %s %q failed: %w", typ, name, err), execAfterCreate())
	}
	err = execAfterCreate()
	if err != nil {
		return err
	}

	// close write handle before syncing local so that temp files or wal files are removed
	err = release()
	if err != nil {
		return err
	}

	// update remote data and metadata
	if err := d.pushToRemote(ctx, name, oldMeta, newMeta); err != nil {
		return fmt.Errorf("create: replicate failed: %w", err)
	}
	d.logger.Debug("create: remote table updated", slog.String("name", name))
	// no errors after this point since background goroutine will eventually sync the local db

	// update local metadata
	err = d.writeTableMeta(name, newMeta)
	if err != nil {
		d.logger.Debug("create: error in writing table meta", slog.String("name", name), slog.String("error", err.Error()))
		return nil
	}

	d.catalog.addTableVersion(name, newMeta, true)
	d.localDirty = false
	return nil
}

func (d *db) MutateTable(ctx context.Context, name string, mutateFn func(ctx context.Context, conn *sqlx.Conn) error) error {
	d.logger.Debug("mutate table", slog.String("name", name))
	err := d.writeSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.writeSem.Release(1)

	// pull latest changes from remote
	err = d.pullFromRemote(ctx, true)
	if err != nil {
		return err
	}

	oldMeta, err := d.catalog.tableMeta(name)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return fmt.Errorf("mutate: Table %q not found", name)
		}
		return fmt.Errorf("mutate: unable to get table meta: %w", err)
	}

	// create new version directory
	newVersion := newVersion()
	newDir := d.localTableDir(name, newVersion)
	err = copyDir(newDir, d.localTableDir(name, oldMeta.Version))
	if err != nil {
		_ = os.RemoveAll(newDir)
		return fmt.Errorf("mutate: copy table failed: %w", err)
	}

	// acquire write connection
	// need to ignore attaching table since it is already present in the db file
	conn, release, err := d.acquireWriteConn(ctx, d.localDBPath(name, newVersion), name, false)
	if err != nil {
		_ = os.RemoveAll(newDir)
		return err
	}

	err = mutateFn(ctx, conn)
	if err != nil {
		_ = os.RemoveAll(newDir)
		_ = release()
		return fmt.Errorf("mutate: mutate failed: %w", err)
	}

	// push to remote
	err = release()
	if err != nil {
		_ = os.RemoveAll(newDir)
		return fmt.Errorf("mutate: failed to close connection: %w", err)
	}
	meta := &tableMeta{
		Name:           name,
		Version:        newVersion,
		CreatedVersion: oldMeta.CreatedVersion,
		Type:           oldMeta.Type,
		SQL:            oldMeta.SQL,
	}
	err = d.pushToRemote(ctx, name, oldMeta, meta)
	if err != nil {
		_ = os.RemoveAll(newDir)
		return fmt.Errorf("mutate: replicate failed: %w", err)
	}
	// no errors after this point since background goroutine will eventually sync the local db

	// update local meta
	err = d.writeTableMeta(name, meta)
	if err != nil {
		d.logger.Debug("mutate: error in writing table meta", slog.String("name", name), slog.String("error", err.Error()))
		return nil
	}

	d.catalog.addTableVersion(name, meta, true)
	d.localDirty = false
	return nil
}

// DropTable implements DB.
func (d *db) DropTable(ctx context.Context, name string) error {
	d.logger.Debug("drop table", slog.String("name", name))
	err := d.writeSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.writeSem.Release(1)

	// pull latest changes from remote
	err = d.pullFromRemote(ctx, true)
	if err != nil {
		return fmt.Errorf("drop: unable to pull from remote: %w", err)
	}

	// check if table exists
	_, err = d.catalog.tableMeta(name)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return fmt.Errorf("drop: Table %q not found", name)
		}
		return fmt.Errorf("drop: unable to get table meta: %w", err)
	}

	// drop the table from remote
	d.localDirty = true
	err = d.deleteRemote(ctx, name, "")
	if err != nil {
		return fmt.Errorf("drop: unable to drop table %q from remote: %w", name, err)
	}
	// no errors after this point since background goroutine will eventually sync the local db

	d.catalog.removeTable(name)
	d.localDirty = false
	return nil
}

func (d *db) RenameTable(ctx context.Context, oldName, newName string) error {
	d.logger.Debug("rename table", slog.String("from", oldName), slog.String("to", newName))
	if strings.EqualFold(oldName, newName) {
		return fmt.Errorf("rename: Table with name %q already exists", newName)
	}
	err := d.writeSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.writeSem.Release(1)

	// pull latest changes from remote
	err = d.pullFromRemote(ctx, true)
	if err != nil {
		return fmt.Errorf("rename: unable to pull from remote: %w", err)
	}

	oldMeta, err := d.catalog.tableMeta(oldName)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return fmt.Errorf("rename: Table %q not found", oldName)
		}
		return fmt.Errorf("rename: unable to get table meta: %w", err)
	}

	newTableOldMeta, err := d.catalog.tableMeta(newName)
	if err != nil && !errors.Is(err, errNotFound) {
		return fmt.Errorf("rename: unable to get table meta for new table: %w", err)
	}

	// copy the old table to new table
	newVersion := newVersion()
	var newDir string
	if oldMeta.Type == "TABLE" {
		newDir = d.localTableDir(newName, newVersion)
		err = copyDir(d.localTableDir(newName, newVersion), d.localTableDir(oldName, oldMeta.Version))
		if err != nil {
			_ = os.RemoveAll(newDir)
			return fmt.Errorf("rename: copy table failed: %w", err)
		}

		// rename the underlying table
		err = renameTable(ctx, d.localDBPath(newName, newVersion), oldName, newName)
		if err != nil {
			_ = os.RemoveAll(newDir)
			return fmt.Errorf("rename: rename table failed: %w", err)
		}
	} else {
		err = copyDir(d.localTableDir(newName, ""), d.localTableDir(oldName, ""))
		if err != nil {
			return fmt.Errorf("rename: copy view failed: %w", err)
		}
	}

	// sync the new table and new version
	meta := &tableMeta{
		Name:           newName,
		Version:        newVersion,
		CreatedVersion: newVersion,
		Type:           oldMeta.Type,
		SQL:            oldMeta.SQL,
	}
	if err := d.pushToRemote(ctx, newName, newTableOldMeta, meta); err != nil {
		if newDir != "" {
			_ = os.RemoveAll(newDir)
		}
		return fmt.Errorf("rename: unable to replicate new table: %w", err)
	}

	// TODO :: fix this
	// at this point db is inconsistent
	// has both old table and new table

	// drop the old table in remote
	err = d.deleteRemote(ctx, oldName, "")
	if err != nil {
		if newDir != "" {
			_ = os.RemoveAll(newDir)
		}
		return fmt.Errorf("rename: unable to delete old table %q from remote: %w", oldName, err)
	}

	// no errors after this point since background goroutine will eventually sync the local db

	// update local meta for new table
	err = d.writeTableMeta(newName, meta)
	if err != nil {
		d.logger.Debug("rename: error in writing table meta", slog.String("name", newName), slog.String("error", err.Error()))
		return nil
	}

	// remove old table from local db
	d.catalog.removeTable(oldName)
	d.catalog.addTableVersion(newName, meta, true)
	d.localDirty = false
	return nil
}

func (d *db) localDBMonitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			// We do not want the localDBMonitor to compete with write operations so we return early if writeSem is not available.
			// Anyways if a write operation is in progress it will sync the local db
			if !d.writeSem.TryAcquire(1) {
				continue
			}
			if !d.localDirty {
				d.writeSem.Release(1)
				// all good
				continue
			}
			err := d.pullFromRemote(d.ctx, true)
			if err != nil && !errors.Is(err, context.Canceled) {
				d.logger.Error("localDBMonitor: error in pulling from remote", slog.String("error", err.Error()))
			}
			d.localDirty = false
			d.writeSem.Release(1)
		}
	}
}

func (d *db) Size() int64 {
	var paths []string
	_ = d.iterateLocalTables(false, func(name string, meta *tableMeta) error {
		// this is to avoid counting temp tables during source ingestion
		// in certain cases we only want to compute the size of the serving db files
		if !strings.HasPrefix(name, "__rill_tmp_") {
			paths = append(paths, d.localDBPath(meta.Name, meta.Version))
		}
		return nil
	})
	return fileSize(paths)
}

// acquireWriteConn syncs the write database, initializes the write handle and returns a write connection.
// The release function should be called to release the connection.
// It should be called with the writeMu locked.
func (d *db) acquireWriteConn(ctx context.Context, dsn, table string, attachExisting bool) (*sqlx.Conn, func() error, error) {
	var ignoreTable string
	if !attachExisting {
		ignoreTable = table
	}
	db, err := d.openDBAndAttach(ctx, dsn, ignoreTable, false)
	if err != nil {
		return nil, nil, err
	}
	conn, err := db.Connx(ctx)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	if attachExisting {
		_, err = conn.ExecContext(ctx, "DROP VIEW IF EXISTS "+safeSQLName(table))
		if err != nil {
			_ = conn.Close()
			_ = db.Close()
			return nil, nil, err
		}
	}

	return conn, func() error {
		_ = conn.Close()
		err = db.Close()
		return err
	}, nil
}

func (d *db) openDBAndAttach(ctx context.Context, uri, ignoreTable string, read bool) (*sqlx.DB, error) {
	d.logger.Debug("open db", slog.Bool("read", read), slog.String("uri", uri))
	// open the db
	var settings map[string]string
	dsn, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if read {
		settings = d.opts.ReadSettings
	} else {
		settings = d.opts.WriteSettings
	}
	query := dsn.Query()
	for k, v := range settings {
		query.Set(k, v)
	}
	// Rebuild DuckDB DSN (which should be "path?key=val&...")
	// this is required since spaces and other special characters are valid in db file path but invalid and hence encoded in URL
	connector, err := duckdb.NewConnector(generateDSN(dsn.Path, query.Encode()), func(execer driver.ExecerContext) error {
		for _, qry := range d.opts.InitQueries {
			_, err := execer.ExecContext(ctx, qry, nil)
			if err != nil && strings.Contains(err.Error(), "Failed to download extension") {
				// Retry using another mirror. Based on: https://github.com/duckdb/duckdb/issues/9378
				_, err = execer.ExecContext(ctx, qry+" FROM 'http://nightly-extensions.duckdb.org'", nil)
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	db := sqlx.NewDb(otelsql.OpenDB(connector), "duckdb")
	err = otelsql.RegisterDBStatsMetrics(db.DB, otelsql.WithAttributes(d.opts.OtelAttributes...))
	if err != nil {
		return nil, fmt.Errorf("registering db stats metrics: %w", err)
	}

	conn, err := db.Connx(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	tables := d.catalog.listTables()
	err = d.attachTables(ctx, conn, tables, ignoreTable)
	if err != nil {
		conn.Close()
		db.Close()
		return nil, err
	}
	if err := conn.Close(); err != nil {
		return nil, err
	}

	// 2023-12-11: Hail mary for solving this issue: https://github.com/duckdblabs/rilldata/issues/6.
	// Forces DuckDB to create catalog entries for the information schema up front (they are normally created lazily).
	// Can be removed if the issue persists.
	_, err = db.ExecContext(context.Background(), `
		select
			coalesce(t.table_catalog, current_database()) as "database",
			t.table_schema as "schema",
			t.table_name as "name",
			t.table_type as "type", 
			array_agg(c.column_name order by c.ordinal_position) as "column_names",
			array_agg(c.data_type order by c.ordinal_position) as "column_types",
			array_agg(c.is_nullable = 'YES' order by c.ordinal_position) as "column_nullable"
		from information_schema.tables t
		join information_schema.columns c on t.table_schema = c.table_schema and t.table_name = c.table_name
		group by 1, 2, 3, 4
		order by 1, 2, 3, 4
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func (d *db) attachTables(ctx context.Context, conn *sqlx.Conn, tables []*tableMeta, ignoreTable string) error {
	// sort tables by created_version
	// this is to ensure that views/tables on which other views depend are attached first
	slices.SortFunc(tables, func(a, b *tableMeta) int {
		// all tables should be attached first and can be attached in any order
		if a.Type == "TABLE" && b.Type == "TABLE" {
			return 0
		}
		if a.Type == "TABLE" {
			return -1
		}
		if b.Type == "TABLE" {
			return 1
		}
		// any order for views
		return strings.Compare(a.CreatedVersion, b.CreatedVersion)
	})

	var failedViews []*tableMeta
	// attach database files
	for _, table := range tables {
		if table.Name == ignoreTable {
			continue
		}
		safeTable := safeSQLName(table.Name)
		if table.Type == "VIEW" {
			_, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE VIEW %s AS %s", safeTable, table.SQL))
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				failedViews = append(failedViews, table)
			}
			continue
		}
		safeDBName := safeSQLName(dbName(table.Name, table.Version))
		_, err := conn.ExecContext(ctx, fmt.Sprintf("ATTACH IF NOT EXISTS %s AS %s (READ_ONLY)", safeSQLString(d.localDBPath(table.Name, table.Version)), safeDBName))
		if err != nil {
			return fmt.Errorf("failed to attach table %q: %w", table.Name, err)
		}
		_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE VIEW %s AS SELECT * FROM %s.%s", safeTable, safeDBName, safeTable))
		if err != nil {
			return err
		}
	}

	// retry creating views
	// views may depend on other views, without building a dependency graph we can not recreate them in correct order
	// so we recreate all failed views and collect the ones that failed
	// once a view is created successfully, it may be possible that other views that depend on it can be created in the next iteration
	// if in a iteration no views are created successfully, it means either all views are invalid or there is a circular dependency
	for len(failedViews) > 0 {
		allViewsFailed := true
		size := len(failedViews)
		for i := 0; i < size; i++ {
			table := failedViews[0]
			failedViews = failedViews[1:]
			safeTable := safeSQLName(table.Name)
			_, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE VIEW %s AS %s", safeTable, table.SQL))
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				failedViews = append(failedViews, table)
				continue
			}
			// successfully created view
			allViewsFailed = false
		}
		if !allViewsFailed {
			continue
		}

		// create views that return error on querying
		// may be the view is incompatible with the underlying data due to schema changes
		for i := 0; i < len(failedViews); i++ {
			table := failedViews[i]
			safeTable := safeSQLName(table.Name)
			// capture the error in creating the view
			_, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE VIEW %s AS %s", safeTable, table.SQL))
			if err == nil {
				// not possible but just to be safe
				continue
			}
			safeErr := strings.Trim(safeSQLString(err.Error()), "'")
			_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE VIEW %s AS SELECT error('View %s is incompatible with the underlying data: %s')", safeTable, safeTable, safeErr))
			if err != nil {
				return err
			}
		}
		break
	}
	return nil
}

func (d *db) tableMeta(name string) (*tableMeta, error) {
	contents, err := os.ReadFile(d.localMetaPath(name))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, errNotFound
		}
		return nil, err
	}
	m := &tableMeta{}
	err = json.Unmarshal(contents, m)
	if err != nil {
		return nil, err
	}

	// this is required because release version does not delete entire table directory but only the version directory
	// and hence the meta file may exist but the db file may not
	if m.Type == "TABLE" {
		_, err = os.Stat(d.localDBPath(name, m.Version))
	} else {
		_, err = os.Stat(d.localTableDir(name, m.Version))
	}
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, errNotFound
		}
		return nil, err
	}
	return m, nil
}

func (d *db) writeTableMeta(name string, meta *tableMeta) error {
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("create: marshal meta failed: %w", err)
	}
	err = os.WriteFile(d.localMetaPath(name), metaBytes, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("create: write meta failed: %w", err)
	}
	return nil
}

func (d *db) localTableDir(name, version string) string {
	var path string
	if version == "" {
		path = filepath.Join(d.localPath, name)
	} else {
		path = filepath.Join(d.localPath, name, version)
	}
	return path
}

func (d *db) localMetaPath(table string) string {
	return filepath.Join(d.localPath, table, "meta.json")
}

func (d *db) localDBPath(table, version string) string {
	return filepath.Join(d.localPath, table, version, "data.db")
}

// initLocalTable creates a directory for the table in the local path.
// If version is provided, a version directory is also created.
func (d *db) initLocalTable(name, version string) error {
	err := os.MkdirAll(d.localTableDir(name, version), fs.ModePerm)
	if err != nil {
		return fmt.Errorf("create: unable to create dir %q: %w", name, err)
	}
	return nil
}

// removeTableVersion removes the table version from the catalog and deletes the local table files.
func (d *db) removeTableVersion(ctx context.Context, name, version string) error {
	err := d.metaSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.metaSem.Release(1)

	_, err = d.dbHandle.ExecContext(ctx, "DETACH DATABASE IF EXISTS "+safeSQLName(dbName(name, version)))
	if err != nil {
		return err
	}
	return d.deleteLocalTableFiles(name, version)
}

// deleteLocalTableFiles delete table files for the given table name. If version is provided, only that version is deleted.
func (d *db) deleteLocalTableFiles(name, version string) error {
	return os.RemoveAll(d.localTableDir(name, version))
}

func (d *db) iterateLocalTables(cleanup bool, fn func(name string, meta *tableMeta) error) error {
	entries, err := os.ReadDir(d.localPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		meta, err := d.tableMeta(entry.Name())
		if err != nil {
			if !cleanup {
				continue
			}
			d.logger.Debug("cleanup: remove table", slog.String("table", entry.Name()))
			err = d.deleteLocalTableFiles(entry.Name(), "")
			if err != nil {
				return err
			}
			continue
		}
		// also remove older versions
		if cleanup {
			versions, err := os.ReadDir(d.localTableDir(entry.Name(), ""))
			if err != nil {
				return err
			}
			for _, version := range versions {
				if !version.IsDir() {
					continue
				}
				if version.Name() == meta.Version {
					continue
				}
				d.logger.Debug("cleanup: remove old version", slog.String("table", entry.Name()), slog.String("version", version.Name()))
				err = d.deleteLocalTableFiles(entry.Name(), version.Name())
				if err != nil {
					return err
				}
			}
		}
		err = fn(entry.Name(), meta)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *db) prepareSnapshot(ctx context.Context, conn *sqlx.Conn, s *snapshot) error {
	err := d.metaSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.metaSem.Release(1)

	if s.ready {
		_, err = conn.ExecContext(ctx, "USE "+schemaName(s.id))
		return err
	}

	_, err = conn.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+schemaName(s.id))
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(ctx, "USE "+schemaName(s.id))
	if err != nil {
		return err
	}

	err = d.attachTables(ctx, conn, s.tables, "")
	if err != nil {
		return err
	}
	s.ready = true
	return nil
}

func (d *db) removeSnapshot(ctx context.Context, id int) error {
	err := d.metaSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer d.metaSem.Release(1)

	_, err = d.dbHandle.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName(id)))
	return err
}

func (d *db) migrateDB() error {
	// does not accept context by choice so that migration is not interrupted by context cancel
	// The queries are expected to be fast
	entries, err := os.ReadDir(d.opts.LocalPath)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}

	// presence of meta.json indicates that the db is already migrated
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		_, err := os.Stat(filepath.Join(d.localPath, entry.Name(), "meta.json"))
		if err == nil {
			// already migrated
			return nil
		}
	}

	// files are in old structure
	// Table migration requires following things:
	// 1. Move the db file named <version>.db to the version folder
	// 2. Rename the table from "default" to the table name
	// 3. Create meta.json
	//
	// Views are directly present in main.db and are not versioned in old structure
	tables := make(map[string]*tableMeta)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(d.localPath, entry.Name(), "version.txt"))
		if err != nil {
			// version.txt not found, skip this directory, also safe to delete this directory
			_ = os.RemoveAll(filepath.Join(d.localPath, entry.Name()))
			continue
		}
		// get version
		version := strings.TrimSpace(string(contents))
		err = d.initLocalTable(entry.Name(), version)
		if err != nil {
			return err
		}

		err = os.Rename(filepath.Join(d.localPath, entry.Name(), fmt.Sprintf("%v.db", version)), filepath.Join(d.localPath, entry.Name(), version, "data.db"))
		if err != nil {
			return err
		}
		_ = os.RemoveAll(filepath.Join(d.localPath, entry.Name(), "version.txt"))
		err = renameTable(context.Background(), filepath.Join(d.localPath, entry.Name(), version, "data.db"), "default", entry.Name())
		if err != nil {
			return err
		}
		// create meta.json file
		meta := &tableMeta{
			Name:           entry.Name(),
			Version:        version,
			CreatedVersion: version,
			Type:           "TABLE",
		}
		err = d.writeTableMeta(entry.Name(), meta)
		if err != nil {
			return err
		}
		tables[entry.Name()] = meta
	}

	// handle views
	// present directly in main.db file
	if err := d.migrateViews(tables); err != nil {
		return err
	}
	// drop the old db files
	_ = os.RemoveAll(filepath.Join(d.localPath, "main.db"))
	_ = os.RemoveAll(filepath.Join(d.localPath, "main.db.wal"))
	return err
}

func (d *db) migrateViews(existingTables map[string]*tableMeta) error {
	db, err := sql.Open("duckdb", filepath.Join(d.localPath, "main.db"))
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query("SELECT view_name, sql FROM duckdb_views() WHERE database_name = current_database() AND schema_name = current_schema() AND internal = false")
	if err != nil {
		return err
	}
	var viewName, viewSQL string
	for rows.Next() {
		err = rows.Scan(&viewName, &viewSQL)
		if err != nil {
			return err
		}
		if _, ok := existingTables[viewName]; ok {
			// view on a table, skip
			continue
		}
		err := d.initLocalTable(viewName, "")
		if err != nil {
			return err
		}
		version := newVersion()
		// create meta.json file
		meta := &tableMeta{
			Name:           viewName,
			Version:        version,
			CreatedVersion: version,
			Type:           "VIEW",
			SQL:            viewSQL,
		}
		err = d.writeTableMeta(viewName, meta)
		if err != nil {
			return err
		}
	}
	return rows.Err()
}

func (d *db) duckdbOnGCS() (bool, error) {
	contents, err := os.ReadFile(filepath.Join(d.localPath, "_duckdb_on_gcs_.txt"))
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(contents)) == "true", nil
}

type tableMeta struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	CreatedVersion string `json:"created_version"`
	Type           string `json:"type"` // either TABLE or VIEW
	SQL            string `json:"sql"`  // populated for views
}

func renameTable(ctx context.Context, dbFile, old, newName string) error {
	db, err := sql.Open("duckdb", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	// TODO :: create temporary views when attaching tables to write connection to avoid left views in .db file
	// In that case this will not be required.
	_, err = db.ExecContext(ctx, fmt.Sprintf("DROP VIEW IF EXISTS %s", safeSQLName(newName)))
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s RENAME TO %s", safeSQLName(old), safeSQLName(newName)))
	return err
}

func newVersion() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

func dbName(table, version string) string {
	return fmt.Sprintf("%s__%s__db", table, version)
}

// Regex to parse human-readable size returned by DuckDB
// nolint
var humanReadableSizeRegex = regexp.MustCompile(`^([\d.]+)\s*(\S+)$`)

// Reversed logic of StringUtil::BytesToHumanReadableString
// see https://github.com/cran/duckdb/blob/master/src/duckdb/src/common/string_util.cpp#L157
// Examples: 1 bytes, 2 bytes, 1KB, 1MB, 1TB, 1PB
// nolint
func humanReadableSizeToBytes(sizeStr string) (float64, error) {
	var multiplier float64

	match := humanReadableSizeRegex.FindStringSubmatch(sizeStr)

	if match == nil {
		return 0, fmt.Errorf("invalid size format: '%s'", sizeStr)
	}

	sizeFloat, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, err
	}

	switch match[2] {
	case "byte", "bytes":
		multiplier = 1
	case "KB":
		multiplier = 1000
	case "MB":
		multiplier = 1000 * 1000
	case "GB":
		multiplier = 1000 * 1000 * 1000
	case "TB":
		multiplier = 1000 * 1000 * 1000 * 1000
	case "PB":
		multiplier = 1000 * 1000 * 1000 * 1000 * 1000
	case "KiB":
		multiplier = 1024
	case "MiB":
		multiplier = 1024 * 1024
	case "GiB":
		multiplier = 1024 * 1024 * 1024
	case "TiB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "PiB":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown size unit '%s' in '%s'", match[2], sizeStr)
	}

	return sizeFloat * multiplier, nil
}

func schemaName(gen int) string {
	return fmt.Sprintf("main_%v", gen)
}

func generateDSN(path, encodedQuery string) string {
	if encodedQuery == "" {
		return path
	}
	return path + "?" + encodedQuery
}
