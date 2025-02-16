package admin

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/pkg/urlutil"
)

// URLs centralizes parsing and formatting of URLs for the admin service.
//
// There are several complexities around URL handling in Rill:
//  1. The frontend may run on a different host than the admin service (e.g. ui.rilldata.com vs. admin.rilldata.com).
//  2. We support custom domains for specific orgs (e.g. analytics.mycompany.com instead of ui.rilldata.com/mycompany).
//  3. The admin service sends transactional emails that link to the frontend, such as project invites.
//  4. The admin service is also responsible for sending transactional emails on behalf of the runtime, which also link to the frontend, such as for alerts and reports.
//  5. We need to ensure correct redirects and callbacks for the auth service (on auth.rilldata.com) and Github.
//     These services have fixed callback URLs on the admin service's primary external URL, which complicates custom domain handling.
//
// For orgs with a custom domain configured (using the CLI command `rill sudo org set-custom-domain`),
// we require the admin service and frontend to be reachable on the custom domain using the following load balancer rules:
//  1. The admin service must be reachable at the `/api` path prefix on the custom domain.
//     The `/api` prefix should be removed by the load balancer before proxying to the admin service.
//  2. The frontend must be reachable at all other paths on the custom domain.
type URLs struct {
	external string // The primary external URL for the admin service (with scheme).
	frontend string // The primary frontend URL for the admin service (with scheme).
	custom   string // Custom domain for the current org. Can optionally be set with WithCustomDomain.
	https    bool   // True if HTTPS should be used.
}

// NewURLs creates a new URLs. The provided URLs should include the scheme, host, optional port, and optional path prefix.
// The provided URLs should be the primary external and frontend URL for the Rill service. The returned *URLs will rewrite them as needed for custom domains.
func NewURLs(externalURL, frontendURL string) (*URLs, error) {
	// NOTE: This is a temporary hack for local development.
	// In local development, the gRPC and REST endpoints are served on different ports.
	// And the external URL is configured to point to the gRPC endpoint.
	// TODO: Move both gRPC and REST to the same port for local development.
	if strings.HasPrefix(externalURL, "http://localhost:9090") {
		externalURL = "http://localhost:8080"
	}

	eu, err := url.Parse(externalURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse external URL: %w", err)
	}

	_, err = url.Parse(frontendURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontend URL: %w", err)
	}

	return &URLs{
		external: externalURL,
		frontend: frontendURL,
		https:    eu.Scheme == "https",
	}, nil
}

// WithCustomDomain returns a copy that generates URLs for the provided custom domain (as described in the type doc).
// The result automatically generates correct URLs also for the few endpoints that must always use the non-custom external URL (such as AuthLogin).
func (u *URLs) WithCustomDomain(domain string) *URLs {
	if u.custom != "" {
		panic(fmt.Errorf("nested calls to WithCustomDomain are not allowed"))
	}

	if domain == "" {
		return u
	}

	custom := &url.URL{
		Scheme: "https",
		Host:   domain,
	}
	if !u.https {
		custom.Scheme = "http"
	}

	return &URLs{
		external: u.external,
		frontend: u.frontend,
		custom:   custom.String(),
		https:    u.https,
	}
}

// WithCustomDomainFromURL attempts to infer a custom domain from a redirect URL.
// If it succeeds, it passes the custom domain to WithCustomDomain and returns the result.
// If it does not detect a custom domain in the redirect URL, or the redirect URL is invalid, it fails silently by returning itself unchanged.
func (u *URLs) WithCustomDomainFromRedirectURL(redirectURL string) *URLs {
	u2, err := url.Parse(redirectURL)
	if err != nil {
		// Ignoring err as per docstring.
		return u
	}

	// Skip if there's no host in the redirect URL.
	if u2.Host == "" {
		return u
	}

	// Skip if it points to the primary external or frontend URL.
	if strings.HasPrefix(redirectURL, u.external) || strings.HasPrefix(redirectURL, u.frontend) {
		return u
	}

	return u.WithCustomDomain(u2.Host)
}

// IsHTTPS returns true if the admin service's external URL uses HTTPS.
func (u *URLs) IsHTTPS() bool {
	return u.https
}

// External returns the external URL for the admin service.
func (u *URLs) External() string {
	if u.custom != "" {
		// As described in the type doc, the admin service is required to be reachable at the `/api` path prefix on a custom domain.
		return urlutil.MustJoinURL(u.custom, "api")
	}
	return u.external
}

// Frontend returns the frontend URL for the admin service.
func (u *URLs) Frontend() string {
	if u.custom != "" {
		return u.custom
	}
	return u.frontend
}

// AuthLogin returns the URL that starts the redirects to the auth service for login.
func (u *URLs) AuthLogin(redirect string) string {
	res := urlutil.MustJoinURL(u.external, "/auth/login") // NOTE: Always using the primary external URL.
	if redirect != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"redirect": redirect})
	}
	return res
}

