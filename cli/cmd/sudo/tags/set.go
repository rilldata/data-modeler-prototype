package tags

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetCmd(ch *cmdutil.Helper) *cobra.Command {
	var tags []string
	setCmd := &cobra.Command{
		Use:   "set <organization> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Set Tags for project in an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ch.Config

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			// if tags is empty, prompt for a warning
			if len(tags) == 0 {
				ch.Printer.PrintlnWarn("Warn: Setting an empty tag list will remove all tags from the project")
				if !cmdutil.ConfirmPrompt("Do you want to continue?", "", false) {
					return nil
				}
			}

			res, err := client.SudoUpdateTags(ctx, &adminv1.SudoUpdateTagsRequest{
				Organization: args[0],
				Project:      args[1],
				Tags:         tags,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Project: %s\n", res.Project.Name)
			fmt.Printf("Organization: %s\n", res.Project.OrgName)
			fmt.Printf("Tags: %s\n", strings.Join(res.Project.Tags, ","))

			return nil
		},
	}
	setCmd.Flags().StringSliceVar(&tags, "tag", []string{}, "Tags to set on the project")

	return setCmd
}
