package druid

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"

	// Load calcite avatica driver for druid
	_ "github.com/apache/calcite-avatica-go/v5"
)

func init() {
	drivers.Register("druid", driver{})
}

type driver struct{}

// Open connects to Druid using Avatica.
// Note that the Druid connection string must have the form "http://host/druid/v2/sql/avatica-protobuf/".
func (d driver) Open(dsn string, logger *zap.Logger) (drivers.Connection, error) {
	db, err := sqlx.Open("avatica", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(40)

	conn := &connection{db: db}
	return conn, nil
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
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) CatalogStore() (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *connection) RepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *connection) OLAPStore() (drivers.OLAPStore, bool) {
	return c, true
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}
