package artifacts_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/sql"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
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
					Connector: "local_file",
					Properties: toProtoStruct(map[string]any{
						"path":   "data/source.csv",
						"format": "csv",
						"duckdb": map[string]any{"delim": "'|'"},
					}),
				},
			},
			`type: local_file
path: data/source.csv
format: csv
duckdb:
    delim: '''|'''
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
						"path":   "s3://bucket/path/file.csv",
						"region": "us-east-2",
					}),
				},
			},
			`type: s3
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
					Name:              "MetricsView",
					Model:             "Model",
					TimeDimension:     "time",
					SmallestTimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
					DefaultTimeRange:  "P1D",
					Dimensions: []*runtimev1.MetricsView_Dimension{
						{
							Name:        "dim0",
							Column:      "prop0",
							Label:       "Dim0_L",
							Description: "Dim0_D",
						},
						{
							Name:        "dim1",
							Column:      "prop1",
							Label:       "Dim1_L",
							Description: "Dim1_D",
						},
					},
					Measures: []*runtimev1.MetricsView_Measure{
						{
							Name:                "measure_0",
							Label:               "Mea0_L",
							Expression:          "count(c0)",
							Description:         "Mea0_D",
							Format:              "humanise",
							ValidPercentOfTotal: true,
						},
						{
							Name:        "avg_measure",
							Label:       "Mea1_L",
							Expression:  "avg(c1)",
							Description: "Mea1_D",
							Format:      "humanise",
						},
					},
					Label:       "dashboard name",
					Description: "long description for dashboard",
				},
			},
			`title: dashboard name
description: long description for dashboard
model: Model
timeseries: time
smallest_time_grain: day
default_time_range: P1D
dimensions:
    - name: dim0
      label: Dim0_L
      column: prop0
      description: Dim0_D
    - name: dim1
      label: Dim1_L
      column: prop1
      description: Dim1_D
measures:
    - label: Mea0_L
      name: measure_0
      expression: count(c0)
      description: Mea0_D
      format_preset: humanise
      valid_percent_of_total: true
    - label: Mea1_L
      name: avg_measure
      expression: avg(c1)
      description: Mea1_D
      format_preset: humanise
`,
		},
	}

	dir := t.TempDir()
	fileStore, err := drivers.Open("file", map[string]any{"dsn": dir}, false, zap.NewNop())
	require.NoError(t, err)
	repoStore, _ := fileStore.AsRepoStore("")
	ctx := context.Background()

	for _, tt := range catalogs {
		t.Run(fmt.Sprint(tt.Name), func(t *testing.T) {
			err := artifacts.Write(ctx, repoStore, "test", tt.Catalog)
			require.NoError(t, err)

			readCatalog, err := artifacts.Read(ctx, repoStore, registryStore(t), "test", tt.Catalog.Path)
			require.NoError(t, err)
			require.Equal(t, readCatalog, tt.Catalog)

			b, err := os.ReadFile(path.Join(dir, tt.Catalog.Path))
			require.NoError(t, err)
			require.Equal(t, tt.Raw, string(b))
		})
	}
}

func TestCsvDelimiterBackwardCompatibility(t *testing.T) {
	catalog := &drivers.CatalogEntry{
		Name: "Source",
		Path: "sources/Source.yaml",
		Type: drivers.ObjectTypeSource,
		Object: &runtimev1.Source{
			Name:      "Source",
			Connector: "local_file",
			Properties: toProtoStruct(map[string]any{
				"path":          "data/source.csv",
				"format":        "csv",
				"csv.delimiter": "|",
			}),
		},
	}
	raw := `type: local_file
path: data/source.csv
format: csv
duckdb:
    delim: '''|'''
`
	dir := t.TempDir()
	fileStore, err := drivers.Open("file", map[string]any{"dsn": dir}, false, zap.NewNop())
	require.NoError(t, err)
	repoStore, _ := fileStore.AsRepoStore("")
	ctx := context.Background()

	err = artifacts.Write(ctx, repoStore, "test", catalog)
	require.NoError(t, err)

	readCatalog, err := artifacts.Read(ctx, repoStore, registryStore(t), "test", catalog.Path)
	require.Equal(t, readCatalog, &drivers.CatalogEntry{
		Name: "Source",
		Path: "sources/Source.yaml",
		Type: drivers.ObjectTypeSource,
		Object: &runtimev1.Source{
			Name:      "Source",
			Connector: "local_file",
			Properties: toProtoStruct(map[string]any{
				"path":   "data/source.csv",
				"format": "csv",
				"duckdb": map[string]any{"delim": "'|'"},
			}),
		},
	})
	require.NoError(t, err)

	err = artifacts.Write(ctx, repoStore, "test", readCatalog)
	require.NoError(t, err)

	b, err := os.ReadFile(path.Join(dir, readCatalog.Path))
	require.NoError(t, err)
	require.Equal(t, raw, string(b))
}

