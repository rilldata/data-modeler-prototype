package token

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func TokenCmd(cfg *config.Config) *cobra.Command {
	var service string
	tokenCmd := &cobra.Command{
		Use:               "token",
		Short:             "Manage service tokens",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
	}

	tokenCmd.PersistentFlags().StringVar(&service, "service", "", "Service Name")
	tokenCmd.AddCommand(IssueCmd(cfg))
	tokenCmd.AddCommand(ListCmd(cfg))
	tokenCmd.AddCommand(RevokeCmd(cfg))

	return tokenCmd
}
