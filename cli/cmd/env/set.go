package env

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// SetCmd is sub command for env. Sets the variable for a project
func SetCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectPath, projectName, enviornment string

	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Args:  cobra.ExactArgs(2),
		Short: "Set variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]
			ctx := cmd.Context()
			client, err := ch.Client()
			if err != nil {
				return err
			}

			// Find the cloud project name
			if projectName == "" {
				projectName, err = ch.InferProjectName(cmd.Context(), ch.Org, projectPath)
				if err != nil {
					return err
				}
			}

			_, err = client.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				Organization: ch.Org,
				Project:      projectName,
				Environment:  enviornment,
				Variables:    map[string]string{key: value},
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Updated project variables\n")
			return nil
		},
	}

	setCmd.Flags().StringVar(&projectName, "project", "", "Cloud project name (will attempt to infer from Git remote if not provided)")
	setCmd.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	setCmd.Flags().StringVar(&enviornment, "environment", "production", "Environment for which variables apply (options: 'production', 'development', '')")

	return setCmd
}
