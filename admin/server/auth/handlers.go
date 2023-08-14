package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	cookieName             = "auth"
	cookieFieldState       = "state"
	cookieFieldRedirect    = "redirect"
	cookieFieldAccessToken = "access_token"
)

// RegisterEndpoints adds HTTP endpoints for auth.
// The mux must be served on the ExternalURL of the Authenticator since the logic in these handlers relies on knowing the full external URIs.
// Note that these are not gRPC handlers, just regular HTTP endpoints that we mount on the gRPC-gateway mux.
func (a *Authenticator) RegisterEndpoints(mux *http.ServeMux, limiter ratelimit.Limiter) {
	// checkLimit needs access to limiter
	checkLimit := func(route string, req *http.Request) error {
		claims := GetClaims(req.Context())
		if claims == nil || claims.OwnerType() == OwnerTypeAnon {
			limitKey := ratelimit.AnonLimitKey(route, observability.HTTPPeer(req))
			if err := limiter.Limit(req.Context(), limitKey, ratelimit.Sensitive); err != nil {
				if errors.As(err, &ratelimit.QuotaExceededError{}) {
					return middleware.NewHTTPError(http.StatusTooManyRequests, err.Error())
				}
				return err
			}
		}
		return nil
	}
	// TODO: Add helper utils to clean this up
	inner := http.NewServeMux()
	inner.Handle("/auth/signup", otelhttp.WithRouteTag("/auth/signup", middleware.RequestHTTPHandler("/auth/signup", checkLimit, http.HandlerFunc(a.authSignup))))
	inner.Handle("/auth/login", otelhttp.WithRouteTag("/auth/login", middleware.RequestHTTPHandler("/auth/login", checkLimit, http.HandlerFunc(a.authLogin))))
	inner.Handle("/auth/callback", otelhttp.WithRouteTag("/auth/callback", middleware.RequestHTTPHandler("/auth/callback", checkLimit, http.HandlerFunc(a.authLoginCallback))))
	inner.Handle("/auth/with-token", otelhttp.WithRouteTag("/auth/with-token", middleware.RequestHTTPHandler("/auth/with-token", checkLimit, http.HandlerFunc(a.authWithToken))))
	inner.Handle("/auth/logout", otelhttp.WithRouteTag("/auth/logout", middleware.RequestHTTPHandler("/auth/logout", checkLimit, http.HandlerFunc(a.authLogout))))
	inner.Handle("/auth/logout/callback", otelhttp.WithRouteTag("/auth/logout/callback", middleware.RequestHTTPHandler("/auth/logout/callback", checkLimit, http.HandlerFunc(a.authLogoutCallback))))
	inner.Handle("/auth/oauth/device_authorization", otelhttp.WithRouteTag("/auth/oauth/device_authorization", middleware.RequestHTTPHandler("/auth/oauth/device_authorization", checkLimit, http.HandlerFunc(a.handleDeviceCodeRequest))))
	inner.Handle("/auth/oauth/device", otelhttp.WithRouteTag("/auth/oauth/device", a.HTTPMiddleware(middleware.RequestHTTPHandler("/auth/oauth/device", checkLimit, http.HandlerFunc(a.handleUserCodeConfirmation))))) // NOTE: Uses auth middleware
	inner.Handle("/auth/oauth/token", otelhttp.WithRouteTag("/auth/oauth/token", middleware.RequestHTTPHandler("/auth/oauth/token", checkLimit, http.HandlerFunc(a.getAccessToken))))
	mux.Handle("/auth/", observability.Middleware("admin", a.logger, inner))
}

// authSignup redirects the users to signup page after starting the 0Auth and OIDC flow
func (a *Authenticator) authSignup(w http.ResponseWriter, r *http.Request) {
	a.authStart(w, r, true)
}

// authLogin redirects the users to login page after starting the 0Auth and OIDC flow
func (a *Authenticator) authLogin(w http.ResponseWriter, r *http.Request) {
	a.authStart(w, r, false)
}

