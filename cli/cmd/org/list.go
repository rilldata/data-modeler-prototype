package org

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List",
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.GetSpinner(4, "Listing orgs...")
			sp.Start()
			// Just for spinner, will have to remove it
			time.Sleep(1 * time.Second)

			client, err := client.New(cfg.AdminURL, cfg.GetAdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			orgs, err := client.ListOrganizations(context.Background(), &adminv1.ListOrganizationsRequest{})
			if err != nil {
				return err
			}

			fmt.Printf("Organizations list: %v\n", orgs)
			sp.Stop()
			return nil
		},
	}

	return listCmd
}
