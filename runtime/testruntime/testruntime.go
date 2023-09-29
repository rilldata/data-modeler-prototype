package testruntime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"

	"github.com/c2h5oh/datasize"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	// Load database drivers for testing.
	_ "github.com/rilldata/rill/runtime/drivers/bigquery"
	_ "github.com/rilldata/rill/runtime/drivers/druid"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/gcs"
	_ "github.com/rilldata/rill/runtime/drivers/github"
	_ "github.com/rilldata/rill/runtime/drivers/https"
	_ "github.com/rilldata/rill/runtime/drivers/postgres"
	_ "github.com/rilldata/rill/runtime/drivers/s3"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	_ "github.com/rilldata/rill/runtime/reconcilers"
)

// TestingT satisfies both *testing.T and *testing.B.
type TestingT interface {
	Name() string
	TempDir() string
	FailNow()
	Errorf(format string, args ...interface{})
	Cleanup(f func())
}

// New returns a runtime configured for use in tests.
func New(t TestingT) *runtime.Runtime {
	opts := &runtime.Options{
		MetastoreConnector: "metastore",
		SystemConnectors: []*runtimev1.Connector{
			{
				Type: "sqlite",
				Name: "metastore",
				// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
				// "cache=shared" is needed to prevent threading problems.
				Config: map[string]string{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())},
			},
		},
		ConnectionCacheSize:     100,
		QueryCacheSizeBytes:     int64(datasize.MB * 100),
		SecurityEngineCacheSize: 100,
		AllowHostAccess:         true,
	}

	logger := zap.NewNop()
	// nolint
	// logger, err := zap.NewDevelopment()
	// require.NoError(t, err)

	rt, err := runtime.New(context.Background(), opts, logger, activity.NewNoopClient())
	require.NoError(t, err)
	t.Cleanup(func() { rt.Close() })

	return rt
}

// InstanceOptions enables configuration of the instance options that are configurable in tests.
type InstanceOptions struct {
	Files                        map[string]string
	Variables                    map[string]string
	IngestionLimitBytes          int64
	WatchRepo                    bool
	StageChanges                 bool
	ModelDefaultMaterialize      bool
	ModelMaterializeDelaySeconds uint32
}

// NewInstanceWithOptions creates a runtime and an instance for use in tests.
// The instance's repo is a temp directory that will be cleared when the tests finish.
func NewInstanceWithOptions(t TestingT, opts InstanceOptions) (*runtime.Runtime, string) {
	rt := New(t)

	tmpDir := t.TempDir()
	inst := &drivers.Instance{
		OLAPConnector: "duckdb",
		RepoConnector: "repo",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": tmpDir},
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: map[string]string{"dsn": ""},
			},
		},
		Variables:                    opts.Variables,
		EmbedCatalog:                 true,
		IngestionLimitBytes:          opts.IngestionLimitBytes,
		WatchRepo:                    opts.WatchRepo,
		StageChanges:                 opts.StageChanges,
		ModelDefaultMaterialize:      opts.ModelDefaultMaterialize,
		ModelMaterializeDelaySeconds: opts.ModelMaterializeDelaySeconds,
	}

	for path, data := range opts.Files {
		abs := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(abs), os.ModePerm))
		require.NoError(t, os.WriteFile(abs, []byte(data), 0o644))
	}

	err := rt.CreateInstance(context.Background(), inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	ctrl, err := rt.Controller(inst.ID)
	require.NoError(t, err)

	_, err = ctrl.Get(context.Background(), runtime.GlobalProjectParserName, false)
	require.NoError(t, err)

	err = ctrl.WaitUntilReady(context.Background())
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(context.Background(), opts.WatchRepo)
	require.NoError(t, err)

	return rt, inst.ID
}

// NewInstance is a convenience wrapper around NewInstanceWithOptions, using defaults sensible for most tests.
func NewInstance(t TestingT) (*runtime.Runtime, string) {
	return NewInstanceWithOptions(t, InstanceOptions{
		Files: map[string]string{"rill.yaml": ""},
	})
}

// NewInstanceWithModel creates a runtime and an instance for use in tests.
// The passed model name and SQL SELECT statement will be loaded into the instance.
func NewInstanceWithModel(t TestingT, name, sql string) (*runtime.Runtime, string) {
	path := filepath.Join("models", name+".sql")
	return NewInstanceWithOptions(t, InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			path:        sql,
		},
	})
}

// NewInstanceForProject creates a runtime and an instance for use in tests.
// The passed name should match a test project in the testdata folder.
// You should not do mutable repo operations on the returned instance.
func NewInstanceForProject(t TestingT, name string) (*runtime.Runtime, string) {
	rt := New(t)

	_, currentFile, _, _ := goruntime.Caller(0)
	projectPath := filepath.Join(currentFile, "..", "testdata", name)

	inst := &drivers.Instance{
		OLAPConnector: "duckdb",
		RepoConnector: "repo",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": projectPath},
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: map[string]string{"dsn": ""},
			},
		},
		EmbedCatalog: true,
	}

	err := rt.CreateInstance(context.Background(), inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	ctrl, err := rt.Controller(inst.ID)
	require.NoError(t, err)

	_, err = ctrl.Get(context.Background(), runtime.GlobalProjectParserName, false)
	require.NoError(t, err)

	err = ctrl.WaitUntilReady(context.Background())
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(context.Background(), false)
	require.NoError(t, err)

	return rt, inst.ID
}
