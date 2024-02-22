package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/telemetry"
	"github.com/spf13/cobra"
)

func verifyInstallCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "verify-install",
		Short:  "Verify installation (called by install script)",
		Hidden: !ch.IsDev(),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Emit telemetry event
			tel := telemetry.New(ch.Version)
			tel.Emit(telemetry.ActionInstallSuccess)

			// Flush telemetry with a 10s timeout
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()
			err := tel.Flush(ctx)
			if err != nil {
				fmt.Printf("Failed to verify installation: %v\n", err)
			}

			return nil
		},
	}

	return cmd
}
