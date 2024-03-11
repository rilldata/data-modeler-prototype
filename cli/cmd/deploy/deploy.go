package deploy

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-git/go-git/v5"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/deviceauth"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/telemetry"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	pollTimeout  = 10 * time.Minute
	pollInterval = 5 * time.Second
)

// DeployCmd is the guided tour for deploying rill projects to rill cloud.
func DeployCmd(ch *cmdutil.Helper) *cobra.Command {
	opts := &Options{}

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy project to Rill Cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			return DeployFlow(cmd.Context(), ch, opts)
		},
	}

	deployCmd.Flags().SortFlags = false
	deployCmd.Flags().StringVar(&opts.GitPath, "path", ".", "Path to project repository (default: current directory)") // This can also be a remote .git URL (undocumented feature)
	deployCmd.Flags().StringVar(&opts.SubPath, "subpath", "", "Relative path to project in the repository (for monorepos)")
	deployCmd.Flags().StringVar(&opts.RemoteName, "remote", "", "Remote name (default: first Git remote)")
	deployCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Org to deploy project in")
	deployCmd.Flags().StringVar(&opts.Name, "project", "", "Project name (default: Git repo name)")
	deployCmd.Flags().StringVar(&opts.Description, "description", "", "Project description")
	deployCmd.Flags().BoolVar(&opts.Public, "public", false, "Make dashboards publicly accessible")
	deployCmd.Flags().StringVar(&opts.Region, "region", "", "Deployment region")
	deployCmd.Flags().StringVar(&opts.ProdBranch, "prod-branch", "", "Git branch to deploy from (default: the default Git branch)")
	deployCmd.Flags().IntVar(&opts.Slots, "prod-slots", 2, "Slots to allocate for production deployments")
	if !ch.IsDev() {
		if err := deployCmd.Flags().MarkHidden("prod-slots"); err != nil {
			panic(err)
		}
	}

	// 2024-02-19: We have deprecated configuration of the OLAP DB using flags in favor of using rill.yaml.
	// When the migration is complete, we can remove the flags as well as the admin-server support for them.
	deployCmd.Flags().StringVar(&opts.DBDriver, "prod-db-driver", "duckdb", "Database driver")
	deployCmd.Flags().StringVar(&opts.DBDSN, "prod-db-dsn", "", "Database driver configuration")
	if err := deployCmd.Flags().MarkHidden("prod-db-driver"); err != nil {
		panic(err)
	}
	if err := deployCmd.Flags().MarkHidden("prod-db-dsn"); err != nil {
		panic(err)
	}

	return deployCmd
}

type Options struct {
	GitPath     string
	SubPath     string
	RemoteName  string
	Name        string
	Description string
	Public      bool
	Region      string
	ProdBranch  string
	DBDriver    string
	DBDSN       string
	Slots       int
}

