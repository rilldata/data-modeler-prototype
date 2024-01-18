package annotations

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func GetCmd(ch *cmdutil.Helper) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Get annotations for project in an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ch.Config

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()
			res, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				OrganizationName: args[0],
				Name:             args[1],
			})
			if err != nil {
				return err
			}

			if len(res.Project.Annotations) == 0 {
				ch.Printer.PrintlnWarn("No annotations found")
				return nil
			}

			for k, v := range res.Project.Annotations {
				fmt.Printf("%s=%s\n", k, v)
			}

			return nil
		},
	}

	return getCmd
}
