package source

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/local"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

// addCmd represents the add command, it requires min 1 args as source name
func AddCmd(cfg *config.Config) *cobra.Command {
	var olapDriver string
	var olapDSN string
	var projectPath string
	var sourceName string
	var force bool
	var verbose bool

	addCmd := &cobra.Command{
		Use:   "add <file>",
		Short: "Add a local file source",
		Long:  "Add a local file source. Supported file types include .parquet, .csv, .tsv.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataPath := args[0]
			if !filepath.IsAbs(dataPath) {
				relPath, err := filepath.Rel(projectPath, dataPath)
				if err != nil {
					return err
				}
				dataPath = relPath
			}

			app, err := local.NewApp(cmd.Context(), cfg.Version, verbose, olapDriver, olapDSN, projectPath, local.LogFormatConsole, nil)
			if err != nil {
				return err
			}
			defer app.Close()

			if !app.IsProjectInit() {
				return fmt.Errorf("not a valid Rill project")
			}

			if sourceName == "" {
				sourceName = fileutil.Stem(dataPath)
			}

			props := map[string]any{"path": dataPath}

			propsPB, err := structpb.NewStruct(props)
			if err != nil {
				return fmt.Errorf("can't serialize artifact: %w", err)
			}

			src := &runtimev1.Source{
				Name:       artifacts.SanitizedName(sourceName),
				Connector:  "local_file",
				Properties: propsPB,
			}

			repo, err := app.Runtime.Repo(cmd.Context(), app.Instance.ID)
			if err != nil {
				panic(err) // Should never happen
			}

			c := rillv1beta.New(repo, app.Instance.ID)
			sourcePath, err := c.PutSource(cmd.Context(), repo, app.Instance.ID, src, force)
			if err != nil {
				if os.IsExist(err) {
					return fmt.Errorf("source already exists (pass --force to overwrite)")
				}
				return fmt.Errorf("write source: %w", err)
			}

			err = app.ReconcileSource(sourcePath)
			if err != nil {
				return fmt.Errorf("reconcile source: %w", err)
			}

			return nil
		},
	}

	addCmd.Flags().SortFlags = false
	addCmd.Flags().StringVar(&sourceName, "name", "", "Source name (defaults to file name)")
	addCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite the source if it already exists")
	addCmd.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	addCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "Database DSN")
	addCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "Database driver")
	addCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")

	return addCmd
}
