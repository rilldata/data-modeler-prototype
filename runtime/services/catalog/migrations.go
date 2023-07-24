package catalog

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/arrayutil"
	"github.com/rilldata/rill/runtime/pkg/dag"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	// Load migrators
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/sql"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/metricsviews"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/models"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
)

type ReconcileConfig struct {
	DryRun            bool
	Strict            bool
	ChangedPaths      []string
	ForcedPaths       []string
	SafeSourceRefresh bool
}

type ReconcileResult struct {
	AddedObjects   []*drivers.CatalogEntry
	UpdatedObjects []*drivers.CatalogEntry
	DroppedObjects []*drivers.CatalogEntry
	AffectedPaths  []string
	Errors         []*runtimev1.ReconcileError
}

func NewReconcileResult() *ReconcileResult {
	return &ReconcileResult{
		AddedObjects:   make([]*drivers.CatalogEntry, 0),
		UpdatedObjects: make([]*drivers.CatalogEntry, 0),
		DroppedObjects: make([]*drivers.CatalogEntry, 0),
		AffectedPaths:  make([]string, 0),
		Errors:         make([]*runtimev1.ReconcileError, 0),
	}
}

func (r *ReconcileResult) collectAffectedPaths() {
	pathDuplicates := make(map[string]bool)
	for _, added := range r.AddedObjects {
		r.AffectedPaths = append(r.AffectedPaths, added.Path)
		pathDuplicates[added.Path] = true
	}
	for _, updated := range r.UpdatedObjects {
		if pathDuplicates[updated.Path] {
			continue
		}
		r.AffectedPaths = append(r.AffectedPaths, updated.Path)
		pathDuplicates[updated.Path] = true
	}
	for _, deleted := range r.DroppedObjects {
		if pathDuplicates[deleted.Path] {
			continue
		}
		r.AffectedPaths = append(r.AffectedPaths, deleted.Path)
		pathDuplicates[deleted.Path] = true
	}
	for _, errored := range r.Errors {
		if pathDuplicates[errored.FilePath] {
			continue
		}
		r.AffectedPaths = append(r.AffectedPaths, errored.FilePath)
		pathDuplicates[errored.FilePath] = true
	}
}

type ArtifactError struct {
	Error error
	Path  string
}

// TODO: support loading existing projects

func (s *Service) Reconcile(ctx context.Context, conf ReconcileConfig) (*ReconcileResult, error) {
	s.Meta.lock.Lock()
	defer s.Meta.lock.Unlock()

	result := NewReconcileResult()
	s.logger.Info("reconcile started", zap.Any("config", conf))

	if len(conf.ChangedPaths) == 0 || slices.Contains(conf.ChangedPaths, "rill.yaml") {
		// check the project is initialized
		c := rillv1beta.New(s.Repo, s.InstID)
		if !c.IsInit(ctx) {
			return nil, fmt.Errorf("not a valid project: rill.yaml not found")
		}

		// set project variables from rill.yaml in instance
		if err := s.setProjectVariables(ctx); err != nil {
			return nil, err
		}
	}

	// collect repos and create migration items
	migrationMap, reconcileErrors, err := s.getMigrationMap(ctx, conf)
	if err != nil {
		return nil, err
	}
	result.Errors = reconcileErrors

	// order the items to have parents before children
	migrations, reconcileErrors := s.collectMigrationItems(migrationMap)
	result.Errors = append(result.Errors, reconcileErrors...)

	err = s.runMigrationItems(ctx, conf, migrations, result)
	if err != nil {
		return nil, err
	}

	if !conf.DryRun {
		// TODO: changes to the file will not be picked up if done while running migration
		s.Meta.LastMigration = time.Now()
		s.Meta.hasMigrated = true
	}
	result.collectAffectedPaths()
	s.logger.Info("reconcile completed", zap.Any("result", result))
	return result, nil
}

