package env

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
)

func ConfigureCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectPath, projectName string
	var redeploy bool

	configureCommand := &cobra.Command{
		Use:   "configure",
		Short: "Configures connector variables for all sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := ch.Client()
			if err != nil {
				return err
			}

			// If projectPath is provided, normalize it
			if projectPath != "" {
				var err error
				projectPath, err = normalizeProjectPath(projectPath)
				if err != nil {
					return err
				}
			}

			// Verify that the projectPath contains a Rill project
			if !rillv1beta.HasRillProject(projectPath) {
				ch.PrintfWarn("Directory at %q doesn't contain a valid Rill project.\n", projectPath)
				ch.PrintfWarn("Run `rill env configure` from a Rill project directory or use `--path` to pass a project path.\n")
				return nil
			}

			// Infer project name from project path
			if projectName == "" && ch.Interactive {
				var err error
				projectName, err = ch.InferProjectName(ctx, ch.Org, projectPath)
				if err != nil {
					return err
				}
			}

			variables, err := VariablesFlow(ctx, ch, projectPath)
			if err != nil {
				return fmt.Errorf("failed to get variables: %w", err)
			}

			// get existing variables
			varResp, err := client.GetProjectVariables(ctx, &adminv1.GetProjectVariablesRequest{
				OrganizationName: ch.Org,
				Name:             projectName,
			})
			if err != nil {
				return fmt.Errorf("failed to list existing variables %w", err)
			}

			if varResp.Variables == nil {
				varResp.Variables = make(map[string]string)
			}

			// update with new variables
			for key, value := range variables {
				varResp.Variables[key] = value
			}

			_, err = client.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				OrganizationName: ch.Org,
				Name:             projectName,
				Variables:        varResp.Variables,
			})
			if err != nil {
				return fmt.Errorf("failed to update variables %w", err)
			}
			ch.PrintfSuccess("Updated project variables\n")

			if redeploy {
				_, err = client.TriggerRedeploy(ctx, &adminv1.TriggerRedeployRequest{Organization: ch.Org, Project: projectName})
				if err != nil {
					ch.PrintfWarn("Redeploy trigger failed. Trigger redeploy again with `rill project reconcile --reset=true` if required.\n")
					return err
				}
				ch.PrintfSuccess("Redeploy triggered successfully.\n")
			}
			return nil
		},
	}

	configureCommand.Flags().SortFlags = false
	configureCommand.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	configureCommand.Flags().StringVar(&projectName, "project", "", "")
	configureCommand.Flags().BoolVar(&redeploy, "redeploy", false, "Redeploy project")

	return configureCommand
}

func VariablesFlow(ctx context.Context, ch *cmdutil.Helper, projectPath string) (map[string]string, error) {
	// Parse the project's connectors
	repo, instanceID, err := cmdutil.RepoForProjectPath(projectPath)
	if err != nil {
		return nil, err
	}
	parser, err := rillv1.Parse(ctx, repo, instanceID, "prod", "duckdb")
	if err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}
	connectors := parser.AnalyzeConnectors(ctx)
	for _, c := range connectors {
		if c.Err != nil {
			return nil, fmt.Errorf("failed to extract connectors: %w", err)
		}
	}

	// Remove the default DuckDB connector we always add
	for i, c := range connectors {
		if c.Name == "duckdb" {
			connectors = slices.Delete(connectors, i, i+1)
			break
		}
	}

	// Exit early if all connectors can be used anonymously
	foundNotAnonymous := false
	for _, c := range connectors {
		if !c.AnonymousAccess {
			foundNotAnonymous = true
		}
	}
	if !foundNotAnonymous {
		return nil, nil
	}

	// Emit start telemetry
	ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventDataAccessStart)

	// Start the flow
	fmt.Printf("Finish deploying your project by providing access to the connectors. Rill requires credentials for the following connectors:\n\n")
	for _, c := range connectors {
		if c.AnonymousAccess {
			continue
		}
		fmt.Printf(" - %s", c.Name)
		if len(c.Resources) == 1 {
			fmt.Printf(" (used by %s)", c.Resources[0].Name.Name)
		} else if len(c.Resources) > 1 {
			fmt.Printf(" (used by %s and others)", c.Resources[0].Name.Name)
		}
		fmt.Print("\n")
	}

	// Prompt for credentials
	variables := make(map[string]string)
	for _, c := range connectors {
		if c.AnonymousAccess {
			continue
		}
		if len(c.Spec.ConfigProperties) == 0 {
			continue
		}

		fmt.Printf("\nConfiguring connector %q:\n", c.Name)
		if c.Spec.DocsURL != "" {
			fmt.Printf("For instructions on how to configure, see: %s\n", c.Spec.DocsURL)
		}

		for i := range c.Spec.ConfigProperties {
			prop := c.Spec.ConfigProperties[i] // TODO: Move into range and turn into pointer
			if prop.NoPrompt {
				continue
			}

			key := fmt.Sprintf("connector.%s.%s", c.Name, prop.Key)
			msg := key
			if prop.Hint != "" {
				msg = fmt.Sprintf(msg+" (%s)", prop.Hint)
			}

			question := &survey.Question{}
			if prop.Secret {
				question.Prompt = &survey.Password{Message: msg}
			} else {
				question.Prompt = &survey.Input{Message: msg, Default: prop.Default}
			}

			if prop.Type == drivers.FilePropertyType {
				question.Transform = fileTransformFunc
				question.Validate = fileValidateFunc
			}

			answer := ""
			if err := survey.Ask([]*survey.Question{question}, &answer); err != nil {
				return nil, fmt.Errorf("variables prompt failed with error: %w", err)
			}

			if answer != "" {
				variables[key] = answer
			}
		}
	}

	// Emit end telemetry
	ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventDataAccessSuccess)

	// Continue with the flow
	fmt.Println("")

	return variables, nil
}

func fileValidateFunc(any interface{}) error {
	val := any.(string)
	if val == "" {
		// user can chhose to leave empty for public sources
		return nil
	}

	path, err := fileutil.ExpandHome(strings.TrimSpace(val))
	if err != nil {
		return err
	}

	_, err = os.Stat(path)
	return err
}

func fileTransformFunc(any interface{}) interface{} {
	val := any.(string)
	if val == "" {
		return ""
	}

	path, err := fileutil.ExpandHome(strings.TrimSpace(val))
	if err != nil {
		return err
	}
	// ignoring error since PathError is already validated
	content, _ := os.ReadFile(path)
	return string(content)
}
