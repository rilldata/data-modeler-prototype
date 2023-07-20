package rillv1beta

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	"github.com/rilldata/rill/runtime/services/catalog/migrator/models"
	"github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"go.uber.org/zap"
)

// TODO :: return this to build support for all kind of variables
type Variables struct {
	ProjectVariables []drivers.PropertySchema
	Connectors       []*Connector
}

type Connector struct {
	Name            string
	Type            string
	Sources         []*runtimev1.Source
	Spec            drivers.Spec
	AnonymousAccess bool
}

func ExtractConnectors(ctx context.Context, projectPath string) ([]*Connector, error) {
	allSources := make([]*runtimev1.Source, 0)

	// get sources from files
	sourcesPath := filepath.Join(projectPath, "sources")
	sourceFiles, err := doublestar.Glob(os.DirFS(sourcesPath), "*.{yaml,yml}", doublestar.WithFailOnPatternNotExist())
	if err != nil {
		return nil, err
	}
	for _, fileName := range sourceFiles {
		src, err := readSource(ctx, filepath.Join(sourcesPath, fileName))
		if err != nil {
			return nil, fmt.Errorf("error in reading source file %v : %w", fileName, err)
		}
		allSources = append(allSources, src)
	}

	// get embedded sources from models
	modelsPath := filepath.Join(projectPath, "models")
	modelFiles, err := doublestar.Glob(os.DirFS(modelsPath), "*.sql", doublestar.WithFailOnPatternNotExist())
	if err != nil {
		return nil, err
	}
	for _, fileName := range modelFiles {
		srces, err := readEmbeddedSources(ctx, filepath.Join(modelsPath, fileName))
		if err != nil {
			return nil, fmt.Errorf("error in reading source file %v : %w", fileName, err)
		}

		allSources = append(allSources, srces...)
	}

	// keeping a map to dedup connectors
	connectorMap := make(map[key][]*runtimev1.Source)
	for _, src := range allSources {
		connector, ok := drivers.Connectors[src.Connector]
		if !ok {
			return nil, fmt.Errorf("no source connector defined for type %q", src.Connector)
		}
		// ignoring error since failure to resolve this should not break the deployment flow
		// this can fail under cases such as full or host/bucket of URI is a variable
		access, _ := connector.HasAnonymousSourceAccess(ctx, source(src.Connector, src), zap.NewNop())
		c := key{Name: src.Connector, Type: src.Connector, AnonymousAccess: access}
		srcs, ok := connectorMap[c]
		if !ok {
			srcs = make([]*runtimev1.Source, 0)
		}
		srcs = append(srcs, src)
		connectorMap[c] = srcs
	}

	result := make([]*Connector, 0)
	for k, v := range connectorMap {
		connector := drivers.Connectors[k.Type]
		result = append(result, &Connector{
			Name:            k.Name,
			Type:            k.Type,
			Spec:            connector.Spec(),
			AnonymousAccess: k.AnonymousAccess,
			Sources:         v,
		})
	}
	return result, nil
}

func readSource(ctx context.Context, path string) (*runtimev1.Source, error) {
	catalog, err := read(ctx, path)
	if err != nil {
		return nil, err
	}

	return catalog.GetSource(), nil
}

func readEmbeddedSources(ctx context.Context, path string) ([]*runtimev1.Source, error) {
	catalog, err := read(ctx, path)
	if err != nil {
		return nil, err
	}

	apiModel := catalog.GetModel()
	dependencies := models.ExtractTableNames(apiModel.Sql)

	embeddedSourcesMap := make(map[string]*runtimev1.Source)
	embeddedSources := make([]*runtimev1.Source, 0)

	for _, dependency := range dependencies {
		source, ok := sources.ParseEmbeddedSource(dependency)
		if !ok {
			continue
		}
		if _, ok := embeddedSourcesMap[source.Name]; ok {
			continue
		}

		embeddedSourcesMap[source.Name] = source
		embeddedSources = append(embeddedSources, source)
	}

	return embeddedSources, nil
}

// read artifact as is. artifacts.Read will fail since it needs a lot more that wont be present in user's terminal
func read(ctx context.Context, path string) (*drivers.CatalogEntry, error) {
	artifact, ok := artifacts.Artifacts[fileutil.FullExt(path)]
	if !ok {
		return nil, fmt.Errorf("no artifact found for %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error in reading file %v : %w", path, err)
	}

	catalog, err := artifact.DeSerialise(ctx, path, string(content), false)
	if err != nil {
		return nil, err
	}

	catalog.Path = path
	return catalog, nil
}

type key struct {
	Name            string
	Type            string
	AnonymousAccess bool
}

func source(connector string, src *runtimev1.Source) drivers.Source {
	props := src.Properties.AsMap()
	switch connector {
	case "s3":
		return &drivers.BucketSource{
			Properties: props,
		}
	case "gcs":
		return &drivers.BucketSource{
			Properties: props,
		}
	case "https":
		return &drivers.FileSource{
			Properties: props,
		}
	case "local_file":
		return &drivers.FileSource{
			Properties: props,
		}
	case "motherduck":
		return &drivers.DatabaseSource{}
	case "bigquery":
		return &drivers.DatabaseSource{
			Props: props,
		}
	default:
		return nil
	}
}
