package billing

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func BillingCmd(ch *cmdutil.Helper) *cobra.Command {
	billingCmd := &cobra.Command{
		Use:   "billing",
		Short: "Billing update for customers",
	}

	billingCmd.AddCommand(SetCmd(ch))

	return billingCmd
}
