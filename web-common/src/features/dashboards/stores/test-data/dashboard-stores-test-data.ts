import type { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { createStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getLocalIANA } from "@rilldata/web-common/lib/time/timezone";
import {
  getOffset,
  getStartOfPeriod,
} from "@rilldata/web-common/lib/time/transforms";
import {
  DashboardTimeControls,
  Period,
  TimeOffsetType,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  RpcStatus,
  TypeCode,
  V1Expression,
  V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  type V1StructType,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/query-core";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { QueryClient } from "@tanstack/svelte-query";
import { deepClone } from "@vitest/utils";
import { get, writable } from "svelte/store";
import { expect } from "vitest";

export const AD_BIDS_NAME = "AdBids";
export const AD_BIDS_SOURCE_NAME = "AdBids_Source";
export const AD_BIDS_MIRROR_NAME = "AdBids_mirror";

export const AD_BIDS_IMPRESSIONS_MEASURE = "impressions";
export const AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN = "impressions_no_grain";
export const AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN = "impressions_day_grain";
export const AD_BIDS_IMPRESSIONS_MEASURE_WINDOW = "impressions_window";
export const AD_BIDS_BID_PRICE_MEASURE = "bid_price";
export const AD_BIDS_PUBLISHER_COUNT_MEASURE = "publisher_count";
export const AD_BIDS_PUBLISHER_DIMENSION = "publisher";
export const AD_BIDS_DOMAIN_DIMENSION = "domain";
export const AD_BIDS_COUNTRY_DIMENSION = "country";
export const AD_BIDS_PUBLISHER_IS_NULL_DOMAIN = "publisher_is_null";
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
export const AD_BIDS_ADVANCED_MEASURES: MetricsViewSpecMeasureV2[] = [
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE,
    expression: "count(*)",
    requiredDimensions: [],
  },
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN,
    expression: "",
    requiredDimensions: [
      {
        name: AD_BIDS_TIMESTAMP_DIMENSION,
        timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
      },
    ],
  },
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
    expression: "",
    requiredDimensions: [
      {
        name: AD_BIDS_TIMESTAMP_DIMENSION,
        timeGrain: V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      },
    ],
  },
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE_WINDOW,
    expression: "",
    window: {
      partition: true,
    },
    requiredDimensions: [],
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
export const TestTimeOffsetConstants = {
  NOW: getOffsetByHour(TestTimeConstants.NOW),
  LAST_6_HOURS: getOffsetByHour(TestTimeConstants.LAST_6_HOURS),
  LAST_12_HOURS: getOffsetByHour(TestTimeConstants.LAST_12_HOURS),
  LAST_18_HOURS: getOffsetByHour(TestTimeConstants.LAST_18_HOURS),
  LAST_DAY: getOffsetByHour(TestTimeConstants.LAST_DAY),
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
};
export const AD_BIDS_TIME_RANGE_SUMMARY: V1MetricsViewTimeRangeResponse = {
  timeRangeSummary: {
    min: TestTimeConstants.LAST_DAY.toISOString(),
    max: TestTimeConstants.NOW.toISOString(),
    interval: V1TimeGrain.TIME_GRAIN_MINUTE as any,
  },
};

export const AD_BIDS_INIT: V1MetricsViewSpec = {
  title: "AdBids",
  table: "AdBids_Source",
  measures: AD_BIDS_INIT_MEASURES,
  dimensions: AD_BIDS_INIT_DIMENSIONS,
};
export const AD_BIDS_INIT_WITH_TIME: V1MetricsViewSpec = {
  ...AD_BIDS_INIT,
  timeDimension: AD_BIDS_TIMESTAMP_DIMENSION,
};
export const AD_BIDS_WITH_DELETED_MEASURE: V1MetricsViewSpec = {
  title: "AdBids",
  table: "AdBids_Source",
  measures: [
    {
      name: AD_BIDS_IMPRESSIONS_MEASURE,
      expression: "count(*)",
    },
  ],
  dimensions: AD_BIDS_INIT_DIMENSIONS,
};
export const AD_BIDS_WITH_THREE_MEASURES: V1MetricsViewSpec = {
  title: "AdBids",
  table: "AdBids_Source",
  measures: AD_BIDS_THREE_MEASURES,
  dimensions: AD_BIDS_INIT_DIMENSIONS,
};
export const AD_BIDS_WITH_DELETED_DIMENSION: V1MetricsViewSpec = {
  title: "AdBids",
  table: "AdBids_Source",
  measures: AD_BIDS_INIT_MEASURES,
  dimensions: [
    {
      name: AD_BIDS_PUBLISHER_DIMENSION,
    },
  ],
};
export const AD_BIDS_WITH_THREE_DIMENSIONS: V1MetricsViewSpec = {
  title: "AdBids",
  table: "AdBids_Source",
  measures: AD_BIDS_INIT_MEASURES,
  dimensions: AD_BIDS_THREE_DIMENSIONS,
};
export const AD_BIDS_WITH_BOOL_DIMENSION: V1MetricsViewSpec = {
  title: "AdBids",
  table: "AdBids_Source",
  measures: AD_BIDS_INIT_MEASURES,
  dimensions: [
    ...AD_BIDS_INIT_DIMENSIONS,
    {
      name: AD_BIDS_PUBLISHER_IS_NULL_DOMAIN,
      expression: "case when publisher is null then true else false end",
    },
  ],
};

