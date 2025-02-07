package server_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestServer_TestSimpleSQLQueryResolver(t *testing.T) {
	t.Parallel()

	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
			// Normal model
			"ad_bids.sql": `SELECT now() AS time, 'DA' AS country, 3.141 as price`,
			// Create a non-default duckdb connector
			"custom_duckdb.yaml": `
type: connector
driver: duckdb
`,
		},
	})

	// Create some externally managed tables
	olapExecAdhoc(t, rt, instanceID, "duckdb", "CREATE TABLE IF NOT EXISTS foo AS SELECT now() AS time, 'DA' AS country, 3.141 as price")
	olapExecAdhoc(t, rt, instanceID, "custom_duckdb", "CREATE TABLE IF NOT EXISTS foo AS SELECT now() AS time, 'DA' AS country, 3.141 as price")

	ctx, cancel := context.WithTimeout(testCtx(), 25*time.Second)
	defer cancel()

	_, release, err := rt.Repo(ctx, instanceID)
	require.NoError(t, err)
	defer release()

	server, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// resolverProperties map[string]interface{} // Resolver properties
	// resolverArgs       map[string]interface{} // Resolver arguments
	tt := []struct {
		name               string           // Test case name
		resolver           string           // Resolver name (e.g. ad_bids.sql - ad_bids)
		resolverProperties *structpb.Struct // Resolver properties
		resolverArgs       *structpb.Struct // Resolver arguments
		contains           []string         // Expected strings in the output
		expectError        bool             // Expect an error
		code               codes.Code       // Expected gRPC error code
	}{
		{
			name:        "should fail with invalid resolver",
			resolver:    "invalid_resolver",
			expectError: true,
			code:        codes.Internal,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(testCtx(), 25*time.Second)
			defer cancel()

			req := &runtimev1.QueryResolverRequest{
				InstanceId:         instanceID,
				Resolver:           tc.resolver,
				ResolverProperties: tc.resolverProperties,
				ResolverArgs:       tc.resolverArgs,
			}
			res, err := server.QueryResolver(ctx, req)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			for _, s := range tc.contains {
				require.Contains(t, res.Data, s)
			}
		})
	}
}
