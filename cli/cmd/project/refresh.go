package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/spf13/cobra"
)

func RefreshCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string
	var local bool
	var models, modelSplits, sources, alerts, reports []string
	var all, full, erroredSplits, parser bool

	refreshCmd := &cobra.Command{
		Use:               "refresh [<project-name>]",
		Args:              cobra.MaximumNArgs(1),
		Short:             "Refresh one or more resources",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine project name
			if len(args) > 0 {
				project = args[0]
			}
			if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				var err error
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return err
				}
			}

			// Connect to the runtime
			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, project, local)
			if err != nil {
				return err
			}

			// If only meta flags are set, default to an incremental refresh of all sources and models.
			var numMetaFlags int
			if cmd.Flags().Changed("project") {
				numMetaFlags++
			}
			if cmd.Flags().Changed("path") {
				numMetaFlags++
			}
			if cmd.Flags().Changed("local") {
				numMetaFlags++
			}
			if numMetaFlags == cmd.Flags().NFlag() {
				all = true
			}

			// Build non-model resources
			var resources []*runtimev1.ResourceName
			for _, s := range sources {
				resources = append(resources, &runtimev1.ResourceName{Kind: runtime.ResourceKindSource, Name: s})
			}
			for _, a := range alerts {
				resources = append(resources, &runtimev1.ResourceName{Kind: runtime.ResourceKindAlert, Name: a})
			}
			for _, r := range reports {
				resources = append(resources, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: r})
			}

			// Build model triggers
			if len(modelSplits) > 0 && len(models) != 1 {
				return fmt.Errorf("must specify exactly one --model when using --split")
			}
			if erroredSplits && len(models) != 1 {
				return fmt.Errorf("must specify exactly one --model when using --errored-splits")
			}
			var modelTriggers []*runtimev1.RefreshModelTrigger
			for _, m := range models {
				modelTriggers = append(modelTriggers, &runtimev1.RefreshModelTrigger{
					Model:            m,
					Full:             full,
					AllErroredSplits: erroredSplits,
					Splits:           modelSplits,
				})
			}

			// Send request
			_, err = rt.CreateTrigger(cmd.Context(), &runtimev1.CreateTriggerRequest{
				InstanceId:           instanceID,
				Resources:            resources,
				Models:               modelTriggers,
				Parser:               parser,
				AllSourcesModels:     all && !full,
				AllSourcesModelsFull: all && full,
			})
			if err != nil {
				return fmt.Errorf("failed to create trigger: %w", err)
			}

			// Print status
			if local {
				ch.Printf("Refresh initiated. Check the project logs for status updates.\n")
			} else {
				ch.Printf("Refresh initiated. To check the status, run `rill project status` or `rill project logs`.\n")
			}

			return nil
		},
	}

	refreshCmd.Flags().SortFlags = false
	refreshCmd.Flags().StringVar(&project, "project", "", "Project name")
	refreshCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	refreshCmd.Flags().BoolVar(&local, "local", false, "Target locally running Rill")
	refreshCmd.Flags().BoolVar(&all, "all", false, "Refresh all sources and models (default)")
	refreshCmd.Flags().BoolVar(&full, "full", false, "Fully reload the targeted models (use with --all or --model)")
	refreshCmd.Flags().StringSliceVar(&models, "model", nil, "Refresh a model")
	refreshCmd.Flags().StringSliceVar(&modelSplits, "split", nil, "Refresh a model split (must set --model)")
	refreshCmd.Flags().BoolVar(&erroredSplits, "errored-splits", false, "Refresh all model splits with errors (must set --model)")
	refreshCmd.Flags().StringSliceVar(&sources, "source", nil, "Refresh a source")
	refreshCmd.Flags().StringSliceVar(&alerts, "alert", nil, "Refresh an alert")
	refreshCmd.Flags().StringSliceVar(&reports, "report", nil, "Refresh a report")
	refreshCmd.Flags().BoolVar(&parser, "parser", false, "Refresh the parser (forces a pull from Github)")

	return refreshCmd
}
