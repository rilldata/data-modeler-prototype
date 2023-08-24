import {
  MetricsExplorerEntity,
  metricsExplorerStore,
} from "@rilldata/web-common/features/dashboards/dashboard-stores";
import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  MetricsViewDimension,
  MetricsViewMeasure,
  RpcStatus,
  V1MetricsView,
  V1MetricsViewFilter,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/query-core";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { get, writable } from "svelte/store";
import { expect } from "vitest";

export const AD_BIDS_NAME = "AdBids";
export const AD_BIDS_SOURCE_NAME = "AdBids_Source";
export const AD_BIDS_MIRROR_NAME = "AdBids_mirror";

export const AD_BIDS_IMPRESSIONS_MEASURE = "impressions";
export const AD_BIDS_BID_PRICE_MEASURE = "bid_price";
export const AD_BIDS_PUBLISHER_COUNT_MEASURE = "publisher_count";
export const AD_BIDS_PUBLISHER_DIMENSION = "publisher";
export const AD_BIDS_DOMAIN_DIMENSION = "domain";
export const AD_BIDS_COUNTRY_DIMENSION = "country";
export const AD_BIDS_TIMESTAMP_DIMENSION = "timestamp";

export const AD_BIDS_INIT_MEASURES = [
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE,
    expression: "count(*)",
  },
  {
    name: AD_BIDS_BID_PRICE_MEASURE,
    expression: "avg(bid_price)",
  },
];
export const AD_BIDS_THREE_MEASURES = [
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE,
    expression: "count(*)",
  },
  {
    name: AD_BIDS_BID_PRICE_MEASURE,
    expression: "avg(bid_price)",
  },
  {
    name: AD_BIDS_PUBLISHER_COUNT_MEASURE,
    expression: "count_distinct(publisher)",
  },
];
export const AD_BIDS_INIT_DIMENSIONS = [
  {
    name: AD_BIDS_PUBLISHER_DIMENSION,
  },
  {
    name: AD_BIDS_DOMAIN_DIMENSION,
  },
];
export const AD_BIDS_THREE_DIMENSIONS = [
  {
    name: AD_BIDS_PUBLISHER_DIMENSION,
  },
  {
    name: AD_BIDS_DOMAIN_DIMENSION,
  },
  {
    name: AD_BIDS_COUNTRY_DIMENSION,
  },
];

const Hour = 1000 * 60 * 60;
export const TestTimeConstants = {
  NOW: new Date(),
  LAST_6_HOURS: new Date(Date.now() - Hour * 6),
  LAST_12_HOURS: new Date(Date.now() - Hour * 12),
  LAST_18_HOURS: new Date(Date.now() - Hour * 18),
  LAST_DAY: new Date(Date.now() - Hour * 24),
};
export const AD_BIDS_DEFAULT_TIME_RANGE = {
  name: TimeRangePreset.ALL_TIME,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
  start: TestTimeConstants.LAST_DAY,
  end: new Date(TestTimeConstants.NOW.getTime() + 1),
};
export const AD_BIDS_DEFAULT_URL_TIME_RANGE = {
  name: TimeRangePreset.ALL_TIME,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
  start: undefined,
  end: undefined,
};

