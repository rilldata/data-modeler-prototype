package project

import (
	"context"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(cfg *config.Config) *cobra.Command {
	var name, path string

	showCmd := &cobra.Command{
		Use:   "show [<project-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Show project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && cfg.Interactive {
				name, err = inferProjectName(cmd.Context(), client, cfg.Org, path)
				if err != nil {
					return err
				}
			}

			proj, err := client.GetProject(context.Background(), &adminv1.GetProjectRequest{
				OrganizationName: cfg.Org,
				Name:             name,
			})
			if err != nil {
				return err
			}

			cmdutil.TablePrinter(toRow(proj.Project))
			return nil
		},
	}

	showCmd.Flags().SortFlags = false
	showCmd.Flags().StringVar(&name, "project", "", "Name")
	showCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return showCmd
}
