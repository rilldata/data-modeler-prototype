package project

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(cfg *config.Config) *cobra.Command {
	var name, description, prodBranch, path string
	var public bool

	editCmd := &cobra.Command{
		Use:   "edit <project-name>",
		Args:  cobra.MaximumNArgs(1),
		Short: "Edit the project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && cfg.Interactive {
				names, err := cmdutil.ProjectNamesByOrg(ctx, client, cfg.Org)
				if err != nil {
					return err
				}

				// prompt for name from user
				name = cmdutil.SelectPrompt("Select project", names, "")
			}

			resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: cfg.Org, Name: name})
			if err != nil {
				return err
			}

			proj := resp.Project

			if cfg.Interactive {
				if !cmd.Flags().Changed("description") {
					description, err = cmdutil.InputPrompt("Enter the project description", proj.Description)
					if err != nil {
						return err
					}
				}

				if !cmd.Flags().Changed("prod-branch") {
					prodBranch, err = cmdutil.InputPrompt("Enter the production branch", proj.ProdBranch)
					if err != nil {
						return err
					}
				}

				if !cmd.Flags().Changed("public") {
					prompt := &survey.Confirm{
						Message: fmt.Sprintf("Do you want the project \"%s\" to public?", color.YellowString(name)),
					}

					err = survey.AskOne(prompt, &public)
					if err != nil {
						return err
					}
				}
			}

			// Todo: Need to add prompt for repo_path <path_for_monorepo>

			updatedProj, err := client.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
				Id:               proj.Id,
				OrganizationName: cfg.Org,
				Name:             proj.Name,
				Description:      description,
				Public:           public,
				ProdBranch:       prodBranch,
				GithubUrl:        proj.GithubUrl,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Updated project")
			cmdutil.TablePrinter(toRow(updatedProj.Project))
			return nil
		},
	}

	editCmd.Flags().SortFlags = false
	editCmd.Flags().StringVar(&name, "project", "", "Name")
	editCmd.Flags().StringVar(&description, "description", "", "Description")
	editCmd.Flags().StringVar(&prodBranch, "prod-branch", "noname", "Production branch name")
	editCmd.Flags().BoolVar(&public, "public", false, "Public")
	editCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return editCmd
}