// AuthLoginCallback returns the URL for the OAuth2 callback.
func (u *URLs) AuthLoginCallback() string {
	return urlutil.MustJoinURL(u.external, "/auth/callback") // NOTE: Always using the primary external URL.
}

// AuthLogout returns the URL that starts the logout redirects.
func (u *URLs) AuthLogout() string {
	return urlutil.MustJoinURL(u.External(), "/auth/logout") // NOTE: Uses custom domain if set to correctly clear cookies.
}

// AuthLogoutProvider returns the URL that starts the logout redirects against the external auth provider.
func (u *URLs) AuthLogoutProvider(redirect string) string {
	res := urlutil.MustJoinURL(u.external, "/auth/logout/provider") // NOTE: Always using the primary external URL.
	if redirect != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"redirect": redirect})
	}
	return res
}

// AuthLogoutCallback returns the URL for the logout callback.
func (u *URLs) AuthLogoutCallback() string {
	return urlutil.MustJoinURL(u.external, "/auth/logout/callback") // NOTE: Always using the primary external URL.
}

// AuthWithToken returns a URL that sets the auth cookie to the provided token.
// Providing a redirect URL is optional.
func (u *URLs) AuthWithToken(tokenStr, redirect string) string {
	res := urlutil.MustJoinURL(u.External(), "/auth/with-token") // NOTE: Uses custom domain if set.
	res = urlutil.MustWithQuery(res, map[string]string{"token": tokenStr})
	if redirect != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"redirect": redirect})
	}
	return res
}

// AuthVerifyEmailUI returns the frontend URL for the verify email page.
func (u *URLs) AuthVerifyEmailUI() string {
	return urlutil.MustJoinURL(u.Frontend(), "/-/auth/verify-email")
}

// AuthVerifyDeviceUI returns the frontend URL for the device auth verification page.
func (u *URLs) AuthVerifyDeviceUI(query map[string]string) string {
	return urlutil.MustWithQuery(urlutil.MustJoinURL(u.Frontend(), "/-/auth/device"), query)
}

// AuthCLISuccessUI returns the frontend URL to redirect to after successful CLI authentication.
func (u *URLs) AuthCLISuccessUI() string {
	return urlutil.MustJoinURL(u.Frontend(), "/-/auth/cli/success")
}

// GithubConnect returns the URL that starts the Github connect redirects.
func (u *URLs) GithubConnect(remote string) string {
	res := urlutil.MustJoinURL(u.external, "/github/connect") // NOTE: Always using the primary external URL.
	if remote != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"remote": remote})
	}
	return res
}

// GithubAuth returns the URL that starts the Github auth redirects.
func (u *URLs) GithubAuth(remote string) string {
	res := urlutil.MustJoinURL(u.external, "/github/auth/login") // NOTE: Always using the primary external URL.
	if remote != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"remote": remote})
	}
	return res
}

// GithubAuthCallback returns the URL for the Github auth callback.
func (u *URLs) GithubAuthCallback() string {
	return urlutil.MustJoinURL(u.external, "/github/auth/callback") // NOTE: Always using the primary external URL.
}

// GithubConnectUI returns the page in the Rill frontend for starting the Github connect flow.
func (u *URLs) GithubConnectUI(redirect string) string {
	res := urlutil.MustJoinURL(u.frontend, "/-/github/connect") // NOTE: Always using the primary frontend URL.
	if redirect != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"redirect": redirect})
	}
	return res
}

// GithubConnectRetryUI returns the page in the Rill frontend for retrying the Github connect flow.
func (u *URLs) GithubConnectRetryUI(remote string) string {
	res := urlutil.MustJoinURL(u.frontend, "/-/github/connect/retry-install") // NOTE: Always using the primary frontend URL.
	if remote != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"remote": remote})
	}
	return res
}

// GithubConnectRequestUI returns the page in the Rill frontend for requesting a Github connect.
func (u *URLs) GithubConnectRequestUI(remote string) string {
	res := urlutil.MustJoinURL(u.frontend, "/-/github/connect/request") // NOTE: Always using the primary frontend URL.
	if remote != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"remote": remote})
	}
	return res
}

// GithubConnectSuccessUI returns the page in the Rill frontend for a successful Github connect.
func (u *URLs) GithubConnectSuccessUI(autoclose bool) string {
	res := urlutil.MustJoinURL(u.frontend, "/-/github/connect/success") // NOTE: Always using the primary frontend URL.
	if autoclose {
		res = urlutil.MustWithQuery(res, map[string]string{"autoclose": "true"})
	}
	return res
}

// GithubRetryAuthUI returns the page in the Rill frontend for retrying the Github auth flow.
func (u *URLs) GithubRetryAuthUI(remote, username string) string {
	res := urlutil.MustJoinURL(u.frontend, "/-/github/connect/retry-auth") // NOTE: Always using the primary frontend URL.
	if remote != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"remote": remote})
	}
	if username != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"githubUsername": username})
	}
	return res
}

// Asset creates a URL for downloading the user-uploaded asset with the given ID.
func (u *URLs) Asset(assetID string) string {
	return urlutil.MustJoinURL(u.External(), "/v1/assets", assetID, "download")
}

