import { QueryClient } from "@rilldata/svelte-query";
import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { createStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_NAME,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import {
  AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
  AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
  AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
  AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
  AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_SORT_ASC_BY_BID_PRICE,
  AD_BIDS_SORT_BY_PERCENT_VALUE,
  AD_BIDS_SORT_DESC_BY_IMPRESSIONS,
  AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
  AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD,
  AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
  AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
  applyMutationsToDashboard,
  type TestDashboardMutation,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import {
  clearExploreSessionStore,
  getExplorePresetForWebView,
} from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-store";
import ExploreStateTestComponent from "@rilldata/web-common/features/dashboards/url-state/ExploreStateTestComponent.svelte";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import {
  applyURLToExploreState,
  getCleanMetricsExploreForAssertion,
} from "@rilldata/web-common/features/dashboards/url-state/url-state-variations.spec";
import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import {
  type V1ExplorePreset,
  V1ExploreWebView,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { Page } from "@sveltejs/kit";
import { render } from "@testing-library/svelte";
import { get, type Readable, readable } from "svelte/store";
import {
  afterAll,
  beforeAll,
  beforeEach,
  describe,
  expect,
  it,
  vi,
} from "vitest";

const pageMock: Readable<Page> = vi.hoisted(() => ({}) as any);

vi.mock("$app/stores", () => {
  return {
    page: pageMock,
  };
});
vi.stubEnv("TZ", "UTC");

type TestView = {
  view: V1ExploreWebView;
  additionalParams?: string;
  mutations: TestDashboardMutation[];
  expectedUrl: string;
};
const TestCases: {
  title: string;
  initView: TestView;
  view: TestView;
}[] = [
  {
    title: "Explore <=> tdd",
    initView: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
      mutations: [],
      expectedUrl:
        "http://localhost/explore/AdBids_explore?tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&sort_by=bid_price&sort_type=percent&sort_dir=ASC",
    },
    view: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
      additionalParams: "&measure=" + AD_BIDS_IMPRESSIONS_MEASURE,
      mutations: [AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
      expectedUrl:
        "http://localhost/explore/AdBids_explore?view=tdd&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measure=impressions&chart_type=stacked_bar",
    },
  },
  {
    title: "dimension table <=> tdd",
    initView: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
      mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
      expectedUrl:
        "http://localhost/explore/AdBids_explore?tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&expand_dim=publisher&sort_by=bid_price&sort_type=percent&sort_dir=ASC",
    },
    view: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
      additionalParams: "&measure=" + AD_BIDS_IMPRESSIONS_MEASURE,
      mutations: [AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
      expectedUrl:
        "http://localhost/explore/AdBids_explore?view=tdd&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measure=impressions&chart_type=stacked_bar",
    },
  },

  {
    title: "Explore <=> Pivot",
    initView: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
      mutations: [],
      expectedUrl:
        "http://localhost/explore/AdBids_explore?tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&sort_by=bid_price&sort_type=percent&sort_dir=ASC",
    },
    view: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
      mutations: [
        AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
        AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
      ],
      expectedUrl:
        "http://localhost/explore/AdBids_explore?view=pivot&tr=P7D&compare_tr=rill-PP&f=publisher+IN+%28%27Google%27%29&rows=publisher%2Ctime.hour&cols=domain%2Ctime.day%2Cimpressions&sort_by=time.day&sort_dir=ASC",
    },
  },
  {
    title: "dimension table <=> Pivot",
    initView: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
      mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
      expectedUrl:
        "http://localhost/explore/AdBids_explore?tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&expand_dim=publisher&sort_by=bid_price&sort_type=percent&sort_dir=ASC",
    },
    view: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
      mutations: [
        AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
        AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
      ],
      expectedUrl:
        "http://localhost/explore/AdBids_explore?view=pivot&tr=P7D&compare_tr=rill-PP&f=publisher+IN+%28%27Google%27%29&rows=publisher%2Ctime.hour&cols=domain%2Ctime.day%2Cimpressions&sort_by=time.day&sort_dir=ASC",
    },
  },
  {
    title: "tdd <=> Pivot",
    initView: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
      additionalParams: "&measure=" + AD_BIDS_IMPRESSIONS_MEASURE,
      mutations: [AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
      expectedUrl:
        "http://localhost/explore/AdBids_explore?view=tdd&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measure=impressions&chart_type=stacked_bar",
    },
    view: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
      mutations: [
        AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
        AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
      ],
      expectedUrl:
        "http://localhost/explore/AdBids_explore?view=pivot&tr=P7D&compare_tr=rill-PP&f=publisher+IN+%28%27Google%27%29&rows=publisher%2Ctime.hour&cols=domain%2Ctime.day%2Cimpressions&sort_by=time.day&sort_dir=ASC",
    },
  },
];

