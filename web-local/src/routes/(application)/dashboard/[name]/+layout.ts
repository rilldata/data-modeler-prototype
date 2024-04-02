import {
  queryServiceMetricsViewTimeRange,
  getQueryServiceMetricsViewTimeRangeQueryKey,
} from "@rilldata/web-common/runtime-client";
import type { QueryFunction } from "@tanstack/svelte-query";

import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export const load = async ({ parent, params }) => {
  const parentData = await parent();

  const instanceId = parentData.instance?.instanceId ?? "default";

  const dashboardName = params.name;

  const timeRangeQuery: QueryFunction<
    Awaited<ReturnType<typeof queryServiceMetricsViewTimeRange>>
  > = ({ signal }) =>
    queryServiceMetricsViewTimeRange(instanceId, dashboardName, {}, signal);

  const timeRange = await queryClient.fetchQuery({
    queryFn: timeRangeQuery,
    queryKey: getQueryServiceMetricsViewTimeRangeQueryKey(
      instanceId,
      dashboardName,
      {},
    ),
  });

  return {
    timeRange: timeRange.timeRangeSummary,
  };
};
