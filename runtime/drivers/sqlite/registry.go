package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers"
)

// FindInstances implements drivers.RegistryStore
func (c *connection) FindInstances(ctx context.Context) ([]*drivers.Instance, error) {
	return c.findInstances(ctx, "")
}

// FindInstance implements drivers.RegistryStore
func (c *connection) FindInstance(ctx context.Context, id string) (*drivers.Instance, bool, error) {
	is, err := c.findInstances(ctx, "WHERE id = $1", id)
	if err != nil {
		return nil, false, err
	}
	if len(is) == 0 {
		return nil, false, nil
	}
	return is[0], true, nil
}

func (c *connection) findInstances(ctx context.Context, whereClause string, args ...any) ([]*drivers.Instance, error) {
	sql := fmt.Sprintf("SELECT id, olap_driver, olap_dsn, repo_driver, repo_dsn, embed_catalog, created_on, updated_on FROM instances %s ORDER BY id", whereClause)

	rows, err := c.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*drivers.Instance
	for rows.Next() {
		i := &drivers.Instance{}
		err := rows.Scan(&i.ID, &i.OLAPDriver, &i.OLAPDSN, &i.RepoDriver, &i.RepoDSN, &i.EmbedCatalog, &i.CreatedOn, &i.UpdatedOn)
		if err != nil {
			return nil, err
		}
		res = append(res, i)
	}

	return res, nil
}

// CreateInstance implements drivers.RegistryStore
func (c *connection) CreateInstance(ctx context.Context, inst *drivers.Instance) error {
	if inst.ID == "" {
		inst.ID = uuid.NewString()
	}

	now := time.Now()
	_, err := c.db.ExecContext(
		ctx,
		"INSERT INTO instances(id, olap_driver, olap_dsn, repo_driver, repo_dsn, embed_catalog, created_on, updated_on) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $7)",
		inst.ID,
		inst.OLAPDriver,
		inst.OLAPDSN,
		inst.RepoDriver,
		inst.RepoDSN,
		inst.EmbedCatalog,
		now,
	)
	if err != nil {
		return err
	}

	// We assign manually instead of using RETURNING because it doesn't work for timestamps in SQLite
	inst.CreatedOn = now
	inst.UpdatedOn = now
	return nil
}

// DeleteInstance implements drivers.RegistryStore
func (c *connection) DeleteInstance(ctx context.Context, id string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM instances WHERE id=$1", id)
	return err
}