// authStart starts an OAuth and OIDC flow that redirects the user for authentication with the auth provider.
// After auth, the user is redirected back to authLoginCallback, which in turn will redirect the user to "/".
// You can override the redirect destination by passing a `?redirect=URI` query to this endpoint.
func (a *Authenticator) authStart(w http.ResponseWriter, r *http.Request, signup bool) {
	// Generate random state for CSRF
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate state: %s", err), http.StatusInternalServerError)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	// Get auth cookie
	sess := a.cookies.Get(r, cookieName)

	// Set state in cookie
	sess.Values[cookieFieldState] = state

	// Set redirect URL in cookie to enable custom redirects after auth has completed
	redirect := r.URL.Query().Get("redirect")
	if redirect != "" {
		sess.Values[cookieFieldRedirect] = redirect
	}

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	// Redirect to auth provider
	redirectURL := a.oauth2.AuthCodeURL(state)
	if signup {
		// Set custom parameters for signup using AuthCodeOption
		customOption := oauth2.SetAuthURLParam("screen_hint", "signup")
		redirectURL = a.oauth2.AuthCodeURL(state, customOption)
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// authLoginCallback is called after the user has successfully authenticated with the auth provider.
// It validates the OAuth info, fetches user profile info, and creates/updates the user in our DB.
// It then issues a new user auth token and saves it in a cookie.
// Finally, it redirects the user to the location specified in the initial call to authLogin.
func (a *Authenticator) authLoginCallback(w http.ResponseWriter, r *http.Request) {
	// Get auth cookie
	sess := a.cookies.Get(r, cookieName)
	// Check that random state matches (for CSRF protection)
	if r.URL.Query().Get("state") != sess.Values[cookieFieldState] {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}
	delete(sess.Values, cookieFieldState)

	// Exchange authorization code for an oauth2 token
	oauthToken, err := a.oauth2.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to convert authorization code into a token: %s", err), http.StatusUnauthorized)
		return
	}

	// Extract and verify ID token (which contains the user's identity info)
	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token field in oauth2 token", http.StatusUnauthorized)
		return
	}

	oidcConfig := &oidc.Config{
		ClientID: a.oauth2.ClientID,
	}

	idToken, err := a.oidc.Verifier(oidcConfig).Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to verify ID token: %s", err), http.StatusInternalServerError)
		return
	}

	// Extract user profile information
	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	email, ok := profile["email"].(string)
	if !ok || email == "" {
		http.Error(w, "claim 'email' not found", http.StatusInternalServerError)
		return
	}
	emailVerified, ok := profile["email_verified"].(bool)
	if !ok {
		http.Error(w, "claim 'email_verified' not found", http.StatusInternalServerError)
		return
	}
	name, ok := profile["name"].(string)
	if !ok {
		http.Error(w, "claim 'name' not found", http.StatusInternalServerError)
		return
	}
	photoURL, ok := profile["picture"].(string)
	if !ok {
		http.Error(w, "claim 'picture' not found", http.StatusInternalServerError)
		return
	}

	// Check that the user's email is verified
	if !emailVerified {
		errorRedirect, err := url.JoinPath(a.opts.FrontendURL, "/-/auth/verify-email")
		if err != nil {
			internalServerError(w, fmt.Errorf("failed to email verify uri: %w", err))
			return
		}

		http.Redirect(w, r, errorRedirect, http.StatusTemporaryRedirect)
		return
	}

	// Create (or update) user in our DB
	user, err := a.admin.CreateOrUpdateUser(r.Context(), email, name, photoURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update user: %s", err), http.StatusInternalServerError)
		return
	}

	// If there's already a token in the cookie, revoke it (since we're now issuing a new one)
	oldAuthToken, ok := sess.Values[cookieFieldAccessToken].(string)
	if ok && oldAuthToken != "" {
		err := a.admin.RevokeAuthToken(r.Context(), oldAuthToken)
		if err != nil {
			a.logger.Info("failed to revoke old user auth token during new auth", zap.Error(err), observability.ZapCtx(r.Context()))
			// The old token was probably manually revoked. We can still continue.
		}
	}

	// Issue a new persistent auth token
	authToken, err := a.admin.IssueUserAuthToken(r.Context(), user.ID, database.AuthClientIDRillWeb, "Browser session", nil, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to issue API token: %s", err), http.StatusInternalServerError)
		return
	}

	// Set auth token in cookie
	sess.Values[cookieFieldAccessToken] = authToken.Token().String()

	// Get redirect destination
	redirect, ok := sess.Values[cookieFieldRedirect].(string)
	if !ok || redirect == "" {
		redirect = "/"
	}
	delete(sess.Values, cookieFieldRedirect)

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to UI (usually)
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

