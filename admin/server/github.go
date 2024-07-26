package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/archive"
	"github.com/rilldata/rill/admin/pkg/gitutil"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	githubcookieName        = "github_auth"
	githubcookieFieldState  = "github_state"
	githubcookieFieldRemote = "github_remote"
	archivePullTimeout      = 10 * time.Minute
)

var allowedPaths = []string{
	".git",
	"README.md",
	"LICENSE",
}

func (s *Server) GetGithubUserStatus(ctx context.Context, req *adminv1.GetGithubUserStatusRequest) (*adminv1.GetGithubUserStatusResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if user.GithubUsername == "" {
		// If we don't have user's github username we navigate user to installtion assuming they never installed github app
		grantAccessURL, err := urlutil.WithQuery(s.urls.githubConnect, nil)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create redirect URL: %s", err)
		}

		return &adminv1.GetGithubUserStatusResponse{
			HasAccess:      false,
			GrantAccessUrl: grantAccessURL,
		}, nil
	}
	token, refreshToken, err := s.userAccessToken(ctx, user.GithubRefreshToken)
	if err != nil {
		// token not valid or expired, take auth again
		grantAccessURL, err := urlutil.WithQuery(s.urls.githubAuth, nil)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create redirect URL: %s", err)
		}

		return &adminv1.GetGithubUserStatusResponse{
			HasAccess:      false,
			GrantAccessUrl: grantAccessURL,
		}, nil
	}

	// refresh token changes after using it for getting a new token
	// so saving the updated refresh token
	user, err = s.admin.DB.UpdateUser(ctx, claims.OwnerID(), &database.UpdateUserOptions{
		DisplayName:         user.DisplayName,
		PhotoURL:            user.PhotoURL,
		GithubUsername:      user.GithubUsername,
		GithubRefreshToken:  refreshToken,
		QuotaSingleuserOrgs: user.QuotaSingleuserOrgs,
		PreferenceTimeZone:  user.PreferenceTimeZone,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	userInstallationPermission := adminv1.GithubPermission_GITHUB_PERMISSION_UNSPECIFIED
	installation, _, err := s.admin.Github.AppClient().Apps.FindUserInstallation(ctx, user.GithubUsername)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return nil, fmt.Errorf("failed to get user installation: %w", err)
		}
	} else {
		// older git app would ask for Contents=read permission whereas new one asks for Contents=write and && Administration=write
		if installation.Permissions != nil && installation.Permissions.Contents != nil && strings.EqualFold(*installation.Permissions.Contents, "read") {
			userInstallationPermission = adminv1.GithubPermission_GITHUB_PERMISSION_READ
		}

		if installation.Permissions != nil && installation.Permissions.Contents != nil && installation.Permissions.Administration != nil && strings.EqualFold(*installation.Permissions.Administration, "write") && strings.EqualFold(*installation.Permissions.Contents, "write") {
			userInstallationPermission = adminv1.GithubPermission_GITHUB_PERMISSION_WRITE
		}
	}

	client := github.NewTokenClient(ctx, token)
	// List all the private organizations for the authenticated user
	orgs, _, err := client.Organizations.List(ctx, "", nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user organizations: %s", err.Error())
	}
	// List all the public organizations for the authenticated user
	publicOrgs, _, err := client.Organizations.List(ctx, user.GithubUsername, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user organizations: %s", err.Error())
	}

	orgs = append(orgs, publicOrgs...)
	allOrgs := make([]string, 0)

	orgInstallationPermission := make(map[string]adminv1.GithubPermission)
	for _, org := range orgs {
		// dedupe orgs
		if _, ok := orgInstallationPermission[org.GetLogin()]; ok {
			continue
		}
		allOrgs = append(allOrgs, org.GetLogin())

		i, _, err := s.admin.Github.AppClient().Apps.FindOrganizationInstallation(ctx, org.GetLogin())
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				orgInstallationPermission[org.GetLogin()] = adminv1.GithubPermission_GITHUB_PERMISSION_UNSPECIFIED
				continue
			}
			return nil, status.Errorf(codes.Internal, "failed to get organization installation: %s", err.Error())
		}
		permission := adminv1.GithubPermission_GITHUB_PERMISSION_UNSPECIFIED
		// older git app would ask for Contents=read permission whereas new one asks for Contents=write and && Administration=write
		if i.Permissions != nil && i.Permissions.Contents != nil && strings.EqualFold(*i.Permissions.Contents, "read") {
			permission = adminv1.GithubPermission_GITHUB_PERMISSION_READ
		}

		if i.Permissions != nil && i.Permissions.Contents != nil && i.Permissions.Administration != nil && strings.EqualFold(*i.Permissions.Administration, "write") && strings.EqualFold(*i.Permissions.Contents, "write") {
			permission = adminv1.GithubPermission_GITHUB_PERMISSION_WRITE
		}

		orgInstallationPermission[org.GetLogin()] = permission
	}

	return &adminv1.GetGithubUserStatusResponse{
		HasAccess:                           true,
		GrantAccessUrl:                      s.urls.githubConnect,
		AccessToken:                         token,
		Account:                             user.GithubUsername,
		Organizations:                       allOrgs,
		UserInstallationPermission:          userInstallationPermission,
		OrganizationInstallationPermissions: orgInstallationPermission,
	}, nil
}