export const AD_BIDS_SCHEMA: V1StructType = {
  fields: [
    {
      name: AD_BIDS_PUBLISHER_DIMENSION,
      type: {
        code: TypeCode.CODE_STRING,
      },
    },
    {
      name: AD_BIDS_DOMAIN_DIMENSION,
      type: {
        code: TypeCode.CODE_STRING,
      },
    },
    {
      name: AD_BIDS_COUNTRY_DIMENSION,
      type: {
        code: TypeCode.CODE_STRING,
      },
    },
    {
      name: AD_BIDS_PUBLISHER_IS_NULL_DOMAIN,
      type: {
        code: TypeCode.CODE_BOOL,
      },
    },
    {
      name: AD_BIDS_IMPRESSIONS_MEASURE,
      type: {
        code: TypeCode.CODE_INT64,
      },
    },
    {
      name: AD_BIDS_BID_PRICE_MEASURE,
      type: {
        code: TypeCode.CODE_FLOAT64,
      },
    },
  ],
};

export function resetDashboardStore() {
  metricsExplorerStore.remove(AD_BIDS_NAME);
  metricsExplorerStore.remove(AD_BIDS_MIRROR_NAME);
  initAdBidsInStore();
  initAdBidsMirrorInStore();
}

export function initAdBidsInStore() {
  metricsExplorerStore.init(
    AD_BIDS_NAME,
    AD_BIDS_INIT,
    AD_BIDS_TIME_RANGE_SUMMARY,
  );
}
export function initAdBidsMirrorInStore() {
  metricsExplorerStore.init(
    AD_BIDS_MIRROR_NAME,
    {
      measures: AD_BIDS_INIT_MEASURES,
      dimensions: AD_BIDS_INIT_DIMENSIONS,
    },
    AD_BIDS_TIME_RANGE_SUMMARY,
  );
}

export function createDashboardState(
  name: string,
  metrics: V1MetricsViewSpec,
  whereFilter: V1Expression = createAndExpression([]),
  timeRange: DashboardTimeControls = AD_BIDS_DEFAULT_TIME_RANGE,
): MetricsExplorerEntity {
  const explorer = getDefaultMetricsExplorerEntity(name, metrics, undefined);
  explorer.whereFilter = whereFilter;
  explorer.selectedTimeRange = timeRange;
  return explorer;
}

export function createAdBidsMirrorInStore(metrics: V1MetricsViewSpec) {
  const proto = get(metricsExplorerStore).entities[AD_BIDS_NAME].proto ?? "";
  // actual url is not relevant here
  metricsExplorerStore.syncFromUrl(
    AD_BIDS_MIRROR_NAME,
    proto,
    metrics ?? { measures: [], dimensions: [] },
    AD_BIDS_SCHEMA,
  );
}

export function createMetricsMetaQueryMock(
  shouldInit = true,
): CreateQueryResult<V1MetricsViewSpec, RpcStatus> & {
  setMeasures: (measures: Array<MetricsViewSpecMeasureV2>) => void;
  setDimensions: (dimensions: Array<MetricsViewSpecDimensionV2>) => void;
} {
  const { update, subscribe } = writable<
    QueryObserverResult<V1MetricsViewSpec, RpcStatus>
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
    setDimensions: (dimensions: Array<MetricsViewSpecDimensionV2>) =>
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
  filters: V1Expression = createAndExpression([]),
  timeRange: DashboardTimeControls = AD_BIDS_DEFAULT_TIME_RANGE,
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
  timeRange: DashboardTimeControls | null,
  selectedMeasure: string,
) {
  const metricsView = get(metricsExplorerStore).entities[name];
  expect(metricsView.whereFilter).toEqual(filters);
  expect(metricsView.selectedTimeRange).toEqual(timeRange);
  expect(metricsView.leaderboardMeasureName).toEqual(selectedMeasure);
}

export function assertVisiblePartsOfMetricsView(
  name: string,
  measures: Array<string> | undefined,
  dimensions: Array<string> | undefined,
) {
  const metricsView = get(metricsExplorerStore).entities[name];
  if (measures)
    expect([...metricsView.visibleMeasureKeys].sort()).toEqual(measures.sort());
  if (dimensions)
    expect([...metricsView.visibleDimensionKeys].sort()).toEqual(
      dimensions.sort(),
    );
}

export function getOffsetByHour(time: Date) {
  return getOffset(
    getStartOfPeriod(time, Period.HOUR, getLocalIANA()),
    Period.HOUR,
    TimeOffsetType.ADD,
    getLocalIANA(),
  );
}

export function initStateManagers(
  dashboardFetchMocks?: DashboardFetchMocks,
  resp?: V1MetricsViewSpec,
  name?: string,
) {
  initAdBidsInStore();
  if (dashboardFetchMocks && resp)
    dashboardFetchMocks.mockMetricsView(name ?? AD_BIDS_NAME, resp);

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
    metricsViewName: name ?? AD_BIDS_NAME,
  });

  return { stateManagers, queryClient };
}

export const AD_BIDS_BASE_FILTER = createAndExpression([
  createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google", "Facebook"]),
  createInExpression(AD_BIDS_DOMAIN_DIMENSION, ["google.com"]),
]);

export const AD_BIDS_EXCLUDE_FILTER = createAndExpression([
  createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google", "Facebook"], true),
  createInExpression(AD_BIDS_DOMAIN_DIMENSION, ["google.com"]),
]);

export const AD_BIDS_CLEARED_FILTER = createAndExpression([
  createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google", "Facebook"], true),
]);

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
