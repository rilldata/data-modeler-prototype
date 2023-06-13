package user

import (
	"github.com/rilldata/rill/cli/cmd/user/whitelist"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func UserCmd(cfg *config.Config) *cobra.Command {
	userCmd := &cobra.Command{
		Use:               "user",
		Short:             "Manage users",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
	}

	userCmd.AddCommand(ListCmd(cfg))
	userCmd.AddCommand(AddCmd(cfg))
	userCmd.AddCommand(RemoveCmd(cfg))
	userCmd.AddCommand(SetRoleCmd(cfg))
	userCmd.AddCommand(whitelist.WhitelistCmd(cfg))

	return userCmd
}

var userRoles = []string{"admin", "viewer"}