func (s *Server) GetGithubRepoStatus(ctx context.Context, req *adminv1.GetGithubRepoStatusRequest) (*adminv1.GetGithubRepoStatusResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.github_url", req.GithubUrl),
	)

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	// Check whether we have the access to the repo
	installationID, err := s.admin.GetGithubInstallation(ctx, req.GithubUrl)
	if err != nil {
		if !errors.Is(err, admin.ErrGithubInstallationNotFound) {
			return nil, status.Errorf(codes.InvalidArgument, "failed to check Github access: %s", err.Error())
		}

		// If no access, return instructions for granting access
		grantAccessURL, err := urlutil.WithQuery(s.urls.githubConnect, map[string]string{"remote": req.GithubUrl})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create redirect URL: %s", err)
		}

		res := &adminv1.GetGithubRepoStatusResponse{
			HasAccess:      false,
			GrantAccessUrl: grantAccessURL,
		}
		return res, nil
	}

	// we have access need to check if user is a collaborator and has authorised app on their account
	userID := claims.OwnerID()
	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// user has not authorized github app
	if user.GithubUsername == "" {
		redirectURL, err := urlutil.WithQuery(s.urls.githubAuth, map[string]string{"remote": req.GithubUrl})
		if err != nil {
			return nil, err
		}

		res := &adminv1.GetGithubRepoStatusResponse{
			HasAccess:      false,
			GrantAccessUrl: redirectURL,
		}
		return res, nil
	}

	// Get repo info for user and return.
	repository, err := s.admin.LookupGithubRepoForUser(ctx, installationID, req.GithubUrl, user.GithubUsername)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsNotCollaborator) {
			// may be user authorised from another username
			redirectURL, err := urlutil.WithQuery(s.urls.githubAuthRetry, map[string]string{"remote": req.GithubUrl, "githubUsername": user.GithubUsername})
			if err != nil {
				return nil, err
			}

			res := &adminv1.GetGithubRepoStatusResponse{
				HasAccess:      false,
				GrantAccessUrl: redirectURL,
			}
			return res, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &adminv1.GetGithubRepoStatusResponse{
		HasAccess:     true,
		DefaultBranch: *repository.DefaultBranch,
	}
	return res, nil
}

