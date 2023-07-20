package service

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RenameCmd(cfg *config.Config) *cobra.Command {
	var newName string

	renameCmd := &cobra.Command{
		Use:   "rename",
		Args:  cobra.MaximumNArgs(1),
		Short: "Rename service",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.UpdateService(cmd.Context(), &adminv1.UpdateServiceRequest{
				Name:             args[0],
				OrganizationName: cfg.Org,
				NewName:          &newName,
			})
			if err != nil {
				return err
			}

			cmdutil.PrintlnSuccess("Renamed service")
			cmdutil.TablePrinter(toRow(res.Service))

			return nil
		},
	}
	renameCmd.Flags().SortFlags = false
	renameCmd.Flags().StringVar(&newName, "new-name", "", "New Service Name")

	return renameCmd
}
