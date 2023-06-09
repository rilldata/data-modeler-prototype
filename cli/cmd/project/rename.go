package project

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RenameCmd(cfg *config.Config) *cobra.Command {
	var name, newName string

	renameCmd := &cobra.Command{
		Use:   "rename",
		Args:  cobra.NoArgs,
		Short: "Rename project",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			fmt.Println("Warn: Renaming an project would invalidate dashboard URLs")

			if !cmd.Flags().Changed("project") && cfg.Interactive {
				projectNames, err := cmdutil.ProjectNamesByOrg(ctx, client, cfg.Org)
				if err != nil {
					return err
				}

				name = cmdutil.SelectPrompt("Select project to rename", projectNames, "")
			}

			if cfg.Interactive {
				err = cmdutil.SetFlagsByInputPrompts(*cmd, "new-name")
				if err != nil {
					return err
				}
			}

			msg := fmt.Sprintf("Do you want to rename the project \"%s\" to \"%s\"?", color.YellowString(name), color.YellowString(newName))
			if !cmdutil.ConfirmPrompt(msg, "", false) {
				return nil
			}

			resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: cfg.Org, Name: name})
			if err != nil {
				return err
			}

			proj := resp.Project

			updatedProj, err := client.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
				Id:               proj.Id,
				OrganizationName: cfg.Org,
				Name:             newName,
				Description:      proj.Description,
				Public:           proj.Public,
				ProdBranch:       proj.ProdBranch,
				GithubUrl:        proj.GithubUrl,
				ProdSlots:        proj.ProdSlots,
				Region:           proj.Region,
			})
			if err != nil {
				return err
			}

			cmdutil.PrintlnSuccess("Renamed project")
			cmdutil.PrintlnSuccess(fmt.Sprintf("New web url is: %s\n", updatedProj.Project.FrontendUrl))
			cmdutil.TablePrinter(toRow(updatedProj.Project))

			return nil
		},
	}

	renameCmd.Flags().SortFlags = false
	renameCmd.Flags().StringVar(&name, "project", "", "Current Project Name")
	renameCmd.Flags().StringVar(&newName, "new-name", "", "New Project Name")

	return renameCmd
}
