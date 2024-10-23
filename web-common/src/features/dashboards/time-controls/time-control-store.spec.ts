import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_NAME,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { initStateManagers } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import TimeControlsStoreTest from "./TimeControlsStoreTest.svelte";
import {
  type TimeControlState,
  type TimeControlStore,
  createTimeControlStore,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  getLocalUserPreferences,
  initLocalUserPreferenceStore,
} from "@rilldata/web-common/features/dashboards/user-preferences";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { waitUntil } from "@rilldata/utils";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { render } from "@testing-library/svelte";
import { get } from "svelte/store";
import { beforeAll, beforeEach, describe, expect, it } from "vitest";

describe("time-control-store", () => {
  runtime.set({
    host: "http://localhost",
    instanceId: "default",
  });
  const dashboardFetchMocks = DashboardFetchMocks.useDashboardFetchMocks();

  beforeAll(() => {
    initLocalUserPreferenceStore(AD_BIDS_EXPLORE_NAME);
  });

  beforeEach(() => {
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
    getLocalUserPreferences().updateTimeZone("UTC");
  });

  it("Switching from no timestamp column to having one", async () => {
    const { unmount, queryClient, timeControlsStore } =
      initTimeControlStoreTest(AD_BIDS_METRICS_INIT);
    await waitUntil(() => !get(timeControlsStore).isFetching);

    const state = get(timeControlsStore);
    expect(state.isFetching).toBeFalsy();
    expect(state.ready).toBeTruthy();
    assertStartAndEnd(state, undefined, undefined, undefined, undefined);

    dashboardFetchMocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT_WITH_TIME,
      AD_BIDS_EXPLORE_INIT,
    );
    dashboardFetchMocks.mockTimeRangeSummary(AD_BIDS_NAME, {
      min: "2022-01-01",
      max: "2022-03-31",
    });
    await queryClient.refetchQueries({
      type: "active",
    });

    await waitForUpdate(timeControlsStore, "2022-01-01T00:00:00.000Z");
    assertStartAndEnd(
      get(timeControlsStore),
      "2022-01-01T00:00:00.000Z",
      "2022-03-31T00:00:00.001Z",
      "2021-12-31T00:00:00.000Z",
      "2022-04-01T00:00:00.000Z",
    );

    dashboardFetchMocks.mockTimeRangeSummary(AD_BIDS_NAME, {
      min: "2023-01-01",
      max: "2023-03-31",
    });
    await queryClient.refetchQueries({
      type: "active",
    });

    await waitForUpdate(timeControlsStore, "2023-01-01T00:00:00.000Z");
    // Updating time range updates the start and end
    assertStartAndEnd(
      get(timeControlsStore),
      "2023-01-01T00:00:00.000Z",
      "2023-03-31T00:00:00.001Z",
      "2022-12-31T00:00:00.000Z",
      "2023-04-01T00:00:00.000Z",
    );

    unmount();
  });

  it("Switching selected time range", async () => {
    dashboardFetchMocks.mockTimeRangeSummary(AD_BIDS_NAME, {
      min: "2022-01-01",
      max: "2022-03-31",
    });
    const { unmount, timeControlsStore } = initTimeControlStoreTest(
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    await waitForUpdate(timeControlsStore, "2022-01-01T00:00:00.000Z");

    metricsExplorerStore.setSelectedTimeRange(AD_BIDS_EXPLORE_NAME, {
      name: TimeRangePreset.LAST_24_HOURS,
      start: undefined,
      end: undefined,
      interval: V1TimeGrain.TIME_GRAIN_HOUR,
    });
    assertStartAndEnd(
      get(timeControlsStore),
      "2022-03-30T01:00:00.000Z",
      "2022-03-31T01:00:00.000Z",
      "2022-03-30T00:00:00.000Z",
      "2022-03-31T02:00:00.000Z",
    );

    metricsExplorerStore.setSelectedTimeRange(AD_BIDS_EXPLORE_NAME, {
      name: TimeRangePreset.CUSTOM,
      start: new Date("2022-03-20T01:00:00.000Z"),
      end: new Date("2022-03-22T01:00:00.000Z"),
      interval: V1TimeGrain.TIME_GRAIN_MONTH,
    });
    let state = get(timeControlsStore);
    assertStartAndEnd(
      state,
      "2022-03-20T01:00:00.000Z",
      "2022-03-22T01:00:00.000Z",
      "2022-03-20T00:00:00.000Z",
      "2022-03-22T02:00:00.000Z",
    );
    // invalid time grain of month is reset to hour
    expect(state.selectedTimeRange!.interval).toEqual(
      V1TimeGrain.TIME_GRAIN_HOUR,
    );

    metricsExplorerStore.setSelectedTimeRange(AD_BIDS_EXPLORE_NAME, {
      name: TimeRangePreset.LAST_7_DAYS,
      start: new Date("2021-01-01"),
      end: new Date("2021-03-31"),
      interval: V1TimeGrain.TIME_GRAIN_HOUR,
    });
    state = get(timeControlsStore);
    // start and end from selected time range is ignored.
    assertStartAndEnd(
      state,
      "2022-03-25T00:00:00.000Z",
      "2022-04-01T00:00:00.000Z",
      "2022-03-24T23:00:00.000Z",
      "2022-04-01T01:00:00.000Z",
    );
    // valid time grain of hour is retained
    expect(state.selectedTimeRange!.interval).toEqual(
      V1TimeGrain.TIME_GRAIN_HOUR,
    );

    unmount();
  });

  it("Switching selected comparison time range", async () => {
    dashboardFetchMocks.mockTimeRangeSummary(AD_BIDS_NAME, {
      min: "2022-01-01",
      max: "2022-03-31",
    });
    const { unmount, timeControlsStore } = initTimeControlStoreTest(
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    await waitForUpdate(timeControlsStore, "2022-01-01T00:00:00.000Z");

    metricsExplorerStore.displayTimeComparison(AD_BIDS_EXPLORE_NAME, true);
    metricsExplorerStore.setSelectedTimeRange(AD_BIDS_EXPLORE_NAME, {
      name: TimeRangePreset.LAST_24_HOURS,
      start: undefined,
      end: undefined,
      interval: V1TimeGrain.TIME_GRAIN_HOUR,
    });
    metricsExplorerStore.setSelectedComparisonRange(
      AD_BIDS_EXPLORE_NAME,
      {} as any,
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    assertComparisonStartAndEnd(
      get(timeControlsStore),
      // Sets to default comparison
      "rill-PD",
      "2022-03-29T01:00:00.000Z",
      "2022-03-30T01:00:00.000Z",
      "2022-03-29T00:00:00.000Z",
      "2022-03-30T02:00:00.000Z",
    );

    metricsExplorerStore.setSelectedTimeRange(AD_BIDS_EXPLORE_NAME, {
      name: TimeRangePreset.LAST_12_MONTHS,
      start: undefined,
      end: undefined,
      interval: V1TimeGrain.TIME_GRAIN_DAY,
    });
    metricsExplorerStore.setSelectedComparisonRange(
      AD_BIDS_EXPLORE_NAME,
      {} as any,
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );

    metricsExplorerStore.setSelectedTimeRange(AD_BIDS_EXPLORE_NAME, {
      name: TimeRangePreset.LAST_7_DAYS,
      start: undefined,
      end: undefined,
      interval: V1TimeGrain.TIME_GRAIN_DAY,
    });
    metricsExplorerStore.setSelectedComparisonRange(
      AD_BIDS_EXPLORE_NAME,
      {} as any,
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    assertComparisonStartAndEnd(
      get(timeControlsStore),
      // Sets to the one selected
      "rill-PW",
      "2022-03-18T00:00:00.000Z",
      "2022-03-25T00:00:00.000Z",
      "2022-03-17T00:00:00.000Z",
      "2022-03-26T00:00:00.000Z",
    );

    unmount();
  });

  it("Switching time zones", async () => {
    dashboardFetchMocks.mockTimeRangeSummary(AD_BIDS_NAME, {
      min: "2022-01-01",
      max: "2022-03-31",
    });
    const { unmount, timeControlsStore } = initTimeControlStoreTest(
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    await waitForUpdate(timeControlsStore, "2022-01-01T00:00:00.000Z");

    metricsExplorerStore.setTimeZone(AD_BIDS_EXPLORE_NAME, "IST");
    assertStartAndEnd(
      get(timeControlsStore),
      "2022-01-01T00:00:00.000Z",
      "2022-03-31T00:00:00.001Z",
      "2021-12-30T18:30:00.000Z",
      "2022-03-31T18:30:00.000Z",
    );

    metricsExplorerStore.displayTimeComparison(AD_BIDS_EXPLORE_NAME, true);
    metricsExplorerStore.setSelectedTimeRange(AD_BIDS_EXPLORE_NAME, {
      name: TimeRangePreset.LAST_24_HOURS,
      start: undefined,
      end: undefined,
      interval: V1TimeGrain.TIME_GRAIN_HOUR,
    });
    metricsExplorerStore.setSelectedComparisonRange(
      AD_BIDS_EXPLORE_NAME,
      {} as any,
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    assertStartAndEnd(
      get(timeControlsStore),
      "2022-03-30T00:30:00.000Z",
      "2022-03-31T00:30:00.000Z",
      "2022-03-29T23:30:00.000Z",
      "2022-03-31T01:30:00.000Z",
    );
    assertComparisonStartAndEnd(
      get(timeControlsStore),
      // Sets to default comparison
      "rill-PD",
      "2022-03-29T00:30:00.000Z",
      "2022-03-30T00:30:00.000Z",
      "2022-03-28T23:30:00.000Z",
      "2022-03-30T01:30:00.000Z",
    );

    unmount();
  });

  it("Scrubbing to zoom", async () => {
    dashboardFetchMocks.mockTimeRangeSummary(AD_BIDS_NAME, {
      min: "2022-01-01",
      max: "2022-03-31",
    });
    const { unmount, timeControlsStore } = initTimeControlStoreTest(
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    await waitForUpdate(timeControlsStore, "2022-01-01T00:00:00.000Z");
    metricsExplorerStore.displayTimeComparison(AD_BIDS_EXPLORE_NAME, true);
    metricsExplorerStore.setSelectedComparisonRange(
      AD_BIDS_EXPLORE_NAME,
      {
        name: TimeComparisonOption.MONTH,
      } as any,
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );

    metricsExplorerStore.setSelectedScrubRange(AD_BIDS_EXPLORE_NAME, {
      start: new Date("2022-02-01 UTC"),
      end: new Date("2022-02-10 UTC"),
      isScrubbing: true,
    });
    assertStartAndEnd(
      get(timeControlsStore),
      "2022-01-01T00:00:00.000Z",
      "2022-03-31T00:00:00.001Z",
      "2021-12-31T00:00:00.000Z",
      "2022-04-01T00:00:00.000Z",
    );
    assertComparisonStartAndEnd(
      get(timeControlsStore),
      // Sets to default comparison
      "rill-PM",
      "2021-12-01T00:00:00.000Z",
      "2022-02-28T00:00:00.001Z",
      "2021-11-30T00:00:00.000Z",
      "2022-03-01T00:00:00.000Z",
    );

    metricsExplorerStore.setSelectedScrubRange(AD_BIDS_EXPLORE_NAME, {
      start: new Date("2022-02-01 UTC"),
      end: new Date("2022-02-10 UTC"),
      isScrubbing: false,
    });
    assertStartAndEnd(
      get(timeControlsStore),
      "2022-02-01T00:00:00.000Z",
      "2022-02-10T00:00:00.000Z",
      "2021-12-31T00:00:00.000Z",
      "2022-04-01T00:00:00.000Z",
    );
    assertComparisonStartAndEnd(
      get(timeControlsStore),
      // Sets to default comparison
      "rill-PM",
      "2022-01-01T00:00:00.000Z",
      "2022-01-10T00:00:00.000Z",
      "2021-11-30T00:00:00.000Z",
      "2022-03-01T00:00:00.000Z",
    );

    unmount();
  });

  it("Default time ranges", async () => {
    dashboardFetchMocks.mockTimeRangeSummary(AD_BIDS_NAME, {
      min: "2022-01-01",
      max: "2022-03-31",
    });
    const { unmount, timeControlsStore, queryClient } =
      initTimeControlStoreTest(AD_BIDS_METRICS_INIT_WITH_TIME);
    await waitForUpdate(timeControlsStore, "2022-01-01T00:00:00.000Z");
    assertStartAndEnd(
      get(timeControlsStore),
      "2022-01-01T00:00:00.000Z",
      "2022-03-31T00:00:00.001Z",
      "2021-12-31T00:00:00.000Z",
      "2022-04-01T00:00:00.000Z",
    );

    console.log("refetch");
    dashboardFetchMocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT_WITH_TIME,
      {
        ...AD_BIDS_EXPLORE_INIT,
        defaultPreset: {
          timeRange: "P4W",
        },
      },
    );
    await queryClient.refetchQueries({
      type: "active",
    });
    await waitForDefaultUpdate(timeControlsStore, "2022-03-07T00:00:00.000Z");
    expect(get(timeControlsStore).defaultTimeRange).toEqual({
      name: TimeRangePreset.LAST_4_WEEKS,
      start: new Date("2022-03-07T00:00:00.000Z"),
      end: new Date("2022-04-04T00:00:00.000Z"),
    });

    dashboardFetchMocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT_WITH_TIME,
      {
        ...AD_BIDS_EXPLORE_INIT,
        defaultPreset: { timeRange: "P2W" },
      },
    );
    await queryClient.refetchQueries({
      type: "active",
    });
    await waitForDefaultUpdate(timeControlsStore, "2022-03-21T00:00:00.000Z");
    expect(get(timeControlsStore).defaultTimeRange).toEqual({
      name: "P2W",
      start: new Date("2022-03-21T00:00:00.000Z"),
      end: new Date("2022-04-04T00:00:00.000Z"),
    });

    unmount();
  });

  function initTimeControlStoreTest(metricsView: V1MetricsViewSpec) {
    dashboardFetchMocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      metricsView,
      AD_BIDS_EXPLORE_INIT,
    );
    const { stateManagers, queryClient } = initStateManagers();
    const timeControlsStore = createTimeControlStore(stateManagers);

    const { unmount } = render(TimeControlsStoreTest, {
      timeControlsStore,
    });

    return { unmount, queryClient, timeControlsStore };
  }
});

function assertStartAndEnd(
  timeControlsSate: TimeControlState,
  start: string | undefined,
  end: string | undefined,
  adjustedStart: string | undefined,
  adjustedEnd: string | undefined,
) {
  expect(timeControlsSate.timeStart).toEqual(start);
  expect(timeControlsSate.timeEnd).toEqual(end);
  expect(timeControlsSate.adjustedStart).toEqual(adjustedStart);
  expect(timeControlsSate.adjustedEnd).toEqual(adjustedEnd);
}

function assertComparisonStartAndEnd(
  timeControlsSate: TimeControlState,
  name: string,
  start: string,
  end: string,
  adjustedStart: string,
  adjustedEnd: string,
) {
  expect(timeControlsSate.selectedComparisonTimeRange?.name).toBe(name);
  expect(timeControlsSate.comparisonTimeStart).toBe(start);
  expect(timeControlsSate.comparisonTimeEnd).toBe(end);
  expect(timeControlsSate.comparisonAdjustedStart).toBe(adjustedStart);
  expect(timeControlsSate.comparisonAdjustedEnd).toBe(adjustedEnd);
}

async function waitForUpdate(
  timeControlsStore: TimeControlStore,
  startTime: string,
) {
  await waitUntil(
    () => get(timeControlsStore).timeStart === startTime,
    1000,
    20,
  );
  expect(get(timeControlsStore).timeStart).toBe(startTime);
}

async function waitForDefaultUpdate(
  timeControlsStore: TimeControlStore,
  startTime: string,
) {
  return waitUntil(
    () =>
      get(timeControlsStore).defaultTimeRange?.start?.toISOString() ===
      startTime,
    1000,
    20,
  );
}
