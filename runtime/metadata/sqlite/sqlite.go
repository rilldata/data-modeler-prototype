package sqlite

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/metadata"
	_ "modernc.org/sqlite"
)

func init() {
	metadata.Register("sqlite", driver{})
}

type driver struct{}

func (d driver) Open(dsn string) (metadata.DB, error) {
	db, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return &connection{db: db}, nil
}

type connection struct {
	db *sqlx.DB
}

func (c *connection) Close() error {
	return c.db.Close()
}

func (c *connection) FindMigrationVersion(ctx context.Context) (int, error) {
	var version int
	err := c.db.QueryRowxContext(ctx, fmt.Sprintf("select version from %s", migrationVersionTable)).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}
