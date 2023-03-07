package project

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(cfg *config.Config) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show",
		Args:  cobra.ExactArgs(1),
		Short: "Show",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.New(cfg.AdminURL, cfg.GetAdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			proj, err := client.GetProject(context.Background(), &adminv1.GetProjectRequest{
				Organization: cfg.DefaultOrg,
				Name:         args[0],
			})
			if err != nil {
				return err
			}

			fmt.Printf("Found project: %v\n", proj)
			return nil
		},
	}
	return showCmd
}
