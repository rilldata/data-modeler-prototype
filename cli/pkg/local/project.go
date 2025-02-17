package local

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
)

func ParseDotenv(ctx context.Context, projectPath string) (map[string]string, error) {
	repo, instanceID, err := cmdutil.RepoForProjectPath(projectPath)
	if err != nil {
		return nil, err
	}
	parser, err := rillv1.Parse(ctx, repo, instanceID, "prod", "duckdb")
	if err != nil {
		return nil, err
	}
	if parser.RillYAML == nil {
		return nil, fmt.Errorf("not a valid Rill project (missing a rill.yaml file)")
	}

	env := parser.GetDotEnv()

	return env, nil
}
