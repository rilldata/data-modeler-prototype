package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rilldata/rill/cli/cmd/admin"
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/cmd/billing"
	"github.com/rilldata/rill/cli/cmd/deploy"
	"github.com/rilldata/rill/cli/cmd/devtool"
	"github.com/rilldata/rill/cli/cmd/docs"
	"github.com/rilldata/rill/cli/cmd/env"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/cmd/project"
	"github.com/rilldata/rill/cli/cmd/publicurl"
	"github.com/rilldata/rill/cli/cmd/query"
	"github.com/rilldata/rill/cli/cmd/runtime"
	"github.com/rilldata/rill/cli/cmd/service"
	"github.com/rilldata/rill/cli/cmd/start"
	"github.com/rilldata/rill/cli/cmd/sudo"
	"github.com/rilldata/rill/cli/cmd/uninstall"
	"github.com/rilldata/rill/cli/cmd/upgrade"
	"github.com/rilldata/rill/cli/cmd/user"
	"github.com/rilldata/rill/cli/cmd/usergroup"
	versioncmd "github.com/rilldata/rill/cli/cmd/version"
	"github.com/rilldata/rill/cli/cmd/whoami"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/printer"
	"github.com/rilldata/rill/cli/pkg/update"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"
)

func init() {
	cobra.EnableCommandSorting = false
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "rill <command> [flags]",
	Short: "A CLI for Rill",
	Long:  `Work with Rill projects directly from the command line.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context, ver cmdutil.Version) {
	err := runCmd(ctx, ver)
	if err != nil {
		errMsg := err.Error()
		// check for known messages
		if strings.Contains(errMsg, "org not found") {
			fmt.Println("Org not found. Run `rill org list` to see the orgs. Run `rill org switch` to default org.")
		} else if strings.Contains(errMsg, "project not found") {
			fmt.Println("Project not found. Run `rill project list` to check the list of projects.")
		} else if strings.Contains(errMsg, "auth token not found") {
			fmt.Println("Auth token is invalid/expired. Login again with `rill login`.")
		} else if strings.Contains(errMsg, "not authenticated as a user") {
			fmt.Println("Please log in or sign up for Rill with `rill login`.")
		} else {
			if s, ok := status.FromError(err); ok {
				// rpc error
				fmt.Printf("Error: %s (%v)\n", s.Message(), s.Code())
			} else {
				fmt.Printf("Error: %s\n", err.Error())
			}
		}
		os.Exit(1)
	}
}

func runCmd(ctx context.Context, ver cmdutil.Version) error {
	// Create cmdutil Helper
	ch := &cmdutil.Helper{
		Printer:     printer.NewPrinter(printer.FormatHuman),
		Version:     ver,
		Interactive: true,
	}
	defer ch.Close()

	// Load base admin config from ~/.rill
	err := ch.ReloadAdminConfig()
	if err != nil {
		return err
	}

	// Load default org
	defaultOrg, err := dotrill.GetDefaultOrg()
	if err != nil {
		return fmt.Errorf("could not parse default org from ~/.rill: %w", err)
	}
	ch.Org = defaultOrg

	// Check version
	err = update.CheckVersion(ctx, ver.Number)
	if err != nil {
		ch.PrintfWarn("Warning: version check failed: %v\n\n", err)
	}

	// Print warning if currently acting as an assumed user
	representingUser, err := dotrill.GetRepresentingUser()
	if err != nil {
		ch.PrintfWarn("Could not parse representing user email\n\n")
	}
	if representingUser != "" {
		ch.PrintfWarn("Warning: Running action as %q\n\n", representingUser)
	}

	// Cobra config
	rootCmd.Version = ver.String()
	// silence usage, usage string will only show up if missing arguments/flags
	rootCmd.SilenceUsage = true
	// we want to override some error messages
	rootCmd.SilenceErrors = true
	rootCmd.PersistentFlags().BoolP("help", "h", false, "Print usage") // Overrides message for help
	rootCmd.PersistentFlags().BoolVar(&ch.Interactive, "interactive", true, "Prompt for missing required parameters")
	rootCmd.PersistentFlags().Var(&ch.Printer.Format, "format", `Output format (options: "human", "json", "csv")`)
	rootCmd.PersistentFlags().StringVar(&ch.AdminURLOverride, "api-url", ch.AdminURLOverride, "Base URL for the cloud API")
	if !ch.IsDev() {
		if err := rootCmd.PersistentFlags().MarkHidden("api-url"); err != nil {
			panic(err)
		}
	}
	rootCmd.PersistentFlags().StringVar(&ch.AdminTokenOverride, "api-token", "", "Token for authenticating with the cloud API")
	rootCmd.Flags().BoolP("version", "v", false, "Show rill version") // Adds option to get version by passing --version or -v

	// Command Groups

	// Project commands
	cmdutil.AddGroup(rootCmd, "Project", false,
		start.StartCmd(ch),
		deploy.DeployCmd(ch),
		query.QueryCmd(ch),
		project.ProjectCmd(ch),
		publicurl.PublicURLCmd(ch),
		env.EnvCmd(ch),
	)

	// Organization commands
	cmdutil.AddGroup(rootCmd, "Organization", false,
		org.OrgCmd(ch),
		user.UserCmd(ch),
		usergroup.UsergroupCmd(ch),
		service.ServiceCmd(ch),
		billing.BillingCmd(ch),
	)

	// Auth commands
	cmdutil.AddGroup(rootCmd, "Auth", false,
		auth.LoginCmd(ch),
		auth.LogoutCmd(ch),
		whoami.WhoamiCmd(ch),
	)

	// Internal commands
	cmdutil.AddGroup(rootCmd, "Internal", !ch.IsDev(),
		// These commands are hidden from the help menu
		admin.AdminCmd(ch),
		runtime.RuntimeCmd(ch),
		devtool.DevtoolCmd(ch),
		sudo.SudoCmd(ch),
		verifyInstallCmd(ch),
	)

	// Additional sub-commands
	rootCmd.AddCommand(
		completionCmd,
		docs.DocsCmd(ch, rootCmd),
		versioncmd.VersionCmd(),
		upgrade.UpgradeCmd(ch),
		uninstall.UninstallCmd(ch),
	)

	return rootCmd.ExecuteContext(ctx)
}
