package cmdutil

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/deviceauth"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/dotrillcloud"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/printer"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultAdminURL = "https://admin.rilldata.com"

	telemetryServiceName    = "cli"
	telemetryIntakeURL      = "https://intake.rilldata.io/events/data-modeler-metrics"
	telemetryIntakeUser     = "data-modeler"
	telemetryIntakePassword = "lkh8T90ozWJP/KxWnQ81PexRzpdghPdzuB0ly2/86TeUU8q/bKiVug==" // nolint:gosec // secret is safe for public use
)

var ErrInvalidProject = errors.New("invalid project")

type Helper struct {
	*printer.Printer
	Version            Version
	Interactive        bool
	Org                string
	AdminURLDefault    string
	AdminURLOverride   string
	AdminTokenDefault  string
	AdminTokenOverride string

	adminClient        *client.Client
	adminClientHash    string
	activityClient     *activity.Client
	activityClientHash string
}

func (h *Helper) Close() error {
	grp := errgroup.Group{}

	if h.adminClient != nil {
		grp.Go(h.adminClient.Close)
	}

	if h.activityClient != nil {
		grp.Go(func() error {
			// We'll give ourselves 5s to flush any remaining events.
			// We don't use the cmd context because it might already be cancelled.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// We don't return the error because telemetry errors shouldn't become user-facing errors.
			_ = h.activityClient.Close(ctx)
			return nil
		})
	}

	return grp.Wait()
}

func (h *Helper) IsDev() bool {
	return h.Version.IsDev()
}

func (h *Helper) IsAuthenticated() bool {
	return h.AdminToken() != ""
}

// ReloadAdminConfig populates the helper's AdminURL, AdminTokenDefault, and Org properties from ~/.rill.
func (h *Helper) ReloadAdminConfig() error {
	adminToken, err := dotrill.GetAccessToken()
	if err != nil {
		return fmt.Errorf("could not parse access token from ~/.rill: %w", err)
	}

	adminURL, err := dotrill.GetDefaultAdminURL()
	if err != nil {
		return fmt.Errorf("could not parse default api URL from ~/.rill: %w", err)
	}
	if adminURL == "" {
		adminURL = defaultAdminURL
	}

	h.AdminURLDefault = adminURL
	h.AdminTokenDefault = adminToken

	return nil
}

func (h *Helper) AdminToken() string {
	if h.AdminTokenOverride != "" {
		return h.AdminTokenOverride
	}
	return h.AdminTokenDefault
}

func (h *Helper) AdminURL() string {
	if h.AdminURLOverride != "" {
		return h.AdminURLOverride
	}
	return h.AdminURLDefault
}

func (h *Helper) Client() (*client.Client, error) {
	// The admin token and URL may have changed (e.g. if the user did a separate login or env switch).
	// Reload the admin config from disk to get the latest values.
	err := h.ReloadAdminConfig()
	if err != nil {
		return nil, err
	}

	// Compute and cache a hash of the admin config values to detect changes.
	// If the hash has changed, we should close the existing client.
	hash := hashStr(h.AdminToken(), h.AdminURL())
	if h.adminClient != nil && h.adminClientHash != hash {
		_ = h.adminClient.Close()
		h.adminClient = nil
		h.adminClientHash = hash
	}
	h.adminClientHash = hash

	// Make a new client if we don't have one.
	if h.adminClient == nil {
		cliVersion := h.Version.Number
		if cliVersion == "" {
			cliVersion = "unknown"
		}

		userAgent := fmt.Sprintf("rill-cli/%v", cliVersion)
		c, err := client.New(h.AdminURL(), h.AdminToken(), userAgent)
		if err != nil {
			return nil, err
		}

		h.adminClient = c
	}

	return h.adminClient, nil
}

