import { isMetricsViewQuery } from "@rilldata/web-local/lib/svelte-query/invalidation";
import type { QueryClient } from "@tanstack/svelte-query";

export function cancelDashboardQueries(
  queryClient: QueryClient,
  metricsViewName: string
) {
  return queryClient.cancelQueries({
    fetchStatus: "fetching",
    predicate: (query) => {
      return isMetricsViewQuery(query.queryHash, metricsViewName);
    },
  });
}
