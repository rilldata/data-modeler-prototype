package env

import (
	"fmt"
	"maps"
	"path/filepath"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
)

func PushCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectPath, projectName string

	pullCmd := &cobra.Command{
		Use:   "push",
		Short: "Push local .env contents to cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If projectPath is provided, normalize it
			if projectPath != "" {
				var err error
				projectPath, err = fileutil.ExpandHome(projectPath)
				if err != nil {
					return err
				}
				projectPath, err = filepath.Abs(projectPath)
				if err != nil {
					return err
				}
			}

			// Parse and verify the project directory
			repo, instanceID, err := cmdutil.RepoForProjectPath(projectPath)
			if err != nil {
				return err
			}
			parser, err := rillv1.Parse(cmd.Context(), repo, instanceID, "prod", "duckdb")
			if err != nil {
				return fmt.Errorf("failed to parse project: %w", err)
			}
			if parser.RillYAML == nil {
				return fmt.Errorf("not a valid Rill project (missing a rill.yaml file)")
			}

			// Find the cloud project name
			if projectName == "" {
				projectName, err = ch.InferProjectName(cmd.Context(), ch.Org, projectPath)
				if err != nil {
					return err
				}
			}

			// Fetch the project variables from the cloud
			client, err := ch.Client()
			if err != nil {
				return err
			}
			res, err := client.GetProjectVariables(cmd.Context(), &adminv1.GetProjectVariablesRequest{
				OrganizationName: ch.Org,
				Name:             projectName,
			})
			if err != nil {
				return err
			}

			// Merge the current .env file with the cloud variables
			vars := make(map[string]string)
			for k, v := range res.Variables {
				vars[k] = v
			}
			for k, v := range parser.DotEnv {
				vars[k] = v
			}

			// Write the merged variables back to the cloud project
			if !maps.Equal(res.Variables, vars) {
				_, err = client.UpdateProjectVariables(cmd.Context(), &adminv1.UpdateProjectVariablesRequest{
					OrganizationName: ch.Org,
					Name:             projectName,
					Variables:        vars,
				})
				if err != nil {
					return err
				}
			}

			ch.Printf("Updated cloud env for project %q with variables from %q.\n", projectName, filepath.Join(projectPath, ".env"))
			return nil
		},
	}

	pullCmd.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	pullCmd.Flags().StringVar(&projectName, "project", "", "Cloud project name (will attempt to infer from Git remote if not provided)")

	return pullCmd
}