func (s *Server) ListGithubUserRepos(ctx context.Context, req *adminv1.ListGithubUserReposRequest) (*adminv1.ListGithubUserReposResponse, error) {
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	userID := claims.OwnerID()
	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// user has not authorized github app
	if user.GithubUsername == "" {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	token, refreshToken, err := s.userAccessToken(ctx, user.GithubRefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// refresh token changes after using it for getting a new token
	// so saving the updated refresh token
	_, err = s.admin.DB.UpdateUser(ctx, claims.OwnerID(), &database.UpdateUserOptions{
		DisplayName:         user.DisplayName,
		PhotoURL:            user.PhotoURL,
		GithubUsername:      user.GithubUsername,
		GithubRefreshToken:  refreshToken,
		QuotaSingleuserOrgs: user.QuotaSingleuserOrgs,
		PreferenceTimeZone:  user.PreferenceTimeZone,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	client := github.NewTokenClient(ctx, token)

	// use a client with user's token to get installations
	repos, err := s.fetchReposForUser(ctx, client)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.ListGithubUserReposResponse{
		Repos: repos,
	}, nil
}

func (s *Server) ConnectProjectToGithub(ctx context.Context, req *adminv1.ConnectProjectToGithubRequest) (*adminv1.ConnectProjectToGithubResponse, error) {
	// TODO: telemetry

	// Find project
	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to delete project")
	}

	if safeStr(proj.GithubURL) != "" {
		return nil, status.Error(codes.InvalidArgument, "project already connected to github")
	}

	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	token, refreshToken, err := s.userAccessToken(ctx, user.GithubRefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// refresh token changes after using it for getting a new token
	// so saving the updated refresh token
	_, err = s.admin.DB.UpdateUser(ctx, claims.OwnerID(), &database.UpdateUserOptions{
		DisplayName:         user.DisplayName,
		PhotoURL:            user.PhotoURL,
		GithubUsername:      user.GithubUsername,
		GithubRefreshToken:  refreshToken,
		QuotaSingleuserOrgs: user.QuotaSingleuserOrgs,
		PreferenceTimeZone:  user.PreferenceTimeZone,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	asset, err := s.admin.DB.FindAsset(ctx, *proj.ArchiveAssetID)
	if err != nil {
		return nil, err
	}

	downloadURL, err := s.generateV4GetObjectSignedURL(asset.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.pushArchiveToGit(ctx, downloadURL, req.Repo, req.Branch, req.Subpath, token, req.Force)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = s.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
		OrganizationName: org.Name,
		Name:             proj.Name,
		ProdBranch:       &req.Branch,
		GithubUrl:        &req.Repo,
		Subpath:          &req.Subpath,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.ConnectProjectToGithubResponse{}, nil
}

// registerGithubEndpoints registers the non-gRPC endpoints for the Github integration.
func (s *Server) registerGithubEndpoints(mux *http.ServeMux) {
	// TODO: Add helper utils to clean this up
	inner := http.NewServeMux()
	observability.MuxHandle(inner, "/github/webhook", http.HandlerFunc(s.githubWebhook))
	observability.MuxHandle(inner, "/github/connect", s.authenticator.HTTPMiddleware(middleware.Check(s.checkGithubRateLimit("/github/connect"), http.HandlerFunc(s.githubConnect))))
	observability.MuxHandle(inner, "/github/connect/callback", s.authenticator.HTTPMiddleware(middleware.Check(s.checkGithubRateLimit("/github/connect/callback"), http.HandlerFunc(s.githubConnectCallback))))
	observability.MuxHandle(inner, "/github/auth/login", s.authenticator.HTTPMiddleware(middleware.Check(s.checkGithubRateLimit("github/auth/login"), http.HandlerFunc(s.githubAuthLogin))))
	observability.MuxHandle(inner, "/github/auth/callback", s.authenticator.HTTPMiddleware(middleware.Check(s.checkGithubRateLimit("github/auth/callback"), http.HandlerFunc(s.githubAuthCallback))))
	observability.MuxHandle(inner, "/github/post-auth-redirect", s.authenticator.HTTPMiddleware(middleware.Check(s.checkGithubRateLimit("github/post-auth-redirect"), http.HandlerFunc(s.githubStatus))))
	mux.Handle("/github/", observability.Middleware("admin", s.logger, inner))
}

// githubConnect starts an installation flow of the Github App.
// It's implemented as a non-gRPC endpoint mounted directly on /github/connect.
// It redirects the user to Github to authorize Rill to access one or more repositories.
// After the Github flow completes, the user is redirected back to githubConnectCallback.
func (s *Server) githubConnect(w http.ResponseWriter, r *http.Request) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() != auth.OwnerTypeUser {
		// redirect to the auth site, with a redirect back to here after successful auth.
		s.redirectLogin(w, r)
		return
	}

	query := r.URL.Query()
	remote := query.Get("remote")
	if remote == "" {
		http.Redirect(w, r, s.urls.githubAppInstallation, http.StatusTemporaryRedirect)
		return
	}

	redirectURL, err := urlutil.WithQuery(s.urls.githubAppInstallation, map[string]string{"state": remote})
	if err != nil {
		http.Error(w, "failed to generate URL", http.StatusInternalServerError)
		return
	}

	// Redirect to Github App for installation
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// githubConnectCallback is called after a Github App authorization flow initiated by githubConnect has completed.
// This call can originate from users who are not logged in in cases like admin user accepting installation request, removing existing installation etc.
// It's implemented as a non-gRPC endpoint mounted directly on /github/connect/callback.
// High level flow:
// User installation
//   - Save user's github username in the users table
//   - verify the user is a collaborator else return unauthorised
//   - verify the user installed the app on the right repo else navigate to retry
//   - navigate to success page
//
// If user requests the app
//   - Save user's github username in the users table
//   - navigate to request page
func (s *Server) githubConnectCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract info from query string
	qry := r.URL.Query()
	setupAction := qry.Get("setup_action")
	if setupAction != "install" && setupAction != "update" && setupAction != "request" {
		http.Error(w, fmt.Sprintf("unexpected setup_action=%q", setupAction), http.StatusBadRequest)
		return
	}

	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() != auth.OwnerTypeUser {
		s.redirectLogin(w, r)
		return
	}

	code := qry.Get("code")
	if code == "" {
		if setupAction == "install" || !qry.Has("state") {
			http.Error(w, "unable to verify user's identity", http.StatusInternalServerError)
			return
		}

		redirectURL, err := urlutil.WithQuery(s.urls.githubConnectRequest, map[string]string{"remote": qry.Get("state")})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create retry request url: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// exchange code to get an auth token and create a github client with user auth
	githubClient, refreshToken, err := s.userAuthGithubClient(ctx, code)
	if err != nil {
		http.Error(w, "unauthorised user", http.StatusUnauthorized)
		return
	}

	githubUser, _, err := githubClient.Users.Get(ctx, "")
	if err != nil {
		// todo :: can this throw Requires authentication error ??
		http.Error(w, "unauthorised user", http.StatusUnauthorized)
		return
	}

	// save github user name
	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		// user is always guaranteed to exist if it reaches here
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user, err = s.admin.DB.UpdateUser(ctx, user.ID, &database.UpdateUserOptions{
		DisplayName:         user.DisplayName,
		PhotoURL:            user.PhotoURL,
		GithubUsername:      githubUser.GetLogin(),
		GithubRefreshToken:  refreshToken,
		QuotaSingleuserOrgs: user.QuotaSingleuserOrgs,
		PreferenceTimeZone:  user.PreferenceTimeZone,
	})
	if err != nil {
		s.logger.Error("failed to update user's github username")
	}

	remoteURL := qry.Get("state")
	account, repo, ok := gitutil.SplitGithubURL(remoteURL)
	if !ok {
		// request without state can come in multiple ways like
		// 	- if user changes app installation directly on the settings page
		//  - if admin user accepts the installation request
		http.Redirect(w, r, s.urls.githubConnectSuccess, http.StatusTemporaryRedirect)
		return
	}

	if setupAction == "request" {
		// access requested
		redirectURL, err := urlutil.WithQuery(s.urls.githubConnectRequest, map[string]string{"remote": remoteURL})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create connect request url: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// verify there is no spoofing and the user is a collaborator to the repo
	isCollaborator, err := s.isCollaborator(ctx, account, repo, githubClient, githubUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to verify ownership: %s", err), http.StatusUnauthorized)
		return
	}

	if !isCollaborator {
		redirectURL, err := urlutil.WithQuery(s.urls.githubAuthRetry, map[string]string{"remote": remoteURL, "githubUsername": user.GithubUsername})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect to retry page
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// install/update setupAction
	// Verify that user installed the app on the right repo and we have access now
	_, err = s.admin.GetGithubInstallation(ctx, remoteURL)
	if err != nil {
		if !errors.Is(err, admin.ErrGithubInstallationNotFound) {
			http.Error(w, fmt.Sprintf("failed to check github repo status: %s", err), http.StatusInternalServerError)
			return
		}

		// no access
		// Redirect to UI retry page
		redirectURL, err := urlutil.WithQuery(s.urls.githubConnectRetry, map[string]string{"remote": remoteURL})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create retry request url: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Redirect to UI success page
	http.Redirect(w, r, s.urls.githubConnectSuccess, http.StatusTemporaryRedirect)
}

// githubAuthLogin starts user authorization of github app.
// In case github app is installed by another user, other users of the repo need to separately authorise github app
// where this flow comes into picture.
// Some implementation details are copied from auth package.
// It's implemented as a non-gRPC endpoint mounted directly on /github/auth/login.
func (s *Server) githubAuthLogin(w http.ResponseWriter, r *http.Request) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() != auth.OwnerTypeUser {
		// Redirect to the auth site, with a redirect back to here after successful auth.
		s.redirectLogin(w, r)
		return
	}

	// Generate random state for CSRF
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate state: %s", err), http.StatusInternalServerError)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	// Get auth cookie
	sess := s.cookies.Get(r, githubcookieName)
	// Set state in cookie
	sess.Values[githubcookieFieldState] = state
	remote := r.URL.Query().Get("remote")
	if remote != "" {
		sess.Values[githubcookieFieldRemote] = remote
	}

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	oauthConf := &oauth2.Config{
		ClientID:     s.opts.GithubClientID,
		ClientSecret: s.opts.GithubClientSecret,
		Endpoint:     githuboauth.Endpoint,
		RedirectURL:  s.urls.githubAuthCallback,
	}
	// Redirect to github login page
	http.Redirect(w, r, oauthConf.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusTemporaryRedirect)
}

// githubAuthCallback is called after a user authorizes github app on their account
// It's implemented as a non-gRPC endpoint mounted directly on /github/auth/callback.
func (s *Server) githubAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() != auth.OwnerTypeUser {
		http.Error(w, "unidentified user", http.StatusUnauthorized)
		return
	}

	// Get auth cookie
	sess := s.cookies.Get(r, githubcookieName)
	// Check that random state matches (for CSRF protection)
	qry := r.URL.Query()
	if qry.Get("state") != sess.Values[githubcookieFieldState] {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}
	delete(sess.Values, githubcookieFieldState)

	// verify user's identity with github
	code := qry.Get("code")
	if code == "" {
		http.Error(w, "unauthorised user", http.StatusUnauthorized)
		return
	}

	// exchange code to get an auth token and create a github client with user auth
	c, refreshToken, err := s.userAuthGithubClient(ctx, code)
	if err != nil {
		// todo :: check for unauthorised user error
		http.Error(w, fmt.Sprintf("internal error %s", err.Error()), http.StatusInternalServerError)
		return
	}

	gitUser, _, err := c.Users.Get(ctx, "")
	if err != nil {
		// todo :: check for unauthorised user error
		http.Error(w, fmt.Sprintf("internal error %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// save the github user name
	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		// can this happen ??
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, "unidentified user", http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("internal error %s", err.Error()), http.StatusInternalServerError)
		return
	}

	_, err = s.admin.DB.UpdateUser(ctx, user.ID, &database.UpdateUserOptions{
		DisplayName:         user.DisplayName,
		PhotoURL:            user.PhotoURL,
		GithubUsername:      gitUser.GetLogin(),
		GithubRefreshToken:  refreshToken,
		QuotaSingleuserOrgs: user.QuotaSingleuserOrgs,
		PreferenceTimeZone:  user.PreferenceTimeZone,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to save user information %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// if there is a remote set, verify the user is a collaborator the repo
	remote := ""
	if value, ok := sess.Values[githubcookieFieldRemote]; ok {
		remote = value.(string)
	}
	delete(sess.Values, githubcookieFieldRemote)

	account, repo, ok := gitutil.SplitGithubURL(remote)
	if !ok {
		http.Redirect(w, r, s.urls.githubConnectSuccess, http.StatusTemporaryRedirect)
		return
	}

	ok, err = s.isCollaborator(ctx, account, repo, c, gitUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("user identification failed with error %s", err.Error()), http.StatusUnauthorized)
		return
	}

	if !ok {
		redirectURL, err := urlutil.WithQuery(s.urls.githubAuthRetry, map[string]string{"remote": remote, "githubUsername": user.GithubUsername})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect to retry page
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to UI success page
	http.Redirect(w, r, s.urls.githubConnectSuccess, http.StatusTemporaryRedirect)
}

// githubWebhook is called by Github to deliver events about new pushes, pull requests, changes to a repository, etc.
// It's implemented as a non-gRPC endpoint mounted directly on /github/webhook.
// Note that Github webhooks have a timeout of 10 seconds. Webhook processing is moved to the background to prevent timeouts.
func (s *Server) githubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "expected a POST request", http.StatusBadRequest)
		return
	}

	payload, err := github.ValidatePayload(r, []byte(s.opts.GithubAppWebhookSecret))
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid github payload: %s", err), http.StatusUnauthorized)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid webhook payload: %s", err), http.StatusBadRequest)
		return
	}

	err = s.admin.ProcessGithubEvent(context.Background(), event)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to process event: %s", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// githubStatus is a http wrapper over [GetGithubRepoStatus]/[GetGithubUserStatus] depending upon whether `remote` query is passed.
// It redirects to the grantAccessURL if there is no access.
// It's implemented as a non-gRPC endpoint mounted directly on /github/post-auth-redirect.
func (s *Server) githubStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		s.redirectLogin(w, r)
		return
	}

	var (
		hasAccess      bool
		grantAccessURL string
		remote         = r.URL.Query().Get("remote")
	)

	if remote == "" {
		resp, err := s.GetGithubUserStatus(ctx, &adminv1.GetGithubUserStatusRequest{})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch user status: %s", err), http.StatusInternalServerError)
			return
		}
		hasAccess = resp.HasAccess
		grantAccessURL = resp.GrantAccessUrl
	} else {
		resp, err := s.GetGithubRepoStatus(ctx, &adminv1.GetGithubRepoStatusRequest{GithubUrl: remote})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch github repo status: %s", err), http.StatusInternalServerError)
			return
		}
		hasAccess = resp.HasAccess
		grantAccessURL = resp.GrantAccessUrl
	}

	if hasAccess {
		http.Redirect(w, r, s.urls.githubConnectSuccess, http.StatusTemporaryRedirect)
		return
	}

	redirectURL, err := urlutil.WithQuery(s.urls.githubConnectUI, map[string]string{"redirect": grantAccessURL})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create redirect URL: %s", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (s *Server) userAuthGithubClient(ctx context.Context, code string) (*github.Client, string, error) {
	oauthConf := &oauth2.Config{
		ClientID:     s.opts.GithubClientID,
		ClientSecret: s.opts.GithubClientSecret,
		Endpoint:     githuboauth.Endpoint,
	}

	token, err := oauthConf.Exchange(ctx, code)
	if err != nil {
		return nil, "", err
	}

	oauthClient := oauthConf.Client(ctx, token)
	return github.NewClient(oauthClient), token.RefreshToken, nil
}

// isCollaborator checks if the user is a collaborator of the repository identified by owner and repo
// client must be authorized with user's auth token
func (s *Server) isCollaborator(ctx context.Context, owner, repo string, client *github.Client, user *github.User) (bool, error) {
	githubUserName := user.GetLogin()
	// repo belongs to the user's personal account
	if owner == githubUserName {
		return true, nil
	}

	// repo belongs to an org
	isCollaborator, resp, err := client.Repositories.IsCollaborator(ctx, owner, repo, user.GetLogin())
	if err != nil {
		// user client does not have access to the repository
		if resp != nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
			return false, nil
		}
		return false, err
	}
	return isCollaborator, nil
}

func (s *Server) redirectLogin(w http.ResponseWriter, r *http.Request) {
	redirectURL, err := urlutil.WithQuery(s.urls.authLogin, map[string]string{"redirect": r.URL.RequestURI()})
	if err != nil {
		http.Error(w, "failed to generate URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (s *Server) checkGithubRateLimit(route string) middleware.CheckFunc {
	return func(req *http.Request) error {
		claims := auth.GetClaims(req.Context())
		if claims == nil || claims.OwnerType() == auth.OwnerTypeAnon {
			limitKey := ratelimit.AnonLimitKey(route, observability.HTTPPeer(req))
			if err := s.limiter.Limit(req.Context(), limitKey, ratelimit.Sensitive); err != nil {
				if errors.As(err, &ratelimit.QuotaExceededError{}) {
					return httputil.Error(http.StatusTooManyRequests, err)
				}
				return err
			}
		}
		return nil
	}
}

func (s *Server) userAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	if refreshToken == "" {
		return "", "", errors.New("refresh token is empty")
	}

	oauthConf := &oauth2.Config{
		ClientID:     s.opts.GithubClientID,
		ClientSecret: s.opts.GithubClientSecret,
		Endpoint:     githuboauth.Endpoint,
	}

	src := oauthConf.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	oauthToken, err := src.Token()
	if err != nil {
		return "", "", err
	}

	return oauthToken.AccessToken, oauthToken.RefreshToken, nil
}

func (s *Server) fetchReposForUser(ctx context.Context, client *github.Client) ([]*adminv1.ListGithubUserReposResponse_Repo, error) {
	repos := make([]*adminv1.ListGithubUserReposResponse_Repo, 0)
	page := 1

	for {
		installations, httpResp, err := client.Apps.ListUserInstallations(ctx, &github.ListOptions{Page: page, PerPage: 100})
		if err != nil {
			return nil, err
		}

		for _, installation := range installations {
			reposForInst, err := s.fetchReposForInstallation(ctx, client, *installation.ID)
			if err != nil {
				return nil, err
			}
			repos = append(repos, reposForInst...)
		}

		if httpResp.NextPage == 0 {
			break
		}
		page = httpResp.NextPage
	}

	return repos, nil
}

func (s *Server) fetchReposForInstallation(ctx context.Context, client *github.Client, instID int64) ([]*adminv1.ListGithubUserReposResponse_Repo, error) {
	repos := make([]*adminv1.ListGithubUserReposResponse_Repo, 0)
	page := 1

	for {
		reposResp, httpResp, err := client.Apps.ListUserRepos(ctx, instID, &github.ListOptions{Page: page, PerPage: 100})
		if err != nil {
			return nil, err
		}

		for _, repo := range reposResp.Repositories {
			var owner string
			if repo.Owner != nil {
				owner = fromStringPtr(repo.Owner.Login)
			}
			var branch string
			if repo.DefaultBranch != nil {
				branch = fromStringPtr(repo.DefaultBranch)
			} else {
				branch = fromStringPtr(repo.MasterBranch)
			}
			repos = append(repos, &adminv1.ListGithubUserReposResponse_Repo{
				Name:          fromStringPtr(repo.Name),
				Owner:         owner,
				Description:   fromStringPtr(repo.Description),
				Url:           fromStringPtr(repo.HTMLURL),
				DefaultBranch: branch,
			})
		}

		if httpResp.NextPage == 0 {
			break
		}
		page = httpResp.NextPage
	}

	return repos, nil
}

func (s *Server) pushArchiveToGit(ctx context.Context, downloadUrl, repo, branch, subpath, token string, force bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), archivePullTimeout)
	defer cancel()

	// generate a temp dir to extract the archive
	dir, err := os.MkdirTemp(os.TempDir(), "extracted_archives")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	downloadDst := filepath.Join(dir, "admin_driver_zipped_repo.tar.gz")
	// use a subfolder for working with git
	gitPath := filepath.Join(dir, "proj")
	// projPath is the target for extracting the archive
	projPath := gitPath
	if subpath != "" {
		projPath = filepath.Join(projPath, subpath)
	}
	err = os.MkdirAll(projPath, fs.ModePerm)
	if err != nil {
		return err
	}

	gitAuth := &githttp.BasicAuth{Username: "x-access-token", Password: token}

	var ghRepo *git.Repository
	empty := false
	ghRepo, err = git.PlainClone(gitPath, false, &git.CloneOptions{
		URL:           repo,
		Auth:          gitAuth,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
	})
	if err != nil {
		if errors.Is(err, transport.ErrEmptyRemoteRepository) {
			empty = true
			ghRepo, err = git.PlainInitWithOptions(gitPath, &git.PlainInitOptions{
				InitOptions: git.InitOptions{
					DefaultBranch: plumbing.NewBranchReferenceName(branch),
				},
				Bare: false,
			})
			if err != nil {
				return fmt.Errorf("failed to init git repo: %w", err)
			}
		} else {
			return fmt.Errorf("failed to init git repo: %w", err)
		}
	}

	wt, err := ghRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	var wtc worktreeContents
	if !empty {
		wtc, err = readWorktree(wt, subpath)
		if err != nil {
			return fmt.Errorf("failed to read worktree: %w", err)
		}

		if len(wtc.otherPaths) > 0 && !force {
			return fmt.Errorf("worktree has additional contents")
		}
	}

	// remove all the other paths
	for _, path := range wtc.otherPaths {
		err = os.RemoveAll(filepath.Join(projPath, path))
		if err != nil {
			return err
		}
	}

	err = archive.Download(ctx, downloadUrl, downloadDst, projPath, false)
	if err != nil {
		return err
	}

	// add back the older gitignore contents if present
	if wtc.gitignore != "" {
		gi, err := os.ReadFile(filepath.Join(projPath, ".gitignore"))
		if err != nil {
			return err
		}

		// if the new gitignore is not the same then it was overwritten during extract
		if string(gi) != wtc.gitignore {
			// append the new contents to the end
			gi = append([]byte(fmt.Sprintf("%s\n", wtc.gitignore)), gi...)

			err = os.WriteFile(filepath.Join(projPath, ".gitignore"), gi, fs.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	// git add .
	if err := wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	// git commit -m
	_, err = wt.Commit("Auto committed by Rill", &git.CommitOptions{All: true})
	if err != nil {
		if !errors.Is(err, git.ErrEmptyCommit) {
			return fmt.Errorf("failed to commit files to git: %w", err)
		}
	}

	if empty {
		_, err = ghRepo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{repo}})
		if err != nil {
			return fmt.Errorf("failed to create remote: %w", err)
		}
	}

	if err := ghRepo.PushContext(ctx, &git.PushOptions{Auth: gitAuth}); err != nil {
		return fmt.Errorf("failed to push to remote %q : %w", repo, err)
	}

	return nil
}

func fromStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type worktreeContents struct {
	gitignore  string
	otherPaths []string
}

func readWorktree(wt *git.Worktree, subpath string) (worktreeContents, error) {
	var wtc worktreeContents

	files, err := wt.Filesystem.ReadDir(subpath)
	if err != nil {
		return worktreeContents{}, err
	}
	for _, file := range files {
		if file.Name() == ".gitignore" {
			f, err := wt.Filesystem.Open(filepath.Join(subpath, file.Name()))
			if err != nil {
				return worktreeContents{}, err
			}
			wtc.gitignore, err = readFile(f)
			if err != nil {
				return worktreeContents{}, err
			}
		} else {
			found := false
			for _, path := range allowedPaths {
				if file.Name() == path {
					found = true
					break
				}
			}
			if !found {
				wtc.otherPaths = append(wtc.otherPaths, file.Name())
			}
		}
	}

	return wtc, nil
}

func readFile(f billy.File) (string, error) {
	defer f.Close()
	buf := make([]byte, 0, 32*1024)
	c := ""

	for {
		n, err := f.Read(buf[:cap(buf)])
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
		if n == 0 {
			continue
		}
		buf = buf[:n]
		c = c + string(buf)
	}

	return c, nil
}
