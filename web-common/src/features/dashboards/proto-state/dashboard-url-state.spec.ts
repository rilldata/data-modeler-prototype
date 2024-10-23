import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import {
  useDashboardDefaultProto,
  useDashboardUrlSync,
} from "@rilldata/web-common/features/dashboards/proto-state/dashboard-url-state";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DEFAULT_URL_TIME_RANGE,
  AD_BIDS_EXCLUDE_FILTER,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_SCHEMA,
  AD_BIDS_TIMESTAMP_DIMENSION,
  AD_BIDS_METRICS_WITH_DELETED_MEASURE,
  TestTimeConstants,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_EXPLORE_WITH_DELETED_MEASURE,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import DashboardTestComponent from "@rilldata/web-common/features/dashboards/stores/DashboardTestComponent.svelte";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  assertMetricsView,
  createDashboardState,
  initStateManagers,
  resetDashboardStore,
} from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { waitUntil } from "@rilldata/utils";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { Page } from "@sveltejs/kit";
import { render } from "@testing-library/svelte";
import { get, type Readable, writable } from "svelte/store";
import {
  beforeAll,
  beforeEach,
  describe,
  expect,
  it,
  type MockInstance,
  vi,
} from "vitest";

const pageMock: PageMock = vi.hoisted(() => ({}) as any);

vi.mock("$app/navigation", () => {
  return {
    goto: (url) => pageMock.goto(url),
  };
});
vi.mock("$app/stores", () => {
  return {
    page: pageMock,
  };
});

describe("useDashboardUrlSync", () => {
  runtime.set({
    host: "http://localhost",
    instanceId: "default",
  });
  const dashboardFetchMocks = DashboardFetchMocks.useDashboardFetchMocks();

  beforeAll(() => {
    createPageMock();
    initLocalUserPreferenceStore(AD_BIDS_EXPLORE_NAME);
  });

  beforeEach(() => {
    resetDashboardStore();
    pageMock.goto(`/explore/${AD_BIDS_EXPLORE_NAME}`);
    pageMock.gotoSpy.mockClear();
  });

  it("Changes to dashboard through interactions", async () => {
    const { teardown, defaultProtoStore, stateManagers } =
      await initDashboardUrlState();
    const {
      actions: {
        dimensionsFilter: { toggleDimensionValueSelection },
      },
    } = stateManagers;

    expect(pageMock.gotoSpy).toBeCalledTimes(0);

    toggleDimensionValueSelection(AD_BIDS_PUBLISHER_DIMENSION, "Google");
    await wait();
    assertUrlState(
      get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].proto!,
    );
    const protoWithFilter =
      get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].proto!;
    expect(pageMock.gotoSpy).toBeCalledTimes(1);

    pageMock.goto("/dashboard/AdBids");
    expect(
      get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].proto,
    ).toEqual(get(defaultProtoStore).proto);
    expect(
      get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].whereFilter,
    ).toEqual(createAndExpression([]));
    expect(pageMock.gotoSpy).toBeCalledTimes(2);

    pageMock.updateState(protoWithFilter);
    await wait();
    assertUrlState(
      get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].proto!,
    );
    expect(
      get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].whereFilter,
    ).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google"]),
      ]),
    );
    expect(pageMock.gotoSpy).toBeCalledTimes(2);

    teardown();
  });

  it("Changes to dashboard config", async () => {
    const { teardown, queryClient } = await initDashboardUrlState();
    expect(pageMock.gotoSpy).toBeCalledTimes(0);

    dashboardFetchMocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      {
        ...AD_BIDS_METRICS_WITH_DELETED_MEASURE,
        timeDimension: AD_BIDS_TIMESTAMP_DIMENSION,
      },
      AD_BIDS_EXPLORE_WITH_DELETED_MEASURE,
    );
    await queryClient.refetchQueries({
      type: "active",
    });
    await wait();
    // Goto not called still since defaultProto has changed
    expect(pageMock.gotoSpy).toBeCalledTimes(0);
    expect(get(pageMock).url.searchParams.has("state")).toBeFalsy();
    // This is not updated since the sync is called in a component
    // TODO: We should add tests for the sync component
    expect([
      ...get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME]
        .visibleMeasureKeys,
    ]).toEqual([AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE]);

    teardown();
  });

  it("Init load from url", async () => {
    gotoDashboardState(
      createDashboardState(
        AD_BIDS_EXPLORE_NAME,
        AD_BIDS_METRICS_INIT,
        AD_BIDS_EXPLORE_INIT,
        AD_BIDS_EXCLUDE_FILTER,
      ),
    );
    const { teardown } = await initDashboardUrlState();

    assertUrlState(
      get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].proto!,
    );
    assertMetricsView(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_EXCLUDE_FILTER,
      AD_BIDS_DEFAULT_URL_TIME_RANGE as DashboardTimeControls,
    );

    teardown();
  });

  async function initDashboardUrlState() {
    dashboardFetchMocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT_WITH_TIME,
      AD_BIDS_EXPLORE_INIT,
    );
    const { queryClient, stateManagers } = initStateManagers();
    dashboardFetchMocks.mockTimeRangeSummary(AD_BIDS_NAME, {
      min: TestTimeConstants.LAST_DAY.toISOString(),
      max: TestTimeConstants.NOW.toISOString(),
    });

    const { unmount } = render(DashboardTestComponent, {
      ctx: stateManagers,
    });

    const defaultProtoStore = useDashboardDefaultProto(stateManagers);
    await waitUntil(() => !get(defaultProtoStore).isFetching, 1000, 5);

    const unsubscribe = useDashboardUrlSync(stateManagers, AD_BIDS_SCHEMA);
    await wait();

    return {
      teardown: () => {
        unmount();
        unsubscribe();
      },
      stateManagers,
      queryClient,
      defaultProtoStore,
    };
  }
});

type PageMock = Readable<Page> & {
  updateState: (state: string) => void;
  goto: (path: string) => void;
  gotoSpy: MockInstance;
};
function createPageMock() {
  const { update, subscribe } = writable<Page>({
    url: new URL("http://localhost/explore/AdBids_explore"),
    params: { name: "AdBids_explore" },
  } as any);

  pageMock.subscribe = subscribe;
  pageMock.updateState = (state: string) => {
    update((page) => {
      if (state) {
        page.url = new URL(
          `http://localhost/explore/AdBids_explore?state=${encodeURIComponent(
            state,
          )}`,
        );
      } else {
        page.url = new URL("http://localhost/explore/AdBids_explore");
      }
      return page;
    });
  };
  pageMock.goto = (path: string) => {
    update((page) => {
      page.url = new URL(`http://localhost${path}`);
      return page;
    });
  };
  pageMock.gotoSpy = vi.spyOn(pageMock, "goto");
}

function assertUrlState(expected: string) {
  const actual = decodeURIComponent(
    get(pageMock).url.searchParams.get("state")!,
  );
  expect(actual).toEqual(expected);
}

function wait() {
  return new Promise((resolve) => setTimeout(resolve, 10));
}

function gotoDashboardState(state: MetricsExplorerEntity) {
  pageMock.goto(
    `/dashboard/${state.name}?state=${encodeURIComponent(
      getProtoFromDashboardState(state),
    )}`,
  );
}