func TestHivePartitioningBackwardCompatibility(t *testing.T) {
	catalog := &drivers.CatalogEntry{
		Name: "Source",
		Path: "sources/Source.yaml",
		Type: drivers.ObjectTypeSource,
		Object: &runtimev1.Source{
			Name:      "Source",
			Connector: "local_file",
			Properties: toProtoStruct(map[string]any{
				"path":              "data/source.csv",
				"format":            "csv",
				"hive_partitioning": true,
			}),
		},
	}
	raw := `type: local_file
path: data/source.csv
format: csv
duckdb:
    hive_partitioning: true
`
	dir := t.TempDir()
	fileStore, err := drivers.Open("file", map[string]any{"dsn": dir}, false, zap.NewNop())
	require.NoError(t, err)
	repoStore, _ := fileStore.AsRepoStore("")
	ctx := context.Background()

	err = artifacts.Write(ctx, repoStore, "test", catalog)
	require.NoError(t, err)

	readCatalog, err := artifacts.Read(ctx, repoStore, registryStore(t), "test", catalog.Path)
	require.Equal(t, readCatalog, &drivers.CatalogEntry{
		Name: "Source",
		Path: "sources/Source.yaml",
		Type: drivers.ObjectTypeSource,
		Object: &runtimev1.Source{
			Name:      "Source",
			Connector: "local_file",
			Properties: toProtoStruct(map[string]any{
				"path":   "data/source.csv",
				"format": "csv",
				"duckdb": map[string]any{"hive_partitioning": true},
			}),
		},
	})
	require.NoError(t, err)

	err = artifacts.Write(ctx, repoStore, "test", readCatalog)
	require.NoError(t, err)

	b, err := os.ReadFile(path.Join(dir, readCatalog.Path))
	require.NoError(t, err)
	require.Equal(t, raw, string(b))
}

func TestMetricsLabelBackwardsCompatibility(t *testing.T) {
	dir := t.TempDir()
	fileStore, err := drivers.Open("file", map[string]any{"dsn": dir}, false, zap.NewNop())
	require.NoError(t, err)
	repoStore, _ := fileStore.AsRepoStore("")
	ctx := context.Background()

	require.NoError(t, repoStore.Put(ctx, "dashboards/MetricsView.yaml", bytes.NewReader([]byte(`display_name: dashboard name
description: long description for dashboard
model: Model
timeseries: time
smallest_time_grain: day
default_time_range: P1D
dimensions: []
measures: []
`))))

	readCatalog, err := artifacts.Read(ctx, repoStore, registryStore(t), "test", "dashboards/MetricsView.yaml")
	require.NoError(t, err)
	require.Equal(t, &drivers.CatalogEntry{
		Name: "MetricsView",
		Path: "dashboards/MetricsView.yaml",
		Type: drivers.ObjectTypeMetricsView,
		Object: &runtimev1.MetricsView{
			Name:              "MetricsView",
			Model:             "Model",
			TimeDimension:     "time",
			SmallestTimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
			DefaultTimeRange:  "P1D",
			Label:             "dashboard name",
			Description:       "long description for dashboard",
		},
	}, readCatalog)
}

