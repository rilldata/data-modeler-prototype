package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-github/v50/github"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/gitutil"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	cookieName       = "github_auth"
	cookieFieldState = "github_state"
)

func (s *Server) GetGithubRepoStatus(ctx context.Context, req *adminv1.GetGithubRepoStatusRequest) (*adminv1.GetGithubRepoStatusResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	// Check whether we have the access to the repo
	userID := claims.OwnerID()
	installationID, ok, err := s.admin.GetGithubInstallation(ctx, userID, req.GithubUrl)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to check Github access: %s", err.Error())
	}

	// If no access, return instructions for granting access
	if !ok {
		grantAccessURL, err := url.JoinPath(s.opts.ExternalURL, "/github/connect")
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
	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// user has not authorized github app
	if user.GithubUserName == "" {
		authoriseURL := s.urls.GithubAuthorise()
		qry := authoriseURL.Query()
		qry.Set("auto_redirect", "true")
		authoriseURL.RawQuery = qry.Encode()
		res := &adminv1.GetGithubRepoStatusResponse{
			HasAccess:            false,
			UserAuthorisationUrl: authoriseURL.String(),
		}
		return res, nil
	}

	// Get repo info for user and return.
	repository, err := s.admin.LookupGithubRepoForUser(ctx, installationID, req.GithubUrl, user.GithubUserName)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsNotCollaborator) {
			authoriseURL := s.urls.GithubAuthorise()
			msg := fmt.Sprintf("You are not a collaborator. Click %s to re-authorise/authorise another account.", authoriseURL.String())
			return nil, status.Error(codes.PermissionDenied, msg)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &adminv1.GetGithubRepoStatusResponse{
		HasAccess:     true,
		DefaultBranch: *repository.DefaultBranch,
	}
	return res, nil
}

// registerGithubEndpoints registers the non-gRPC endpoints for the Github integration.
func (s *Server) registerGithubEndpoints(mux *gateway.ServeMux) error {
	err := mux.HandlePath("POST", "/github/webhook", s.githubWebhook)
	if err != nil {
		return err
	}

	err = mux.HandlePath("GET", "/github/connect", s.authenticator.HTTPMiddleware(s.githubConnect))
	if err != nil {
		return err
	}

	err = mux.HandlePath("GET", "/github/connect/callback", s.authenticator.HTTPMiddleware(s.githubConnectCallback))
	if err != nil {
		return err
	}

	err = mux.HandlePath("GET", "/github/auth/login", s.authenticator.HTTPMiddleware(s.githubAuthLogin))
	if err != nil {
		return err
	}

	err = mux.HandlePath("GET", "/github/auth/callback", s.authenticator.HTTPMiddleware(s.githubAuthCallback))
	if err != nil {
		return err
	}

	return nil
}

// githubConnect starts an installation flow of the Github App.
// It's implemented as a non-gRPC endpoint mounted directly on /github/connect.
// It redirects the user to Github to authorize Rill to access one or more repositories.
// After the Github flow completes, the user is redirected back to githubConnectCallback.
func (s *Server) githubConnect(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() != auth.OwnerTypeUser {
		// TODO: It should redirect to the auth site, with a redirect back to here after successful auth.
		http.Error(w, "only authenticated users can connect to github", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	remote := query.Get("remote")
	// Should we add any other validation for remote ?
	// Should we return bad request if remote not set ?
	if remote == "" {
		http.Error(w, "no remote set", http.StatusBadRequest)
		return
	}

	redirectURL := s.urls.GithubAppInstallationURL()
	values := redirectURL.Query()
	// `state` query parameter will be passed through to githubConnectCallback.
	// we will use this state parameter to verify that the user installed the app on right repo
	values.Add("state", remote)
	redirectURL.RawQuery = values.Encode()

	// Redirect to Github App for installation
	http.Redirect(w, r, redirectURL.String(), http.StatusTemporaryRedirect)
}

// githubConnectCallback is called after a Github App authorization flow initiated by githubConnect has completed.
// It's implemented as a non-gRPC endpoint mounted directly on /github/connect/callback.
func (s *Server) githubConnectCallback(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	ctx := r.Context()

	// Extract info from query string
	qry := r.URL.Query()
	setupAction := qry.Get("setup_action")
	if setupAction != "install" && setupAction != "update" && setupAction != "request" { // TODO: Also handle "request"
		http.Error(w, fmt.Sprintf("unexpected setup_action=%q", setupAction), http.StatusBadRequest)
		return
	}

	// For request flows this can originate from some user who is not authenticated to rill cloud
	// claims := auth.GetClaims(r.Context())
	// if claims.OwnerType() != auth.OwnerTypeUser {
	// 	http.Error(w, "only authenticated users can connect to github", http.StatusUnauthorized)
	// 	return
	// }

	remoteURL := qry.Get("state")
	account, repo, ok := gitutil.SplitGithubURL(remoteURL)
	if !ok {
		// request without state can come in multiple ways like
		// if user changes app installation directly on the settings page
		// if admin user accepts the installation request
		// nothing to be done in such cases
		// may be redirect to some rill page ??
		w.WriteHeader(http.StatusOK)
		return
	}

	if setupAction == "request" {
		// access requested
		redirectURL := s.urls.GithubConnectRequest()
		qry := redirectURL.Query()
		qry.Set("remote", remoteURL)
		redirectURL.RawQuery = qry.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusTemporaryRedirect)
		return
	}

	// Verify that user installed the app on the right repo and we have access now
	_, resp, err := s.admin.Github.Apps.FindRepositoryInstallation(r.Context(), account, repo)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			// no access
			// Redirect to UI retry page
			redirectURL := s.urls.GithubConnectRetry()
			qry := redirectURL.Query()
			qry.Set("remote", remoteURL)
			redirectURL.RawQuery = qry.Encode()
			http.Redirect(w, r, redirectURL.String(), http.StatusTemporaryRedirect)
		}
		http.Error(w, fmt.Sprintf("failed to check github repo status: %s", err), http.StatusInternalServerError)
		return
	}

	code := qry.Get("code")
	if code == "" {
		http.Error(w, "unauthorised user", http.StatusUnauthorized)
		return
	}

	githubClient, err := s.userAuthGithubClient(ctx, code)
	if err != nil {
		http.Error(w, "unauthorised user", http.StatusUnauthorized)
		return
	}

	// verify there is no spoofing and the user is a collaborator to the repo
	gitUser, isCollaborator, err := s.isCollaborator(ctx, account, repo, githubClient)
	if err != nil {
		// todo :: separate unauthorised user error from other errors
		http.Error(w, fmt.Sprintf("failed to verify ownership: %s", err), http.StatusUnauthorized)
		return
	}

	if !isCollaborator {
		http.Error(w, "unauthorised user", http.StatusUnauthorized)
		return
	}

	claims := auth.GetClaims(r.Context())
	if claims.OwnerID() != "" {
		// find user, can be called by user not logged into rill cloud as well
		if user, err := s.admin.DB.FindUser(ctx, claims.OwnerID()); err == nil { // ignoring error
			_, err := s.admin.DB.UpdateUser(ctx, user.ID, user.DisplayName, user.PhotoURL, gitUser.GetLogin())
			if err != nil {
				s.logger.Error("failed to update user's github username")
			}
		}
	}

	// Redirect to UI success page
	redirectURL, err := url.JoinPath(s.opts.FrontendURL, "/github/connect/success")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create redirect URL: %s", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// copied from package auth