// Embed creates a URL for embedding the frontend in an iframe.
func (u *URLs) Embed(query map[string]string) (string, error) {
	return urlutil.WithQuery(urlutil.MustJoinURL(u.Frontend(), "-", "embed"), query)
}

// Organization returns the URL for an org in the frontend.
func (u *URLs) Organization(org string) string {
	return urlutil.MustJoinURL(u.Frontend(), org)
}

// OrganizationInviteAccept returns the URL for accepting an organization invite.
func (u *URLs) OrganizationInviteAccept(org string) string {
	redirect := urlutil.MustJoinURL(u.Frontend(), org)                                                                     // NOTE: Redirecting to the custom domain if set.
	return urlutil.MustWithQuery(urlutil.MustJoinURL(u.external, "/auth/signup"), map[string]string{"redirect": redirect}) // NOTE: Always using the primary external URL.
}

// Project returns the URL for a project in the frontend.
func (u *URLs) Project(org, project string) string {
	return urlutil.MustJoinURL(u.Frontend(), org, project)
}

// ProjectInviteAccept returns the URL for accepting a project invite.
func (u *URLs) ProjectInviteAccept(org, project string) string {
	redirect := urlutil.MustJoinURL(u.Frontend(), org, project)                                                            // NOTE: Redirecting to the custom domain if set.
	return urlutil.MustWithQuery(urlutil.MustJoinURL(u.external, "/auth/signup"), map[string]string{"redirect": redirect}) // NOTE: Always using the primary external URL.
}

// MagicAuthTokenOpen returns the frontend URL for opening a magic auth token.
func (u *URLs) MagicAuthTokenOpen(org, project, token string) string {
	return urlutil.MustJoinURL(u.Frontend(), org, project, "-", "share", token)
}

// ApproveProjectAccess returns the frontend URL for approving a project access request.
func (u *URLs) ApproveProjectAccess(org, project, id string) string {
	return urlutil.MustJoinURL(u.Frontend(), org, project, "-", "request-access", id, "approve")
}

// DenyProjectAccess returns the frontend URL for denying a project access request.
func (u *URLs) DenyProjectAccess(org, project, id string) string {
	return urlutil.MustJoinURL(u.Frontend(), org, project, "-", "request-access", id, "deny")
}

// ReportOpen returns the URL for opening a report in the frontend.
func (u *URLs) ReportOpen(org, project, report, token string, executionTime time.Time) string {
	if token == "" {
		return urlutil.MustWithQuery(urlutil.MustJoinURL(u.Frontend(), org, project, "-", "reports", report, "open"), map[string]string{"execution_time": executionTime.UTC().Format(time.RFC3339)})
	}
	return urlutil.MustWithQuery(urlutil.MustJoinURL(u.Frontend(), org, project, "-", "reports", report, "open"), map[string]string{"execution_time": executionTime.UTC().Format(time.RFC3339), "token": token})
}

// ReportExport returns the URL for exporting a report in the frontend.
func (u *URLs) ReportExport(org, project, report, token string) string {
	return urlutil.MustWithQuery(urlutil.MustJoinURL(u.Frontend(), org, project, "-", "reports", report, "export"), map[string]string{"token": token})
}

// ReportUnsubscribe returns the URL for unsubscribing from the report.
func (u *URLs) ReportUnsubscribe(org, project, report, token, email string) string {
	queryParams := map[string]string{"token": token}
	if email != "" {
		queryParams["email"] = email
	}
	return urlutil.MustWithQuery(urlutil.MustJoinURL(u.Frontend(), org, project, "-", "reports", report, "unsubscribe"), queryParams)
}

// ReportEdit returns the URL for editing a report in the frontend.
func (u *URLs) ReportEdit(org, project, report string) string {
	return urlutil.MustJoinURL(u.Frontend(), org, project, "-", "reports", report)
}

// AlertOpen returns the URL for opening an alert in the frontend.
func (u *URLs) AlertOpen(org, project, alert string) string {
	return urlutil.MustJoinURL(u.Frontend(), org, project, "-", "alerts", alert, "open")
}

// AlertEdit returns the URL for editing an alert in the frontend.
func (u *URLs) AlertEdit(org, project, alert string) string {
	return urlutil.MustJoinURL(u.Frontend(), org, project, "-", "alerts", alert)
}

// Billing returns the landing page url that optionally shows the upgrade modal.
func (u *URLs) Billing(org string, upgrade bool) string {
	bu := urlutil.MustJoinURL(u.Frontend(), org, "-", "settings", "billing")
	if upgrade {
		return urlutil.MustWithQuery(bu, map[string]string{"upgrade": "true"})
	}
	return bu
}

// PaymentPortal returns the landing page url that redirects user to payment portal
// Since the payment link can expire it is generated in this landing page on demand.
func (u *URLs) PaymentPortal(org string) string {
	return urlutil.MustJoinURL(u.Frontend(), org, "-", "settings", "billing", "payment")
}
