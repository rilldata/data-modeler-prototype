package rillv1

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	_ "github.com/rilldata/rill/runtime/drivers/file"
)

func TestRillYAML(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: `
display_name: Hello world
description: This project says hello to the world

connectors:
- name: my-s3
  type: s3
  defaults:
    region: us-east-1

env:
  foo: bar

mock_users:
- email: foo@bar.com
  custom_attribute: yeah
`,
	})

	res, err := ParseRillYAML(ctx, repo, "")
	require.NoError(t, err)

	require.Equal(t, res.DisplayName, "Hello world")
	require.Equal(t, res.Description, "This project says hello to the world")

	require.Len(t, res.Connectors, 1)
	require.Equal(t, "my-s3", res.Connectors[0].Name)
	require.Equal(t, "s3", res.Connectors[0].Type)
	require.Len(t, res.Connectors[0].Defaults, 1)
	require.Equal(t, "us-east-1", res.Connectors[0].Defaults["region"])

	require.Len(t, res.Variables, 1)
	require.Equal(t, "foo", res.Variables[0].Name)
	require.Equal(t, "bar", res.Variables[0].Default)
}

func TestRillYAMLFeatures(t *testing.T) {
	tt := []struct {
		yaml    string
		want    map[string]bool
		wantErr bool
	}{
		{
			yaml: ` `,
			want: nil,
		},
		{
			yaml:    `features: 10`,
			wantErr: true,
		},
		{
			yaml: `features: []`,
			want: map[string]bool{},
		},
		{
			yaml: `features: {}`,
			want: map[string]bool{},
		},
		{
			yaml: `
features:
  foo: true
  bar: false
`,
			want: map[string]bool{"foo": true, "bar": false},
		},
		{
			yaml: `
features:
- foo
- bar
`,
			want: map[string]bool{"foo": true, "bar": true},
		},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("case=%d", i), func(t *testing.T) {
			ctx := context.Background()
			repo := makeRepo(t, map[string]string{
				`rill.yaml`: tc.yaml,
			})

			res, err := ParseRillYAML(ctx, repo, "")
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, res.FeatureFlags)
		})
	}
}

