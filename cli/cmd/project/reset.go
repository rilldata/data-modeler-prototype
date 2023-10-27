package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ResetCmd(cfg *config.Config) *cobra.Command {
	var project, path string

	resetCmd := &cobra.Command{
		Use:               "reset [<project-name>]",
		Args:              cobra.MaximumNArgs(1),
		Short:             "Reset project",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				project = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && cfg.Interactive {
				var err error
				project, err = inferProjectName(ctx, client, cfg.Org, path)
				if err != nil {
					return err
				}
			}

			_, err = client.TriggerRedeploy(ctx, &adminv1.TriggerRedeployRequest{Organization: cfg.Org, Project: project})
			if err != nil {
				return err
			}

			fmt.Printf("Triggered project reset. To see status, run `rill project status --project %s`.\n", project)

			return nil
		},
	}

	resetCmd.Flags().SortFlags = false
	resetCmd.Flags().StringVar(&project, "project", "", "Project name")
	resetCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return resetCmd
}