// TODO: add tests by wrapping DashboardURLStateSync.svelte
describe.skip("ExploreWebViewStore", () => {
  runtime.set({
    host: "http://localhost",
    instanceId: "default",
  });
  const dashboardFetchMocks = DashboardFetchMocks.useDashboardFetchMocks();
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
  let unmountTestComponent: () => void;

  beforeAll(async () => {
    const { subscribe } = readable({
      url: new URL("http://localhost/explore/" + AD_BIDS_EXPLORE_NAME),
    });
    pageMock.subscribe = subscribe as any;
    initLocalUserPreferenceStore(AD_BIDS_EXPLORE_NAME);

    dashboardFetchMocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
      AD_BIDS_EXPLORE_INIT,
    );
    const validSpecQuery = useExploreValidSpec(
      "default",
      AD_BIDS_EXPLORE_NAME,
      {
        queryClient,
      },
    );
    const { unmount } = render(ExploreStateTestComponent, {
      validSpecQuery,
    });
    unmountTestComponent = unmount;
    await waitUntil(() => !get(validSpecQuery).isLoading);
  });

  afterAll(() => {
    unmountTestComponent?.();
  });

  beforeEach(() => {
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
    clearExploreSessionStore(AD_BIDS_NAME, undefined);
  });

  for (const { title, initView, view } of TestCases) {
    it(title, () => {
      metricsExplorerStore.init(
        AD_BIDS_EXPLORE_NAME,
        getInitExploreStateForTest(
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
          AD_BIDS_EXPLORE_INIT,
          AD_BIDS_TIME_RANGE_SUMMARY,
        ),
      );
      createStateManagers({
        queryClient,
        metricsViewName: AD_BIDS_NAME,
        exploreName: AD_BIDS_EXPLORE_NAME,
      });
      const defaultExplorePreset = getDefaultExplorePreset(
        AD_BIDS_EXPLORE_INIT,
        AD_BIDS_TIME_RANGE_SUMMARY,
      );

      // apply mutations to main view to setup the initial state
      applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, [
        AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
        AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
        AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
        AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
        AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
        AD_BIDS_SORT_DESC_BY_IMPRESSIONS,
        AD_BIDS_SORT_BY_PERCENT_VALUE,
        AD_BIDS_SORT_ASC_BY_BID_PRICE,
      ]);

      // simulate going to the init view's url
      applyURLToExploreState(
        getUrlForWebView(
          initView.view,
          defaultExplorePreset,
          initView.additionalParams,
        ),
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
      );
      // apply any mutations in the init view
      applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, initView.mutations);
      const initState = getCleanMetricsExploreForAssertion();

      // simulate going to the view's url
      applyURLToExploreState(
        getUrlForWebView(
          view.view,
          defaultExplorePreset,
          view.additionalParams,
        ),
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
      );
      // apply any mutations in the view
      applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, view.mutations);

      const backToInitUrl = getUrlForWebView(
        initView.view,
        defaultExplorePreset,
      );
      expect(backToInitUrl.toString()).toEqual(initView.expectedUrl);
      applyURLToExploreState(
        backToInitUrl,
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
      );

      const backToViewUrl = getUrlForWebView(view.view, defaultExplorePreset);
      expect(backToViewUrl.toString()).toEqual(view.expectedUrl);
      const currentState = getCleanMetricsExploreForAssertion();
      expect(initState).toEqual(currentState);
    });
  }
});

function getUrlForWebView(
  view: V1ExploreWebView,
  defaultExplorePreset: V1ExplorePreset,
  additionalParams: string | undefined = undefined,
) {
  const newUrl = new URL("http://localhost/explore/AdBids_explore");

  const explorePresetFromSessionStorage = getExplorePresetForWebView(
    AD_BIDS_EXPLORE_NAME,
    undefined,
    view,
  );
  if (!explorePresetFromSessionStorage) {
    return newUrl;
  }

  const { partialExploreState } = convertPresetToExploreState(
    AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
    AD_BIDS_EXPLORE_INIT,
    explorePresetFromSessionStorage,
    [],
  );

  const exploreState = partialExploreState as MetricsExplorerEntity;
  newUrl.search =
    convertExploreStateToURLSearchParams(
      exploreState,
      AD_BIDS_EXPLORE_INIT,
      getTimeControlState(
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_INIT,
        AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
        exploreState,
      ),
      defaultExplorePreset,
    ) + (additionalParams ?? "");
  return newUrl;
}
