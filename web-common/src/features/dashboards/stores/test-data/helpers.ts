import { QueryClient } from "@rilldata/svelte-query";
import { createStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  AD_BIDS_DEFAULT_TIME_RANGE,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_MIRROR_NAME,
  AD_BIDS_NAME,
  AD_BIDS_SCHEMA,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import type { ExploreValidSpecResponse } from "@rilldata/web-common/features/explores/selectors";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  type V1ExploreSpec,
  type V1Expression,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";
import { deepClone } from "@vitest/utils";
import { get } from "svelte/store";
import { expect } from "vitest";

export function resetDashboardStore() {
  metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  metricsExplorerStore.remove(AD_BIDS_MIRROR_NAME);
  initAdBidsInStore();
  initAdBidsMirrorInStore();
}

export function initAdBidsInStore() {
  metricsExplorerStore.init(
    AD_BIDS_EXPLORE_NAME,
    getInitExploreStateForTest(
      AD_BIDS_METRICS_INIT,
      AD_BIDS_EXPLORE_INIT,
      AD_BIDS_TIME_RANGE_SUMMARY,
    ),
  );
}

export function initAdBidsMirrorInStore() {
  metricsExplorerStore.init(
    AD_BIDS_MIRROR_NAME,
    getInitExploreStateForTest(
      AD_BIDS_METRICS_INIT,
      AD_BIDS_EXPLORE_INIT,
      AD_BIDS_TIME_RANGE_SUMMARY,
    ),
  );
}

export function getInitExploreStateForTest(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1MetricsViewTimeRangeResponse | undefined = undefined,
) {
  const defaultExplorePreset = getDefaultExplorePreset(
    exploreSpec,
    timeRangeSummary,
  );
  const { partialExploreState } = convertPresetToExploreState(
    metricsViewSpec,
    exploreSpec,
    defaultExplorePreset,
    [],
  );
  return partialExploreState;
}

export function createAdBidsMirrorInStore({
  metricsView,
  explore,
}: ExploreValidSpecResponse) {
  const proto =
    get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].proto ?? "";
  // actual url is not relevant here
  metricsExplorerStore.syncFromUrl(
    AD_BIDS_MIRROR_NAME,
    proto,
    metricsView ?? { measures: [], dimensions: [] },
    explore ?? { metricsView: AD_BIDS_NAME, measures: [], dimensions: [] },
    AD_BIDS_SCHEMA,
  );
}

// Wrapper function to simplify assert call
export function assertMetricsView(
  name: string,
  filters: V1Expression = createAndExpression([]),
  timeRange = {
    name: AD_BIDS_DEFAULT_TIME_RANGE.name,
    interval: AD_BIDS_DEFAULT_TIME_RANGE.interval,
  } as DashboardTimeControls,
  selectedMeasure = AD_BIDS_IMPRESSIONS_MEASURE,
) {
  assertMetricsViewRaw(name, filters, timeRange, selectedMeasure);
}

// Raw assert function without any optional params.
// This allows us to assert for `undefined`
// TODO: find a better solution that this hack
export function assertMetricsViewRaw(
  name: string,
  filters: V1Expression,
  timeRange: DashboardTimeControls | undefined,
  selectedMeasure: string,
) {
  const metricsView = get(metricsExplorerStore).entities[name];
  expect(metricsView.whereFilter).toEqual(filters);
  expect(metricsView.selectedTimeRange).toEqual(timeRange);
  expect(metricsView.leaderboardMeasureName).toEqual(selectedMeasure);
}

export function initStateManagers(metricsViewName?: string) {
  metricsViewName ??= AD_BIDS_NAME;
  const exploreName = metricsViewName + "_explore";

  initAdBidsInStore();

  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        refetchOnMount: false,
        refetchOnReconnect: false,
        refetchOnWindowFocus: false,
        retry: false,
        networkMode: "always",
      },
    },
  });
  const stateManagers = createStateManagers({
    queryClient,
    metricsViewName,
    exploreName,
  });

  return { stateManagers, queryClient };
}

export function getPartialDashboard(
  name: string,
  keys: (keyof MetricsExplorerEntity)[],
) {
  const dashboard = get(metricsExplorerStore).entities[name];
  const partialDashboard = {} as MetricsExplorerEntity;
  keys.forEach(
    (key) => ((partialDashboard as any)[key] = deepClone(dashboard[key])),
  );
  return partialDashboard;
}
