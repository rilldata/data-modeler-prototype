import { expect } from "@jest/globals";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  V1MetricsViewFilter,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export const AdBidsName = "AdBids";
export const AdBidsMirrorName = "AdBids_mirror";

export const AdBidsImpressionsMeasure = "impressions";
export const AdBidsBidPriceMeasure = "bid_price";
export const AdBidsPublisherDimension = "publisher";
export const AdBidsDomainDimension = "domain";

const Hour = 1000 * 60 * 60;
export const TestTimeConstants = {
  Now: new Date(),
  Last6Hours: new Date(Date.now() - Hour * 6),
  Last12Hours: new Date(Date.now() - Hour * 12),
  Last18Hours: new Date(Date.now() - Hour * 18),
  LastDay: new Date(Date.now() - Hour * 24),
};

export function createAdBidsInStore() {
  metricsExplorerStore.sync(AdBidsName, {
    name: "AdBids",
    measures: [
      {
        name: AdBidsImpressionsMeasure,
        expression: "count(*)",
      },
      {
        name: AdBidsBidPriceMeasure,
        expression: "sum(bid_price)",
      },
    ],
    dimensions: [
      {
        name: AdBidsPublisherDimension,
      },
      {
        name: AdBidsDomainDimension,
      },
    ],
  });
  // clear everything if already created
  metricsExplorerStore.clearFilters(AdBidsName);
  metricsExplorerStore.setSelectedTimeRange(AdBidsName, {
    name: TimeRangePreset.ALL_TIME,
    interval: V1TimeGrain.TIME_GRAIN_MINUTE,
    start: TestTimeConstants.LastDay,
    end: TestTimeConstants.Now,
  });
}

export function createAdBidsMirrorInStore() {
  const proto = get(metricsExplorerStore).entities[AdBidsName].proto;
  // actual url is not relevant here
  metricsExplorerStore.syncFromUrl(
    AdBidsMirrorName,
    new URL(`http://localhost/dashboard?state=${proto}`)
  );
}

export function assertMetricsView(
  name: string,
  filters: V1MetricsViewFilter = {
    include: [],
    exclude: [],
  },
  timeRange: DashboardTimeControls = {
    name: TimeRangePreset.ALL_TIME,
    interval: V1TimeGrain.TIME_GRAIN_MINUTE,
    start: TestTimeConstants.LastDay,
    end: TestTimeConstants.Now,
  },
  selectedMeasure = AdBidsImpressionsMeasure
) {
  const metricsView = get(metricsExplorerStore).entities[name];
  expect(metricsView.filters).toEqual(filters);
  expect(metricsView.selectedTimeRange).toEqual(timeRange);
  expect(metricsView.leaderboardMeasureName).toEqual(selectedMeasure);
}
