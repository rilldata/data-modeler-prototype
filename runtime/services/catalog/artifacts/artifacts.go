package artifacts

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

var Artifacts = make(map[string]Artifact)

var FileReadError = errors.New("failed to read artifact")
var InvalidFileName = errors.New("invalid file name")

func Register(name string, artifact Artifact) {
	if Artifacts[name] != nil {
		panic(fmt.Errorf("already registered artifact type with name '%s'", name))
	}
	Artifacts[name] = artifact
}

type Artifact interface {
	DeSerialise(ctx context.Context, filePath string, blob string) (*drivers.CatalogEntry, error)
	Serialise(ctx context.Context, catalogObject *drivers.CatalogEntry) (string, error)
}

func Read(ctx context.Context, repoStore drivers.RepoStore, instID string, filePath string) (*drivers.CatalogEntry, error) {
	extension := fileutil.FullExt(filePath)
	artifact, ok := Artifacts[extension]
	if !ok {
		return nil, fmt.Errorf("no artifact found for %s", extension)
	}

	blob, err := repoStore.Get(ctx, instID, filePath)
	if err != nil {
		return nil, FileReadError
	}

	catalog, err := artifact.DeSerialise(ctx, filePath, blob)
	if err != nil {
		return nil, err
	}

	if !isValidName(fileutil.Stem(filePath)) {
		return nil, InvalidFileName
	}

	catalog.Path = filePath
	return catalog, nil
}

func Write(ctx context.Context, repoStore drivers.RepoStore, instID string, catalog *drivers.CatalogEntry) error {
	extension := fileutil.FullExt(catalog.Path)
	artifact, ok := Artifacts[extension]
	if !ok {
		return fmt.Errorf("no artifact found for %s", extension)
	}

	blob, err := artifact.Serialise(ctx, catalog)
	if err != nil {
		return err
	}

	return repoStore.Put(ctx, instID, catalog.Path, strings.NewReader(blob))
}

var regex = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")

func isValidName(itemName string) bool {
	return regex.MatchString(itemName)
}
