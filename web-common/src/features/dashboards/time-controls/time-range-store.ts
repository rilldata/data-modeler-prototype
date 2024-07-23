import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  getAvailableComparisonsForTimeRange,
  getComparisonRange,
  getTimeComparisonParametersForComponent,
} from "@rilldata/web-common/lib/time/comparisons";
import {
  PREVIOUS_COMPLETE_DATE_RANGES,
  type TimeRangeMetaSet,
} from "@rilldata/web-common/lib/time/config";
import {
  DEFAULT_TIME_RANGES,
  LATEST_WINDOW_TIME_RANGES,
  PERIOD_TO_DATE_RANGES,
} from "@rilldata/web-common/lib/time/config";
import { getChildTimeRanges } from "@rilldata/web-common/lib/time/ranges";
import { isoDurationToTimeRangeMeta } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import type {
  DashboardTimeControls,
  TimeRangeOption,
} from "@rilldata/web-common/lib/time/types";
import {
  TimeComparisonOption,
  TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  RpcStatus,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/svelte-query";

export type TimeRangeControlsState = {
  latestWindowTimeRanges: Array<TimeRangeOption>;
  periodToDateRanges: Array<TimeRangeOption>;
  previousCompleteDateRanges: Array<TimeRangeOption>;
  showDefaultItem: boolean;
};

export function timeRangeSelectionsSelector([
  metricsView,
  timeRangeResponse,
  explorer,
]: [
  QueryObserverResult<V1MetricsViewSpec, RpcStatus>,
  QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>,
  MetricsExplorerEntity,
]): TimeRangeControlsState {
  if (!metricsView.data || !timeRangeResponse?.data?.timeRangeSummary)
    return {
      latestWindowTimeRanges: [],
      periodToDateRanges: [],
      previousCompleteDateRanges: [],
      showDefaultItem: false,
    };

  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(timeRangeResponse.data.timeRangeSummary.min ?? 0),
    end: new Date(timeRangeResponse.data.timeRangeSummary.max ?? 0),
  };
  const minTimeGrain =
    (metricsView.data.smallestTimeGrain as V1TimeGrain) ||
    V1TimeGrain.TIME_GRAIN_UNSPECIFIED;

  let latestWindowTimeRanges: TimeRangeMetaSet = {};
  let periodToDateRanges: TimeRangeMetaSet = {};
  let previousCompleteDateRanges: TimeRangeMetaSet = {};
  let hasDefaultInRanges = false;

  if (metricsView.data.availableTimeRanges?.length) {
    for (const availableTimeRange of metricsView.data.availableTimeRanges) {
      if (!availableTimeRange.range) continue;

      // default time range is part of availableTimeRanges.
      // this is used to not show a separate selection for the default
      if (metricsView.data.defaultTimeRange === availableTimeRange.range) {
        hasDefaultInRanges = true;
      }
      if (availableTimeRange.range in LATEST_WINDOW_TIME_RANGES) {
        latestWindowTimeRanges[availableTimeRange.range] =
          LATEST_WINDOW_TIME_RANGES[availableTimeRange.range];
      } else if (availableTimeRange.range in PERIOD_TO_DATE_RANGES) {
        periodToDateRanges[availableTimeRange.range] =
          PERIOD_TO_DATE_RANGES[availableTimeRange.range];
      } else if (availableTimeRange.range in PREVIOUS_COMPLETE_DATE_RANGES) {
        previousCompleteDateRanges[availableTimeRange.range] =
          PREVIOUS_COMPLETE_DATE_RANGES[availableTimeRange.range];
      } else {
        latestWindowTimeRanges[availableTimeRange.range] =
          isoDurationToTimeRangeMeta(
            availableTimeRange.range,
            availableTimeRange.comparisonOffsets?.[0]
              ?.offset as TimeComparisonOption,
          );
      }
    }
  } else {
    latestWindowTimeRanges = LATEST_WINDOW_TIME_RANGES;
    periodToDateRanges = PERIOD_TO_DATE_RANGES;
    previousCompleteDateRanges = PREVIOUS_COMPLETE_DATE_RANGES;
    hasDefaultInRanges =
      !!metricsView.data.defaultTimeRange &&
      (metricsView.data.defaultTimeRange in LATEST_WINDOW_TIME_RANGES ||
        metricsView.data.defaultTimeRange in PERIOD_TO_DATE_RANGES ||
        metricsView.data.defaultTimeRange in PREVIOUS_COMPLETE_DATE_RANGES);
  }

  return {
    latestWindowTimeRanges: getChildTimeRanges(
      allTimeRange.start,
      allTimeRange.end,
      latestWindowTimeRanges,
      minTimeGrain,
      explorer.selectedTimezone,
    ),
    periodToDateRanges: getChildTimeRanges(
      allTimeRange.start,
      allTimeRange.end,
      periodToDateRanges,
      minTimeGrain,
      explorer.selectedTimezone,
    ),
    previousCompleteDateRanges: getChildTimeRanges(
      allTimeRange.start,
      allTimeRange.end,
      previousCompleteDateRanges,
      minTimeGrain,
      explorer.selectedTimezone,
    ),
    showDefaultItem: !!metricsView.data.defaultTimeRange && !hasDefaultInRanges,
  };
}

