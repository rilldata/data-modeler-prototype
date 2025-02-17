import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getValidComparisonOption } from "@rilldata/web-common/features/dashboards/time-controls/time-range-store";
import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
import {
  getComparionRangeForScrub,
  getComparisonRange,
  getTimeComparisonParametersForComponent,
} from "@rilldata/web-common/lib/time/comparisons";
import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
import {
  checkValidTimeGrain,
  findValidTimeGrain,
  getAllowedTimeGrains,
  getDefaultTimeGrain,
} from "@rilldata/web-common/lib/time/grains";
import {
  convertTimeRangePreset,
  getAdjustedFetchTime,
} from "@rilldata/web-common/lib/time/ranges";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  type DashboardTimeControls,
  type ScrubRange,
  TimeComparisonOption,
  type TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  type V1ExploreSpec,
  type V1ExploreTimeRange,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  V1TimeGrain,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import type { Readable } from "svelte/store";
import { derived } from "svelte/store";
import { memoizeMetricsStore } from "../state-managers/memoize-metrics-store";

export type TimeRangeState = {
  // Selected ranges with start and end filled based on time range type
  selectedTimeRange?: DashboardTimeControls;
  // In all of our queries we do a check on hasTime and pass in undefined for start and end if false.
  // Using these directly will simplify those usages since this store will take care of marking them undefined.
  timeStart?: string;
  adjustedStart?: string;
  timeEnd?: string;
  adjustedEnd?: string;
};
export type ComparisonTimeRangeState = {
  showTimeComparison?: boolean;
  selectedComparisonTimeRange?: DashboardTimeControls;
  comparisonTimeStart?: string;
  comparisonAdjustedStart?: string;
  comparisonTimeEnd?: string;
  comparisonAdjustedEnd?: string;
};
export type TimeControlState = {
  isFetching: boolean;

  // Computed properties from all time range query
  minTimeGrain?: V1TimeGrain;
  allTimeRange?: TimeRange;
  defaultTimeRange?: TimeRange;
  timeDimension?: string;

  ready?: boolean;
} & TimeRangeState &
  ComparisonTimeRangeState;
export type TimeControlStore = Readable<TimeControlState>;

/**
 * Returns a TimeControlState. Calls getTimeControlState internally.
 *
 * Consumers of this will have a QueryObserverResult for time range summary.
 * They will need `isFetching` to wait for this response.
 */
export const timeControlStateSelector = ([
  metricsView,
  explore,
  timeRangeResponse,
  metricsExplorer,
]: [
  V1MetricsViewSpec | undefined,
  V1ExploreSpec | undefined,
  QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>,
  MetricsExplorerEntity,
]): TimeControlState => {
  const hasTimeSeries = Boolean(metricsView?.timeDimension);
  if (
    !metricsView ||
    !explore ||
    !metricsExplorer ||
    !timeRangeResponse ||
    !timeRangeResponse.isSuccess ||
    !hasTimeSeries
  ) {
    return {
      isFetching: timeRangeResponse.isRefetching,
      ready: !metricsExplorer || !hasTimeSeries,
    } as TimeControlState;
  }

  const state = getTimeControlState(
    metricsView,
    explore,
    timeRangeResponse.data?.timeRangeSummary,
    metricsExplorer,
  );
  if (!state) {
    return {
      ready: false,
      isFetching: false,
    };
  }

  return {
    ...state,
    isFetching: false,
    ready: true,
  } as TimeControlState;
};

/**
 * Generates TimeControlState
 *
 * Consumers of this will already have a V1TimeRangeSummary.
 */
export function getTimeControlState(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
  exploreState: MetricsExplorerEntity,
) {
  const hasTimeSeries = Boolean(metricsViewSpec.timeDimension);
  const timeDimension = metricsViewSpec.timeDimension;
  if (!hasTimeSeries || !timeRangeSummary?.max || !timeRangeSummary?.min)
    return undefined;

  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(timeRangeSummary.min),
    end: new Date(timeRangeSummary.max),
  };
  const minTimeGrain =
    (metricsViewSpec.smallestTimeGrain as V1TimeGrain) ||
    V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  const defaultTimeRange = isoDurationToFullTimeRange(
    exploreSpec.defaultPreset?.timeRange,
    allTimeRange.start,
    allTimeRange.end,
    exploreState.selectedTimezone,
  );

  const {
    selectedTimeRange,
    selectedComparisonTimeRange,
    selectedTimezone,
    lastDefinedScrubRange,
    showTimeComparison,
  } = exploreState;

  const timeRangeState = calculateTimeRangePartial(
    allTimeRange,
    selectedTimeRange,
    lastDefinedScrubRange,
    selectedTimezone,
    defaultTimeRange,
    minTimeGrain,
  );
  if (!timeRangeState) {
    return undefined;
  }

  const comparisonTimeRangeState = calculateComparisonTimeRangePartial(
    exploreSpec.timeRanges,
    allTimeRange,
    selectedComparisonTimeRange,
    selectedTimezone,
    lastDefinedScrubRange,
    showTimeComparison,
    timeRangeState,
  );

  return {
    minTimeGrain,
    allTimeRange,
    defaultTimeRange,
    timeDimension,

    ...timeRangeState,

    ...comparisonTimeRangeState,
  } as TimeControlState;
}

