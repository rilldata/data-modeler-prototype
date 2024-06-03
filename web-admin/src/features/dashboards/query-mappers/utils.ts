import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { PreviousCompleteRangeMap } from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  getQueryServiceMetricsViewAggregationQueryKey,
  queryServiceMetricsViewAggregation,
  type QueryServiceMetricsViewAggregationBody,
  type V1Expression,
  type V1TimeRange,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

// We are manually sending in duration, offset and round to grain for previous complete ranges.
// This is to map back that split
const PreviousCompleteRangeReverseMap: Record<string, TimeRangePreset> = {};
for (const preset in PreviousCompleteRangeMap) {
  const range: V1TimeRange = PreviousCompleteRangeMap[preset];
  PreviousCompleteRangeReverseMap[
    `${range.isoDuration}_${range.isoOffset}_${range.roundToGrain}`
  ] = preset as TimeRangePreset;
}

export function fillTimeRange(
  dashboard: MetricsExplorerEntity,
  reqTimeRange: V1TimeRange | undefined,
  reqComparisonTimeRange: V1TimeRange | undefined,
  timeRangeSummary: V1TimeRangeSummary,
  executionTime: string,
) {
  if (reqTimeRange) {
    dashboard.selectedTimeRange = getSelectedTimeRange(
      reqTimeRange,
      timeRangeSummary,
      reqTimeRange.isoDuration ?? "",
      executionTime,
    );
  }

  if (reqComparisonTimeRange) {
    if (
      !reqComparisonTimeRange.isoOffset &&
      reqComparisonTimeRange.isoDuration
    ) {
      dashboard.selectedComparisonTimeRange = {
        name: TimeComparisonOption.CONTIGUOUS,
        start: undefined as unknown as Date,
        end: undefined as unknown as Date,
      };
    } else {
      dashboard.selectedComparisonTimeRange = getSelectedTimeRange(
        reqComparisonTimeRange,
        timeRangeSummary,
        reqComparisonTimeRange.isoOffset,
        executionTime,
      );
    }

    if (dashboard.selectedComparisonTimeRange) {
      dashboard.selectedComparisonTimeRange.interval =
        dashboard.selectedTimeRange?.interval;
    }
    dashboard.showTimeComparison = true;
  }
}

export function getSelectedTimeRange(
  timeRange: V1TimeRange,
  timeRangeSummary: V1TimeRangeSummary,
  duration: string | undefined,
  executionTime: string,
): DashboardTimeControls | undefined {
  let selectedTimeRange: DashboardTimeControls;

  const fullRangeKey = `${timeRange.isoDuration ?? ""}_${timeRange.isoOffset ?? ""}_${timeRange.roundToGrain ?? ""}`;
  if (fullRangeKey in PreviousCompleteRangeReverseMap) {
    duration = PreviousCompleteRangeReverseMap[fullRangeKey];
  }

  if (timeRange.start && timeRange.end) {
    selectedTimeRange = {
      name: TimeRangePreset.CUSTOM,
      start: new Date(timeRange.start),
      end: new Date(timeRange.end),
    };
  } else if (duration && timeRangeSummary.min) {
    selectedTimeRange = isoDurationToFullTimeRange(
      duration,
      new Date(timeRangeSummary.min),
      new Date(executionTime),
    );
  } else {
    return undefined;
  }

  selectedTimeRange.interval = timeRange.roundToGrain;

  return selectedTimeRange;
}

export async function convertExprToToplist(
  queryClient: QueryClient,
  instanceId: string,
  metricsView: string,
  dimensionName: string,
  measureName: string,
  timeRange: V1TimeRange | undefined,
  comparisonTimeRange: V1TimeRange | undefined,
  executionTime: string,
  where: V1Expression | undefined,
  having: V1Expression,
) {
  const toplistBody: QueryServiceMetricsViewAggregationBody = {
    measures: [
      {
        name: measureName,
      },
      ...(comparisonTimeRange
        ? [
            {
              name: measureName + ComparisonDeltaAbsoluteSuffix,
              comparisonDelta: { measure: measureName },
            },
            {
              name: measureName + ComparisonDeltaRelativeSuffix,
              comparisonRatio: { measure: measureName },
            },
          ]
        : []),
    ],
    dimensions: [{ name: dimensionName }],
    ...(timeRange
      ? {
          timeRange: {
            ...timeRange,
            end: executionTime,
          },
        }
      : {}),
    ...(comparisonTimeRange
      ? {
          comparisonTimeRange: {
            ...comparisonTimeRange,
            end: executionTime,
          },
        }
      : {}),
    where,
    having,
    sort: [
      {
        name: measureName,
        desc: false,
      },
    ],
    limit: "250",
  };
  const toplist = await queryClient.fetchQuery({
    queryKey: getQueryServiceMetricsViewAggregationQueryKey(
      instanceId,
      metricsView,
      toplistBody,
    ),
    queryFn: () =>
      queryServiceMetricsViewAggregation(instanceId, metricsView, toplistBody),
  });
  if (!toplist.data) {
    return undefined;
  }
  return createInExpression(
    dimensionName,
    toplist.data.map((t) => t[dimensionName]),
  );
}