func TestDimensionAndMeasureNameAutoFill(t *testing.T) {
	dir := t.TempDir()
	fileStore, err := drivers.Open("file", map[string]any{"dsn": dir}, false, zap.NewNop())
	require.NoError(t, err)
	repoStore, _ := fileStore.AsRepoStore("")
	ctx := context.Background()

	require.NoError(t, repoStore.Put(ctx, "dashboards/MetricsView.yaml", bytes.NewReader([]byte(`display_name: dashboard name
description: long description for dashboard
model: Model
timeseries: time
smallest_time_grain: day
default_time_range: P1D
dimensions:
    - label: Dim0_L
      column: prop0
      description: Dim0_D
    - name: dim1
      label: Dim1_L
      description: Dim1_D
measures:
    - label: Mea0_L
      name: measure_imp
      expression: count(c0)
      description: Mea0_D
      format_preset: humanise
    - label: Mea1_L
      expression: avg(c1)
      description: Mea1_D
      format_preset: humanise
`))))

	readCatalog, err := artifacts.Read(ctx, repoStore, registryStore(t), "test", "dashboards/MetricsView.yaml")
	require.NoError(t, err)
	require.Equal(t, &drivers.CatalogEntry{
		Name: "MetricsView",
		Path: "dashboards/MetricsView.yaml",
		Type: drivers.ObjectTypeMetricsView,
		Object: &runtimev1.MetricsView{
			Name:              "MetricsView",
			Model:             "Model",
			TimeDimension:     "time",
			SmallestTimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
			DefaultTimeRange:  "P1D",
			Label:             "dashboard name",
			Description:       "long description for dashboard",
			Dimensions: []*runtimev1.MetricsView_Dimension{
				{
					Name:        "prop0",
					Column:      "prop0",
					Label:       "Dim0_L",
					Description: "Dim0_D",
				},
				{
					Name:        "dim1",
					Label:       "Dim1_L",
					Description: "Dim1_D",
				},
			},
			Measures: []*runtimev1.MetricsView_Measure{
				{
					Name:        "measure_imp",
					Label:       "Mea0_L",
					Expression:  "count(c0)",
					Description: "Mea0_D",
					Format:      "humanise",
				},
				{
					Name:        "measure_1",
					Label:       "Mea1_L",
					Expression:  "avg(c1)",
					Description: "Mea1_D",
					Format:      "humanise",
				},
			},
		},
	}, readCatalog)
}

func TestDimensionColumnBackwardsCompatibility(t *testing.T) {
	dir := t.TempDir()
	fileStore, err := drivers.Open("file", map[string]any{"dsn": dir}, false, zap.NewNop())
	require.NoError(t, err)
	repoStore, _ := fileStore.AsRepoStore("")
	ctx := context.Background()

	require.NoError(t, repoStore.Put(ctx, "dashboards/MetricsView.yaml", bytes.NewReader([]byte(`display_name: dashboard name
description: long description for dashboard
model: Model
timeseries: time
smallest_time_grain: day
default_time_range: P1D
dimensions:
    - label: Dim0_L
      property: prop0
    - name: dim1
      property: prop1
      column: col1
      label: Dim1_L
    - label: Dim2
      column: col2
measures: []
`))))

	readCatalog, err := artifacts.Read(ctx, repoStore, registryStore(t), "test", "dashboards/MetricsView.yaml")
	require.NoError(t, err)
	require.Equal(t, &drivers.CatalogEntry{
		Name: "MetricsView",
		Path: "dashboards/MetricsView.yaml",
		Type: drivers.ObjectTypeMetricsView,
		Object: &runtimev1.MetricsView{
			Name:              "MetricsView",
			Model:             "Model",
			TimeDimension:     "time",
			SmallestTimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
			DefaultTimeRange:  "P1D",
			Label:             "dashboard name",
			Description:       "long description for dashboard",
			Dimensions: []*runtimev1.MetricsView_Dimension{
				{
					Name:   "prop0",
					Column: "prop0",
					Label:  "Dim0_L",
				},
				{
					Name:   "dim1",
					Column: "col1",
					Label:  "Dim1_L",
				},
				{
					Name:   "col2",
					Column: "col2",
					Label:  "Dim2",
				},
			},
		},
	}, readCatalog)
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
			`type: local_file
  uri: data/source.csv
`,
		},
	}

	dir := t.TempDir()
	fileStore, err := drivers.Open("file", map[string]any{"dsn": dir}, false, zap.NewNop())
	require.NoError(t, err)
	repoStore, _ := fileStore.AsRepoStore("")
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

			_, err = artifacts.Read(ctx, repoStore, registryStore(t), "test", tt.Path)
			require.Error(t, err)
		})
	}
}

func TestSanitizedName(t *testing.T) {
	variations := []struct {
		fileName     string
		expectedName string
	}{
		{"table", "table"},
		{"table.parquet", "table"},
		{"table.v1.parquet", "table"},
		{"table.parquet.tgz", "table"},
		{"22-02-10.parquet", "_22_02_10"},
		{"-22-02-11.parquet", "_22_02_11"},
		{"_22-02-12.parquet", "_22_02_12"},
	}

	for _, variation := range variations {
		filePathVariations := []struct {
			filePath     string
			expectedName string
		}{
			{variation.fileName, variation.expectedName},
			{"/" + variation.fileName, variation.expectedName},
			{"./" + variation.fileName, variation.expectedName},
			{"path/to/file/" + variation.fileName, variation.expectedName},
			{"/path/to/file/" + variation.fileName, variation.expectedName},
			{"./path/to/file/" + variation.fileName, variation.expectedName},
		}

		for _, filePathVariation := range filePathVariations {
			t.Run(filePathVariation.filePath, func(t *testing.T) {
				require.Equal(t, filePathVariation.expectedName, artifacts.SanitizedName(filePathVariation.filePath))
			})
		}
	}
}