func DeployFlow(ctx context.Context, ch *cmdutil.Helper, opts *Options) error {
	// Setup telemetry
	tel := telemetry.New(ch.Version)
	if ch.IsAuthenticated() {
		user, _ := ch.CurrentUser(ctx)
		if user != nil {
			tel.WithUserID(user.Id)
		}
	}
	tel.Emit(telemetry.ActionDeployStart)
	defer func() {
		// give 5s for emitting events over the parent context.
		// this will make sure if user cancelled the command events are still fired.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// telemetry errors shouldn't fail deploy command
		_ = tel.Flush(ctx)
	}()

	// The gitPath can be either a local path or a remote .git URL.
	// Determine which it is.
	var isLocalGitPath bool
	var githubURL string
	if opts.GitPath != "" {
		u, err := url.Parse(opts.GitPath)
		if err != nil || u.Scheme == "" {
			isLocalGitPath = true
		} else {
			githubURL, err = gitutil.RemoteToGithubURL(opts.GitPath)
			if err != nil {
				return fmt.Errorf("failed to parse path as a Github remote: %w", err)
			}
		}
	}

	// If the Git path is local, we'll do some extra steps to infer the githubURL.
	var localGitPath, localProjectPath string
	if isLocalGitPath {
		var err error
		if opts.GitPath != "" {
			localGitPath, err = fileutil.ExpandHome(opts.GitPath)
			if err != nil {
				return err
			}
		}
		localGitPath, err = filepath.Abs(localGitPath)
		if err != nil {
			return err
		}

		if opts.SubPath == "" {
			localProjectPath = localGitPath
		} else {
			localProjectPath = filepath.Join(localGitPath, opts.SubPath)
		}

		// Verify that localProjectPath contains a Rill project.
		// If not, we still navigate user to login and then fail afterwards.
		if !rillv1beta.HasRillProject(localProjectPath) {
			if !ch.IsAuthenticated() {
				err := loginWithTelemetry(ctx, ch, "", tel)
				if err != nil {
					ch.PrintfWarn("Login failed with error: %s\n", err.Error())
				}
				fmt.Println()
			}

			ch.PrintfWarn("Directory %q doesn't contain a valid Rill project.\n", localProjectPath)
			ch.PrintfWarn("Run `rill deploy` from a Rill project directory or use `--path` to pass a project path.\n")
			ch.PrintfWarn("Run `rill start` to initialize a new Rill project.\n")
			return nil
		}

		// Extract the Git remote and infer the githubURL.
		var remote *gitutil.Remote
		remote, githubURL, err = gitutil.ExtractGitRemote(localGitPath, opts.RemoteName, false)
		if err != nil {
			// It's not a valid remote for Github. But we still navigate user to login and then fail afterwards.
			if !ch.IsAuthenticated() {
				err := loginWithTelemetry(ctx, ch, "", tel)
				if err != nil {
					ch.PrintfWarn("Login failed with error: %s\n", err.Error())
				}
				fmt.Println()
			}
			if errors.Is(err, gitutil.ErrGitRemoteNotFound) || errors.Is(err, git.ErrRepositoryNotExists) {
				ch.PrintfBold(githubSetupMsg)
				return nil
			}
			return err
		}

		// Error if the repository is not in sync with the remote
		if !repoInSyncFlow(ch, localGitPath, opts.ProdBranch, remote.Name) {
			ch.PrintfBold("You can run `rill deploy` again when you have pushed your local changes to the remote.\n")
			return nil
		}
	}

	// We now have a githubURL.

	// Extract Github account and repo name from the githubURL
	ghAccount, ghRepo, ok := gitutil.SplitGithubURL(githubURL)
	if !ok {
		ch.PrintfError("Invalid Github URL %q\n", githubURL)
		return nil
	}

	// If user is not authenticated, run login flow.
	// To prevent opening the browser twice, we make it directly redirect to the Github flow.
	silentGitFlow := false
	if !ch.IsAuthenticated() {
		silentGitFlow = true
		authURL := ch.AdminURL
		if strings.Contains(authURL, "http://localhost:9090") {
			authURL = "http://localhost:8080"
		}
		redirectURL, err := urlutil.WithQuery(urlutil.MustJoinURL(authURL, "/github/post-auth-redirect"), map[string]string{"remote": githubURL})
		if err != nil {
			return err
		}
		if err := loginWithTelemetry(ctx, ch, redirectURL, tel); err != nil {
			return err
		}
	}

	client, err := ch.Client()
	if err != nil {
		return err
	}

	// Run flow for access to the Github remote (if necessary)
	ghRes, err := githubFlow(ctx, ch, githubURL, silentGitFlow, tel)
	if err != nil {
		return fmt.Errorf("failed Github flow: %w", err)
	}

	if opts.ProdBranch == "" {
		opts.ProdBranch = ghRes.DefaultBranch
	}

	// If no project name was provided, default to Git repo name
	if opts.Name == "" {
		opts.Name = ghRepo
	}

	// Set a default org for the user if necessary
	// (If user is not in an org, we'll create one based on their Github account later in the flow.)
	if ch.Org == "" {
		res, err := client.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		if err != nil {
			return fmt.Errorf("listing orgs failed: %w", err)
		}

		if len(res.Organizations) == 1 {
			ch.Org = res.Organizations[0].Name
			if err := dotrill.SetDefaultOrg(ch.Org); err != nil {
				return err
			}
		} else if len(res.Organizations) > 1 {
			orgName, err := org.SwitchSelectFlow(res.Organizations)
			if err != nil {
				return fmt.Errorf("org selection failed %w", err)
			}

			ch.Org = orgName
			if err := dotrill.SetDefaultOrg(ch.Org); err != nil {
				return err
			}
		}
	}

	// If no default org is set by now, it means the user is not in an org yet.
	// We create a default org based on their Github account name.
	if ch.Org == "" {
		err := createOrgFlow(ctx, ch, ghAccount)
		if err != nil {
			return fmt.Errorf("org creation failed with error: %w", err)
		}
		ch.PrintfSuccess("Created org %q. Run `rill org edit` to change name if required.\n\n", ch.Org)
	} else {
		ch.PrintfBold("Using org %q.\n\n", ch.Org)
	}

	// Check if a project matching githubURL already exists in this org
	projectNameExists := false
	projects, err := ch.ProjectNamesByGithubURL(ctx, ch.Org, githubURL)
	if err == nil && len(projects) != 0 { // ignoring error since this is just for a confirmation prompt
		for _, p := range projects {
			if strings.EqualFold(opts.Name, p) {
				projectNameExists = true
				break
			}
		}

		ch.PrintfWarn("Another project %q already deploys from %q.\n", projects[0], githubURL)
		ch.PrintfBold("- To force the existing project to rebuild, press 'n' and run `rill project reconcile --reset`\n")
		ch.PrintfBold("- To delete the existing project, press 'n' and run `rill project delete`\n")
		ch.PrintfBold("- To deploy the repository as a new project under another name, press 'y' or enter\n")
		if !cmdutil.ConfirmPrompt("Do you want to continue?", "", true) {
			ch.PrintfWarn("Aborted\n")
			return nil
		}
	}

	// If the project name already exists, prompt for another name
	if projectNameExists {
		opts.Name, err = projectNamePrompt(ctx, ch, ch.Org)
		if err != nil {
			return err
		}
	}

	// Create the project (automatically deploys prod branch)
	res, err := createProjectFlow(ctx, ch, &adminv1.CreateProjectRequest{
		OrganizationName: ch.Org,
		Name:             opts.Name,
		Description:      opts.Description,
		Region:           opts.Region,
		ProdOlapDriver:   opts.DBDriver,
		ProdOlapDsn:      opts.DBDSN,
		ProdSlots:        int64(opts.Slots),
		Subpath:          opts.SubPath,
		ProdBranch:       opts.ProdBranch,
		Public:           opts.Public,
		GithubUrl:        githubURL,
	})
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.PermissionDenied {
			ch.PrintfError("You do not have the permissions needed to create a project in org %q. Please reach out to your Rill admin.\n", ch.Org)
			return nil
		}
		return fmt.Errorf("create project failed with error %w", err)
	}

	// Success!
	ch.PrintfSuccess("Created project \"%s/%s\". Use `rill project rename` to change name if required.\n\n", ch.Org, res.Project.Name)
	ch.PrintfSuccess("Rill projects deploy continuously when you push changes to Github.\n")

	// If the Git path is local, we can parse the project and check if credentials are available for the connectors used by the project.
	if isLocalGitPath {
		variablesFlow(ctx, ch, localProjectPath, opts.SubPath, opts.Name)
	}

	// Open browser
	if res.Project.FrontendUrl != "" {
		ch.PrintfSuccess("Your project can be accessed at: %s\n", res.Project.FrontendUrl)
		ch.PrintfSuccess("Opening project in browser...\n")
		time.Sleep(3 * time.Second)
		_ = browser.Open(res.Project.FrontendUrl)
	}

	tel.Emit(telemetry.ActionDeploySuccess)
	return nil
}

