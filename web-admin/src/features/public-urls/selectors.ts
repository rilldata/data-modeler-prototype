import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { createRuntimeServiceGetResource } from "@rilldata/web-common/runtime-client";

// Use the ListResources API to get the target dashboard
// The JWT generated via a "magic" token will only have access to one dashboard, so we can assume the first one is the correct one
export function useShareableURLMetricsView(
  instanceId: string,
  metricsViewName: string,
  enabled: boolean,
) {
  return createRuntimeServiceGetResource(
    instanceId,
    {
      "name.kind": ResourceKind.MetricsView,
      "name.name": metricsViewName,
    },
    {
      query: {
        select: (data) => data?.resource,
        enabled: !!instanceId && !!metricsViewName && enabled,
      },
    },
  );
}
