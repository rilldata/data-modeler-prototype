package server_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAPI(t *testing.T) {
	// API YAMLs
	api := `
type: api
sql: SELECT COUNT(*) AS count FROM model
`
	apiWithSecurity := `
types: api
sql: SELECT COUNT(*) AS count FROM model
security:
  access: "{{ .user.admin }}"
`

	// Load the test runtime with mock files
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml":     "",                    // rill config
			"model.sql":     "SELECT 'foo' AS bar", // model
			"api.yaml":      api,                   // api file
			"security.yaml": apiWithSecurity,       // api file with security
		},
	})

	// Create externally managed tables
	olapExecAdhoc(t, rt, instanceID, "duckdb", "CREATE TABLE IF NOT EXISTS foo AS SELECT now() AS time, 'DA' AS country, 3.141 as price")

	// Context
	ctx, cancel := context.WithTimeout(testCtx(), 25*time.Second)
	defer cancel()

	// Repo
	repo, release, err := rt.Repo(ctx, instanceID)
	require.NoError(t, err)
	defer release()

	// Issuer and Audience
	audienceURL := "http://example.org"
	iss, _, close := testIssuerAndAudience(t, audienceURL)
	defer close()

	// Server options
	options := &server.Options{
		AuthEnable:      true,
		AuthIssuerURL:   iss.GetIssuerURL(),
		AuthAudienceURL: audienceURL,
	}

	// Create a Server
	server, err := server.NewServer(ctx, options, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	defer server.Close()

	// Test cases
	tt := []struct {
		name       string // test name
		api        string // api name "/v1/instances/{instance_id}/api/{name...}"
		method     string // http method
		endpoint   string // endpoint to call
		statusCode int    // expected status code
		auth       bool   // whether to authenticate
	}{{
		name:       "no security baseline",
		api:        "api.yaml",
		method:     "GET",
		endpoint:   fmt.Sprintf("/v1/instances/%s/api/api", instanceID),
		statusCode: http.StatusOK, // 200
		auth:       true,
	}, {
		name:       "with security",
		api:        "security.yaml",
		method:     "GET",
		endpoint:   fmt.Sprintf("/v1/instances/%s/api/security", instanceID),
		statusCode: http.StatusOK,
		auth:       true,
	}}

	// Test loop
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Expect the files to be resolved
			_, err := repo.Get(ctx, tc.api)
			require.NoError(t, err)

			// Call the endpoint
			req, err := http.NewRequest(tc.method, tc.endpoint, nil)
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			if tc.auth {
				token, err := iss.NewToken(auth.TokenOptions{
					AudienceURL:       audienceURL,
					Subject:           "token",
					TTL:               time.Hour,
					SystemPermissions: []auth.Permission{},
				})
				require.NoError(t, err)

				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				t.Logf("token: %s", token)
			}

			// Call the endpoint
			w := httptest.NewRecorder()
			res := w.Result()
			data, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if res.StatusCode != tc.statusCode {
				t.Errorf("expected status code %d, got %d", tc.statusCode, res.StatusCode)
				t.Logf("response: %s", string(data))
			}
		})
	}
}
