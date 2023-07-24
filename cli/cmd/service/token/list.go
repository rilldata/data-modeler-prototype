package token

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	var name string
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.ListServiceAuthTokens(cmd.Context(), &adminv1.ListServiceAuthTokensRequest{
				ServiceName:      name,
				OrganizationName: cfg.Org,
			})
			if err != nil {
				return err
			}

			cmdutil.TablePrinter(res.Tokens)
			return nil
		},
	}

	listCmd.Flags().SortFlags = false
	listCmd.Flags().StringVar(&name, "service", "", "Service Name")

	return listCmd
}
