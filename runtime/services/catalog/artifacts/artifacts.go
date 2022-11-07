package artifacts

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
)

var Artifacts = make(map[string]Artifact)

func Register(name string, artifact Artifact) {
	if Artifacts[name] != nil {
		panic(fmt.Errorf("already registered artifact type with name '%s'", name))
	}
	Artifacts[name] = artifact
}

type Artifact interface {
	DeSerialise(ctx context.Context, blob string) (*drivers.CatalogObject, error)
	Serialise(ctx context.Context, catalogObject *drivers.CatalogObject) (string, error)
}

func Read(ctx context.Context, repoStore drivers.RepoStore, repo *api.Repo, filePath string) (*drivers.CatalogObject, error) {
	extension := filepath.Ext(filePath)
	artifact, ok := Artifacts[extension]
	if !ok {
		return nil, fmt.Errorf("no artifact found for %s", extension)
	}

	blob, err := repoStore.Get(ctx, repo.RepoId, filePath)
	if err != nil {
		return nil, err
	}

	catalog, err := artifact.DeSerialise(ctx, blob)
	if err != nil {
		return nil, err
	}

	catalog.Path = filePath
	return catalog, nil
}

func Write(ctx context.Context, repoStore drivers.RepoStore, repo *api.Repo, catalog *drivers.CatalogObject) error {
	extension := filepath.Ext(catalog.Path)
	artifact, ok := Artifacts[extension]
	if !ok {
		return fmt.Errorf("no artifact found for %s", extension)
	}

	blob, err := artifact.Serialise(ctx, catalog)
	if err != nil {
		return err
	}

	return repoStore.PutBlob(ctx, repo.RepoId, catalog.Path, blob)
}