func TestComplete(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// init.sql
		`init.sql`: `
{{ configure "max_version" 2 }}
INSTALL 'hello';
`,
		// source s1
		`sources/s1.yaml`: `
connector: s3
path: hello
`,
		// source s2
		`sources/s2.sql`: `
-- @connector: postgres
-- @refresh.cron: 0 0 * * *
SELECT 1
`,
		// model m1
		`models/m1.sql`: `
SELECT 1
`,
		// model m2
		`models/m2.sql`: `
SELECT * FROM m1
`,
		`models/m2.yaml`: `
materialize: true
`,
		// dashboard d1
		`metrics/d1.yaml`: `
version: 1
type: metrics_view
model: m2
dimensions:
  - name: a
    column: a
measures:
  - name: b
    expression: count(*)
    format_d3: "0,0"
    format_d3_locale:
        currency: ["£", ""]
first_day_of_week: 7
first_month_of_year: 3
`,
		// explore e1
		`explores/e1.yaml`: `
type: explore
display_name: E1
metrics_view: d1
measures:
  - b
time_ranges:
  - P2W
  - range: P4W
  - range: P2M
    comparison_offsets:
      - P1M
      - offset: P4M
        range: P2M
defaults:
  time_range: P4W
`,
		// migration c1
		`custom/c1.yml`: `
type: migration
max_version: 3
sql: |
  CREATE TABLE a(a integer);
`,
		// model c2
		`custom/c2.sql`: `
{{ configure "type" "model" }}
{{ configure "materialize" true }}
SELECT * FROM {{ ref "m2" }}
`,
		// connector s3
		`connectors/s3.yaml`: `
type: connector
region: us-east-1
driver: s3
dev:
  bucket: "my-bucket-dev"
prod:
  bucket: "my-bucket-prod"
`,
		// connector postgres
		`connectors/postgres.yaml`: `
driver: postgres
database: postgres
schema: default
`,
	}

	timeRange := "P4W"
	resources := []*Resource{
		// init.sql
		{
			Name:  ResourceName{Kind: ResourceKindMigration, Name: "init"},
			Paths: []string{"/init.sql"},
			MigrationSpec: &runtimev1.MigrationSpec{
				Connector: "duckdb",
				Version:   2,
				Sql:       strings.TrimSpace(files["init.sql"]),
			},
		},
		// source s1
		{
			Name:  ResourceName{Kind: ResourceKindSource, Name: "s1"},
			Paths: []string{"/sources/s1.yaml"},
			SourceSpec: &runtimev1.SourceSpec{
				SourceConnector: "s3",
				SinkConnector:   "duckdb",
				Properties:      must(structpb.NewStruct(map[string]any{"path": "hello"})),
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
		// source s2
		{
			Name:  ResourceName{Kind: ResourceKindSource, Name: "s2"},
			Paths: []string{"/sources/s2.sql"},
			SourceSpec: &runtimev1.SourceSpec{
				SourceConnector: "postgres",
				SinkConnector:   "duckdb",
				Properties:      must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["sources/s2.sql"])})),
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true, Cron: "0 0 * * *"},
			},
		},
		// model m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m1.sql"])})),
				OutputConnector: "duckdb",
			},
		},
		// model m2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}},
			Paths: []string{"/models/m2.yaml", "/models/m2.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule:  &runtimev1.Schedule{RefUpdate: true},
				InputConnector:   "duckdb",
				InputProperties:  must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m2.sql"])})),
				OutputConnector:  "duckdb",
				OutputProperties: must(structpb.NewStruct(map[string]any{"materialize": true})),
			},
		},
		// dashboard d1
		{
			Name:  ResourceName{Kind: ResourceKindMetricsView, Name: "d1"},
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m2"}},
			Paths: []string{"/metrics/d1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector:   "duckdb",
				Model:       "m2",
				DisplayName: "D1",
				Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
					{Name: "a", DisplayName: "A", Column: "a"},
				},
				Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
					{
						Name:           "b",
						DisplayName:    "B",
						Expression:     "count(*)",
						Type:           runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
						FormatD3:       "0,0",
						FormatD3Locale: must(structpb.NewStruct(map[string]any{"currency": []any{"£", ""}})),
					},
				},
				FirstDayOfWeek:   7,
				FirstMonthOfYear: 3,
			},
		},
		// explore e1
		{
			Name:  ResourceName{Kind: ResourceKindExplore, Name: "e1"},
			Refs:  []ResourceName{{Kind: ResourceKindMetricsView, Name: "d1"}},
			Paths: []string{"/explores/e1.yaml"},
			ExploreSpec: &runtimev1.ExploreSpec{
				DisplayName:        "E1",
				MetricsView:        "d1",
				DimensionsSelector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
				Measures:           []string{"b"},
				TimeRanges: []*runtimev1.ExploreTimeRange{
					{Range: "P2W"},
					{Range: "P4W"},
					{
						Range: "P2M",
						ComparisonTimeRanges: []*runtimev1.ExploreComparisonTimeRange{
							{Offset: "P1M"},
							{Offset: "P4M", Range: "P2M"},
						},
					},
				},
				DefaultPreset: &runtimev1.ExplorePreset{
					DimensionsSelector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
					MeasuresSelector:   &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
					TimeRange:          &timeRange,
					ComparisonMode:     runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_NONE,
				},
			},
		},
		// migration c1
		{
			Name:  ResourceName{Kind: ResourceKindMigration, Name: "c1"},
			Paths: []string{"/custom/c1.yml"},
			MigrationSpec: &runtimev1.MigrationSpec{
				Connector: "duckdb",
				Version:   3,
				Sql:       "CREATE TABLE a(a integer);",
			},
		},
		// model c2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "c2"},
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m2"}},
			Paths: []string{"/custom/c2.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{
					"sql": strings.TrimSpace(files["custom/c2.sql"]),
				})),
				OutputConnector:  "duckdb",
				OutputProperties: must(structpb.NewStruct(map[string]any{"materialize": true})),
			},
		},
		// postgres
		{
			Name:  ResourceName{Kind: ResourceKindConnector, Name: "postgres"},
			Paths: []string{"/connectors/postgres.yaml"},
			ConnectorSpec: &runtimev1.ConnectorSpec{
				Driver: "postgres",
				Properties: map[string]string{
					"database": "postgres",
					"schema":   "default",
				},
			},
		},
		// s3
		{
			Name:  ResourceName{Kind: ResourceKindConnector, Name: "s3"},
			Paths: []string{"/connectors/s3.yaml"},
			ConnectorSpec: &runtimev1.ConnectorSpec{
				Driver: "s3",
				Properties: map[string]string{
					"region": "us-east-1",
				},
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestLocationError(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// source s1
		`sources/s1.yaml`: `
connector: s3
path: hello
  world: foo
`,
		// model m1
		`/models/m1.sql`: `
-- @materialize: true
SELECT *

FRO m1
`,
	}

	errors := []*runtimev1.ParseError{
		{
			Message:       " mapping values are not allowed in this context",
			FilePath:      "/sources/s1.yaml",
			StartLocation: &runtimev1.CharLocation{Line: 4},
		},
		{
			Message:       "syntax error at or near \"m1\"",
			FilePath:      "/models/m1.sql",
			StartLocation: &runtimev1.CharLocation{Line: 5},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, nil, errors)
}

func TestUniqueSourceModelName(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// source s1
		`sources/s1.yaml`: `
connector: s3
`,
		// model s1
		`/models/s1.sql`: `
SELECT 1
`,
	}

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindSource, Name: "s1"},
			Paths: []string{"/sources/s1.yaml"},
			SourceSpec: &runtimev1.SourceSpec{
				SourceConnector: "s3",
				SinkConnector:   "duckdb",
				Properties:      must(structpb.NewStruct(map[string]any{})),
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
	}

	errors := []*runtimev1.ParseError{
		{
			Message:  "model name collides with source \"s1\"",
			FilePath: "/models/s1.sql",
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, errors)
}

func TestReparse(t *testing.T) {
	// Prepare
	ctx := context.Background()

	// Create empty project
	repo := makeRepo(t, map[string]string{`rill.yaml`: ``})
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, nil, nil)

	// Add a source
	putRepo(t, repo, map[string]string{
		`sources/s1.yaml`: `
connector: s3
path: hello
`,
	})
	s1 := &Resource{
		Name:  ResourceName{Kind: ResourceKindSource, Name: "s1"},
		Paths: []string{"/sources/s1.yaml"},
		SourceSpec: &runtimev1.SourceSpec{
			SourceConnector: "s3",
			SinkConnector:   "duckdb",
			Properties:      must(structpb.NewStruct(map[string]any{"path": "hello"})),
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}
	diff, err := p.Reparse(ctx, s1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1}, nil)
	require.Equal(t, &Diff{
		Added: []ResourceName{s1.Name},
	}, diff)

	// Add a model
	putRepo(t, repo, map[string]string{
		`models/m1.sql`: `
SELECT * FROM foo
`,
	})
	m1 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Paths: []string{"/models/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			InputConnector:  "duckdb",
			InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT * FROM foo"})),
			OutputConnector: "duckdb",
		},
	}
	diff, err = p.Reparse(ctx, m1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1, m1}, nil)
	require.Equal(t, &Diff{
		Added: []ResourceName{m1.Name},
	}, diff)

	// Annotate the model with a YAML file
	putRepo(t, repo, map[string]string{
		`models/m1.yaml`: `
materialize: true
`,
	})
	m1.Paths = []string{"/models/m1.sql", "/models/m1.yaml"}
	m1.ModelSpec.OutputProperties = must(structpb.NewStruct(map[string]any{"materialize": true}))
	diff, err = p.Reparse(ctx, []string{"/models/m1.yaml"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1, m1}, nil)
	require.Equal(t, &Diff{
		Modified: []ResourceName{m1.Name},
	}, diff)

	// Modify the model's SQL
	putRepo(t, repo, map[string]string{
		`models/m1.sql`: `
SELECT * FROM bar
`,
	})
	m1.ModelSpec.InputProperties = must(structpb.NewStruct(map[string]any{"sql": "SELECT * FROM bar"}))
	diff, err = p.Reparse(ctx, []string{"/models/m1.sql"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1, m1}, nil)
	require.Equal(t, &Diff{
		Modified: []ResourceName{m1.Name},
	}, diff)

	// Rename the model to collide with the source
	putRepo(t, repo, map[string]string{
		`models/m1.sql`: `
-- @name: s1
SELECT * FROM bar
`,
	})
	diff, err = p.Reparse(ctx, []string{"/models/m1.sql"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1}, []*runtimev1.ParseError{
		{
			Message:  "model name collides with source \"s1\"",
			FilePath: "/models/m1.sql",
		},
		{
			Message:  "model name collides with source \"s1\"",
			FilePath: "/models/m1.yaml",
		},
	})
	require.Equal(t, &Diff{
		Deleted: []ResourceName{m1.Name},
	}, diff)

	// Put m1 back and add a syntax error in the source
	putRepo(t, repo, map[string]string{
		`models/m1.sql`: `
SELECT * FROM bar
`,
		`sources/s1.yaml`: `
connector: s3
path: hello
  world: path
`,
	})
	diff, err = p.Reparse(ctx, []string{"/models/m1.sql", "/sources/s1.yaml"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1}, []*runtimev1.ParseError{{
		Message:       "mapping values are not allowed in this context", // note: approximate string match
		FilePath:      "/sources/s1.yaml",
		StartLocation: &runtimev1.CharLocation{Line: 4},
	}})
	require.Equal(t, &Diff{
		Added:   []ResourceName{m1.Name},
		Deleted: []ResourceName{s1.Name},
	}, diff)

	// Delete the source
	deleteRepo(t, repo, s1.Paths[0])
	diff, err = p.Reparse(ctx, s1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1}, nil)
	require.Equal(t, &Diff{}, diff)
}

