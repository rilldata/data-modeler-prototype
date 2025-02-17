package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var prodSlots int

	editCmd := &cobra.Command{
		Use:   "edit <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Edit the project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			req := &adminv1.UpdateProjectRequest{
				OrganizationName:     args[0],
				Name:                 args[1],
				SuperuserForceAccess: true,
			}

			isEditRequested := false
			if cmd.Flags().Changed("prod-slots") {
				prodSlotsInt64 := int64(prodSlots)
				req.ProdSlots = &prodSlotsInt64
				isEditRequested = true
			}

			if !isEditRequested {
				ch.Printf("No edit requested\n")
				return nil
			}

			if *req.ProdSlots <= 0 {
				return fmt.Errorf("--prod-slots must be greater than zero")
			}

			client, err := ch.Client()
			if err != nil {
				return err
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

	editCmd.Flags().IntVar(&prodSlots, "prod-slots", 0, "Slots to allocate for production deployments")
	return editCmd
}
