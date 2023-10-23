package token

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var name string
	listCmd := &cobra.Command{
		Use:   "list [<service>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "List tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ch.Config
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				name = args[0]
			}

			res, err := client.ListServiceAuthTokens(cmd.Context(), &adminv1.ListServiceAuthTokensRequest{
				ServiceName:      name,
				OrganizationName: cfg.Org,
			})
			if err != nil {
				return err
			}

			if len(res.Tokens) == 0 {
				ch.Printer.PrintlnWarn("No tokens found")
				return nil
			}

			err = ch.Printer.PrintResource(toTable(res.Tokens))
			if err != nil {
				return err
			}
			return nil
		},
	}

	listCmd.Flags().SortFlags = false
	listCmd.Flags().StringVar(&name, "service", "", "Service Name")

	return listCmd
}