func TestReparseSourceModelCollision(t *testing.T) {
	// Create project with model m1
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`models/m1.sql`: `
SELECT 10
		`,
	})
	m1 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Paths: []string{"/models/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			InputConnector:  "duckdb",
			InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT 10"})),
			OutputConnector: "duckdb",
		},
	}
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1}, nil)

	// Add colliding source m1
	putRepo(t, repo, map[string]string{
		`sources/m1.yaml`: `
connector: s3
path: hello
`,
	})
	s1 := &Resource{
		Name:  ResourceName{Kind: ResourceKindSource, Name: "m1"},
		Paths: []string{"/sources/m1.yaml"},
		SourceSpec: &runtimev1.SourceSpec{
			SourceConnector: "s3",
			SinkConnector:   "duckdb",
			Properties:      must(structpb.NewStruct(map[string]any{"path": "hello"})),
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}
	diff, err := p.Reparse(ctx, s1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1}, []*runtimev1.ParseError{
		{
			Message:  "model name collides with source \"m1\"",
			FilePath: "/models/m1.sql",
		},
	})
	require.Equal(t, &Diff{
		Added:   []ResourceName{s1.Name},
		Deleted: []ResourceName{m1.Name},
	}, diff)

	// Remove colliding source, verify model is restored
	deleteRepo(t, repo, "/sources/m1.yaml")
	diff, err = p.Reparse(ctx, s1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1}, nil)
	require.Equal(t, &Diff{
		Added:   []ResourceName{m1.Name},
		Deleted: []ResourceName{s1.Name},
	}, diff)
}

func TestReparseNameCollision(t *testing.T) {
	// Create project with model m1
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`models/m1.sql`: `
SELECT 10
		`,
		`models/nested/m1.sql`: `
SELECT 20
		`,
		`models/m2.sql`: `
SELECT * FROM m1
		`,
	})
	m1 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Paths: []string{"/models/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			InputConnector:  "duckdb",
			InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT 10"})),
			OutputConnector: "duckdb",
		},
	}
	m1Nested := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Paths: []string{"/models/nested/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			InputConnector:  "duckdb",
			InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT 20"})),
			OutputConnector: "duckdb",
		},
	}
	m2 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
		Paths: []string{"/models/m2.sql"},
		Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}},
		ModelSpec: &runtimev1.ModelSpec{
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			InputConnector:  "duckdb",
			InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT * FROM m1"})),
			OutputConnector: "duckdb",
		},
	}
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1, m2}, []*runtimev1.ParseError{
		{
			Message:  "name collision",
			FilePath: "/models/nested/m1.sql",
			External: true,
		},
	})

	// Remove colliding model, verify things still work
	deleteRepo(t, repo, "/models/m1.sql")
	diff, err := p.Reparse(ctx, m1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1Nested, m2}, nil)
	require.Equal(t, &Diff{
		Modified: []ResourceName{m1.Name, m2.Name}, // m2 due to ref re-inference
	}, diff)
}

