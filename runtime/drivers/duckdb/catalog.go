package duckdb

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/proto"
)

func (c *connection) FindEntries(ctx context.Context, instanceID string, typ drivers.ObjectType) ([]*drivers.CatalogEntry, error) {
	if typ == drivers.ObjectTypeUnspecified {
		return c.findEntries(ctx, "")
	}
	return c.findEntries(ctx, "WHERE type = ?", typ)
}

func (c *connection) FindEntry(ctx context.Context, instanceID, name string) (*drivers.CatalogEntry, error) {
	// Names are stored with case everywhere, but the checks should be case-insensitive.
	// Hence, the translation to lower case here.
	es, err := c.findEntries(ctx, "WHERE LOWER(name) = LOWER(?)", name)
	if err != nil {
		return nil, err
	}
	if len(es) == 0 {
		return nil, drivers.ErrNotFound
	}
	return es[0], nil
}

func (c *connection) findEntries(ctx context.Context, whereClause string, args ...any) ([]*drivers.CatalogEntry, error) {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	sql := fmt.Sprintf("SELECT name, type, object, path, bytes_ingested, embedded, created_on, updated_on, refreshed_on FROM rill.catalog %s ORDER BY lower(name)", whereClause)
	rows, err := conn.QueryxContext(ctx, sql, args...)
	if err != nil {
		return nil, c.checkErr(err)
	}
	defer rows.Close()

	var res []*drivers.CatalogEntry
	for rows.Next() {
		var objBlob []byte
		e := &drivers.CatalogEntry{}

		err := rows.Scan(&e.Name, &e.Type, &objBlob, &e.Path, &e.BytesIngested, &e.Embedded, &e.CreatedOn, &e.UpdatedOn, &e.RefreshedOn)
		if err != nil {
			return nil, c.checkErr(err)
		}

		// Parse object protobuf
		if objBlob != nil {
			switch e.Type {
			case drivers.ObjectTypeTable:
				e.Object = &runtimev1.Table{}
			case drivers.ObjectTypeSource:
				e.Object = &runtimev1.Source{}
			case drivers.ObjectTypeModel:
				e.Object = &runtimev1.Model{}
			case drivers.ObjectTypeMetricsView:
				e.Object = &runtimev1.MetricsView{}
			default:
				panic(fmt.Errorf("unexpected object type: %v", e.Type))
			}

			err = proto.Unmarshal(objBlob, e.Object)
			if err != nil {
				panic(err)
			}
		}

		res = append(res, e)
	}

	return res, nil
}

func (c *connection) CreateEntry(ctx context.Context, instanceID string, e *drivers.CatalogEntry) error {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	var present bool
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) > 0 AS present FROM rill.catalog WHERE name = ?", e.Name).Scan(&present); err != nil {
		return err
	}

	// adding a application side check instead of unique index bcz of duckdb limitations on indexes
	// https://duckdb.org/docs/sql/indexes#over-eager-unique-constraint-checking
	if present {
		return fmt.Errorf("catalog entry with name %q already exists", e.Name)
	}

	// Serialize object
	obj, err := proto.Marshal(e.Object)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = conn.ExecContext(
		ctx,
		"INSERT INTO rill.catalog(name, type, object, path, bytes_ingested, embedded, created_on, updated_on, refreshed_on) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		e.Name,
		e.Type,
		obj,
		e.Path,
		e.BytesIngested,
		e.Embedded,
		now,
		now,
		now,
	)
	if err != nil {
		return c.checkErr(err)
	}

	e.CreatedOn = now
	e.UpdatedOn = now
	e.RefreshedOn = now
	return nil
}

func (c *connection) UpdateEntry(ctx context.Context, instanceID string, e *drivers.CatalogEntry) error {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	// Serialize object
	obj, err := proto.Marshal(e.Object)
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(
		ctx,
		"UPDATE rill.catalog SET type = ?, object = ?, path = ?, bytes_ingested = ?, embedded = ?, updated_on = ?, refreshed_on = ? WHERE name = ?",
		e.Type,
		obj,
		e.Path,
		e.BytesIngested,
		e.Embedded,
		e.UpdatedOn, // TODO: Use time.Now()
		e.RefreshedOn,
		e.Name,
	)
	if err != nil {
		return c.checkErr(err)
	}

	return nil
}

func (c *connection) DeleteEntry(ctx context.Context, instanceID, name string) error {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	_, err = conn.ExecContext(ctx, "DELETE FROM rill.catalog WHERE LOWER(name) = LOWER(?)", name)
	return c.checkErr(err)
}

// DeleteEntries deletes the entire catalog table.
// This will be handled by dropping the entire rill db file when deleting instance.
// But implementing this from completeness pov.
func (c *connection) DeleteEntries(ctx context.Context, instanceID string) error {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	_, err = conn.ExecContext(ctx, "DELETE FROM rill.catalog")
	return c.checkErr(err)
}