export const AD_BIDS_INIT: V1MetricsView = {
  name: "AdBids",
  model: "AdBids_Source",
  measures: AD_BIDS_INIT_MEASURES,
  dimensions: AD_BIDS_INIT_DIMENSIONS,
};
export const AD_BIDS_INIT_WITH_TIME: V1MetricsView = {
  ...AD_BIDS_INIT,
  timeDimension: AD_BIDS_TIMESTAMP_DIMENSION,
};
export const AD_BIDS_WITH_DELETED_MEASURE = {
  name: "AdBids",
  measures: [
    {
      name: AD_BIDS_IMPRESSIONS_MEASURE,
      expression: "count(*)",
    },
  ],
  dimensions: AD_BIDS_INIT_DIMENSIONS,
};
export const AD_BIDS_WITH_THREE_MEASURES = {
  name: "AdBids",
  measures: AD_BIDS_THREE_MEASURES,
  dimensions: AD_BIDS_INIT_DIMENSIONS,
};
export const AD_BIDS_WITH_DELETED_DIMENSION = {
  name: "AdBids",
  measures: AD_BIDS_INIT_MEASURES,
  dimensions: [
    {
      name: AD_BIDS_PUBLISHER_DIMENSION,
    },
  ],
};
export const AD_BIDS_WITH_THREE_DIMENSIONS = {
  name: "AdBids",
  measures: AD_BIDS_INIT_MEASURES,
  dimensions: AD_BIDS_THREE_DIMENSIONS,
};

export function resetDashboardStore() {
  metricsExplorerStore.remove(AD_BIDS_NAME);
  metricsExplorerStore.remove(AD_BIDS_MIRROR_NAME);
  initAdBidsInStore();
  initAdBidsMirrorInStore();
}

export function initAdBidsInStore() {
  metricsExplorerStore.init(AD_BIDS_NAME, AD_BIDS_INIT, {
    timeRangeSummary: {
      min: TestTimeConstants.LAST_DAY.toISOString(),
      max: TestTimeConstants.NOW.toISOString(),
      interval: V1TimeGrain.TIME_GRAIN_MINUTE as any,
    },
  });
}
export function initAdBidsMirrorInStore() {
  metricsExplorerStore.init(
    AD_BIDS_MIRROR_NAME,
    {
      name: AD_BIDS_MIRROR_NAME,
      measures: AD_BIDS_INIT_MEASURES,
      dimensions: AD_BIDS_INIT_DIMENSIONS,
    },
    {
      timeRangeSummary: {
        min: TestTimeConstants.LAST_DAY.toISOString(),
        max: TestTimeConstants.NOW.toISOString(),
        interval: V1TimeGrain.TIME_GRAIN_MINUTE as any,
      },
    }
  );
}

export function createDashboardState(
  name: string,
  metrics: V1MetricsView,
  filters: V1MetricsViewFilter = {
    include: [],
    exclude: [],
  },
  timeRange: DashboardTimeControls = AD_BIDS_DEFAULT_TIME_RANGE
): MetricsExplorerEntity {
  return {
    name,

    selectedMeasureNames: [],
    visibleDimensionKeys: new Set(metrics.dimensions.map((d) => d.name)),
    allDimensionsVisible: true,
    visibleMeasureKeys: new Set(metrics.measures.map((m) => m.name)),
    allMeasuresVisible: true,

    filters,
    dimensionFilterExcludeMode: new Map(),

    leaderboardMeasureName: metrics.measures[0]?.name,
    leaderboardContextColumn: LeaderboardContextColumn.HIDDEN,

    selectedTimeRange: timeRange,
  };
}

export function createAdBidsMirrorInStore(metrics: V1MetricsView) {
  const proto = get(metricsExplorerStore).entities[AD_BIDS_NAME].proto;
  // actual url is not relevant here
  metricsExplorerStore.syncFromUrl(
    AD_BIDS_MIRROR_NAME,
    proto,
    metrics ?? { measures: [], dimensions: [] }
  );
}

export function createMetricsMetaQueryMock(
  shouldInit = true
): CreateQueryResult<V1MetricsView, RpcStatus> & {
  setMeasures: (measures: Array<MetricsViewMeasure>) => void;
  setDimensions: (dimensions: Array<MetricsViewDimension>) => void;
} {
  const { update, subscribe } = writable<
    QueryObserverResult<V1MetricsView, RpcStatus>
  >({
    data: undefined,
    isSuccess: false,
    isRefetching: false,
  } as any);

  const mock = {
    subscribe,
    setMeasures: (measures) =>
      update((value) => {
        value.isSuccess = true;
        value.data ??= {
          measures: [],
          dimensions: [],
        };
        value.data.measures = measures;
        return value;
      }),
    setDimensions: (dimensions: Array<MetricsViewDimension>) =>
      update((value) => {
        value.isSuccess = true;
        value.data ??= {
          measures: [],
          dimensions: [],
        };
        value.data.dimensions = dimensions;
        return value;
      }),
  };

  if (shouldInit) {
    mock.setMeasures(AD_BIDS_INIT_MEASURES);
    mock.setDimensions(AD_BIDS_INIT_DIMENSIONS);
  }

  return mock;
}

