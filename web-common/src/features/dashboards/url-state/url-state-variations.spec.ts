import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  AD_BIDS_DIMENSION_TABLE_PRESET,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_PIVOT_PRESET,
  AD_BIDS_PRESET,
  AD_BIDS_TIME_DIMENSION_DETAILS_PRESET,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import {
  AD_BIDS_CLOSE_DIMENSION_TABLE,
  AD_BIDS_CLOSE_TDD,
  AD_BIDS_DISABLE_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_OPEN_DOM_DIMENSION_TABLE,
  AD_BIDS_OPEN_DOMAIN_BID_PRICE_PIVOT,
  AD_BIDS_OPEN_IMP_TDD,
  AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
  AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
  AD_BIDS_SET_ALL_TIME_RANGE_FILTER,
  AD_BIDS_SET_KATHMANDU_TIMEZONE,
  AD_BIDS_SET_LA_TIMEZONE,
  AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
  AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
  AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_SORT_ASC_BY_BID_PRICE,
  AD_BIDS_SORT_ASC_BY_IMPRESSIONS,
  AD_BIDS_SORT_BY_DELTA_ABS_VALUE,
  AD_BIDS_SORT_BY_PERCENT_VALUE,
  AD_BIDS_SORT_BY_VALUE,
  AD_BIDS_SORT_DESC_BY_BID_PRICE,
  AD_BIDS_SORT_DESC_BY_IMPRESSIONS,
  AD_BIDS_SORT_PIVOT_BY_IMPRESSIONS_DESC,
  AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
  AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD,
  AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
  AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
  AD_BIDS_TOGGLE_PIVOT,
  applyMutationsToDashboard,
  type TestDashboardMutation,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { convertURLToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  type V1ExplorePreset,
  type V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";
import { deepClone } from "@vitest/utils";
import { get } from "svelte/store";
import { beforeAll, beforeEach, describe, expect, it, vi } from "vitest";

vi.stubEnv("TZ", "UTC");

const TestCases: {
  title: string;
  mutations: TestDashboardMutation[];
  preset?: V1ExplorePreset;
  expectedUrl: string;
  // Mainly tests that close certain views.
  // Closing view would retain some state of the old view in protobuf state
  legacyNotSupported?: boolean;
}[] = [
  {
    title: "Time range without preset",
    mutations: [
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_KATHMANDU_TIMEZONE,
    ],
    expectedUrl: "http://localhost/?tr=P4W&tz=Asia%2FKathmandu&grain=week",
  },
  {
    title: "Time range with preset and state matching preset",
    mutations: [
      AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
      AD_BIDS_SET_KATHMANDU_TIMEZONE,
    ],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title: "Time range with preset and state not matching preset",
    mutations: [AD_BIDS_SET_P4W_TIME_RANGE_FILTER, AD_BIDS_SET_LA_TIMEZONE],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/?tr=P4W&tz=America%2FLos_Angeles&grain=week",
  },
  {
    title: "Time range with preset and ALL_TIME selected",
    mutations: [AD_BIDS_SET_ALL_TIME_RANGE_FILTER],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/?tr=inf",
  },

  {
    title: "Time range comparison without preset",
    mutations: [
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER,
    ],
    expectedUrl: "http://localhost/?tr=P4W&compare_tr=rill-PW&grain=week",
  },
  {
    title: "Time range comparison with preset and state matching preset",
    mutations: [
      AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
    ],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title: "Time range comparison with preset and state not matching preset",
    mutations: [
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER,
    ],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/?tr=P4W&compare_tr=rill-PW&grain=week",
  },
  {
    title: "Time range comparison enable and disable",
    mutations: [
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER,
      AD_BIDS_DISABLE_COMPARE_TIME_RANGE_FILTER,
    ],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/?tr=P4W&grain=week",
    legacyNotSupported: true,
  },

  {
    title:
      "Measures/dimensions visibility with no preset and partially visible measures/dimensions in state",
    mutations: [
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
    ],
    expectedUrl: "http://localhost/?measures=impressions&dims=publisher",
  },
  {
    title:
      "Measures/dimensions visibility with no preset and all measures/dimensions visible in state",
    mutations: [
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
      // re-toggle to show
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
    ],
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Measures/dimensions visibility with preset and partially visible measures/dimensions in state matching preset",
    mutations: [
      // initially hidden due to preset, show them now.
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
      // hide them back.
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
    ],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Measures/dimensions visibility with preset and all measures/dimensions visible in state not matching preset",
    mutations: [
      // initially hidden due to preset, show them now.
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
    ],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/?measures=*&dims=*",
  },

  {
    title:
      "Leaderboard configs with no preset and leaderboard sort measure in state different than default",
    mutations: [AD_BIDS_SORT_BY_DELTA_ABS_VALUE, AD_BIDS_SORT_ASC_BY_BID_PRICE],
    expectedUrl:
      "http://localhost/?sort_by=bid_price&sort_type=delta_abs&sort_dir=ASC",
  },
  {
    title:
      "Leaderboard configs with no preset and leaderboard sort measure in state same as default",
    mutations: [AD_BIDS_SORT_BY_VALUE, AD_BIDS_SORT_DESC_BY_IMPRESSIONS],
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Leaderboard configs with preset and leaderboard sort measure in state same as preset",
    mutations: [AD_BIDS_SORT_BY_PERCENT_VALUE, AD_BIDS_SORT_ASC_BY_IMPRESSIONS],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Leaderboard configs with preset and leaderboard sort measure in state different than preset",
    mutations: [
      AD_BIDS_SORT_BY_DELTA_ABS_VALUE,
      AD_BIDS_SORT_DESC_BY_BID_PRICE,
    ],
    preset: AD_BIDS_PRESET,
    expectedUrl:
      "http://localhost/?sort_by=bid_price&sort_type=delta_abs&sort_dir=DESC",
  },

  {
    title: "Dimension table with no preset and dimension table active in state",
    mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
    expectedUrl: "http://localhost/?expand_dim=publisher",
  },
  {
    title: "Dimension table with no preset and open and close dimension table",
    mutations: [
      AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
      AD_BIDS_CLOSE_DIMENSION_TABLE,
    ],
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Dimension table with preset and with dimension table in state same as preset",
    mutations: [AD_BIDS_OPEN_DOM_DIMENSION_TABLE],
    preset: AD_BIDS_DIMENSION_TABLE_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Dimension table with preset and with dimension table in state different than preset",
    mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
    preset: AD_BIDS_DIMENSION_TABLE_PRESET,
    expectedUrl: "http://localhost/?expand_dim=publisher",
  },
  {
    title:
      "Dimension table with preset and with no dimension table in state different than preset",
    mutations: [AD_BIDS_CLOSE_DIMENSION_TABLE],
    preset: AD_BIDS_DIMENSION_TABLE_PRESET,
    expectedUrl: "http://localhost/?expand_dim=",
    legacyNotSupported: true,
  },

  {
    title:
      "Time dimensional details with no preset and has time dimensional details in state",
    mutations: [AD_BIDS_OPEN_IMP_TDD, AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
    expectedUrl:
      "http://localhost/?view=tdd&measure=impressions&chart_type=stacked_bar",
  },
  {
    title: "Time dimensional details with no preset, open and close TDD",
    mutations: [AD_BIDS_OPEN_IMP_TDD, AD_BIDS_CLOSE_TDD],
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Time dimensional details with preset and has time dimensional details in state same as presets",
    mutations: [AD_BIDS_OPEN_IMP_TDD],
    preset: AD_BIDS_TIME_DIMENSION_DETAILS_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Time dimensional details with preset and has time dimensional details in state different than presets",
    mutations: [AD_BIDS_CLOSE_TDD],
    preset: AD_BIDS_TIME_DIMENSION_DETAILS_PRESET,
    expectedUrl: "http://localhost/?view=explore",
    legacyNotSupported: true,
  },

  {
    title: "Pivot with no preset and has pivot in state",
    mutations: [
      AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
      AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
    ],
    expectedUrl:
      "http://localhost/?view=pivot&rows=publisher%2Ctime.hour&cols=domain%2Ctime.day%2Cimpressions&sort_by=time.day&sort_dir=ASC",
  },
  {
    title: "Pivot with no preset, open and close pivot",
    mutations: [
      AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
      AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
      AD_BIDS_TOGGLE_PIVOT,
    ],
    expectedUrl: "http://localhost/",
    legacyNotSupported: true,
  },
  {
    title: "Pivot with preset and has pivot in state same as preset",
    mutations: [
      AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
      AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
    ],
    preset: AD_BIDS_PIVOT_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title: "Pivot with preset and pivot in state different as preset",
    mutations: [
      AD_BIDS_OPEN_DOMAIN_BID_PRICE_PIVOT,
      AD_BIDS_SORT_PIVOT_BY_IMPRESSIONS_DESC,
    ],
    preset: AD_BIDS_PIVOT_PRESET,
    expectedUrl:
      "http://localhost/?rows=domain%2Ctime.day&cols=impressions&sort_by=impressions&sort_dir=DESC",
  },
  {
    title: "Pivot with preset and no pivot in state different as preset",
    mutations: [AD_BIDS_TOGGLE_PIVOT],
    preset: AD_BIDS_PIVOT_PRESET,
    expectedUrl: "http://localhost/?view=explore",
    legacyNotSupported: true,
  },
];

describe("Human readable URL state variations", () => {
  beforeAll(() => {
    initLocalUserPreferenceStore(AD_BIDS_EXPLORE_NAME);
  });

  beforeEach(() => {
    sessionStorage.clear();
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  });

  describe("Should update url state and restore default state on empty params", () => {
    for (const { title, mutations, preset, expectedUrl } of TestCases) {
      it(title, () => {
        const explore: V1ExploreSpec = {
          ...AD_BIDS_EXPLORE_INIT,
          ...(preset ? { defaultPreset: preset } : {}),
          timeZones: ["UTC", "Asia/Kathmandu"],
        };
        metricsExplorerStore.init(
          AD_BIDS_EXPLORE_NAME,
          getInitExploreStateForTest(
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            explore,
            AD_BIDS_TIME_RANGE_SUMMARY,
          ),
        );
        const initState = getCleanMetricsExploreForAssertion();
        const defaultExplorePreset = getDefaultExplorePreset(
          explore,
          AD_BIDS_TIME_RANGE_SUMMARY,
        );

        applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);

        // load url params with updated metrics state
        const url = new URL("http://localhost");
        url.search = convertExploreStateToURLSearchParams(
          get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
          explore,
          getTimeControlState(
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            explore,
            AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
            get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
          ),
          defaultExplorePreset,
        );
        expect(url.toString()).to.eq(expectedUrl);

        // load empty url into metrics
        const defaultUrl = new URL("http://localhost");
        const errors = applyURLToExploreState(
          defaultUrl,
          explore,
          defaultExplorePreset,
        );
        expect(errors.length).toEqual(0);
        const currentState = getCleanMetricsExploreForAssertion();
        // current state should match the initial state
        expect(currentState).toEqual(initState);
      });
    }
  });

  describe("Should set correct state for legacy protobuf state and restore default state on empty params", () => {
    for (const { title, mutations, preset, legacyNotSupported } of TestCases) {
      if (legacyNotSupported) continue;
      it(title, () => {
        const explore: V1ExploreSpec = {
          ...AD_BIDS_EXPLORE_INIT,
          ...(preset ? { defaultPreset: preset } : {}),
        };
        metricsExplorerStore.init(
          AD_BIDS_EXPLORE_NAME,
          getInitExploreStateForTest(
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            explore,
            AD_BIDS_TIME_RANGE_SUMMARY,
          ),
        );
        const defaultExplorePreset = getDefaultExplorePreset(
          explore,
          AD_BIDS_TIME_RANGE_SUMMARY,
        );

        const initState = getCleanMetricsExploreForAssertion();
        applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);
        const curState =
          getCleanMetricsExploreForAssertion() as MetricsExplorerEntity;

        const url = new URL("http://localhost");
        // load url with legacy protobuf state
        url.searchParams.set("state", getProtoFromDashboardState(curState));
        // get back the entity from url params
        const { partialExploreState: entityFromUrl } = convertURLToExploreState(
          url.searchParams,
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
          explore,
          defaultExplorePreset,
          [],
        );
        expect(entityFromUrl).toEqual(curState);

        // go back to default url
        const defaultUrl = new URL("http://localhost");
        const { partialExploreState: entityFromDefaultUrl } =
          convertURLToExploreState(
            defaultUrl.searchParams,
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            explore,
            defaultExplorePreset,
            [],
          );

        // assert that the entity we got back matches the original
        expect(entityFromDefaultUrl).toEqual(initState);
      });
    }
  });
});

export function applyURLToExploreState(
  url: URL,
  exploreSpec: V1ExploreSpec,
  defaultExplorePreset: V1ExplorePreset,
) {
  const { partialExploreState: partialExploreStateDefaultUrl, errors } =
    convertURLToExploreState(
      url.searchParams,
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
      exploreSpec,
      defaultExplorePreset,
      [],
    );
  metricsExplorerStore.mergePartialExplorerEntity(
    AD_BIDS_EXPLORE_NAME,
    partialExploreStateDefaultUrl,
    AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  );
  return errors;
}

// cleans the metrics explore of any state that is not stored or restored from url state
export function getCleanMetricsExploreForAssertion() {
  // clone the existing state so that any mutations do affect the copy during assertion
  const cleanedState = deepClone(
    get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
  ) as Partial<MetricsExplorerEntity>;
  // these are not cloned
  cleanedState.visibleMeasureKeys = new Set(cleanedState.visibleMeasureKeys);
  cleanedState.visibleDimensionKeys = new Set(
    cleanedState.visibleDimensionKeys,
  );

  delete cleanedState.name;
  delete cleanedState.proto;
  delete cleanedState.dimensionFilterExcludeMode;
  delete cleanedState.temporaryFilterName;
  delete cleanedState.contextColumnWidths;
  if (cleanedState.selectedTimeRange) {
    cleanedState.selectedTimeRange = {
      name: cleanedState.selectedTimeRange?.name ?? "inf",
      interval: cleanedState.selectedTimeRange?.interval,
    } as DashboardTimeControls;
  }
  delete cleanedState.lastDefinedScrubRange;

  // TODO
  delete cleanedState.selectedScrubRange;
  delete cleanedState.leaderboardContextColumn;

  return cleanedState;
}