func TestReparseMultiKindNameCollision(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`:            ``,
		`models/m1.sql`:        `SELECT 10`,
		`models/nested/m1.sql`: `SELECT 20`,
		`sources/m1.yaml`: `
connector: s3
path: hello
`,
	})
	src := &Resource{
		Name:  ResourceName{Kind: ResourceKindSource, Name: "m1"},
		Paths: []string{"/sources/m1.yaml"},
		SourceSpec: &runtimev1.SourceSpec{
			SourceConnector: "s3",
			SinkConnector:   "duckdb",
			Properties:      must(structpb.NewStruct(map[string]any{"path": "hello"})),
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}
	mdl := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Paths: []string{"/models/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			InputConnector:  "duckdb",
			InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT 10"})),
			OutputConnector: "duckdb",
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{src}, []*runtimev1.ParseError{
		{
			Message:  "collides with source",
			FilePath: "/models/m1.sql",
			External: true,
		},
		{
			Message:  "name collision",
			FilePath: "/models/nested/m1.sql",
			External: true,
		},
	})

	// Delete source m1
	deleteRepo(t, repo, "/sources/m1.yaml")
	diff, err := p.Reparse(ctx, src.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{mdl}, []*runtimev1.ParseError{
		{
			Message:  "name collision",
			FilePath: "/models/nested/m1.sql",
			External: true,
		},
	})
	require.Equal(t, &Diff{
		Added:   []ResourceName{mdl.Name},
		Deleted: []ResourceName{src.Name},
	}, diff)
}

func TestReparseRillYAML(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{})

	mdl := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Paths: []string{"/models/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			InputConnector:  "duckdb",
			InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT 10"})),
			OutputConnector: "duckdb",
		},
	}
	perr := &runtimev1.ParseError{
		Message:  "rill.yaml not found",
		FilePath: "/rill.yaml",
	}

	// Parse empty project. Expect rill.yaml error.
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	require.Nil(t, p.RillYAML)
	requireResourcesAndErrors(t, p, nil, []*runtimev1.ParseError{perr})

	// Add rill.yaml. Expect success.
	putRepo(t, repo, map[string]string{
		`rill.yaml`: ``,
	})
	diff, err := p.Reparse(ctx, []string{"/rill.yaml"})
	require.NoError(t, err)
	require.True(t, diff.Reloaded)
	require.NotNil(t, p.RillYAML)
	requireResourcesAndErrors(t, p, nil, nil)

	// Remove rill.yaml and add a model. Expect reloaded.
	deleteRepo(t, repo, "/rill.yaml")
	putRepo(t, repo, map[string]string{"/models/m1.sql": "SELECT 10"})
	diff, err = p.Reparse(ctx, []string{"/rill.yaml", "/models/m1.sql"})
	require.NoError(t, err)
	require.True(t, diff.Reloaded)
	require.Nil(t, p.RillYAML)
	requireResourcesAndErrors(t, p, []*Resource{mdl}, []*runtimev1.ParseError{perr})

	// Edit model. Expect nothing to happen because rill.yaml is still broken.
	putRepo(t, repo, map[string]string{"/models/m1.sql": "SELECT 20"})
	diff, err = p.Reparse(ctx, []string{"/models/m1.sql"})
	require.NoError(t, err)
	require.Equal(t, &Diff{Skipped: true}, diff)
	require.Nil(t, p.RillYAML)
	requireResourcesAndErrors(t, p, []*Resource{mdl}, []*runtimev1.ParseError{perr})

	// Fix rill.yaml. Expect reloaded.
	mdl.ModelSpec.InputProperties = must(structpb.NewStruct(map[string]any{"sql": "SELECT 20"}))
	putRepo(t, repo, map[string]string{"/rill.yaml": ""})
	diff, err = p.Reparse(ctx, []string{"/rill.yaml"})
	require.NoError(t, err)
	require.True(t, diff.Reloaded)
	require.NotNil(t, p.RillYAML)
	requireResourcesAndErrors(t, p, []*Resource{mdl}, nil)
}

func TestRefInferrence(t *testing.T) {
	// Create model referencing "bar"
	foo := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "foo"},
		Paths: []string{"/models/foo.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			InputConnector:  "duckdb",
			InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT * FROM bar"})),
			OutputConnector: "duckdb",
		},
	}
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// model foo
		`models/foo.sql`: `SELECT * FROM bar`,
	})
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{foo}, nil)

	// Add model "bar"
	foo.Refs = []ResourceName{{Kind: ResourceKindModel, Name: "bar"}}
	bar := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "bar"},
		Paths: []string{"/models/bar.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			InputConnector:  "duckdb",
			InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT * FROM baz"})),
			OutputConnector: "duckdb",
		},
	}
	putRepo(t, repo, map[string]string{
		`models/bar.sql`: `SELECT * FROM baz`,
	})
	diff, err := p.Reparse(ctx, []string{"/models/bar.sql"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{foo, bar}, nil)
	require.Equal(t, &Diff{
		Added:    []ResourceName{bar.Name},
		Modified: []ResourceName{foo.Name},
	}, diff)

	// Remove "bar"
	foo.Refs = nil
	deleteRepo(t, repo, bar.Paths[0])
	diff, err = p.Reparse(ctx, []string{"/models/bar.sql"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{foo}, nil)
	require.Equal(t, &Diff{
		Modified: []ResourceName{foo.Name},
		Deleted:  []ResourceName{bar.Name},
	}, diff)
}

func BenchmarkReparse(b *testing.B) {
	ctx := context.Background()
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// model m1
		`models/m1.sql`: `
SELECT 1
`,
		// model m2
		`models/m2.sql`: `
SELECT * FROM m1
`,
		`models/m2.yaml`: `
materialize: true
`,
	}
	resources := []*Resource{
		// m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m1.sql"])})),
				OutputConnector: "duckdb",
			},
		},
		// m2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}},
			Paths: []string{"/models/m2.sql", "/models/m2.yaml"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule:  &runtimev1.Schedule{RefUpdate: true},
				InputConnector:   "duckdb",
				InputProperties:  must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m2.sql"])})),
				OutputConnector:  "duckdb",
				OutputProperties: must(structpb.NewStruct(map[string]any{"materialize": true})),
			},
		},
	}
	repo := makeRepo(b, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(b, err)
	requireResourcesAndErrors(b, p, resources, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files[`models/m2.sql`] = fmt.Sprintf(`SELECT * FROM m1 LIMIT %d`, i)
		_, err = p.Reparse(ctx, []string{`models/m2.sql`})
		require.NoError(b, err)
		require.Empty(b, p.Errors)
	}
}

func TestProjectModelDefaults(t *testing.T) {
	ctx := context.Background()

	files := map[string]string{
		// Provide dashboard defaults in rill.yaml
		`rill.yaml`: `
models:
  materialize: true
`,
		// Model that inherits defaults
		`models/m1.sql`: `
SELECT * FROM t1
`,
		// Model that overrides defaults
		`models/m2.sql`: `
-- @materialize: false
SELECT * FROM t2
`,
	}

	resources := []*Resource{
		// m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule:  &runtimev1.Schedule{RefUpdate: true},
				InputConnector:   "duckdb",
				InputProperties:  must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m1.sql"])})),
				OutputConnector:  "duckdb",
				OutputProperties: must(structpb.NewStruct(map[string]any{"materialize": true})),
			},
		},
		// m2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
			Paths: []string{"/models/m2.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule:  &runtimev1.Schedule{RefUpdate: true},
				InputConnector:   "duckdb",
				InputProperties:  must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m2.sql"])})),
				OutputConnector:  "duckdb",
				OutputProperties: must(structpb.NewStruct(map[string]any{"materialize": false})),
			},
		},
	}

	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestProjectMetricsViewDefaults(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		// Provide metrics view defaults in rill.yaml
		`rill.yaml`: `
metrics_views:
  first_day_of_week: 7
  security:
    access: true
`,
		// Metrics that inherits defaults
		`mv1.yaml`: `
version: 1
type: metrics_view
table: t1
dimensions:
  - name: a
    column: a
measures:
  - name: b
    expression: count(*)
`,
		// Metrics that overrides defaults
		`mv2.yaml`: `
version: 1
type: metrics_view
table: t2
dimensions:
  - name: a
    column: a
measures:
  - name: b
    expression: count(*)
first_day_of_week: 1
security:
  row_filter: true
`,
	})

	resources := []*Resource{
		// metrics view mv1
		{
			Name:  ResourceName{Kind: ResourceKindMetricsView, Name: "mv1"},
			Paths: []string{"/mv1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector:   "duckdb",
				Table:       "t1",
				DisplayName: "Mv1",
				Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
					{Name: "a", DisplayName: "A", Column: "a"},
				},
				Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
					{Name: "b", DisplayName: "B", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
				},
				FirstDayOfWeek: 7,
				SecurityRules: []*runtimev1.SecurityRule{
					{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
						Condition: "true",
						Allow:     true,
					}}},
				},
			},
		},
		// metrics view mv2
		{
			Name:  ResourceName{Kind: ResourceKindMetricsView, Name: "mv2"},
			Paths: []string{"/mv2.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector:   "duckdb",
				Table:       "t2",
				DisplayName: "Mv2",
				Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
					{Name: "a", DisplayName: "A", Column: "a"},
				},
				Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
					{Name: "b", DisplayName: "B", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
				},
				FirstDayOfWeek: 1,
				SecurityRules: []*runtimev1.SecurityRule{
					{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
						Condition: "true",
						Allow:     true,
					}}},
					{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{
						Sql: "true",
					}}},
				},
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestEnvironmentOverrides(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		// Provide dashboard defaults in rill.yaml
		`rill.yaml`: `
dev:
  sources:
    limit: 10000
`,
		// source s1
		`sources/s1.yaml`: `
connector: s3
path: hello
sql: SELECT 10
dev:
  path: world
  sql: SELECT 20 # Override a property from commonYAML
  refresh:
    cron: "0 0 * * *"
    run_in_dev: true
`,
	})

	s1Base := &Resource{
		Name:  ResourceName{Kind: ResourceKindSource, Name: "s1"},
		Paths: []string{"/sources/s1.yaml"},
		SourceSpec: &runtimev1.SourceSpec{
			SourceConnector: "s3",
			SinkConnector:   "duckdb",
			Properties:      must(structpb.NewStruct(map[string]any{"path": "hello", "sql": "SELECT 10"})),
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}

	s1Test := &Resource{
		Name:  ResourceName{Kind: ResourceKindSource, Name: "s1"},
		Paths: []string{"/sources/s1.yaml"},
		SourceSpec: &runtimev1.SourceSpec{
			SourceConnector: "s3",
			SinkConnector:   "duckdb",
			Properties:      must(structpb.NewStruct(map[string]any{"path": "world", "limit": 10000, "sql": "SELECT 20"})),
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true, Cron: "0 0 * * *"},
		},
	}

	// Parse without environment
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1Base}, nil)

	// Parse in environment "dev"
	p, err = Parse(ctx, repo, "", "dev", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1Test}, nil)
}

func TestMetricsViewSecurity(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`metrics/d1.yaml`: `
version: 1
type: metrics_view
table: t1
dimensions:
  - name: a
    column: a
measures:
  - name: b
    expression: count(*)
security:
  row_filter: true
  include:
    - if: "'{{ .user.domain }}' = 'example.com'"
      names: '*'
    - if: true
      names: [a]
  exclude:
    - # Whoopsie empty
    - if: "'{{ .user.domain }}' = 'bad.com'"
      names: '*'
    - if: true
      names: [b]
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindMetricsView, Name: "d1"},
			Paths: []string{"/metrics/d1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector:   "duckdb",
				Table:       "t1",
				DisplayName: "D1",
				Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
					{Name: "a", DisplayName: "A", Column: "a"},
				},
				Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
					{Name: "b", DisplayName: "B", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
				},
				SecurityRules: []*runtimev1.SecurityRule{
					{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
						Condition: "",
						Allow:     false,
					}}},
					{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{
						Sql: "true",
					}}},
					{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
						Condition: "'{{ .user.domain }}' = 'example.com'",
						Allow:     true,
						AllFields: true,
					}}},
					{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
						Condition: "true",
						Allow:     true,
						Fields:    []string{"a"},
					}}},
					{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
						Condition: "'{{ .user.domain }}' = 'bad.com'",
						Allow:     false,
						AllFields: true,
					}}},
					{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
						Condition: "true",
						Allow:     false,
						Fields:    []string{"b"},
					}}},
				},
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestReport(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`reports/r1.yaml`: `
type: report
display_name: My Report

refresh:
  cron: 0 * * * *
  time_zone: America/Los_Angeles

watermark: inherit

intervals:
  duration: PT1H
  limit: 10

query:
  name: MetricsViewToplist
  args:
    metrics_view: mv1

export:
  format: csv
  limit: 10000

email:
  recipients:
    - benjamin@example.com

annotations:
  foo: bar
`,
		`reports/r2.yaml`: `
type: report
display_name: My Report

refresh:
  cron: 0 * * * *
  time_zone: America/Los_Angeles

watermark: inherit

intervals:
  duration: PT1H
  limit: 10

query:
  name: MetricsViewToplist
  args:
    metrics_view: mv1

export:
  format: csv
  limit: 10000


notify:
  email:
    recipients:
      - user_1@example.com
  slack:
    channels:
      - reports
    users:
      - user_2@example.com

annotations:
  foo: bar
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindReport, Name: "r1"},
			Paths: []string{"/reports/r1.yaml"},
			ReportSpec: &runtimev1.ReportSpec{
				DisplayName: "My Report",
				RefreshSchedule: &runtimev1.Schedule{
					RefUpdate: true,
					Cron:      "0 * * * *",
					TimeZone:  "America/Los_Angeles",
				},
				QueryName:     "MetricsViewToplist",
				QueryArgsJson: `{"metrics_view":"mv1"}`,
				ExportFormat:  runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
				ExportLimit:   10000,
				Notifiers: []*runtimev1.Notifier{{
					Connector:  "email",
					Properties: must(structpb.NewStruct(map[string]any{"recipients": []any{"benjamin@example.com"}})),
				}},
				Annotations:          map[string]string{"foo": "bar"},
				WatermarkInherit:     true,
				IntervalsIsoDuration: "PT1H",
				IntervalsLimit:       10,
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindReport, Name: "r2"},
			Paths: []string{"/reports/r2.yaml"},
			ReportSpec: &runtimev1.ReportSpec{
				DisplayName: "My Report",
				RefreshSchedule: &runtimev1.Schedule{
					RefUpdate: true,
					Cron:      "0 * * * *",
					TimeZone:  "America/Los_Angeles",
				},
				QueryName:     "MetricsViewToplist",
				QueryArgsJson: `{"metrics_view":"mv1"}`,
				ExportFormat:  runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
				ExportLimit:   10000,
				Notifiers: []*runtimev1.Notifier{
					{Connector: "email", Properties: must(structpb.NewStruct(map[string]any{"recipients": []any{"user_1@example.com"}}))},
					{Connector: "slack", Properties: must(structpb.NewStruct(map[string]any{"users": []any{"user_2@example.com"}, "channels": []any{"reports"}, "webhooks": []any{}}))},
				},
				Annotations:          map[string]string{"foo": "bar"},
				WatermarkInherit:     true,
				IntervalsIsoDuration: "PT1H",
				IntervalsLimit:       10,
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestAlert(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		// model m1
		`models/m1.sql`: `SELECT 1`,
		`alerts/a1.yaml`: `
type: alert
display_name: My Alert

refs:
  - model/m1

refresh:
  ref_update: false
  cron: '0 * * * *'

watermark: inherit

intervals:
  duration: PT1H
  limit: 10

query:
  name: MetricsViewToplist
  args:
    metrics_view: mv1
  for:
    user_email: benjamin@example.com

on_recover: true
renotify: true
renotify_after: 24h

notify:
  email:
    recipients:
      - benjamin@example.com

annotations:
  foo: bar
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": `SELECT 1`})),
				OutputConnector: "duckdb",
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAlert, Name: "a1"},
			Paths: []string{"/alerts/a1.yaml"},
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}},
			AlertSpec: &runtimev1.AlertSpec{
				DisplayName: "My Alert",
				RefreshSchedule: &runtimev1.Schedule{
					Cron:      "0 * * * *",
					RefUpdate: false,
				},
				WatermarkInherit:     true,
				IntervalsIsoDuration: "PT1H",
				IntervalsLimit:       10,
				Resolver:             "legacy_metrics",
				ResolverProperties: must(structpb.NewStruct(map[string]any{
					"query_name":      "MetricsViewToplist",
					"query_args_json": `{"metrics_view":"mv1"}`,
				})),
				QueryFor:             &runtimev1.AlertSpec_QueryForUserEmail{QueryForUserEmail: "benjamin@example.com"},
				NotifyOnRecover:      true,
				NotifyOnFail:         true,
				Renotify:             true,
				RenotifyAfterSeconds: 24 * 60 * 60,
				Notifiers:            []*runtimev1.Notifier{{Connector: "email", Properties: must(structpb.NewStruct(map[string]any{"recipients": []any{"benjamin@example.com"}}))}},
				Annotations:          map[string]string{"foo": "bar"},
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestMetricsViewAvoidSelfCyclicRef(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		// dashboard d1
		`metrics/d1.yaml`: `
version: 1
type: metrics_view
table: d1
dimensions:
  - name: a
    column: a
measures:
  - name: b
    expression: count(*)
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindMetricsView, Name: "d1"},
			Refs:  nil, // NOTE: This is what we're testing – that it avoids inferring the missing "d1" as a self-reference
			Paths: []string{"/metrics/d1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector:   "duckdb",
				Table:       "d1",
				DisplayName: "D1",
				Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
					{Name: "a", DisplayName: "A", Column: "a"},
				},
				Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
					{Name: "b", DisplayName: "B", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
				},
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestTheme(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		// Theme resource
		`themes/t1.yaml`: `
type: theme

colors:
  primary: red
  secondary: grey
`,
		// Explore referencing the external theme resource
		`explores/e1.yaml`: `
type: explore
metrics_view: missing
theme: t1
`,
		// Explore that defines an inline theme
		`explores/e2.yaml`: `
type: explore
metrics_view: missing
theme:
  colors:
    primary: red
`,
		// Canvas referencing the external theme resource
		`canvases/c1.yaml`: `
type: canvas
theme: t1
`,
		// Canvas that defines an inline theme
		`canvases/c2.yaml`: `
type: canvas
theme:
  colors:
    primary: red
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindTheme, Name: "t1"},
			Paths: []string{"/themes/t1.yaml"},
			ThemeSpec: &runtimev1.ThemeSpec{
				PrimaryColor: &runtimev1.Color{
					Red:   1,
					Green: 0,
					Blue:  0,
					Alpha: 1,
				},
				SecondaryColor: &runtimev1.Color{
					Red:   0.5019608,
					Green: 0.5019608,
					Blue:  0.5019608,
					Alpha: 1,
				},
				PrimaryColorRaw:   "red",
				SecondaryColorRaw: "grey",
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindExplore, Name: "e1"},
			Paths: []string{"/explores/e1.yaml"},
			Refs:  []ResourceName{{Kind: ResourceKindMetricsView, Name: "missing"}, {Kind: ResourceKindTheme, Name: "t1"}},
			ExploreSpec: &runtimev1.ExploreSpec{
				DisplayName:        "E1",
				MetricsView:        "missing",
				DimensionsSelector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
				MeasuresSelector:   &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
				Theme:              "t1",
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindExplore, Name: "e2"},
			Paths: []string{"/explores/e2.yaml"},
			Refs:  []ResourceName{{Kind: ResourceKindMetricsView, Name: "missing"}},
			ExploreSpec: &runtimev1.ExploreSpec{
				DisplayName:        "E2",
				MetricsView:        "missing",
				DimensionsSelector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
				MeasuresSelector:   &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
				EmbeddedTheme: &runtimev1.ThemeSpec{
					PrimaryColor: &runtimev1.Color{
						Red:   1,
						Green: 0,
						Blue:  0,
						Alpha: 1,
					},
					PrimaryColorRaw: "red",
				},
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindCanvas, Name: "c1"},
			Paths: []string{"/canvases/c1.yaml"},
			Refs:  []ResourceName{{Kind: ResourceKindTheme, Name: "t1"}},
			CanvasSpec: &runtimev1.CanvasSpec{
				DisplayName:    "C1",
				Theme:          "t1",
				FiltersEnabled: true,
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindCanvas, Name: "c2"},
			Paths: []string{"/canvases/c2.yaml"},
			CanvasSpec: &runtimev1.CanvasSpec{
				DisplayName: "C2",
				EmbeddedTheme: &runtimev1.ThemeSpec{
					PrimaryColor: &runtimev1.Color{
						Red:   1,
						Green: 0,
						Blue:  0,
						Alpha: 1,
					},
					PrimaryColorRaw: "red",
				},
				FiltersEnabled: true,
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestComponentsAndCanvas(t *testing.T) {
	vegaLiteSpec := normalizeJSON(t, `
  {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "description": "A simple bar chart with embedded data.",
    "mark": "bar",
    "data": {
      "name": "table"
    },
    "encoding": {
      "x": {"field": "time", "type": "nominal", "axis": {"labelAngle": 0}},
      "y": {"field": "total_sales", "type": "quantitative"}
    }
  }`)
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`components/c1.yaml`: fmt.Sprintf(`
type: component
vega_lite:
  spec: '%s'
`, vegaLiteSpec),
		`components/c2.yaml`: fmt.Sprintf(`
type: component
vega_lite:
  spec: '%s'
`, vegaLiteSpec),
		`components/c3.yaml`: `
type: component
kpi:
  metrics_view: foo
  measure: bar
  time_range: P1W
`,
		`canvases/d1.yaml`: `
type: canvas

max_width: 4
gap_x: 1
gap_y: 2

time_ranges:
  - P2W
  - range: P4W
  - range: P2M
    comparison_offsets:
      - P1M
      - offset: P4M
        range: P2M

filters:
  enable: false

defaults:
  time_range: P4W

rows:
- items:
  - component: c1
- height: 100px
  items:
  - component: c2
  - width: 2
    component:
      markdown:
        content: "Foo"
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindComponent, Name: "c1"},
			Paths: []string{"/components/c1.yaml"},
			ComponentSpec: &runtimev1.ComponentSpec{
				DisplayName:        "C1",
				Renderer:           "vega_lite",
				RendererProperties: must(structpb.NewStruct(map[string]any{"spec": vegaLiteSpec})),
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindComponent, Name: "c2"},
			Paths: []string{"/components/c2.yaml"},
			ComponentSpec: &runtimev1.ComponentSpec{
				DisplayName:        "C2",
				Renderer:           "vega_lite",
				RendererProperties: must(structpb.NewStruct(map[string]any{"spec": vegaLiteSpec})),
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindComponent, Name: "c3"},
			Paths: []string{"/components/c3.yaml"},
			Refs:  []ResourceName{{Kind: ResourceKindMetricsView, Name: "foo"}},
			ComponentSpec: &runtimev1.ComponentSpec{
				DisplayName:        "C3",
				Renderer:           "kpi",
				RendererProperties: must(structpb.NewStruct(map[string]any{"metrics_view": "foo", "measure": "bar", "time_range": "P1W"})),
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindComponent, Name: "d1--component-1-1"},
			Paths: []string{"/canvases/d1.yaml"},
			ComponentSpec: &runtimev1.ComponentSpec{
				Renderer:           "markdown",
				RendererProperties: must(structpb.NewStruct(map[string]any{"content": "Foo"})),
				DefinedInCanvas:    true,
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindCanvas, Name: "d1"},
			Paths: []string{"/canvases/d1.yaml"},
			Refs: []ResourceName{
				{Kind: ResourceKindComponent, Name: "c1"},
				{Kind: ResourceKindComponent, Name: "c2"},
				{Kind: ResourceKindComponent, Name: "d1--component-1-1"},
			},
			CanvasSpec: &runtimev1.CanvasSpec{
				DisplayName: "D1",
				MaxWidth:    4,
				GapX:        1,
				GapY:        2,
				TimeRanges: []*runtimev1.ExploreTimeRange{
					{Range: "P2W"},
					{Range: "P4W"},
					{
						Range: "P2M",
						ComparisonTimeRanges: []*runtimev1.ExploreComparisonTimeRange{
							{Offset: "P1M"},
							{Offset: "P4M", Range: "P2M"},
						},
					},
				},
				FiltersEnabled: false,
				DefaultPreset: &runtimev1.CanvasPreset{
					TimeRange:      asPtr("P4W"),
					ComparisonMode: runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_NONE,
				},
				Rows: []*runtimev1.CanvasRow{
					{
						Items: []*runtimev1.CanvasItem{
							{Component: "c1"},
						},
					},
					{
						Height:     100,
						HeightUnit: "px",
						Items: []*runtimev1.CanvasItem{
							{Component: "c2"},
							{Component: "d1--component-1-1", DefinedInCanvas: true, Width: 2},
						},
					},
				},
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestAPI(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		// model m1
		`models/m1.sql`: `SELECT 1`,
		// api a1
		`apis/a1.yaml`: `
type: api
sql: select * from m1
`,
		// api a2
		`apis/a2.yaml`: `
type: api
metrics_sql: select * from m1
`,
		// api a3 with security rules
		`apis/a3.yaml`: `
type: api
sql: select * from m1
security:
  access: true
`,
		// api a4
		`apis/a4.yaml`: `
type: api
metrics_sql: select * from m1
security:
  access: '{{ .user.admin }}'
`,
		// api a5
		`apis/a5.yaml`: `
type: api
metrics_sql: select * from m1
skip_nested_security: true
security:
  access: '{{ .user.admin }}'
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": `SELECT 1`})),
				OutputConnector: "duckdb",
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a1"},
			Paths: []string{"/apis/a1.yaml"},
			APISpec: &runtimev1.APISpec{
				Resolver:           "sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"connector": "duckdb", "sql": "select * from m1"})),
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a2"},
			Paths: []string{"/apis/a2.yaml"},
			APISpec: &runtimev1.APISpec{
				Resolver:           "metrics_sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"sql": "select * from m1"})),
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a3"},
			Paths: []string{"/apis/a3.yaml"},
			APISpec: &runtimev1.APISpec{
				Resolver:           "sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"connector": "duckdb", "sql": "select * from m1"})),
				SecurityRules: []*runtimev1.SecurityRule{
					{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
						Condition: "true",
						Allow:     true,
					}}},
				},
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a4"},
			Paths: []string{"/apis/a4.yaml"},
			APISpec: &runtimev1.APISpec{
				Resolver:           "metrics_sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"sql": "select * from m1"})),
				SecurityRules: []*runtimev1.SecurityRule{
					{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
						Condition: "{{ .user.admin }}",
						Allow:     true,
					}}},
				},
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a5"},
			Paths: []string{"/apis/a5.yaml"},
			APISpec: &runtimev1.APISpec{
				Resolver:           "metrics_sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"sql": "select * from m1"})),
				SecurityRules: []*runtimev1.SecurityRule{
					{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
						Condition: "{{ .user.admin }}",
						Allow:     true,
					}}},
				},
				SkipNestedSecurity: true,
			},
		},
	}
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestKindBackwardsCompatibility(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// source s1
		`sources/s1.yaml`: `
type: s3
`,
		// api a1
		`/apis/a1.yaml`: `
kind: api
refs:
- kind: source
  name: s1
- type: source
  name: s2
sql: select 1
`,
		// migration m1
		`/migrations/m1.sql`: `
-- @kind: migration
select 2
`,
		// migration m2
		`/migrations/m2.sql`: `
-- @type: migration
select 3
`,
	}

	resources := []*Resource{
		// s1
		{
			Name:  ResourceName{Kind: ResourceKindSource, Name: "s1"},
			Paths: []string{"/sources/s1.yaml"},
			SourceSpec: &runtimev1.SourceSpec{
				SourceConnector: "s3",
				SinkConnector:   "duckdb",
				Properties:      must(structpb.NewStruct(map[string]any{})),
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
		// a1
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a1"},
			Paths: []string{"/apis/a1.yaml"},
			Refs:  []ResourceName{{Kind: ResourceKindSource, Name: "s1"}, {Kind: ResourceKindSource, Name: "s2"}},
			APISpec: &runtimev1.APISpec{
				Resolver:           "sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"connector": "duckdb", "sql": "select 1"})),
			},
		},
		// m1
		{
			Name:  ResourceName{Kind: ResourceKindMigration, Name: "m1"},
			Paths: []string{"/migrations/m1.sql"},
			MigrationSpec: &runtimev1.MigrationSpec{
				Connector: "duckdb",
				Sql:       strings.TrimSpace(files["/migrations/m1.sql"]),
			},
		},
		// m2
		{
			Name:  ResourceName{Kind: ResourceKindMigration, Name: "m2"},
			Paths: []string{"/migrations/m2.sql"},
			MigrationSpec: &runtimev1.MigrationSpec{
				Connector: "duckdb",
				Sql:       strings.TrimSpace(files["/migrations/m2.sql"]),
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestAdvancedMeasures(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// dashboard d1
		`metrics/d1.yaml`: `
version: 1
type: metrics_view
table: t1
timeseries: t
dimensions:
  - column: foo
measures:
  - name: a
    type: simple
    expression: count(*)
  - name: b
    type: derived
    expression: a+1
    requires: [a]
  - name: c
    expression: sum(a)
    per: [foo]
    requires: [a]
  - name: d
    type: derived
    expression: a/lag(a)
    window:
      order:
      - name: t
        time_grain: day
      frame: unbounded preceding to current row
    requires:
      - a
`,
	}

	resources := []*Resource{
		// dashboard d1
		{
			Name:  ResourceName{Kind: ResourceKindMetricsView, Name: "d1"},
			Paths: []string{"/metrics/d1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector:     "duckdb",
				Table:         "t1",
				DisplayName:   "D1",
				TimeDimension: "t",
				Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
					{Name: "foo", DisplayName: "Foo", Column: "foo"},
				},
				Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
					{
						Name:        "a",
						DisplayName: "A",
						Expression:  "count(*)",
						Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
					},
					{
						Name:               "b",
						DisplayName:        "B",
						Expression:         "a+1",
						Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED,
						ReferencedMeasures: []string{"a"},
					},
					{
						Name:               "c",
						DisplayName:        "C",
						Expression:         "sum(a)",
						Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED,
						PerDimensions:      []*runtimev1.MetricsViewSpec_DimensionSelector{{Name: "foo"}},
						ReferencedMeasures: []string{"a"},
					},
					{
						Name:        "d",
						DisplayName: "D",
						Expression:  "a/lag(a)",
						Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED,
						Window: &runtimev1.MetricsViewSpec_MeasureWindow{
							Partition:       true,
							OrderBy:         []*runtimev1.MetricsViewSpec_DimensionSelector{{Name: "t", TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY}},
							FrameExpression: "unbounded preceding to current row",
						},
						RequiredDimensions: []*runtimev1.MetricsViewSpec_DimensionSelector{{Name: "t", TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY}},
						ReferencedMeasures: []string{"a"},
					},
				},
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestRefreshInDev(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		// model m1
		`m1.yaml`: `
type: model
sql: SELECT 1
refresh:
  cron: 0 0 * * *
`,
		// model m2
		`m2.yaml`: `
type: model
sql: SELECT 1
refresh:
  cron: 0 0 * * *
  run_in_dev: true
`,
	})

	m1 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Paths: []string{"/m1.yaml"},
		ModelSpec: &runtimev1.ModelSpec{
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true, Cron: "0 0 * * *"},
			InputConnector:  "duckdb",
			InputProperties: must(structpb.NewStruct(map[string]any{"sql": `SELECT 1`})),
			OutputConnector: "duckdb",
		},
	}

	m2 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
		Paths: []string{"/m2.yaml"},
		ModelSpec: &runtimev1.ModelSpec{
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true, Cron: "0 0 * * *"},
			InputConnector:  "duckdb",
			InputProperties: must(structpb.NewStruct(map[string]any{"sql": `SELECT 1`})),
			OutputConnector: "duckdb",
		},
	}

	// Parse for prod and check
	p, err := Parse(ctx, repo, "", "prod", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1, m2}, nil)

	// Clear the cron refresh only for m1
	m1.ModelSpec.RefreshSchedule.Cron = ""

	// Parse for dev and check
	p, err = Parse(ctx, repo, "", "dev", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1, m2}, nil)
}

