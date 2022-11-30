package artifacts_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/sql"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSourceReadWrite(t *testing.T) {
	catalogs := []struct {
		// Adding explicit name and using it in the title,
		// adds the run button on goland for each test case.
		Name    string
		Catalog *drivers.CatalogEntry
		Raw     string
	}{
		{
			"Source",
			&drivers.CatalogEntry{
				Name: "Source",
				Path: "sources/Source.yaml",
				Type: drivers.ObjectTypeSource,
				Object: &runtimev1.Source{
					Name:      "Source",
					Connector: "file",
					Properties: toProtoStruct(map[string]any{
						"path": "data/source.csv",
					}),
				},
			},
			`version: 0.0.1
type: file
path: data/source.csv
`,
		},
		{
			"S3Source",
			&drivers.CatalogEntry{
				Name: "S3Source",
				Path: "sources/S3Source.yaml",
				Type: drivers.ObjectTypeSource,
				Object: &runtimev1.Source{
					Name:      "S3Source",
					Connector: "s3",
					Properties: toProtoStruct(map[string]any{
						"path":       "s3://bucket/path/file.csv",
						"aws.region": "us-east-2",
					}),
				},
			},
			`version: 0.0.1
type: s3
uri: s3://bucket/path/file.csv
region: us-east-2
`,
		},
		{
			"Model",
			&drivers.CatalogEntry{
				Name: "Model",
				Path: "models/Model.sql",
				Type: drivers.ObjectTypeModel,
				Object: &runtimev1.Model{
					Name:    "Model",
					Sql:     "select * from A",
					Dialect: runtimev1.Model_DIALECT_DUCKDB,
				},
			},
			"select * from A",
		},
		{
			"MetricsView",
			&drivers.CatalogEntry{
				Name: "MetricsView",
				Path: "dashboards/MetricsView.yaml",
				Type: drivers.ObjectTypeMetricsView,
				Object: &runtimev1.MetricsView{
					Name:          "MetricsView",
					From:          "Model",
					TimeDimension: "time",
					TimeGrains:    []string{"1 day", "1 month"},
					Dimensions: []*runtimev1.MetricsView_Dimension{
						{
							Name:        "dim0",
							Label:       "Dim0_L",
							Description: "Dim0_D",
						},
						{
							Name:        "dim1",
							Label:       "Dim1_L",
							Description: "Dim1_D",
							Ignore:      true,
						},
					},
					Measures: []*runtimev1.MetricsView_Measure{
						{
							Name:        "measure_0",
							Label:       "Mea0_L",
							Expression:  "count(c0)",
							Description: "Mea0_D",
							Format:      "humanise",
							Ignore:      false,
						},
						{
							Name:        "measure_1",
							Label:       "Mea1_L",
							Expression:  "avg(c1)",
							Description: "Mea1_D",
							Format:      "humanise",
							Ignore:      true,
						},
					},
					DisplayName: "dashboard name",
					Description: "long description for dashboard",
				},
			},
			`version: 0.0.1
display_name: dashboard name
description: long description for dashboard
from: Model
timeseries: time
timegrains:
- 1 day
- 1 month
default_timegrain: ""
dimensions:
- label: Dim0_L
  property: dim0
  description: Dim0_D
- label: Dim1_L
  property: dim1
  description: Dim1_D
  ignore: true
measures:
- label: Mea0_L
  expression: count(c0)
  description: Mea0_D
  format_preset: humanise
- label: Mea1_L
  expression: avg(c1)
  description: Mea1_D
  format_preset: humanise
  ignore: true
`,
		},
	}

	dir := t.TempDir()
	fileStore, err := drivers.Open("file", dir)
	require.NoError(t, err)
	repoStore, _ := fileStore.RepoStore()
	ctx := context.Background()

	for _, tt := range catalogs {
		t.Run(fmt.Sprintf("%s", tt.Name), func(t *testing.T) {
			err := artifacts.Write(ctx, repoStore, "test", tt.Catalog)
			require.NoError(t, err)

			readCatalog, err := artifacts.Read(ctx, repoStore, "test", tt.Catalog.Path)
			require.NoError(t, err)
			require.Equal(t, readCatalog, tt.Catalog)

			b, err := os.ReadFile(path.Join(dir, tt.Catalog.Path))
			require.NoError(t, err)
			require.Equal(t, tt.Raw, string(b))
		})
	}
}

func TestReadFailure(t *testing.T) {
	files := []struct {
		Name string
		Path string
		Raw  string
	}{
		{
			"InvalidSource",
			"sources/InvalidSource.yaml",
			`version: 0.0.1
type: file
  uri: data/source.csv
`,
		},
	}

	dir := t.TempDir()
	fileStore, err := drivers.Open("file", dir)
	require.NoError(t, err)
	repoStore, _ := fileStore.RepoStore()
	ctx := context.Background()

	err = os.MkdirAll(path.Join(dir, "sources"), os.ModePerm)
	require.NoError(t, err)
	err = os.MkdirAll(path.Join(dir, "models"), os.ModePerm)
	require.NoError(t, err)
	err = os.MkdirAll(path.Join(dir, "dashboards"), os.ModePerm)
	require.NoError(t, err)

	for _, tt := range files {
		t.Run(tt.Name, func(t *testing.T) {
			err := os.WriteFile(path.Join(dir, tt.Path), []byte(tt.Raw), os.ModePerm)
			require.NoError(t, err)

			_, err = artifacts.Read(ctx, repoStore, "test", tt.Path)
			require.Error(t, err)
		})
	}
}

func toProtoStruct(obj map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(obj)
	if err != nil {
		panic(err)
	}
	return s
}
