package devtool

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/project"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func SeedCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed {cloud}",
		Short: "Authenticate and deploy a seed project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			preset := args[0]
			if preset != "cloud" {
				return fmt.Errorf("seeding not available for preset %q", preset)
			}

			return project.GitPushFlow(cmd.Context(), ch, &project.DeployOpts{
				GitPath:     "https://github.com/rilldata/rill-examples.git",
				SubPath:     "rill-openrtb-prog-ads",
				Name:        "rill-openrtb-prog-ads",
				ProdVersion: "latest",
				DBDriver:    "duckdb",
				Slots:       2,
			})
		},
	}

	return cmd
}
