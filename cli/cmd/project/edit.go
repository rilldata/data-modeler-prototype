package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, description, prodVersion, prodBranch, subpath, path, provisioner string
	var public bool
	var prodTTL int64

	editCmd := &cobra.Command{
		Use:   "edit [<project-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Edit the project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				names, err := projectNames(ctx, ch)
				if err != nil {
					return err
				}

				// prompt for name from user
				name, err = cmdutil.SelectPrompt("Select project", names, "")
				if err != nil {
					return err
				}
			}
			if name == "" {
				return fmt.Errorf("pass project name as argument or with --project flag")
			}

			req := &adminv1.UpdateProjectRequest{
				OrganizationName: ch.Org,
				Name:             name,
			}
			promptFlagValues := ch.Interactive
			if cmd.Flags().Changed("provisioner") {
				promptFlagValues = false
				req.Provisioner = &provisioner
			}
			if cmd.Flags().Changed("description") {
				promptFlagValues = false
				req.Description = &description
			}
			if cmd.Flags().Changed("prod-version") {
				promptFlagValues = false
				req.ProdVersion = &prodVersion
			}
			if cmd.Flags().Changed("prod-branch") {
				promptFlagValues = false
				req.ProdBranch = &prodBranch
			}
			if cmd.Flags().Changed("subpath") {
				promptFlagValues = false
				req.Subpath = &subpath
			}
			if cmd.Flags().Changed("public") {
				promptFlagValues = false
				req.Public = &public
			}

			if cmd.Flags().Changed("prod-ttl-seconds") {
				promptFlagValues = false
				req.ProdTtlSeconds = &prodTTL
			}

			if promptFlagValues {
				resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: ch.Org, Name: name})
				if err != nil {
					return err
				}
				proj := resp.Project

				description, err = cmdutil.InputPrompt("Enter the description", proj.Description)
				if err != nil {
					return err
				}
				req.Description = &description

				prodBranch, err = cmdutil.InputPrompt("Enter the production branch", proj.ProdBranch)
				if err != nil {
					return err
				}
				req.ProdBranch = &prodBranch

				public, err = cmdutil.ConfirmPrompt("Make project public", "", proj.Public)
				if err != nil {
					return err
				}
				req.Public = &public
			}

			updatedProj, err := client.UpdateProject(ctx, req)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Updated project\n")
			ch.PrintProjects([]*adminv1.Project{updatedProj.Project})

			return nil
		},
	}

	editCmd.Flags().SortFlags = false
	editCmd.Flags().StringVar(&name, "project", "", "Project Name")
	editCmd.Flags().StringVar(&description, "description", "", "Project Description")
	editCmd.Flags().StringVar(&prodBranch, "prod-branch", "", "Production branch name")
	editCmd.Flags().BoolVar(&public, "public", false, "Make dashboards publicly accessible")
	editCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	editCmd.Flags().StringVar(&subpath, "subpath", "", "Relative path to project in the repository (for monorepos)")
	editCmd.Flags().StringVar(&provisioner, "provisioner", "", "Project provisioner (default: current provisioner)")
	editCmd.Flags().Int64Var(&prodTTL, "prod-ttl-seconds", 0, "Prod deployment TTL in seconds")
	editCmd.Flags().StringVar(&prodVersion, "prod-version", "", "Rill version (default: current version)")

	return editCmd
}
