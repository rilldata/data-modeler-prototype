package migrator

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
)

/**
 * Any entity specific actions when a catalog is deleted will be here.
 * EG: on delete sources will drop the table and models will drop the view
 * Any future entity specific cache invalidation will go here as well.
 *
 * TODO: does migrator name fit this?
 * TODO: is this in the right place?
 */

var Migrators = make(map[drivers.ObjectType]EntityMigrator)

func Register(t drivers.ObjectType, artifact EntityMigrator) {
	if Migrators[t] != nil {
		panic(fmt.Errorf("already registered migrator type with name '%v'", t))
	}
	Migrators[t] = artifact
}

type EntityMigrator interface {
	Create(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalog *drivers.CatalogEntry) error
	Update(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalog *drivers.CatalogEntry) error
	Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalog *drivers.CatalogEntry) error
	Delete(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) error
	GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) []string
	Validate(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) error
	// IsEqual checks everything but the name
	IsEqual(ctx context.Context, cat1 *drivers.CatalogEntry, cat2 *drivers.CatalogEntry) bool
	ExistsInOlap(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) (bool, error)
}

func Create(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalog *drivers.CatalogEntry) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Create(ctx, olap, repo, catalog)
}

func Update(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalog *drivers.CatalogEntry) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Update(ctx, olap, repo, catalog)
}

func Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalog *drivers.CatalogEntry) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Rename(ctx, olap, from, catalog)
}

func Delete(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Delete(ctx, olap, catalog)
}

func GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) []string {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return []string{}
	}
	return migrator.GetDependencies(ctx, olap, catalog)
}

// Validate also returns list of dependents
func Validate(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Validate(ctx, olap, catalog)
}

// IsEqual checks everything but the name
func IsEqual(ctx context.Context, cat1 *drivers.CatalogEntry, cat2 *drivers.CatalogEntry) bool {
	if cat1.Type != cat2.Type {
		return false
	}
	migrator, ok := getMigrator(cat1)
	if !ok {
		// no error here. not all migrators are needed
		return false
	}
	return migrator.IsEqual(ctx, cat1, cat2)
}

func ExistsInOlap(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) (bool, error) {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return false, nil
	}
	return migrator.ExistsInOlap(ctx, olap, catalog)
}

func getMigrator(catalog *drivers.CatalogEntry) (EntityMigrator, bool) {
	m, ok := Migrators[catalog.Type]
	return m, ok
}