func loginWithTelemetry(ctx context.Context, ch *cmdutil.Helper, redirectURL string, tel *telemetry.Telemetry) error {
	ch.PrintfBold("Please log in or sign up for Rill. Opening browser...\n")
	time.Sleep(2 * time.Second)

	tel.Emit(telemetry.ActionLoginStart)
	if err := auth.Login(ctx, ch, redirectURL); err != nil {
		if errors.Is(err, deviceauth.ErrAuthenticationTimedout) {
			ch.PrintfWarn("Rill login has timed out as the code was not confirmed in the browser.\n")
			ch.PrintfWarn("Run `rill deploy` again.\n")
			return nil
		} else if errors.Is(err, deviceauth.ErrCodeRejected) {
			ch.PrintfError("Login failed: Confirmation code rejected\n")
			return nil
		}
		return fmt.Errorf("login failed: %w", err)
	}
	user, _ := ch.CurrentUser(ctx)
	if user != nil {
		tel.WithUserID(user.Id)
	}
	// fire this after we potentially get the user id
	tel.Emit(telemetry.ActionLoginSuccess)
	return nil
}

func githubFlow(ctx context.Context, ch *cmdutil.Helper, githubURL string, silent bool, tel *telemetry.Telemetry) (*adminv1.GetGithubRepoStatusResponse, error) {
	// Get the admin client
	c, err := ch.Client()
	if err != nil {
		return nil, err
	}

	// Check for access to the Github repo
	res, err := c.GetGithubRepoStatus(ctx, &adminv1.GetGithubRepoStatusRequest{
		GithubUrl: githubURL,
	})
	if err != nil {
		return nil, err
	}

	// If the user has not already granted access, open browser and poll for access
	if !res.HasAccess {
		tel.Emit(telemetry.ActionGithubConnectedStart)

		// Print instructions to grant access
		if !silent {
			ch.Print("Rill projects deploy continuously when you push changes to Github.\n")
			ch.Print("You need to grant Rill read only access to your repository on Github.\n\n")
			time.Sleep(3 * time.Second)
			ch.Print("Open this URL in your browser to grant Rill access to Github:\n\n")
			ch.Print("\t" + res.GrantAccessUrl + "\n\n")

			// Open browser if possible
			_ = browser.Open(res.GrantAccessUrl)
		} else {
			ch.Printf("Polling for Github access for: %q\n", githubURL)
			ch.Printf("If the browser did not redirect, visit this URL to grant access: %q\n\n", res.GrantAccessUrl)
		}

		// Poll for permission granted
		pollCtx, cancel := context.WithTimeout(ctx, pollTimeout)
		defer cancel()
		for {
			select {
			case <-pollCtx.Done():
				return nil, pollCtx.Err()
			case <-time.After(pollInterval):
				// Ready to check again.
			}

			// Poll for access to the Github URL
			pollRes, err := c.GetGithubRepoStatus(ctx, &adminv1.GetGithubRepoStatusRequest{
				GithubUrl: githubURL,
			})
			if err != nil {
				return nil, err
			}

			if pollRes.HasAccess {
				// Success
				tel.Emit(telemetry.ActionGithubConnectedSuccess)
				_, ghRepo, _ := gitutil.SplitGithubURL(githubURL)
				ch.PrintfSuccess("You have connected to the %q project in Github.\n", ghRepo)
				return pollRes, nil
			}

			// Sleep and poll again
		}
	}

	return res, nil
}

