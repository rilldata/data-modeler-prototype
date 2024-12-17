import type { QueryObserverResult } from "@rilldata/svelte-query";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { timeControlStateSelector } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";

export function getAggregationAPIRequestForExplore(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeQuery: QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>,
  exploreState: MetricsExplorerEntity,
) {
  const timeControls = timeControlStateSelector([
    metricsViewSpec,
    exploreSpec,
    timeRangeQuery,
    exploreState,
  ]);

  const timeRange = {
    start: timeControls.timeStart,
    end: timeControls.timeEnd,
  };

  const comparisonTimeRange = timeControls.showTimeComparison
    ? {
        start: timeControls.comparisonTimeStart,
        end: timeControls.comparisonTimeEnd,
      }
    : undefined;
}