// collectMigrationItems collects all valid MigrationItem
// It will order the items based on dag with parents coming before children.
func (s *Service) collectMigrationItems(
	migrationMap map[string]*MigrationItem,
) ([]*MigrationItem, []*runtimev1.ReconcileError) {
	migrationItems := make([]*MigrationItem, 0)
	reconcileErrors := make([]*runtimev1.ReconcileError, 0)
	visited := make(map[string]int)
	update := make(map[string]bool)

	// temporary local dag for just the items to be migrated
	// this will also help in getting a dag for new items
	// TODO: is there a better way to do this?
	tempDag := dag.NewDAG()
	for name, migration := range migrationMap {
		_, err := tempDag.Add(name, migration.NormalizedDependencies)
		if err != nil {
			reconcileErrors = append(reconcileErrors, &runtimev1.ReconcileError{
				Code:     runtimev1.ReconcileError_CODE_SOURCE,
				Message:  err.Error(),
				FilePath: migration.Path,
			})
		}
	}

	for _, name := range tempDag.TopologicalSort() {
		item, ok := migrationMap[name]
		if !ok {
			continue
		}

		if item.Type == MigrationNoChange {
			if update[name] {
				// items identified as to created/updated because a parent changed
				// but was initially marked no change
				if item.CatalogInStore == nil {
					item.Type = MigrationCreate
				} else {
					item.Type = MigrationUpdate
				}
			} else if _, ok := s.Meta.NameToPath[item.NormalizedName]; ok {
				// this allows parents later in the order to re add children
				visited[name] = -1
				continue
			}
		}

		visited[name] = len(migrationItems)
		migrationItems = append(migrationItems, item)

		newChildren := tempDag.GetDeepChildren(name)

		if item.NewCatalog != nil && item.NewCatalog.Embedded {
			if len(newChildren) == 0 {
				// if embedded item's children is empty then delete it
				item.Type = MigrationDelete
			} else if item.Type == MigrationReportUpdate {
				// do not update children for embedded sources that didn't change
				continue
			}
		}

		// get all the children and make sure they are not present before the parent in the order
		children := arrayutil.Dedupe(append(
			s.Meta.dag.GetDeepChildren(name),
			newChildren...,
		))
		if item.FromName != "" {
			children = append(children, arrayutil.Dedupe(append(
				s.Meta.dag.GetDeepChildren(strings.ToLower(item.FromNormalizedName)),
				tempDag.GetDeepChildren(strings.ToLower(item.FromNormalizedName))...,
			))...)
		}

		for _, child := range children {
			i, ok := visited[child]
			if !ok {
				if item.Type != MigrationNoChange {
					// if not already visited, mark the child as needing update
					update[child] = true
				}
				continue
			}

			var childItem *MigrationItem
			// if a child was already visited push to the end
			// this can happen when there is a rename and the new DAG wont have the old order
			visited[child] = len(migrationItems)
			if i != -1 {
				childItem = migrationItems[i]
				// mark the original position as nil. this is cleaned up later
				migrationItems[i] = nil
			} else {
				childItem = migrationMap[child]
			}

			migrationItems = append(migrationItems, childItem)
			if item.Type == MigrationNoChange {
				continue
			}
			if childItem.Type == MigrationNoChange || childItem.Error != nil {
				// if the child has no change then mark it as update or create based on presence of catalog in store
				if childItem.CatalogInStore == nil {
					childItem.Type = MigrationCreate
				} else {
					childItem.Type = MigrationUpdate
				}
			}
		}
	}

	// cleanup any nil values that occurred by pushing child later in the order
	cleanedMigrationItems := make([]*MigrationItem, 0)
	for _, migration := range migrationItems {
		if migration == nil {
			continue
		}
		cleanedMigrationItems = append(cleanedMigrationItems, migration)
	}

	return cleanedMigrationItems, reconcileErrors
}

// TODO: test changing source make an invalid model valid. should propagate validity to metrics

