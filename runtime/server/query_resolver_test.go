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
			// Model
			"ad_bids.sql": `SELECT now() AS time, 'DA' AS country, 3 as price`,
			// Duckdb connector
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

	tt := []struct {
		name               string             // Test case name
		resolver           string             // Resolver name (e.g. ad_bids.sql - ad_bids)
		resolverProperties *structpb.Struct   // Resolver properties
		resolverArgs       *structpb.Struct   // Resolver arguments
		schema             []string           // Expected schema
		data               []*structpb.Struct // Expected data
		expectError        bool               // Expect an error
		code               codes.Code         // Expected gRPC error code
	}{
		{
			name:        "should fail with invalid resolver",
			resolver:    "invalid_resolver",
			expectError: true,
			code:        codes.Internal,
		},
		{
			name:        "should fail with missing sql query",
			resolver:    "sql",
			expectError: true,
			code:        codes.Internal, // Update the expected error code
		},
		{
			name:     "should succeed with a simple SQL query",
			resolver: "sql",
			resolverProperties: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"sql": structpb.NewStringValue("SELECT country FROM foo limit 1"),
				},
			},
			resolverArgs: &structpb.Struct{},
			schema:       []string{"country"},
			data: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"country": structpb.NewStringValue("DA"),
					},
				},
			},
		},
		// Test multiple columns
		{
			name:     "should succeed with multiple columns",
			resolver: "sql",
			resolverProperties: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"sql": structpb.NewStringValue("SELECT country, price FROM foo limit 1"),
				},
			},
			resolverArgs: &structpb.Struct{},
			schema:       []string{"country", "price"},
			data: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"country": structpb.NewStringValue("DA"),
						"price":   structpb.NewNumberValue(3.141),
					},
				},
			},
		},
		{
			name:     "should succeed with a simple SQL query with a WHERE clause",
			resolver: "sql",
			resolverProperties: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"sql": structpb.NewStringValue("SELECT country FROM foo WHERE country = 'DA'"),
				},
			},
			resolverArgs: &structpb.Struct{},
			schema:       []string{"country"},
			data: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"country": structpb.NewStringValue("DA"),
					},
				},
			},
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

			data := res.GetData()
			schema := res.GetSchema()

			// Check expected schema
			require.Equal(t, len(schema.Fields), len(tc.schema))
			for i, s := range tc.schema {
				require.Equal(t, schema.Fields[i].Name, s)
			}

			// Check expected data
			require.Equal(t, len(data), len(tc.data))
			for i, d := range data {
				require.Equal(t, d, tc.data[i])
			}
		})
	}
}
