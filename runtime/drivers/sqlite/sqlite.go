package sqlite

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"

	// Load sqlite driver
	_ "modernc.org/sqlite"
)

func init() {
	drivers.Register("sqlite", driver{})
}

type driver struct{}

func (d driver) Open(dsn string, logger *zap.Logger) (drivers.Connection, error) {
	// The sqlite driver requires the DSN to contain "_time_format=sqlite" to support TIMESTAMP types in all timezones.
	if !strings.Contains(dsn, "_time_format") {
		if strings.Contains(dsn, "?") {
			dsn += "&_time_format=sqlite"
		} else {
			dsn += "?_time_format=sqlite"
		}
	}

	// Open DB handle
	db, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return &connection{db: db}, nil
}

func (d driver) Drop(dsn string, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

type connection struct {
	db *sqlx.DB
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return c.db.Close()
}

// Registry implements drivers.Connection.
func (c *connection) RegistryStore() (drivers.RegistryStore, bool) {
	return c, true
}

// Catalog implements drivers.Connection.
func (c *connection) CatalogStore() (drivers.CatalogStore, bool) {
	return c, true
}

// Repo implements drivers.Connection.
func (c *connection) RepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *connection) OLAPStore() (drivers.OLAPStore, bool) {
	return nil, false
}
