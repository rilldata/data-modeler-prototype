package project

import (
	"context"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List",
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.Spinner("Listing projects...")
			sp.Start()

			client, err := client.New(cfg.AdminURL, cfg.AdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			proj, err := client.ListProjects(context.Background(), &adminv1.ListProjectsRequest{
				OrganizationName: cfg.Org(),
			})
			if err != nil {
				return err
			}

			sp.Stop()
			cmdutil.TextPrinter("Projects list \n")
			cmdutil.TablePrinter(toTable(proj.Projects))
			return nil
		},
	}
	return listCmd
}