// Wrapper function to simplify assert call
export function assertMetricsView(
  name: string,
  filters: V1MetricsViewFilter = {
    include: [],
    exclude: [],
  },
  timeRange: DashboardTimeControls = AD_BIDS_DEFAULT_TIME_RANGE,
  selectedMeasure = AD_BIDS_IMPRESSIONS_MEASURE
) {
  assertMetricsViewRaw(name, filters, timeRange, selectedMeasure);
}
// Raw assert function without any optional params.
// This allows us to assert for `undefined`
// TODO: find a better solution that this hack
export function assertMetricsViewRaw(
  name: string,
  filters: V1MetricsViewFilter,
  timeRange: DashboardTimeControls,
  selectedMeasure: string
) {
  const metricsView = get(metricsExplorerStore).entities[name];
  expect(metricsView.filters).toEqual(filters);
  expect(metricsView.selectedTimeRange).toEqual(timeRange);
  expect(metricsView.leaderboardMeasureName).toEqual(selectedMeasure);
}

export function assertVisiblePartsOfMetricsView(
  name: string,
  measures: Array<string> | undefined,
  dimensions: Array<string> | undefined
) {
  const metricsView = get(metricsExplorerStore).entities[name];
  if (measures)
    expect([...metricsView.visibleMeasureKeys].sort()).toEqual(measures.sort());
  if (dimensions)
    expect([...metricsView.visibleDimensionKeys].sort()).toEqual(
      dimensions.sort()
    );
}

export const AD_BIDS_BASE_FILTER = {
  include: [
    {
      name: AD_BIDS_PUBLISHER_DIMENSION,
      in: ["Google", "Facebook"],
    },
    {
      name: AD_BIDS_DOMAIN_DIMENSION,
      in: ["google.com"],
    },
  ],
  exclude: [],
};

export const AD_BIDS_EXCLUDE_FILTER = {
  include: [
    {
      name: AD_BIDS_DOMAIN_DIMENSION,
      in: ["google.com"],
    },
  ],
  exclude: [
    {
      name: AD_BIDS_PUBLISHER_DIMENSION,
      in: ["Google", "Facebook"],
    },
  ],
};

export const AD_BIDS_CLEARED_FILTER = {
  include: [],
  exclude: [
    {
      name: AD_BIDS_PUBLISHER_DIMENSION,
      in: ["Google", "Facebook"],
    },
  ],
};

// parsed time controls won't have start & end
export const ALL_TIME_PARSED_TEST_CONTROLS = {
  name: TimeRangePreset.ALL_TIME,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
} as DashboardTimeControls;

export const LAST_6_HOURS_TEST_CONTROLS = {
  name: TimeRangePreset.LAST_SIX_HOURS,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
  start: TestTimeConstants.LAST_6_HOURS,
  end: TestTimeConstants.NOW,
} as DashboardTimeControls;

// parsed time controls won't have start & end
export const LAST_6_HOURS_TEST_PARSED_CONTROLS = {
  name: TimeRangePreset.LAST_SIX_HOURS,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
} as DashboardTimeControls;

export const CUSTOM_TEST_CONTROLS = {
  name: TimeRangePreset.CUSTOM,
  interval: V1TimeGrain.TIME_GRAIN_MINUTE,
  start: TestTimeConstants.LAST_18_HOURS,
  end: TestTimeConstants.LAST_12_HOURS,
} as DashboardTimeControls;
