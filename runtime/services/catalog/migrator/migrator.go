package migrator

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

/**
 * Any entity specific actions when a catalog is touched will be here.
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

type Options struct {
	InstanceEnv               map[string]string
	IngestStorageLimitInBytes int64
}

type EntityMigrator interface {
	Create(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, opts Options, catalog *drivers.CatalogEntry, logger *zap.Logger) error
	Update(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, opts Options, oldCatalog *drivers.CatalogEntry, newCatalog *drivers.CatalogEntry, logger *zap.Logger) error
	Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalog *drivers.CatalogEntry) error
	Delete(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) error
	GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) ([]string, []*drivers.CatalogEntry)
	Validate(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) []*runtimev1.ReconcileError
	// IsEqual checks everything but the name
	IsEqual(ctx context.Context, cat1 *drivers.CatalogEntry, cat2 *drivers.CatalogEntry) bool
	ExistsInOlap(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) (bool, error)
}

func Create(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, opts Options, catalog *drivers.CatalogEntry, logger *zap.Logger) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Create(ctx, olap, repo, opts, catalog, logger)
}

func Update(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, opts Options, oldCatalog, newCatalog *drivers.CatalogEntry, logger *zap.Logger) error {
	migrator, ok := getMigrator(newCatalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Update(ctx, olap, repo, opts, oldCatalog, newCatalog, logger)
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

func GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) ([]string, []*drivers.CatalogEntry) {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return []string{}, nil
	}
	return migrator.GetDependencies(ctx, olap, catalog)
}

// Validate also returns list of dependents.
func Validate(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) []*runtimev1.ReconcileError {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Validate(ctx, olap, catalog)
}

// IsEqual checks everything but the name.
func IsEqual(ctx context.Context, cat1, cat2 *drivers.CatalogEntry) bool {
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

func SetSchema(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) error {
	// TODO: do we need to push this to individual implementations?
	if catalog.Type == drivers.ObjectTypeMetricsView {
		return nil
	}

	table, err := olap.InformationSchema().Lookup(ctx, catalog.Name)
	if err != nil {
		return err
	}

	switch catalog.Type {
	case drivers.ObjectTypeTable:
		catalog.GetTable().Schema = table.Schema
	case drivers.ObjectTypeSource:
		catalog.GetSource().Schema = table.Schema
	case drivers.ObjectTypeModel:
		catalog.GetModel().Schema = table.Schema
	}

	return nil
}

func LastUpdated(ctx context.Context, instID string, repo drivers.RepoStore, catalog *drivers.CatalogEntry) (time.Time, error) {
	// TODO: do we need to push this to individual implementations to handle just local_file source?
	if catalog.Type != drivers.ObjectTypeSource || catalog.GetSource().Connector != "local_file" {
		// return a very old time
		return time.Time{}, nil
	}
	stat, err := repo.Stat(ctx, instID, catalog.GetSource().Properties.Fields["path"].GetStringValue())
	if err != nil {
		return time.Time{}, err
	}
	return stat.LastUpdated, nil
}

func CreateValidationError(filePath, message string) []*runtimev1.ReconcileError {
	return []*runtimev1.ReconcileError{
		{
			Code:     runtimev1.ReconcileError_CODE_VALIDATION,
			FilePath: filePath,
			Message:  message,
		},
	}
}

func getMigrator(catalog *drivers.CatalogEntry) (EntityMigrator, bool) {
	m, ok := Migrators[catalog.Type]
	return m, ok
}
