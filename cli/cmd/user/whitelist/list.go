package whitelist

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	cfg := ch.Config
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List whitelisted email domains for the org",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			whitelistedDomains, err := client.ListWhitelistedDomains(ctx, &adminv1.ListWhitelistedDomainsRequest{Organization: cfg.Org})
			if err != nil {
				return err
			}

			if len(whitelistedDomains.Domains) > 0 {
				ch.Printer.Println(fmt.Sprintf("Whitelisted email domains for %q:", cfg.Org))
				for _, d := range whitelistedDomains.Domains {
					ch.Printer.Println(fmt.Sprintf("%q (%q)", d.Domain, d.Role))
				}
			} else {
				ch.Printer.Println(fmt.Sprintf("No whitelisted email domains for %q", cfg.Org))
			}
			return nil
		},
	}

	listCmd.Flags().StringVar(&cfg.Org, "org", cfg.Org, "Organization")

	return listCmd
}