// Telemetry returns a client for recording events.
// Note: It should only be used for parts of the CLI that run on users' local computer because:
// a) it accesses ~/.rill and adds information about the current user,
// b) it sends events to the public intake endpoint instead of directly to Kafka.
func (h *Helper) Telemetry(ctx context.Context) *activity.Client {
	// If the admin token or URL changes, the user ID of the telemetry client may have changed.
	// We compute and cache a hash of these values to detect changes.
	// If the hash has changed, we refetch the current user and update the client.
	hash := hashStr(h.AdminToken(), h.AdminURL())

	// Return the client if it's already created and the hash hasn't changed.
	if h.activityClient != nil && h.activityClientHash == hash {
		return h.activityClient
	}

	// Now we can update the hash. The user ID will be refetched below.
	h.activityClientHash = hash

	// Load telemetry config
	installID, analyticsEnabled, err := dotrill.AnalyticsInfo()
	if err != nil {
		analyticsEnabled = false
	}

	// Create a client if there isn't one
	if h.activityClient == nil {
		// If analytics are disabled, we'll use a no-op client.
		// We can set it and return early here.
		if !analyticsEnabled {
			h.activityClient = activity.NewNoopClient()
			return h.activityClient
		}

		// Create a sink that sends events to the intake server.
		intakeSink := activity.NewIntakeSink(zap.NewNop(), activity.IntakeSinkOptions{
			IntakeURL:      telemetryIntakeURL,
			IntakeUser:     telemetryIntakeUser,
			IntakePassword: telemetryIntakePassword,
			BufferSize:     50,
			SinkInterval:   time.Second,
		})

		// Wrap the intake sink in a filter sink that omits events we don't want to send from local.
		// (Remember, this telemetry client will only be used on local.)
		sink := activity.NewFilterSink(intakeSink, func(e activity.Event) bool {
			// Omit metrics events (since they are quite chatty and potentially sensitive).
			return e.EventType != activity.EventTypeMetric
		})

		// Create the telemetry client with metadata about the current environment.
		h.activityClient = activity.NewClient(sink, zap.NewNop())
		h.activityClient = h.activityClient.WithServiceName(telemetryServiceName)
		if h.Version.Number != "" || h.Version.Commit != "" {
			h.activityClient = h.activityClient.WithServiceVersion(h.Version.Number, h.Version.Commit)
		}
		if h.Version.IsDev() {
			h.activityClient = h.activityClient.WithIsDev()
		}
		h.activityClient = h.activityClient.WithInstallID(installID)
	}

	// Fetch the current user ID and set it on the telemetry client.
	// We do this outside of the client creation block to reset the user ID if the hash changes.
	var userID string
	if h.AdminToken() != "" {
		userID, _ = h.CurrentUserID(ctx)
	}
	h.activityClient = h.activityClient.WithUserID(userID)

	return h.activityClient
}

// CurrentUserID fetches the ID of the current user.
// It caches the result in ~/.rill, along with a hash of the current admin token for cache invalidation in case of login/logout.
func (h *Helper) CurrentUserID(ctx context.Context) (string, error) {
	if h.AdminToken() == "" {
		return "", nil
	}

	newHash := hashStr(h.AdminToken(), h.AdminURL())

	oldHash, err := dotrill.GetUserCheckHash()
	if err != nil {
		return "", err
	}

	if oldHash == newHash {
		userID, err := dotrill.GetUserID()
		if err != nil {
			return "", err
		}
		return userID, nil
	}

	c, err := h.Client()
	if err != nil {
		return "", err
	}

	res, err := c.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
	if err != nil {
		return "", err
	}

	var userID string
	if res.User != nil {
		userID = res.User.Id
	}

	err = dotrill.SetUserID(userID)
	if err != nil {
		return "", err
	}

	err = dotrill.SetUserCheckHash(newHash)
	if err != nil {
		return "", err
	}

	return userID, nil
}

// LoadProject loads the cloud project identified by the .rillcloud directory at the given path.
// It returns an error if the caller is not authenticated.
// If there is no .rillcloud directory, it returns a nil project an no error.
func (h *Helper) LoadProject(ctx context.Context, path string) (*adminv1.Project, error) {
	if !h.IsAuthenticated() {
		return nil, fmt.Errorf("can't load project because you are not authenticated")
	}

	rc, err := dotrillcloud.GetAll(path, h.AdminURL())
	if err != nil {
		return nil, fmt.Errorf("failed to load .rillcloud: %w", err)
	}
	if rc == nil {
		return nil, nil
	}

	c, err := h.Client()
	if err != nil {
		return nil, err
	}

	res, err := c.GetProjectByID(ctx, &adminv1.GetProjectByIDRequest{
		Id: rc.ProjectID,
	})
	if err != nil {
		// If the project doesn't exist, delete the local project metadata.
		if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
			err = dotrillcloud.Delete(path, h.AdminURL())
			if err != nil {
				return nil, err
			}
		}

		// We'll ignore the error, pretending no .rillcloud metadata was found
		return nil, nil
	}

	return res.Project, nil
}

