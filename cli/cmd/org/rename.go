package org

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RenameCmd(cfg *config.Config) *cobra.Command {
	var name, newName string

	renameCmd := &cobra.Command{
		Use:   "rename",
		Args:  cobra.NoArgs,
		Short: "Rename organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			fmt.Println("Warn: Renaming an org would invalidate dashboard URLs")

			if !cmd.Flags().Changed("org") && cfg.Interactive {
				orgNames, err := cmdutil.OrgNames(ctx, client)
				if err != nil {
					return err
				}

				name = cmdutil.SelectPrompt("Select org to rename", orgNames, "")
			}

			if cfg.Interactive {
				err = cmdutil.SetFlagsByInputPrompts(*cmd, "new-name")
				if err != nil {
					return err
				}
			}

			if newName == "" {
				return fmt.Errorf("please provide valid org new-name, provided: %q", newName)
			}

			msg := fmt.Sprintf("Do you want to rename org \"%s\" to \"%s\"?", color.YellowString(name), color.YellowString(newName))
			if !cmdutil.ConfirmPrompt(msg, "", false) {
				return nil
			}

			resp, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: name})
			if err != nil {
				return err
			}

			org := resp.Organization
			updatedOrg, err := client.UpdateOrganization(ctx, &adminv1.UpdateOrganizationRequest{
				Id:          org.Id,
				Name:        newName,
				Description: org.Description,
			})
			if err != nil {
				return err
			}

			err = dotrill.SetDefaultOrg(newName)
			if err != nil {
				return err
			}

			cmdutil.PrintlnSuccess("Renamed organization")
			cmdutil.TablePrinter(toRow(updatedOrg.Organization))
			return nil
		},
	}
	renameCmd.Flags().SortFlags = false
	renameCmd.Flags().StringVar(&name, "org", cfg.Org, "Current Org Name")
	renameCmd.Flags().StringVar(&newName, "new-name", "", "New Org Name")

	return renameCmd
}
