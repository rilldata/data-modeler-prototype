package whitelist

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AddCmd(ch *cmdutil.Helper) *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add <org> <domain> <role>",
		Args:  cobra.ExactArgs(3),
		Short: "Whitelist users from a domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			org := args[0]
			domain := args[1]
			role := args[2]

			ch.Printer.PrintlnWarn(fmt.Sprintf("Warn: Whitelisting will give all users from domain %q access to the organization %q as %s", domain, org, role))
			if !cmdutil.ConfirmPrompt("Do you want to continue", "", false) {
				ch.Printer.PrintlnWarn("Aborted")
				return nil
			}

			_, err = client.CreateWhitelistedDomain(ctx, &adminv1.CreateWhitelistedDomainRequest{
				Organization: org,
				Domain:       domain,
				Role:         role,
			})
			if err != nil {
				return err
			}

			ch.Printer.PrintlnSuccess("Success")

			return nil
		},
	}

	return addCmd
}