func createOrgFlow(ctx context.Context, ch *cmdutil.Helper, defaultName string) error {
	c, err := ch.Client()
	if err != nil {
		return err
	}

	res, err := c.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
		Name: defaultName,
	})
	if err != nil {
		if !errMsgContains(err, "an org with that name already exists") {
			return err
		}

		ch.PrintfWarn("Rill organizations are derived from the owner of your Github repository.\n")
		ch.PrintfWarn("The %q organization associated with your Github repository already exists.\n", defaultName)
		ch.PrintfWarn("Contact your Rill admin to be added to your org or create a new organization below.\n")

		name, err := orgNamePrompt(ctx, ch)
		if err != nil {
			return err
		}

		res, err = c.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
			Name: name,
		})
		if err != nil {
			return err
		}
	}

	// Switching to the created org
	ch.Org = res.Organization.Name
	err = dotrill.SetDefaultOrg(ch.Org)
	if err != nil {
		return err
	}

	return nil
}

func orgNamePrompt(ctx context.Context, ch *cmdutil.Helper) (string, error) {
	qs := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Enter an org name",
			},
			Validate: func(any interface{}) error {
				// Validate org name doesn't exist already
				name := any.(string)
				if name == "" {
					return fmt.Errorf("empty name")
				}

				exists, err := orgExists(ctx, ch, name)
				if err != nil {
					return fmt.Errorf("org name %q is already taken", name)
				}

				if exists {
					// this should always be true but adding this check from completeness POV
					return fmt.Errorf("org with name %q already exists", name)
				}
				return nil
			},
		},
	}

	name := ""
	if err := survey.Ask(qs, &name); err != nil {
		return "", err
	}

	return name, nil
}

