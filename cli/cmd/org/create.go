package org

import (
	"context"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(cfg *config.Config) *cobra.Command {
	var description string

	createCmd := &cobra.Command{
		Use:   "create <org-name>",
		Short: "Create",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.CreateOrganization(context.Background(), &adminv1.CreateOrganizationRequest{
				Name:        args[0],
				Description: description,
			})
			if err != nil {
				return err
			}

			// Switching to the created org
			err = dotrill.SetDefaultOrg(res.Organization.Name)
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Created organization \n")
			cmdutil.TablePrinter(toRow(res.Organization))
			return nil
		},
	}
	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&description, "description", "", "Description")

	return createCmd
}
