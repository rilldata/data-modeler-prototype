package quota

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetCmd(cfg *config.Config) *cobra.Command {
	var org, email string
	var singleUser, projects, deployments, slotsTotal, slotsPerDeployment, outstandingInvites uint32
	setCmd := &cobra.Command{
		Use:   "set",
		Args:  cobra.NoArgs,
		Short: "Set quota for user or org",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if org != "" {
				req := &adminv1.SudoUpdateOrganizationQuotasRequest{
					OrgName: org,
				}

				if cmd.Flags().Changed("projects") {
					req.Projects = &projects
				}
				if cmd.Flags().Changed("deployments") {
					req.Deployments = &deployments
				}
				if cmd.Flags().Changed("slots-total") {
					req.SlotsTotal = &slotsTotal
				}
				if cmd.Flags().Changed("slots-per-deployment") {
					req.SlotsPerDeployment = &slotsPerDeployment
				}
				if cmd.Flags().Changed("outstanding-invites") {
					req.OutstandingInvites = &outstandingInvites
				}

				res, err := client.SudoUpdateOrganizationQuotas(ctx, req)
				if err != nil {
					return err
				}

				orgQuotas := res.Organization.Quotas
				cmdutil.PrintlnSuccess("Updated organizations quota")
				fmt.Printf("Organization Name: %s\n", org)
				fmt.Printf("Projects: %d\n", orgQuotas.Projects)
				fmt.Printf("Deployments: %d\n", orgQuotas.Deployments)
				fmt.Printf("Slots total: %d\n", orgQuotas.SlotsTotal)
				fmt.Printf("Slots per deployment: %d\n", orgQuotas.SlotsPerDeployment)
				fmt.Printf("Outstanding invites: %d\n", orgQuotas.OutstandingInvites)
			} else if email != "" {
				req := &adminv1.SudoUpdateUserQuotasRequest{
					Email: email,
				}

				if cmd.Flags().Changed("singleuser-orgs") {
					req.SingleuserOrgs = &singleUser
				}

				res, err := client.SudoUpdateUserQuotas(ctx, req)
				if err != nil {
					return err
				}

				userQuotas := res.User.Quotas
				cmdutil.PrintlnSuccess("Updated user's quota")
				fmt.Printf("User: %s\n", email)
				fmt.Printf("Single user orgs: %d\n", userQuotas.SingleuserOrgs)
			} else {
				return fmt.Errorf("Please set --org or --user")
			}

			return nil
		},
	}

	setCmd.Flags().SortFlags = false
	setCmd.Flags().StringVar(&org, "org", "", "Organization Name")
	setCmd.Flags().StringVar(&email, "user", "", "User Email")
	setCmd.Flags().Uint32Var(&singleUser, "singleuser-orgs", 0, "Quota single user org")
	setCmd.Flags().Uint32Var(&projects, "projects", 0, "Quota projects")
	setCmd.Flags().Uint32Var(&deployments, "deployments", 0, "Quota deployments")
	setCmd.Flags().Uint32Var(&slotsTotal, "slots-total", 0, "Quota slots total")
	setCmd.Flags().Uint32Var(&slotsPerDeployment, "slots-per-deployment", 0, "Quota slots per deployment")
	setCmd.Flags().Uint32Var(&outstandingInvites, "outstanding-invites", 0, "Quota outstanding invites")
	return setCmd
}
