package service

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List service",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.ListServices(cmd.Context(), &adminv1.ListServicesRequest{
				OrganizationName: cfg.Org,
			})
			if err != nil {
				return err
			}

			cmdutil.TablePrinter(toTable(res.Services))

			return nil
		},
	}

	return listCmd
}
