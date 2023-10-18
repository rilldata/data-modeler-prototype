package token

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/printer"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func IssueCmd(ch *cmdutil.Helper) *cobra.Command {
	var name string
	issueCmd := &cobra.Command{
		Use:   "issue [<service>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Issue service token",
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

			res, err := client.IssueServiceAuthToken(cmd.Context(), &adminv1.IssueServiceAuthTokenRequest{
				OrganizationName: cfg.Org,
				ServiceName:      name,
			})
			if err != nil {
				return err
			}

			ch.Printer.Println(printer.BoldGreen(fmt.Sprintf("Issued token: %v", res.Token)))

			return nil
		},
	}

	issueCmd.Flags().SortFlags = false
	issueCmd.Flags().StringVar(&name, "service", "", "Service Name")

	return issueCmd
}
