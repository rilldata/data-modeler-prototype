package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func getTestServer(t *testing.T) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstance(t)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func testCtx() context.Context {
	return auth.WithOpen(context.Background())
}

func testIssuerAndAudience(t *testing.T, audienceURL string) (*auth.Issuer, *auth.Audience, func()) {
	// Create Issuer and serve on a test server
	iss, err := auth.NewEphemeralIssuer("")
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.Handle("/.well-known/jwks.json", iss.WellKnownHandler())

	srv := httptest.NewServer(mux)
	iss.SetIssuerURL(srv.URL)

	// Create Audience
	aud, err := auth.OpenAudience(context.Background(), zap.NewNop(), srv.URL, audienceURL)
	require.NoError(t, err)

	return iss, aud, func() { srv.Close(); aud.Close() }
}