func TestReadWithEnvVariables(t *testing.T) {
	repoStore := repoStore(t)
	registryStore := registryStore(t)
	tests := []struct {
		name     string
		filePath string
		content  string
		want     *drivers.CatalogEntry
		wantErr  bool
	}{
		{
			name:     "valid source yaml",
			filePath: "sources/Source.yaml",
			content: `type: s3
uri: "s3://bucket/file"
format: csv
region: {{.env.region}}
duckdb:
    delim: '''{{.env.delimitter}}'''
`,
			want: &drivers.CatalogEntry{
				Name: "Source",
				Path: "sources/Source.yaml",
				Type: drivers.ObjectTypeSource,
				Object: &runtimev1.Source{
					Name:      "Source",
					Connector: "s3",
					Properties: toProtoStruct(map[string]any{
						"path":   "s3://bucket/file",
						"format": "csv",
						"region": "us-east-2",
						"duckdb": map[string]any{"delim": "'|'"},
					}),
				},
			},
			wantErr: false,
		},
		{
			name:     "invalid source yaml no env defined",
			filePath: "sources/Source.yaml",
			content: `type: s3
uri: "s3://bucket/file"
csv.delimiter: {{.env.delimitter}}
format: {{.env.format}}
region: {{.env.region}}
`,
			want:    nil,
			wantErr: true,
		},
		{
			name:     "Model",
			filePath: "models/Model.sql",
			content:  "select * from {{ \"foo\" | upper }} {{.env.limit}}",
			want: &drivers.CatalogEntry{
				Name: "Model",
				Path: "models/Model.sql",
				Type: drivers.ObjectTypeModel,
				Object: &runtimev1.Model{
					Name:    "Model",
					Sql:     "select * from FOO limit 10",
					Dialect: runtimev1.Model_DIALECT_DUCKDB,
				},
			},
			wantErr: false,
		},
		{
			name:     "Model",
			filePath: "models/Model.sql",
			content:  "select {{ env \"SECRET\" }}",
			want:     nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, repoStore.Put(context.Background(), tt.filePath, bytes.NewReader([]byte(tt.content))))
			got, err := artifacts.Read(context.Background(), repoStore, registryStore, "test", tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricsViewAvailableTimeZones(t *testing.T) {
	repoStore := repoStore(t)
	registryStore := registryStore(t)
	tests := []struct {
		name     string
		filePath string
		content  string
		want     *drivers.CatalogEntry
		wantErr  bool
	}{
		{
			name:     "valid time zones",
			filePath: "dashboards/dashboard.yaml",
			content: `
available_time_zones:
- UTC
- Europe/Copenhagen
`,
			want: &drivers.CatalogEntry{
				Name: "dashboard",
				Path: "dashboards/dashboard.yaml",
				Type: drivers.ObjectTypeMetricsView,
				Object: &runtimev1.MetricsView{
					Name:               "dashboard",
					AvailableTimeZones: []string{"UTC", "Europe/Copenhagen"},
				},
			},
			wantErr: false,
		},
		{
			name:     "invalid time zones",
			filePath: "dashboards/dashboard.yaml",
			content: `
available_time_zones:
- foo
`,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, repoStore.Put(context.Background(), tt.filePath, bytes.NewReader([]byte(tt.content))))
			got, err := artifacts.Read(context.Background(), repoStore, registryStore, "test", tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func repoStore(t *testing.T) drivers.RepoStore {
	dir := t.TempDir()
	fileStore, err := drivers.Open("file", map[string]any{"dsn": dir}, false, zap.NewNop())
	require.NoError(t, err)

	repoStore, ok := fileStore.AsRepoStore("")
	require.True(t, ok)

	return repoStore
}

func toProtoStruct(obj map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(obj)
	if err != nil {
		panic(err)
	}
	return s
}

func registryStore(t *testing.T) drivers.RegistryStore {
	ctx := context.Background()
	store, err := drivers.Open("sqlite", map[string]any{"dsn": ":memory:"}, false, zap.NewNop())
	require.NoError(t, err)
	store.Migrate(ctx)
	registry, _ := store.AsRegistry()

	env := map[string]string{"delimitter": "|", "region": "us-east-2", "limit": "limit 10"}
	err = registry.CreateInstance(ctx, &drivers.Instance{ID: "test", Variables: env})
	require.NoError(t, err)

	return registry
}