func TestConnector(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{`rill.yaml`: ``})

	putRepo(t, repo, map[string]string{
		`connectors/clickhouse.yaml`: `
type: connector
driver: clickhouse
`})
	r := &Resource{
		Name:  ResourceName{Kind: ResourceKindConnector, Name: "clickhouse"},
		Paths: []string{"/connectors/clickhouse.yaml"},
		ConnectorSpec: &runtimev1.ConnectorSpec{
			Driver: "clickhouse",
		},
	}
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{r}, nil)

	putRepo(t, repo, map[string]string{
		`connectors/clickhouse.yaml`: `
type: connector
driver: clickhouse
managed: true
`})
	r = &Resource{
		Name:  ResourceName{Kind: ResourceKindConnector, Name: "clickhouse"},
		Paths: []string{"/connectors/clickhouse.yaml"},
		ConnectorSpec: &runtimev1.ConnectorSpec{
			Driver:    "clickhouse",
			Provision: true,
		},
	}
	p, err = Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{r}, nil)

	putRepo(t, repo, map[string]string{
		`connectors/clickhouse.yaml`: `
type: connector
driver: clickhouse
managed:
  hello: world
time_zone: America/Los_Angeles
`})
	r = &Resource{
		Name:  ResourceName{Kind: ResourceKindConnector, Name: "clickhouse"},
		Paths: []string{"/connectors/clickhouse.yaml"},
		ConnectorSpec: &runtimev1.ConnectorSpec{
			Driver:        "clickhouse",
			Properties:    map[string]string{"time_zone": "America/Los_Angeles"},
			Provision:     true,
			ProvisionArgs: must(structpb.NewStruct(map[string]any{"hello": "world"})),
		},
	}
	p, err = Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{r}, nil)

	putRepo(t, repo, map[string]string{
		`connectors/clickhouse.yaml`: `
type: connector
driver: clickhouse
managed: 10
`})
	p, err = Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, nil, []*runtimev1.ParseError{
		{Message: "failed to decode 'managed'", FilePath: "/connectors/clickhouse.yaml"},
	})
}

