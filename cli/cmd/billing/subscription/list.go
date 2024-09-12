package subscription

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List subscription for an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			resp, err := client.GetBillingSubscription(cmd.Context(), &adminv1.GetBillingSubscriptionRequest{
				OrgName: ch.Org,
			})
			if err != nil {
				return err
			}

			if resp.Subscription == nil {
				ch.PrintfWarn("No subscription found for organization %q.\n", ch.Org)
				return nil
			}

			ch.PrintfSuccess("Subscription for organization %q\n", ch.Org)
			ch.PrintSubscriptions([]*adminv1.Subscription{resp.Subscription})
			return nil
		},
	}
	return listCmd
}