// runMigrationItems runs various actions from MigrationItem based on MigrationItem.Type.
func (s *Service) runMigrationItems(
	ctx context.Context,
	conf ReconcileConfig,
	migrations []*MigrationItem,
	result *ReconcileResult,
) error {
	for _, item := range migrations {
		if item.Error != nil {
			result.Errors = append(result.Errors, item.Error)
		}

		var validationErrors []*runtimev1.ReconcileError

		if item.CatalogInFile != nil {
			validationErrors = migrator.Validate(ctx, s.Olap, item.CatalogInFile)
		}

		var err error
		failed := false
		if len(validationErrors) > 0 {
			// do not run migration if validation failed
			result.Errors = append(result.Errors, validationErrors...)
			failed = true
		} else if !conf.DryRun {
			if item.CatalogInStore != nil {
				// make sure store catalog has the correct name
				// could be different in cases like "rename with different case"
				item.CatalogInStore.Name = item.Name
			}
			// only run the actual migration if in dry run
			switch item.Type {
			case MigrationNoChange:
				recErr := s.addToDag(item)
				if recErr != nil {
					result.Errors = append(result.Errors, recErr)
				}
			case MigrationCreate:
				if item.CatalogInFile == nil {
					break
				}
				err = s.createInStore(ctx, item)
				result.AddedObjects = append(result.AddedObjects, item.CatalogInFile)
			case MigrationRename:
				if item.CatalogInFile == nil {
					break
				}
				err = s.renameInStore(ctx, item)
				result.UpdatedObjects = append(result.UpdatedObjects, item.CatalogInFile)
			case MigrationUpdate:
				if item.CatalogInFile == nil {
					break
				}
				err = s.updateInStore(ctx, item)
				result.UpdatedObjects = append(result.UpdatedObjects, item.CatalogInFile)
			case MigrationReportUpdate:
				// only report the update
				// UI needs to know when dag changed. we use this for now to notify it
				// TODO: have a better way to notify UI of DAG changes
				result.UpdatedObjects = append(result.UpdatedObjects, item.CatalogInFile)
				recErr := s.addToDag(item)
				if recErr != nil {
					result.Errors = append(result.Errors, recErr)
				}
			case MigrationDelete:
				err = s.deleteInStore(ctx, item)
				result.DroppedObjects = append(result.DroppedObjects, item.CatalogInStore)
			}
		}

		if err != nil {
			errCode := runtimev1.ReconcileError_CODE_OLAP
			var connectorErr *drivers.PermissionDeniedError
			if errors.As(err, &connectorErr) {
				errCode = runtimev1.ReconcileError_CODE_SOURCE_PERMISSION_DENIED
			}
			result.Errors = append(result.Errors, &runtimev1.ReconcileError{
				Code:     errCode,
				Message:  err.Error(),
				FilePath: item.Path,
			})
			failed = true
		}

		if failed && !conf.DryRun {
			shouldDelete := !conf.SafeSourceRefresh || item.NewCatalog.Type != drivers.ObjectTypeSource
			var err error
			if shouldDelete {
				// remove entity from catalog and OLAP if it failed validation or during migration
				err = s.Catalog.DeleteEntry(ctx, s.InstID, item.Name)
				if err != nil {
					// shouldn't ideally happen
					result.Errors = append(result.Errors, &runtimev1.ReconcileError{
						Code:     runtimev1.ReconcileError_CODE_OLAP,
						Message:  err.Error(),
						FilePath: item.Path,
					})
				}
			}
			_, err = s.Meta.dag.Add(item.NormalizedName, item.NormalizedDependencies)
			if err != nil {
				result.Errors = append(result.Errors, &runtimev1.ReconcileError{
					Code:     runtimev1.ReconcileError_CODE_OLAP,
					Message:  err.Error(),
					FilePath: item.Path,
				})
			}
			if item.CatalogInStore != nil && shouldDelete {
				err := migrator.Delete(ctx, s.Olap, item.CatalogInStore)
				if err != nil {
					// shouldn't ideally happen
					result.Errors = append(result.Errors, &runtimev1.ReconcileError{
						Code:     runtimev1.ReconcileError_CODE_OLAP,
						Message:  err.Error(),
						FilePath: item.Path,
					})
				}
			}
			if conf.Strict {
				return err
			}
		}
	}

	return nil
}