export function createTimeControlStore(ctx: StateManagers) {
  return derived(
    [ctx.validSpecStore, ctx.timeRangeSummaryStore, ctx.dashboardStore],
    ([validSpecResp, timeRangeSummaryResp, dashboardStore]) =>
      timeControlStateSelector([
        validSpecResp.data?.metricsView,
        validSpecResp.data?.explore,
        timeRangeSummaryResp,
        dashboardStore,
      ]),
  );
}

/**
 * Memoized version of the store. Currently, memoized by metrics view name.
 */
export const useTimeControlStore = memoizeMetricsStore<TimeControlStore>(
  (ctx: StateManagers) => createTimeControlStore(ctx),
);

/**
 * Calculates time range and grain from all time range and selected time range name.
 * Also adds start, end and their adjusted counterparts as strings ready to use in requests.
 */
export function calculateTimeRangePartial(
  allTimeRange: DashboardTimeControls,
  currentSelectedTimeRange: DashboardTimeControls | undefined,
  lastDefinedScrubRange: ScrubRange | undefined,
  selectedTimezone: string,
  defaultTimeRange: DashboardTimeControls,
  minTimeGrain: V1TimeGrain,
): TimeRangeState | undefined {
  if (!currentSelectedTimeRange) return undefined;

  const selectedTimeRange = getTimeRange(
    currentSelectedTimeRange,
    selectedTimezone,
    allTimeRange,
    defaultTimeRange,
  );
  if (!selectedTimeRange) return undefined;

  selectedTimeRange.interval = getTimeGrain(
    currentSelectedTimeRange,
    selectedTimeRange,
    minTimeGrain,
  );
  const { start: adjustedStart, end: adjustedEnd } = getAdjustedFetchTime(
    selectedTimeRange.start,
    selectedTimeRange.end,
    selectedTimezone,
    selectedTimeRange.interval,
  );

  let timeStart = selectedTimeRange.start;
  let timeEnd = selectedTimeRange.end;
  if (lastDefinedScrubRange) {
    const { start, end } = getOrderedStartEnd(
      lastDefinedScrubRange.start,
      lastDefinedScrubRange.end,
    );
    timeStart = start;
    timeEnd = end;
  }

  return {
    selectedTimeRange,
    timeStart: timeStart.toISOString(),
    adjustedStart,
    timeEnd: timeEnd.toISOString(),
    adjustedEnd,
  };
}

/**
 * Calculates time range and grain for comparison based on time range and comparison selection.
 * Also adds start, end and their adjusted counterparts as strings ready to use in requests.
 */
export function calculateComparisonTimeRangePartial(
  timeRanges: V1ExploreTimeRange[] | undefined,
  allTimeRange: DashboardTimeControls,
  currentComparisonTimeRange: DashboardTimeControls | undefined,
  selectedTimezone: string,
  lastDefinedScrubRange: ScrubRange | undefined,
  showTimeComparison: boolean,
  timeRangeState: TimeRangeState,
): ComparisonTimeRangeState {
  const selectedComparisonTimeRange = getComparisonTimeRange(
    timeRanges,
    allTimeRange,
    timeRangeState.selectedTimeRange,
    currentComparisonTimeRange,
  );

  let comparisonAdjustedStart: string | undefined = undefined;
  let comparisonAdjustedEnd: string | undefined = undefined;
  if (selectedComparisonTimeRange) {
    const adjustedComparisonTime = getAdjustedFetchTime(
      selectedComparisonTimeRange.start,
      selectedComparisonTimeRange.end,
      selectedTimezone,
      timeRangeState.selectedTimeRange?.interval,
    );
    comparisonAdjustedStart = adjustedComparisonTime.start;
    comparisonAdjustedEnd = adjustedComparisonTime.end;
  }

  let comparisonTimeStart = selectedComparisonTimeRange?.start;
  let comparisonTimeEnd = selectedComparisonTimeRange?.end;
  if (selectedComparisonTimeRange && lastDefinedScrubRange) {
    const { start, end } = getOrderedStartEnd(
      lastDefinedScrubRange.start,
      lastDefinedScrubRange.end,
    );

    if (!timeRangeState.selectedTimeRange?.start) {
      throw new Error("No time range");
    }

    const comparisonRange = getComparionRangeForScrub(
      timeRangeState.selectedTimeRange?.start,
      timeRangeState.selectedTimeRange?.end,
      selectedComparisonTimeRange.start,
      selectedComparisonTimeRange.end,
      start,
      end,
    );
    comparisonTimeStart = comparisonRange.start;
    comparisonTimeEnd = comparisonRange.end;
  }

  return {
    showTimeComparison: showTimeComparison,
    selectedComparisonTimeRange,
    comparisonTimeStart: comparisonTimeStart?.toISOString(),
    comparisonAdjustedStart,
    comparisonTimeEnd: comparisonTimeEnd?.toISOString(),
    comparisonAdjustedEnd,
  };
}