func orgExists(ctx context.Context, ch *cmdutil.Helper, name string) (bool, error) {
	c, err := ch.Client()
	if err != nil {
		return false, err
	}

	_, err = c.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: name})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func createProjectFlow(ctx context.Context, ch *cmdutil.Helper, req *adminv1.CreateProjectRequest) (*adminv1.CreateProjectResponse, error) {
	// Get the admin client
	c, err := ch.Client()
	if err != nil {
		return nil, err
	}

	// Create the project (automatically deploys prod branch)
	res, err := c.CreateProject(ctx, req)
	if err != nil {
		if !errMsgContains(err, "a project with that name already exists in the org") {
			return nil, err
		}

		ch.PrintfWarn("Rill project names are derived from your Github repository name.\n")
		ch.PrintfWarn("The %q project already exists under org %q. Please enter a different name.\n", req.Name, req.OrganizationName)

		// project name already exists, prompt for project name and create project with new name again
		name, err := projectNamePrompt(ctx, ch, req.OrganizationName)
		if err != nil {
			return nil, err
		}

		req.Name = name
		return c.CreateProject(ctx, req)
	}
	return res, err
}

func variablesFlow(ctx context.Context, ch *cmdutil.Helper, gitPath, subPath, projectName string) {
	// Parse the project's connectors
	repo, instanceID, err := cmdutil.RepoForProjectPath(gitPath)
	if err != nil {
		return
	}
	parser, err := rillv1.Parse(ctx, repo, instanceID, "prod", "duckdb")
	if err != nil {
		return
	}
	connectors, err := parser.AnalyzeConnectors(ctx)
	if err != nil {
		return
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
		return
	}

	ch.PrintfWarn("\nCould not access all connectors. Rill requires credentials for the following connectors:\n\n")
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
	if subPath == "" {
		ch.PrintfWarn("\nRun `rill env configure --project %s` to provide credentials.\n\n", projectName)
	} else {
		ch.PrintfWarn("\nRun `rill env configure --project %s` from directory `%s` to provide credentials.\n\n", projectName, gitPath)
	}
	time.Sleep(2 * time.Second)
}

func repoInSyncFlow(ch *cmdutil.Helper, gitPath, branch, remoteName string) bool {
	syncStatus, err := gitutil.GetSyncStatus(gitPath, branch, remoteName)
	if err != nil {
		// ignore errors since check is best effort and can fail in multiple cases
		return true
	}

	switch syncStatus {
	case gitutil.SyncStatusUnspecified:
		return true
	case gitutil.SyncStatusSynced:
		return true
	case gitutil.SyncStatusModified:
		ch.PrintfWarn("Some files have been locally modified. These changes will not be present in the deployed project.\n")
	case gitutil.SyncStatusAhead:
		ch.PrintfWarn("Local commits are not pushed to remote yet. These changes will not be present in the deployed project.\n")
	}

	return cmdutil.ConfirmPrompt("Do you want to continue", "", true)
}

func projectNamePrompt(ctx context.Context, ch *cmdutil.Helper, orgName string) (string, error) {
	questions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Enter a project name",
			},
			Validate: func(any interface{}) error {
				name := any.(string)
				if name == "" {
					return fmt.Errorf("empty name")
				}
				exists, err := projectExists(ctx, ch, orgName, name)
				if err != nil {
					return fmt.Errorf("project already exists at %s/%s", orgName, name)
				}
				if exists {
					// this should always be true but adding this check from completeness POV
					return fmt.Errorf("project with name %q already exists in the org", name)
				}
				return nil
			},
		},
	}

	name := ""
	if err := survey.Ask(questions, &name); err != nil {
		return "", err
	}

	return name, nil
}

func projectExists(ctx context.Context, ch *cmdutil.Helper, orgName, projectName string) (bool, error) {
	c, err := ch.Client()
	if err != nil {
		return false, err
	}

	_, err = c.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: orgName, Name: projectName})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func errMsgContains(err error, msg string) bool {
	if st, ok := status.FromError(err); ok && st != nil {
		return strings.Contains(st.Message(), msg)
	}
	return false
}

const (
	githubSetupMsg = `No git remote was found.

Rill projects deploy continuously when you push changes to Github.
Therefore, your project must be on Github before you deploy it to Rill.

Follow these steps to push your project to Github.
	
1. Initialize git

	git init

2. Add and commit files
	
	git add .
	git commit -m 'initial commit'

3. Create a new GitHub repository on https://github.com/new

4. Link git to the remote repository
	
	git remote add origin https://github.com/your-account/your-repo.git
	
5. Rename master branch to main
	
	git branch -M main

6. Push your repository
	
	git push -u origin main
	
7. Deploy Rill to your repository
	
	rill deploy
	
`
)