export function timeComparisonOptionsSelector([
  metricsView,
  timeRangeResponse,
  explorer,
  selectedTimeRange,
]: [
  QueryObserverResult<V1MetricsViewSpec, RpcStatus>,
  QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>,
  MetricsExplorerEntity,
  DashboardTimeControls | undefined,
]): Array<{
  name: TimeComparisonOption;
  key: number;
  start: Date;
  end: Date;
}> {
  if (
    !metricsView.data ||
    !timeRangeResponse?.data?.timeRangeSummary ||
    !explorer.selectedTimeRange ||
    !selectedTimeRange ||
    !timeRangeResponse.data.timeRangeSummary.min ||
    !timeRangeResponse.data.timeRangeSummary.max
  ) {
    return [];
  }

  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(timeRangeResponse.data.timeRangeSummary.min),
    end: new Date(timeRangeResponse.data.timeRangeSummary.max),
  };

  let allOptions = [...Object.values(TimeComparisonOption)];

  if (metricsView.data.availableTimeRanges?.length) {
    const timeRange = metricsView.data.availableTimeRanges.find(
      (tr) => tr.range === explorer.selectedTimeRange?.name,
    );
    if (timeRange?.comparisonOffsets?.length) {
      allOptions =
        timeRange.comparisonOffsets?.map(
          (co) => co.offset as TimeComparisonOption,
        ) ?? [];
      allOptions.push(TimeComparisonOption.CUSTOM);
    }
  } else if (
    explorer.selectedTimeRange?.name &&
    explorer.selectedTimeRange?.name in PREVIOUS_COMPLETE_DATE_RANGES
  ) {
    // Previous complete ranges should only have previous period.
    // Other options dont make sense with our current wording of the comparison ranges.
    allOptions = [TimeComparisonOption.CONTIGUOUS, TimeComparisonOption.CUSTOM];
  }

  const timeComparisonOptions = getAvailableComparisonsForTimeRange(
    allTimeRange.start,
    allTimeRange.end,
    selectedTimeRange.start,
    selectedTimeRange.end,
    allOptions,
  );

  return timeComparisonOptions.map((co, i) => {
    const comparisonTimeRange = getComparisonRange(
      selectedTimeRange.start,
      selectedTimeRange.end,
      co,
    );
    return {
      name: co,
      key: i,
      start: comparisonTimeRange.start,
      end: comparisonTimeRange.end,
    };
  });
}

export function getValidComparisonOption(
  metricsView: V1MetricsViewSpec,
  selectedTimeRange: TimeRange,
  prevComparisonOption: TimeComparisonOption | undefined,
  allTimeRange: TimeRange,
) {
  if (!metricsView.availableTimeRanges?.length) {
    return DEFAULT_TIME_RANGES[selectedTimeRange.name as TimeRangePreset]
      ?.defaultComparison as TimeComparisonOption;
  }

  const timeRange = metricsView.availableTimeRanges.find(
    (tr) => tr.range === selectedTimeRange.name,
  );
  if (!timeRange) return undefined;

  // If comparisonOffsets are not defined get default from presets.
  if (!timeRange.comparisonOffsets?.length) {
    return DEFAULT_TIME_RANGES[selectedTimeRange.name as TimeRangePreset]
      ?.defaultComparison as TimeComparisonOption;
  }

  const existing = timeRange.comparisonOffsets?.find(
    (co) => co.offset === prevComparisonOption,
  );

  const existingComparison = getTimeComparisonParametersForComponent(
    prevComparisonOption,
    allTimeRange.start,
    allTimeRange.end,
    selectedTimeRange.start,
    selectedTimeRange.end,
  );
  // if currently selected comparison option is in allowed list and is valid select it
  if (existing && existingComparison.isComparisonRangeAvailable) {
    return prevComparisonOption;
  }

  return timeRange.comparisonOffsets[0].offset as TimeComparisonOption;
}
