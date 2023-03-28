package project

import (
	"context"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(cfg *config.Config) *cobra.Command {
	var name, description, prodBranch string
	var public bool

	editCmd := &cobra.Command{
		Use:   "edit",
		Args:  cobra.ExactArgs(1),
		Short: "Edit",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			proj, err := client.UpdateProject(context.Background(), &adminv1.UpdateProjectRequest{
				OrganizationName: cfg.Org,
				Name:             args[0],
				Description:      description,
			})
			if err != nil {
				return err
			}

			cmdutil.TextPrinter("Updated project \n")
			cmdutil.TablePrinter(toRow(proj.Project))
			return nil
		},
	}

	editCmd.Flags().SortFlags = false

	editCmd.Flags().StringVar(&name, "name", "noname", "Name")
	editCmd.Flags().StringVar(&description, "description", "", "Description")
	editCmd.Flags().StringVar(&prodBranch, "prod-branch", "noname", "Production branch name")
	editCmd.Flags().BoolVar(&public, "public", false, "Public")

	return editCmd
}