export function getTimeRange(
  selectedTimeRange: DashboardTimeControls | undefined,
  selectedTimezone: string,
  allTimeRange: DashboardTimeControls,
  defaultTimeRange: DashboardTimeControls,
) {
  if (!selectedTimeRange) return undefined;

  let timeRange: DashboardTimeControls;

  if (selectedTimeRange?.name === TimeRangePreset.CUSTOM) {
    /** set the time range to the fixed custom time range */
    timeRange = {
      name: TimeRangePreset.CUSTOM,
      start: new Date(selectedTimeRange.start),
      end: new Date(selectedTimeRange.end),
    };
  } else if (selectedTimeRange?.name) {
    if (selectedTimeRange?.name in DEFAULT_TIME_RANGES) {
      /** rebuild off of relative time range */
      timeRange = convertTimeRangePreset(
        selectedTimeRange?.name ?? TimeRangePreset.ALL_TIME,
        allTimeRange.start,
        allTimeRange.end,
        selectedTimezone,
      );
    } else {
      timeRange = isoDurationToFullTimeRange(
        selectedTimeRange?.name,
        allTimeRange.start,
        allTimeRange.end,
        selectedTimezone,
      );
    }
  } else {
    /** set the time range to the fixed custom time range */
    timeRange = {
      name: defaultTimeRange.name,
      start: defaultTimeRange.start,
      end: defaultTimeRange.end,
    };
  }

  return timeRange;
}

export function getTimeGrain(
  selectedTimeRange: DashboardTimeControls | undefined,
  timeRange: DashboardTimeControls,
  minTimeGrain: V1TimeGrain,
) {
  const timeGrainOptions = getAllowedTimeGrains(timeRange.start, timeRange.end);
  const isValidTimeGrain = checkValidTimeGrain(
    selectedTimeRange?.interval,
    timeGrainOptions,
    minTimeGrain,
  );

  let timeGrain: V1TimeGrain | undefined;
  if (isValidTimeGrain) {
    timeGrain = selectedTimeRange?.interval;
  } else {
    const defaultTimeGrain = getDefaultTimeGrain(
      timeRange.start,
      timeRange.end,
    ).grain;
    timeGrain = findValidTimeGrain(
      defaultTimeGrain,
      timeGrainOptions,
      minTimeGrain,
    );
  }

  return timeGrain;
}

export function getComparisonTimeRange(
  timeRanges: V1ExploreTimeRange[] | undefined,
  allTimeRange: TimeRange | undefined,
  timeRange: DashboardTimeControls | undefined,
  comparisonTimeRange: DashboardTimeControls | undefined,
) {
  if (!timeRange || !timeRange.name || !allTimeRange) return undefined;

  if (!comparisonTimeRange?.name) {
    const comparisonOption = getValidComparisonOption(
      timeRanges,
      timeRange,
      undefined,
      allTimeRange,
    );
    const range = getTimeComparisonParametersForComponent(
      comparisonOption,
      allTimeRange.start,
      allTimeRange.end,
      timeRange.start,
      timeRange.end,
    );

    if (range.start && range.end) {
      return {
        start: range.start,
        end: range.end,
        name: comparisonOption,
      };
    }
  } else if (
    comparisonTimeRange.name === TimeComparisonOption.CUSTOM ||
    // 1st step towards using a single `Custom` variable
    // TODO: replace the usage of TimeComparisonOption.CUSTOM with TimeRangePreset.CUSTOM
    comparisonTimeRange.name === TimeRangePreset.CUSTOM
  ) {
    return comparisonTimeRange;
  } else {
    // variable time range of some kind.
    const comparisonOption = comparisonTimeRange.name as TimeComparisonOption;
    const range = getComparisonRange(
      timeRange.start,
      timeRange.end,
      comparisonOption,
    );

    return {
      ...range,
      name: comparisonOption,
    };
  }
}

/**
 * Fills in start and end dates based on selected time range and all time range.
 */
export function selectedTimeRangeSelector([
  exploreSpec,
  timeRangeResponse,
  explorer,
]: [
  V1ExploreSpec | undefined,
  QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>,
  MetricsExplorerEntity,
]) {
  if (
    !exploreSpec ||
    !timeRangeResponse.data?.timeRangeSummary ||
    !timeRangeResponse.data.timeRangeSummary.min ||
    !timeRangeResponse.data.timeRangeSummary.max
  ) {
    return undefined;
  }

  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(timeRangeResponse.data.timeRangeSummary.min),
    end: new Date(timeRangeResponse.data.timeRangeSummary.max),
  };
  const defaultTimeRange = isoDurationToFullTimeRange(
    exploreSpec?.defaultPreset?.timeRange,
    allTimeRange.start,
    allTimeRange.end,
    explorer.selectedTimezone,
  );

  return getTimeRange(
    explorer.selectedTimeRange,
    explorer.selectedTimezone,
    allTimeRange,
    defaultTimeRange,
  );
}
