package rillv1_test

import (
	"context"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestEnvParsing tests repo without a .env file
func TestEnvParsing(t *testing.T) {
	t.Run("without env file", func(t *testing.T) {
		ctx := context.Background()
		repo := makeRepo(t, map[string]string{
			`rill.yaml`: ``,
		})

		parser, err := rillv1.Parse(ctx, repo, "", "", "duckdb")
		require.NoError(t, err)
		require.Equal(t, len(parser.DotEnv), 0)
	})

	t.Run("with single env file", func(t *testing.T) {
		ctx := context.Background()
		repo := makeRepo(t, map[string]string{
			"rill.yaml": ``,
			".env": `
TEST=test
FOO=bar
`,
		})

		parser, err := rillv1.Parse(ctx, repo, "", "", "duckdb")
		require.NoError(t, err)

		mergedEnv := parser.GetDotEnv()

		require.Equal(t, "test", mergedEnv["TEST"])
		require.Equal(t, "bar", mergedEnv["FOO"])
	})

	t.Run("with multiple env files", func(t *testing.T) {
		ctx := context.Background()
		repo := makeRepo(t, map[string]string{
			`rill.yaml`: ``,
			".env": `
ROOT_TEST_VAR=root
`,
			"models/.env": `
MODELS_FOLDER_TEST_VAR=models
`,
		})

		parser, err := rillv1.Parse(ctx, repo, "", "", "duckdb")
		require.NoError(t, err)

		require.Empty(t, parser.Errors)
		mergedEnv := parser.GetDotEnv()
		require.Equal(t, "root", mergedEnv["ROOT_TEST_VAR"], "root-only variable should be preserved")
		require.Equal(t, "models", mergedEnv["MODELS_FOLDER_TEST_VAR"], "models-only variable should be preserved")
	})

	t.Run("env value merge behavior", func(t *testing.T) {
		ctx := context.Background()
		repo := makeRepo(t, map[string]string{
			`rill.yaml`: ``,
			".env": `
SHARED_VAR=root_value
ROOT_ONLY=root_value
`,
			"models/.env": `
SHARED_VAR=models_value
MODELS_ONLY=models_value
`,
			"models/nested/.env": `
SHARED_VAR=nested_value
NESTED_ONLY=nested_value
`,
		})

		parser, err := rillv1.Parse(ctx, repo, "", "", "duckdb")
		require.NoError(t, err)
		require.Empty(t, parser.Errors)

		mergedEnv := parser.GetDotEnv()

		// Check that variables from all levels exist
		require.Equal(t, "root_value", mergedEnv["ROOT_ONLY"], "root-only variable should be preserved")
		require.Equal(t, "models_value", mergedEnv["MODELS_ONLY"], "models-only variable should be preserved")
		require.Equal(t, "nested_value", mergedEnv["NESTED_ONLY"], "nested-only variable should be preserved")

		// Check that shared variable takes the value from the deepest .env file
		require.Equal(t, "nested_value", mergedEnv["SHARED_VAR"], "shared variable should take value from deepest .env file")
	})
}

// Helper functions

// makeRepo is a helper function that creates a new repo with the given files
func makeRepo(t testing.TB, files map[string]string) drivers.RepoStore {
	root := t.TempDir()
	handle, err := drivers.Open("file", "default", map[string]any{"dsn": root}, storage.MustNew(root, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	repo, ok := handle.AsRepoStore("")
	require.True(t, ok)

	for path, data := range files {
		err := repo.Put(context.Background(), path, strings.NewReader(data))
		require.NoError(t, err)
	}
	return repo
}
