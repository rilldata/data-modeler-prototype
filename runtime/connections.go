package runtime

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
)

func (r *Runtime) Registry() drivers.RegistryStore {
	registry, ok := r.metastore.AsRegistry()
	if !ok {
		// Verified as registry in New, so this should never happen
		panic("metastore is not a registry")
	}
	return registry
}

func (r *Runtime) Repo(ctx context.Context, instanceID string) (drivers.RepoStore, error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	conn, err := r.connCache.get(ctx, instanceID, inst.RepoDriver, inst.RepoDSN)
	if err != nil {
		return nil, err
	}

	repo, ok := conn.AsRepoStore()
	if !ok {
		// Verified as repo when instance is created, so this should never happen
		return nil, fmt.Errorf("connection for instance '%s' is not a repo", instanceID)
	}

	return repo, nil
}

func (r *Runtime) OLAP(ctx context.Context, instanceID string) (drivers.OLAPStore, error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	conn, err := r.connCache.get(ctx, instanceID, inst.OLAPDriver, inst.OLAPDSN)
	if err != nil {
		return nil, err
	}

	olap, ok := conn.AsOLAP()
	if !ok {
		// Verified as OLAP when instance is created, so this should never happen
		return nil, fmt.Errorf("connection for instance '%s' is not an olap", instanceID)
	}

	return olap, nil
}

func (r *Runtime) Catalog(ctx context.Context, instanceID string) (drivers.CatalogStore, error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	if inst.EmbedCatalog {
		conn, err := r.connCache.get(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
		if err != nil {
			return nil, err
		}

		store, ok := conn.AsCatalogStore()
		if !ok {
			// Verified as CatalogStore when instance is created, so this should never happen
			return nil, fmt.Errorf("instance cannot embed catalog")
		}

		return store, nil
	}

	store, ok := r.metastore.AsCatalogStore()
	if !ok {
		return nil, fmt.Errorf("metastore cannot serve as catalog")
	}
	return store, nil
}

func (r *Runtime) NewCatalogService(ctx context.Context, instanceID string) (*catalog.Service, error) {
	// get all stores
	olapStore, err := r.OLAP(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	catalogStore, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	repoStore, err := r.Repo(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	registry := r.Registry()

	migrationMetadata := r.migrationMetaCache.get(instanceID)
	return catalog.NewService(catalogStore, repoStore, olapStore, registry, instanceID, r.logger, migrationMetadata), nil
}