func (a *Authenticator) authWithToken(w http.ResponseWriter, r *http.Request) {
	// Get new auth token
	newToken := r.URL.Query().Get("token")
	if newToken == "" {
		http.Error(w, "token not provided", http.StatusBadRequest)
		return
	}

	// Get auth cookie
	sess := a.cookies.Get(r, cookieName)

	// If there's already a token in the cookie, and it's not the same one, revoke it (since we're now setting a new one).
	oldAuthToken, ok := sess.Values[cookieFieldAccessToken].(string)
	if ok && oldAuthToken != "" && oldAuthToken != newToken {
		err := a.admin.RevokeAuthToken(r.Context(), oldAuthToken)
		if err != nil {
			a.logger.Info("failed to revoke old user auth token during new auth", zap.Error(err), observability.ZapCtx(r.Context()))
			// The old token was probably manually revoked. We can still continue.
		}
	}

	// Set auth token in cookie
	sess.Values[cookieFieldAccessToken] = newToken

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to UI
	http.Redirect(w, r, a.opts.FrontendURL, http.StatusTemporaryRedirect)
}

// authLogout implements user logout. It revokes the current user auth token, then redirects to the auth provider's logout flow.
// Once the logout has completed, the auth provider will redirect the user to authLogoutCallback.
func (a *Authenticator) authLogout(w http.ResponseWriter, r *http.Request) {
	// Get auth cookie
	sess := a.cookies.Get(r, cookieName)
	// Revoke access token and clear in cookie
	authToken, ok := sess.Values[cookieFieldAccessToken].(string)
	if ok && authToken != "" {
		err := a.admin.RevokeAuthToken(r.Context(), authToken)
		if err != nil {
			a.logger.Info("failed to revoke user auth token during logout", zap.Error(err), observability.ZapCtx(r.Context()))
			// We should still continue to ensure the user is logged out on the auth provider as well.
		}
	}
	delete(sess.Values, cookieFieldAccessToken)

	// Set redirect URL in cookie to enable custom redirects after logout
	redirect := r.URL.Query().Get("redirect")
	if redirect != "" {
		sess.Values[cookieFieldRedirect] = redirect
	}

	// Save cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, fmt.Sprintf("failed to save session: %s", err), http.StatusInternalServerError)
		return
	}

	// Build auth provider logout URL
	logoutURL, err := url.Parse("https://" + a.opts.AuthDomain + "/v2/logout")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build callback endpoint for authLogoutCallback
	returnTo, err := url.JoinPath(a.opts.ExternalURL, "/auth/logout/callback")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to build callback URL: %s", err), http.StatusInternalServerError)
		return
	}

	// Redirect to auth provider's logout
	parameters := url.Values{}
	parameters.Add("returnTo", returnTo)
	parameters.Add("client_id", a.opts.AuthClientID)
	logoutURL.RawQuery = parameters.Encode()

	http.Redirect(w, r, logoutURL.String(), http.StatusTemporaryRedirect)
}

// authLogoutCallback is called when a logout flow iniated by authLogout has completed.
func (a *Authenticator) authLogoutCallback(w http.ResponseWriter, r *http.Request) {
	// Get auth cookie
	sess := a.cookies.Get(r, cookieName)
	// Get redirect destination
	redirect, ok := sess.Values[cookieFieldRedirect].(string)
	if !ok || redirect == "" {
		redirect = "/"
	}
	delete(sess.Values, cookieFieldRedirect)

	// Save updated cookie
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to UI (usually)
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}
