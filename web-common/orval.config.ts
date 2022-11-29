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
          RuntimeService_MetricsViewMeta: {
            query: {
              useQuery: true,
            },
          },
          RuntimeService_EstimateRollupInterval: {
            query: {
              useQuery: true,
            },
          },
          RuntimeService_GetTopK: {
            query: {
              useQuery: true,
            },
          },
          RuntimeService_GenerateTimeSeries: {
            query: {
              useQuery: true,
            },
          },
          RuntimeService_ProfileColumns: {
            query: {
              useQuery: true,
            },
          },
          RuntimeService_MetricsViewTotals: {
            query: {
              useQuery: true,
            },
          },
        },
      },
    },
  },
});
