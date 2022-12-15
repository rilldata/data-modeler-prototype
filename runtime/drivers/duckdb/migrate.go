package duckdb

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

// Embed migrations directory in the binary
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// Name of the table that tracks migrations
var migrationVersionTable = "rill.migration_version"

// Migrate implements drivers.Connection.
// Migrate for DuckDB may not be safe for concurrent use.
func (c *connection) Migrate(ctx context.Context) (err error) {
	// Create rill schema if it doens't exist
	_, err = c.db.ExecContext(ctx, "create schema if not exists rill")
	if err != nil {
		return err
	}

	// Create migrationVersionTable if it doesn't exist
	_, err = c.db.ExecContext(ctx, fmt.Sprintf("create table if not exists %s(version integer not null)", migrationVersionTable))
	if err != nil {
		return err
	}

	// Set the version to 0 if table is empty
	_, err = c.db.ExecContext(ctx, fmt.Sprintf("insert into %s(version) select 0 where 0=(select count(*) from %s)", migrationVersionTable, migrationVersionTable))
	if err != nil {
		return err
	}

	// Get version of latest migration
	var currentVersion int
	err = c.db.QueryRowContext(ctx, fmt.Sprintf("select version from %s", migrationVersionTable)).Scan(&currentVersion)
	if err != nil {
		return err
	}

	// Iterate over migrations (sorted by filename)
	files, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}

	for _, file := range files {
		// Extract version number from filename
		version, err := migrationFilenameToVersion(file.Name())
		if err != nil {
			return fmt.Errorf("unexpected migration filename: %s", file.Name())
		}

		// Skip migrations below current version
		if version <= currentVersion {
			continue
		}

		// Read SQL
		sql, err := migrationsFS.ReadFile(path.Join("migrations", file.Name()))
		if err != nil {
			return err
		}

		err = migrateSingle(ctx, c, file, sql, version)
		if err != nil {
			return err
		}
	}

	return nil
}

func migrateSingle(ctx context.Context, c *connection, file fs.DirEntry, sql []byte, version int) (err error) {
	// Start a transaction
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Run migration
	_, err = tx.ExecContext(ctx, string(sql))
	if err != nil {
		return fmt.Errorf("failed to run migration '%s': %w", file.Name(), err)
	}

	// Update migration version
	_, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET version=?", migrationVersionTable), version)
	if err != nil {
		return err
	}

	// Commit migration
	err = tx.Commit()
	if err != nil {
		return err
	}

	// Force DuckDB to merge WAL into .db file
	_, err = c.db.ExecContext(ctx, "CHECKPOINT;")
	if err != nil {
		return err
	}
	return nil
}

// MigrationStatus implements drivers.Connection
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	// Get current version
	err = c.db.QueryRowxContext(ctx, fmt.Sprintf("select version from %s", migrationVersionTable)).Scan(&current)
	if err != nil {
		return 0, 0, err
	}

	// Set desired to version number of last migration file
	files, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return 0, 0, err
	}
	if len(files) > 0 {
		file := files[len(files)-1]
		version, err := migrationFilenameToVersion(file.Name())
		if err != nil {
			return 0, 0, fmt.Errorf("unexpected migration filename: %s", file.Name())
		}
		desired = version
	}

	return current, desired, nil
}

func migrationFilenameToVersion(name string) (int, error) {
	return strconv.Atoi(strings.TrimSuffix(name, ".sql"))
}
