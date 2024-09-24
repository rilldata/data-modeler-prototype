import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain, V1TimeRange } from "@rilldata/web-common/runtime-client";

// Temporary fix to split previous complete ranges to duration and round to grain to get it working on backend
// TODO: Eventually we should support this in the backend.
export const PreviousCompleteRangeMap: Partial<
  Record<TimeRangePreset, V1TimeRange>
> = {
  [TimeRangePreset.YESTERDAY_COMPLETE]: {
    isoDuration: "P1D",
    roundToGrain: V1TimeGrain.TIME_GRAIN_DAY,
  },
  [TimeRangePreset.PREVIOUS_WEEK_COMPLETE]: {
    isoDuration: "P1W",
    roundToGrain: V1TimeGrain.TIME_GRAIN_WEEK,
  },
  [TimeRangePreset.PREVIOUS_MONTH_COMPLETE]: {
    isoDuration: "P1M",
    roundToGrain: V1TimeGrain.TIME_GRAIN_MONTH,
  },
  [TimeRangePreset.PREVIOUS_QUARTER_COMPLETE]: {
    isoDuration: "P3M",
    roundToGrain: V1TimeGrain.TIME_GRAIN_QUARTER,
  },
  [TimeRangePreset.PREVIOUS_YEAR_COMPLETE]: {
    isoDuration: "P1Y",
    roundToGrain: V1TimeGrain.TIME_GRAIN_YEAR,
  },
};

/**
 * Maps selectedTimeRange to V1TimeRange.
 */
export function mapTimeRange(
  timeControlState: TimeControlState,
  defaultTimeRange: string | undefined,
) {
  if (!timeControlState.selectedTimeRange?.name) return undefined;

  const timeRange: V1TimeRange = {};
  switch (timeControlState.selectedTimeRange.name) {
    case TimeRangePreset.DEFAULT:
      timeRange.isoDuration = defaultTimeRange;
      break;

    case TimeRangePreset.CUSTOM:
      timeRange.start = timeControlState.timeStart;
      timeRange.end = timeControlState.timeEnd;
      break;

    default:
      if (timeControlState.selectedTimeRange.name in PreviousCompleteRangeMap) {
        const prevCompleteTimeRange: V1TimeRange | undefined =
          PreviousCompleteRangeMap[timeControlState.selectedTimeRange.name];
        // Backend doesn't support previous complete ranges since it has offset built in.
        // We add the offset manually as a workaround for now
        timeRange.isoDuration = prevCompleteTimeRange?.isoDuration;
        timeRange.isoOffset = prevCompleteTimeRange?.isoOffset;
        timeRange.roundToGrain = prevCompleteTimeRange?.roundToGrain;
      } else {
        timeRange.isoDuration = timeControlState.selectedTimeRange.name;
      }
      break;
  }

  return timeRange;
}

/**
 * Maps selectedComparisonTimeRange to V1TimeRange if time comparison is enabled.
 */
export function mapComparisonTimeRange(
  dashboardState: MetricsExplorerEntity,
  timeControlState: TimeControlState,
  timeRange: V1TimeRange | undefined,
) {
  if (
    !timeRange ||
    dashboardState.selectedComparisonDimension ||
    !timeControlState.showTimeComparison ||
    !timeControlState.selectedComparisonTimeRange?.name
  ) {
    return undefined;
  }

  const comparisonTimeRange: V1TimeRange = {};
  switch (timeControlState.selectedComparisonTimeRange.name) {
    default:
      comparisonTimeRange.isoOffset =
        timeControlState.selectedComparisonTimeRange.name;
      comparisonTimeRange.isoDuration = timeRange.isoDuration;
      break;
    case TimeComparisonOption.CONTIGUOUS:
      comparisonTimeRange.isoOffset = comparisonTimeRange.isoDuration =
        timeRange.isoDuration;
      break;

    case TimeComparisonOption.CUSTOM:
      comparisonTimeRange.start = timeControlState.comparisonTimeStart;
      comparisonTimeRange.end = timeControlState.comparisonTimeEnd;
      break;
  }
  return comparisonTimeRange;
}