func (s *Server) githubAuthLogin(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// Generate random state for CSRF
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate state: %s", err), http.StatusInternalServerError)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	// Get auth cookie
	sess, err := s.cookies.Get(r, cookieName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get session: %s", err), http.StatusInternalServerError)
		return
	}

	// Set state in cookie
	sess.Values[cookieFieldState] = state

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	redirectURL := s.urls.GithubLoginCallbackURL()
	oauthConf := &oauth2.Config{
		ClientID:     s.opts.GithubClientID,
		ClientSecret: s.opts.GithubClientSecret,
		Endpoint:     githuboauth.Endpoint,
		RedirectURL:  redirectURL.String(),
	}
	// Redirect to github login page
	http.Redirect(w, r, oauthConf.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusTemporaryRedirect)
}

// githubAuthCallback is called after a user authorizes github app on his account
// It's implemented as a non-gRPC endpoint mounted directly on /github/auth/callback.
func (s *Server) githubAuthCallback(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	ctx := r.Context()
	// Get auth cookie
	sess, err := s.cookies.Get(r, cookieName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get session: %s", err), http.StatusInternalServerError)
		return
	}

	// Check that random state matches (for CSRF protection)
	if r.URL.Query().Get("state") != sess.Values[cookieFieldState] {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}
	delete(sess.Values, cookieFieldState)

	// Check there's an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		http.Error(w, "only authenticated users can connect to github", http.StatusUnauthorized)
		return
	}

	// Extract info from query string
	qry := r.URL.Query()
	code := qry.Get("code")
	if code == "" {
		http.Error(w, "unauthorised user", http.StatusUnauthorized)
		return
	}

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

	client, err := s.userAuthGithubClient(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("internal error %s", err.Error()), http.StatusInternalServerError)
		return
	}

	gitUser, _, err := client.Users.Get(ctx, "")
	if err != nil {
		http.Error(w, fmt.Sprintf("internal error %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// save the github user name
	_, err = s.admin.DB.UpdateUser(ctx, user.ID, user.DisplayName, user.PhotoURL, gitUser.GetLogin())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to save user information %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Redirect to UI success page
	redirectURL, err := url.JoinPath(s.opts.FrontendURL, "/github/auth/success")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create redirect URL: %s", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// githubWebhook is called by Github to deliver events about new pushes, pull requests, changes to a repository, etc.
// It's implemented as a non-gRPC endpoint mounted directly on /github/webhook.
// Note that Github webhooks have a timeout of 10 seconds. Webhook processing is moved to the background to prevent timeouts.
func (s *Server) githubWebhook(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
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

func (s *Server) userAuthGithubClient(ctx context.Context, code string) (*github.Client, error) {
	oauthConf := &oauth2.Config{
		ClientID:     s.opts.GithubClientID,
		ClientSecret: s.opts.GithubClientSecret,
		Endpoint:     githuboauth.Endpoint,
	}

	token, err := oauthConf.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	oauthClient := oauthConf.Client(ctx, token)
	return github.NewClient(oauthClient), nil
}

func (s *Server) isCollaborator(ctx context.Context, owner, repo string, client *github.Client) (*github.User, bool, error) {
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, false, err
	}

	githubUserName := user.GetLogin()
	// repo belongs to the user's personal account
	if owner == githubUserName {
		return user, true, nil
	}

	// repo belongs to an org
	isCollaborator, _, err := client.Repositories.IsCollaborator(ctx, owner, repo, user.GetLogin())
	return user, isCollaborator, err
}
