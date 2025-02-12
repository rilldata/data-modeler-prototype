import { defineConfig } from "orval";

export default defineConfig({
  api: {
    input: "../proto/gen/rill/runtime/v1/runtime.swagger.yaml",
    output: {
      workspace: "./src/runtime-client/",
      target: "gen/index.ts",
      client: "svelte-query",
      mode: "tags-split",
      mock: false,
      prettier: true,
      override: {
        mutator: {
          path: "http-client.ts", // Relative to workspace path set above
          name: "httpClient",
        },
        // Override queries and mutations here
        operations: {
          // Turn MetricsViewMeta into a query even though it's a POST request
          QueryService_MetricsViewMeta: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_ColumnRollupInterval: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_ColumnTopK: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_ColumnTimeSeries: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_TableColumns: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_MetricsViewAggregation: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_MetricsViewTotals: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_MetricsViewTimeSeries: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_MetricsViewToplist: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_MetricsViewComparison: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_MetricsViewRows: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_MetricsViewTimeRange: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_MetricsViewSearch: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_MetricsViewTimeRanges: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_ResolveComponent: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          QueryService_ResolveCanvas: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
          RuntimeService_IssueDevJWT: {
            query: {
              useQuery: true,
              signal: true,
            },
          },
        },
      },
    },
  },
});
