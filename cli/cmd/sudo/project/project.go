package project

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func ProjectCmd(ch *cmdutil.Helper) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Project search for support users",
	}

	projectCmd.AddCommand(GetCmd(ch))
	projectCmd.AddCommand(EditCmd(ch))
	projectCmd.AddCommand(SearchCmd(ch))
	projectCmd.AddCommand(HibernateCmd(ch))
	projectCmd.AddCommand(ResetCmd(ch))
	projectCmd.AddCommand(DumpResources(ch))

	return projectCmd
}