func (h *Helper) ProjectNamesByGithubURL(ctx context.Context, org, githubURL, subPath string) ([]string, error) {
	c, err := h.Client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{
		OrganizationName: org,
	})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, p := range resp.Projects {
		if strings.EqualFold(p.GithubUrl, githubURL) && (subPath == "" || strings.EqualFold(p.Subpath, subPath)) {
			names = append(names, p.Name)
		}
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no project with github URL %q exists in org %q", githubURL, org)
	}

	return names, nil
}

func (h *Helper) InferProjectName(ctx context.Context, org, path string) (string, error) {
	// Try loading the project from the .rillcloud directory
	proj, err := h.LoadProject(ctx, path)
	if err != nil {
		return "", err
	}
	if proj != nil {
		return proj.Name, nil
	}

	// Verify projectPath is a Git repo with remote on Github
	_, githubURL, err := gitutil.ExtractGitRemote(path, "", true)
	if err != nil {
		return "", err
	}

	// Fetch project names matching the Github URL
	names, err := h.ProjectNamesByGithubURL(ctx, org, githubURL, "")
	if err != nil {
		return "", err
	}

	if len(names) == 1 {
		return names[0], nil
	}

	return SelectPrompt("Select project", names, "")
}

func (h *Helper) LoginWithTelemetry(ctx context.Context, redirectURL string) error {
	h.PrintfBold("Please log in or sign up for Rill. Opening browser...\n")
	time.Sleep(2 * time.Second)

	h.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventLoginStart)

	if err := auth.Login(ctx, h, redirectURL); err != nil {
		if errors.Is(err, deviceauth.ErrAuthenticationTimedout) {
			h.PrintfWarn("Rill login has timed out as the code was not confirmed in the browser.\n")
			h.PrintfWarn("Run `rill deploy` again.\n")
			return nil
		} else if errors.Is(err, deviceauth.ErrCodeRejected) {
			h.PrintfError("Login failed: Confirmation code rejected\n")
			return nil
		}
		return fmt.Errorf("login failed: %w", err)
	}

	// The cmdutil.Helper automatically detects the login and will add the user's ID to the telemetry.
	h.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventLoginSuccess)

	return nil
}

func (h *Helper) ValidateLocalProject(ctx context.Context, gitPath, subPath string) (string, string, error) {
	var localGitPath string
	var err error
	if gitPath != "" {
		localGitPath, err = fileutil.ExpandHome(gitPath)
		if err != nil {
			return "", "", err
		}
	}
	localGitPath, err = filepath.Abs(localGitPath)
	if err != nil {
		return "", "", err
	}

	var localProjectPath string
	if subPath == "" {
		localProjectPath = localGitPath
	} else {
		localProjectPath = filepath.Join(localGitPath, subPath)
	}

	// Verify that localProjectPath contains a Rill project.
	if rillv1beta.HasRillProject(localProjectPath) {
		return localGitPath, localProjectPath, nil
	}
	// If not, we still navigate user to login and then fail afterwards.
	if !h.IsAuthenticated() {
		err := h.LoginWithTelemetry(ctx, "")
		if err != nil {
			h.PrintfWarn("Login failed with error: %s\n", err.Error())
		}
		fmt.Println()
	}

	h.PrintfWarn("Directory %q doesn't contain a valid Rill project.\n", localProjectPath)
	h.PrintfWarn("Run `rill deploy` from a Rill project directory or use `--path` to pass a project path.\n")
	h.PrintfWarn("Run `rill start` to initialize a new Rill project.\n")
	return "", "", ErrInvalidProject
}

// SetDefaultOrg sets a default org for the user if user is part of any org.
func (h *Helper) SetDefaultOrg(ctx context.Context) error {
	c, err := h.Client()
	if err != nil {
		return err
	}

	res, err := c.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
	if err != nil {
		return fmt.Errorf("listing orgs failed: %w", err)
	}

	if len(res.Organizations) == 1 {
		h.Org = res.Organizations[0].Name
		if err := dotrill.SetDefaultOrg(h.Org); err != nil {
			return err
		}
	} else if len(res.Organizations) > 1 {
		orgName, err := org.SwitchSelectFlow(res.Organizations)
		if err != nil {
			return fmt.Errorf("org selection failed %w", err)
		}

		h.Org = orgName
		if err := dotrill.SetDefaultOrg(h.Org); err != nil {
			return err
		}
	}
	return nil
}

func hashStr(ss ...string) string {
	hash := md5.New()
	for _, s := range ss {
		_, err := hash.Write([]byte(s))
		if err != nil {
			panic(err)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}