func TestNamespace(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`models/m1.yaml`: `
type: model
sql: SELECT 1
`,
		`explores/e1.yaml`: `
type: explore
namespace: foo
metrics_view: missing
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.yaml"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": `SELECT 1`})),
				OutputConnector: "duckdb",
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindExplore, Name: "foo:e1"},
			Paths: []string{"/explores/e1.yaml"},
			Refs:  []ResourceName{{Kind: ResourceKindMetricsView, Name: "missing"}},
			ExploreSpec: &runtimev1.ExploreSpec{
				DisplayName:        "Foo: E1",
				MetricsView:        "missing",
				DimensionsSelector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
				MeasuresSelector:   &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func requireResourcesAndErrors(t testing.TB, p *Parser, wantResources []*Resource, wantErrors []*runtimev1.ParseError) {
	// Check errors
	// NOTE: Assumes there's at most one parse error per file path
	// NOTE: Matches error messages using Contains (exact match not required)
	gotErrors := slices.Clone(p.Errors)
	for _, want := range wantErrors {
		found := false
		for i, got := range gotErrors {
			if want.FilePath == got.FilePath {
				require.Contains(t, got.Message, want.Message, "for path %q", got.FilePath)
				require.Equal(t, want.StartLocation, got.StartLocation, "for path %q", got.FilePath)
				gotErrors = slices.Delete(gotErrors, i, i+1)
				found = true
				break
			}
		}
		require.True(t, found, "missing error for path %q", want.FilePath)
	}
	require.True(t, len(gotErrors) == 0, "unexpected errors: %v", gotErrors)

	// Check resources
	gotResources := maps.Clone(p.Resources)
	for _, want := range wantResources {
		found := false
		for _, got := range gotResources {
			if want.Name == got.Name {
				require.Equal(t, want.Name, got.Name)
				require.ElementsMatch(t, want.Refs, got.Refs, "for resource %q", want.Name)
				require.ElementsMatch(t, want.Paths, got.Paths, "for resource %q", want.Name)
				require.Equal(t, want.SourceSpec, got.SourceSpec, "for resource %q", want.Name)
				require.Equal(t, want.ModelSpec, got.ModelSpec, "for resource %q", want.Name)
				require.Equal(t, want.MetricsViewSpec, got.MetricsViewSpec, "for resource %q", want.Name)
				require.Equal(t, want.ExploreSpec, got.ExploreSpec, "for resource %q", want.Name)
				require.Equal(t, want.MigrationSpec, got.MigrationSpec, "for resource %q", want.Name)
				require.True(t, reflect.DeepEqual(want.ReportSpec, got.ReportSpec), "for resource %q", want.Name)
				require.True(t, reflect.DeepEqual(want.AlertSpec, got.AlertSpec), "for resource %q", want.Name)
				require.Equal(t, want.ThemeSpec, got.ThemeSpec, "for resource %q", want.Name)
				require.Equal(t, want.ComponentSpec, got.ComponentSpec, "for resource %q", want.Name)
				require.Equal(t, want.CanvasSpec, got.CanvasSpec, "for resource %q", want.Name)
				require.Equal(t, want.APISpec, got.APISpec, "for resource %q", want.Name)
				require.Equal(t, want.ConnectorSpec, got.ConnectorSpec, "for resource %q", want.Name)

				delete(gotResources, got.Name)
				found = true
				break
			}
		}
		require.True(t, found, "missing resource %q", want.Name)
	}
	require.True(t, len(gotResources) == 0, "unexpected resources: %v", gotResources)
}

func makeRepo(t testing.TB, files map[string]string) drivers.RepoStore {
	root := t.TempDir()
	handle, err := drivers.Open("file", "default", map[string]any{"dsn": root}, storage.MustNew(root, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	repo, ok := handle.AsRepoStore("")
	require.True(t, ok)

	putRepo(t, repo, files)

	return repo
}

func putRepo(t testing.TB, repo drivers.RepoStore, files map[string]string) {
	for path, data := range files {
		err := repo.Put(context.Background(), path, strings.NewReader(data))
		require.NoError(t, err)
	}
}

func deleteRepo(t testing.TB, repo drivers.RepoStore, files ...string) {
	for _, path := range files {
		err := repo.Delete(context.Background(), path, false)
		require.NoError(t, err)
	}
}

func asPtr[T any](val T) *T {
	return &val
}

func normalizeJSON(t *testing.T, s string) string {
	var v interface{}
	require.NoError(t, json.Unmarshal([]byte(s), &v))
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