func (s *Service) createInStore(ctx context.Context, item *MigrationItem) error {
	s.Meta.NameToPath[item.NormalizedName] = item.Path
	// add the item to dag
	_, err := s.Meta.dag.Add(item.NormalizedName, item.NormalizedDependencies)
	if err != nil {
		return err
	}

	inst, err := s.RegistryStore.FindInstance(ctx, s.InstID)
	if err != nil {
		return err
	}

	// NOTE :: IngestStorageLimitInBytes check will only work if sources are ingested in serial
	ingestionLimit, err := s.getSourceIngestionLimit(ctx, inst)
	if err != nil {
		return err
	}
	opts := migrator.Options{
		InstanceEnv:               inst.ResolveVariables(),
		IngestStorageLimitInBytes: ingestionLimit,
	}

	// create in olap
	err = s.wrapMigrator(item.CatalogInFile, func() error {
		return migrator.Create(ctx, s.Olap, s.Repo, opts, item.CatalogInFile, s.logger)
	})
	if err != nil {
		return err
	}

	// update the catalog object and create it in store
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	_, err = s.Catalog.FindEntry(ctx, s.InstID, item.Name)
	// create or updated
	if err != nil {
		if !errors.Is(err, drivers.ErrNotFound) {
			return err
		}
		return s.Catalog.CreateEntry(ctx, s.InstID, catalog)
	}
	return s.Catalog.UpdateEntry(ctx, s.InstID, catalog)
}

func (s *Service) renameInStore(ctx context.Context, item *MigrationItem) error {
	fromLowerName := strings.ToLower(item.FromName)
	delete(s.Meta.NameToPath, fromLowerName)
	s.Meta.NameToPath[item.NormalizedName] = item.Path

	// delete old item and add new item to dag
	s.Meta.dag.Delete(fromLowerName)
	_, err := s.Meta.dag.Add(item.NormalizedName, item.NormalizedDependencies)
	if err != nil {
		return err
	}

	// rename the item in olap
	err = migrator.Rename(ctx, s.Olap, item.FromName, item.CatalogInFile)
	if err != nil {
		return err
	}

	// delete the old catalog object
	// TODO: do we need a rename here?
	err = s.Catalog.DeleteEntry(ctx, s.InstID, item.FromName)
	if err != nil {
		return err
	}
	// update the catalog object and create it in store
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	return s.Catalog.CreateEntry(ctx, s.InstID, catalog)
}

func (s *Service) updateInStore(ctx context.Context, item *MigrationItem) error {
	s.Meta.NameToPath[item.NormalizedName] = item.Path
	// add the item to dag with new dependencies
	_, err := s.Meta.dag.Add(item.NormalizedName, item.NormalizedDependencies)
	if err != nil {
		return err
	}

	inst, err := s.RegistryStore.FindInstance(ctx, s.InstID)
	if err != nil {
		return err
	}

	// update in olap
	if item.Type == MigrationUpdate {
		ingestionLimit, err := s.getSourceIngestionLimit(ctx, inst)
		if err != nil {
			return err
		}
		err = s.wrapMigrator(item.CatalogInFile, func() error {
			opts := migrator.Options{
				InstanceEnv:               inst.ResolveVariables(),
				IngestStorageLimitInBytes: ingestionLimit,
			}
			return migrator.Update(ctx, s.Olap, s.Repo, opts, item.CatalogInStore, item.CatalogInFile, s.logger)
		})
		if err != nil {
			return err
		}
	}
	// update the catalog object and update it in store
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	return s.Catalog.UpdateEntry(ctx, s.InstID, catalog)
}

func (s *Service) deleteInStore(ctx context.Context, item *MigrationItem) error {
	delete(s.Meta.NameToPath, item.NormalizedName)

	// delete item from dag
	s.Meta.dag.Delete(item.NormalizedName)
	// delete item from olap
	err := migrator.Delete(ctx, s.Olap, item.CatalogInStore)
	if err != nil {
		return err
	}

	// delete from catalog store
	return s.Catalog.DeleteEntry(ctx, s.InstID, item.Name)
}

