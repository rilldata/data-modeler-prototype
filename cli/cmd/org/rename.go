package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RenameCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, newName, displayName string
	var force bool

	renameCmd := &cobra.Command{
		Use:   "rename",
		Args:  cobra.NoArgs,
		Short: "Rename organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if !cmd.Flags().Changed("org") && ch.Interactive && !force {
				orgNames, err := OrgNames(ctx, ch)
				if err != nil {
					return err
				}

				name, err = cmdutil.SelectPrompt("Select org to rename", orgNames, "")
				if err != nil {
					return err
				}
			}

			resp, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: name})
			if err != nil {
				if st, ok := status.FromError(err); ok {
					if st.Code() != codes.NotFound {
						return err
					}
				}
				fmt.Printf("Org name %q doesn't exist, please run `rill org list` to list available orgs\n", name)
				return nil
			}

			org := resp.Organization
			req := &adminv1.UpdateOrganizationRequest{
				Name: org.Name,
			}

			ch.PrintfWarn("\nWarn: Renaming an org would invalidate dashboard URLs.\n")
			hasDisplayName := org.DisplayName != ""

			if cmd.Flags().Changed("display-name") {
				req.DisplayName = &displayName
			} else if ch.Interactive && !force && hasDisplayName {
				displayName, err = cmdutil.InputPrompt("Enter the display name", org.DisplayName)
				if err != nil {
					return err
				}
				req.DisplayName = &displayName
			}

			if cmd.Flags().Changed("new-name") {
				req.NewName = &newName
				if !cmd.Flags().Changed("display-name") && org.DisplayName != "" {
					ch.PrintfWarn("Warn: Changing org name without updating display name. Consider using --display-name to keep names consistent.\n")
				}
			} else if ch.Interactive && !force {
				ok, err := confirmFieldChange("Org name", force)
				if err != nil {
					return err
				}
				if ok {
					newName, err = cmdutil.InputPrompt("Enter new org name", org.Name)
					if err != nil {
						return err
					}
					req.NewName = &newName
				}
			}

			if !cmd.Flags().Changed("new-name") && !cmd.Flags().Changed("display-name") {
				return fmt.Errorf("at least one of --new-name or --display-name flags must be provided")
			}

			updatedOrg, err := client.UpdateOrganization(ctx, req)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Renamed organization\n")
			if req.NewName != nil && *req.NewName != name {
				ch.Printf("Updated name: %s to %s\n", name, updatedOrg.Organization.Name)
			}
			if req.DisplayName != nil && *req.DisplayName != org.DisplayName {
				ch.Printf("Updated display name: %s to %s\n", org.DisplayName, updatedOrg.Organization.DisplayName)
			}

			ch.PrintOrgs([]*adminv1.Organization{updatedOrg.Organization}, "")

			if newName != "" {
				err = dotrill.SetDefaultOrg(newName)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
	renameCmd.Flags().SortFlags = false
	renameCmd.Flags().StringVar(&name, "org", ch.Org, "Current Org Name")
	renameCmd.Flags().StringVar(&newName, "new-name", "", "New Org Name")
	renameCmd.Flags().StringVar(&displayName, "display-name", "", "Org Display Name")
	renameCmd.Flags().BoolVar(&force, "force", false, "Force rename org without confirmation prompt")

	return renameCmd
}

func confirmFieldChange(field string, force bool) (bool, error) {
	if force {
		return true, nil
	}
	return cmdutil.ConfirmPrompt(fmt.Sprintf("Do you want to update the %s", field), "", false)
}