func (s *Service) updateCatalogObject(ctx context.Context, item *MigrationItem) (*drivers.CatalogEntry, error) {
	// get artifact stats
	// stat will not succeed for embedded entries
	repoStat, _ := s.Repo.Stat(ctx, s.InstID, item.Path)

	// convert protobuf to database object
	catalogEntry := item.CatalogInFile
	// NOTE: Previously there was a copy here when using the API types. This might have to reverted.

	// set the UpdatedOn as LastUpdated from the artifact file
	// this will allow to not reprocess unchanged files
	if repoStat != nil {
		catalogEntry.UpdatedOn = repoStat.LastUpdated
	}
	catalogEntry.RefreshedOn = time.Now()

	err := migrator.SetSchema(ctx, s.Olap, catalogEntry)
	if err != nil {
		return nil, err
	}

	return catalogEntry, nil
}

// wrapMigrator is a temporary solution to log source related messages.
func (s *Service) wrapMigrator(catalogEntry *drivers.CatalogEntry, run func() error) error {
	st := time.Now()
	if catalogEntry.Type == drivers.ObjectTypeSource {
		path := catalogEntry.GetSource().Properties.Fields["path"].GetStringValue()
		if path == "" {
			s.logger.Named("console").Info(fmt.Sprintf("Ingesting source %q", catalogEntry.Name))
		} else {
			s.logger.Named("console").Info(fmt.Sprintf(
				"Ingesting source %q from %q",
				catalogEntry.Name, catalogEntry.GetSource().Properties.Fields["path"].GetStringValue(),
			))
		}
	}
	err := run()
	if catalogEntry.Type == drivers.ObjectTypeSource {
		if err != nil {
			s.logger.Error(fmt.Sprintf("Ingestion failed for %q : %s", catalogEntry.Name, err.Error()))
		} else {
			s.logger.Named("console").Info(fmt.Sprintf("Finished ingesting %q, took %v seconds", catalogEntry.Name, time.Since(st).Seconds()))
		}
	}
	return err
}

func (s *Service) addToDag(item *MigrationItem) *runtimev1.ReconcileError {
	if _, ok := s.Meta.NameToPath[item.NormalizedName]; ok {
		return nil
	}
	// this is perhaps an init. so populate cache data
	s.Meta.NameToPath[item.NormalizedName] = item.Path
	_, err := s.Meta.dag.Add(item.NormalizedName, item.NormalizedDependencies)
	if err != nil {
		return &runtimev1.ReconcileError{
			Code:     runtimev1.ReconcileError_CODE_SOURCE,
			Message:  err.Error(),
			FilePath: item.Path,
		}
	}
	return nil
}

func (s *Service) getSourceIngestionLimit(ctx context.Context, inst *drivers.Instance) (int64, error) {
	if inst.IngestionLimitBytes == 0 {
		return math.MaxInt64, nil
	}

	var sizeSoFar int64
	entries, err := s.Catalog.FindEntries(ctx, s.InstID, drivers.ObjectTypeSource)
	if err != nil {
		return 0, err
	}
	for _, entry := range entries {
		sizeSoFar += entry.BytesIngested
	}

	limitInBytes := inst.IngestionLimitBytes
	limitInBytes -= sizeSoFar
	if limitInBytes < 0 {
		return 0, nil
	}
	return limitInBytes, nil
}

func (s *Service) setProjectVariables(ctx context.Context) error {
	c := rillv1beta.New(s.Repo, s.InstID)
	proj, err := c.ProjectConfig(ctx)
	if err != nil {
		return err
	}

	inst, err := s.RegistryStore.FindInstance(ctx, s.InstID)
	if err != nil {
		return err
	}

	inst.ProjectVariables = proj.Variables
	return s.RegistryStore.EditInstance(ctx, inst)
}
