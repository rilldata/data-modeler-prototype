import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaPreviousSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  QueryServiceMetricsViewAggregationBody,
  V1Expression,
  V1MetricsViewAggregationMeasure,
} from "@rilldata/web-common/runtime-client";
import { SortType } from "./proto-state/derived-types";
import type { TimeControlState } from "./time-controls/time-control-store";

const countRegex = /count(?=[^(]*\()/i;
const sumRegex = /sum(?=[^(]*\()/i;
// Clickhouse has uniq and uniqExact, which are summable
const uniqRegex = /uniq(?=[^(]*\()/i;
const uniqExactRegex = /uniqExact(?=[^(]*\()/i;

export function isSummableMeasure(measure: MetricsViewSpecMeasureV2): boolean {
  const expression = measure.expression?.toLowerCase();
  return !!(
    expression?.match(countRegex) ||
    expression?.match(sumRegex) ||
    expression?.match(uniqRegex) ||
    expression?.match(uniqExactRegex)
  );
}

/**
 * Returns a sanitized column name appropriate for use in e.g. filters.
 *
 * Even though this is a one-liner, we externalize it as a function
 * becuase it is used in a few places and we want to make sure we
 * are consistent in how we handle this.
 */
export function getDimensionColumn(
  dimension: MetricsViewSpecDimensionV2,
): string {
  return (dimension?.column || dimension?.name) as string;
}

export function prepareSortedQueryBody(
  dimensionName: string,
  measureNames: string[],
  timeControls: TimeControlState,
  // Note: sortMeasureName may be null if we are sorting by dimension values
  sortMeasureName: string | null,
  sortType: SortType,
  sortAscending: boolean,
  whereFilterForDimension: V1Expression,
  limit: number,
): QueryServiceMetricsViewAggregationBody {
  const measures = measureNames.map(
    (n) =>
      <V1MetricsViewAggregationMeasure>{
        name: n,
      },
  );

  let apiSortName = sortMeasureName;
  if (sortType === SortType.DIMENSION || sortMeasureName === null) {
    apiSortName = dimensionName;
  }

  if (
    timeControls.showTimeComparison &&
    !!timeControls.selectedComparisonTimeRange &&
    sortMeasureName
  ) {
    // insert beside the correct measure
    measures.splice(
      measures.findIndex((m) => m.name === sortMeasureName),
      0,
      ...getComparisonRequestMeasures(sortMeasureName),
    );
    if (apiSortName === sortMeasureName) {
      // only update if the sort was on measure
      switch (sortType) {
        case DashboardState_LeaderboardSortType.DELTA_ABSOLUTE:
          apiSortName += ComparisonDeltaAbsoluteSuffix;
          break;
        case DashboardState_LeaderboardSortType.DELTA_PERCENT:
          apiSortName += ComparisonDeltaRelativeSuffix;
          break;
      }
    }
  }

  return {
    dimensions: [
      {
        name: dimensionName,
      },
    ],
    measures,
    timeRange: {
      start: timeControls.timeStart,
      end: timeControls.timeEnd,
    },
    ...(timeControls.selectedComparisonTimeRange &&
    timeControls.showTimeComparison
      ? {
          comparisonTimeRange: {
            start: timeControls.comparisonTimeStart,
            end: timeControls.comparisonTimeEnd,
          },
        }
      : {}),
    sort: apiSortName
      ? [
          {
            desc: !sortAscending,
            name: apiSortName,
          },
        ]
      : [],
    where: sanitiseExpression(whereFilterForDimension, undefined),
    limit: limit.toString(),
    offset: "0",
  };
}

/**
 * Gets comparison based measures used in MetricsViewAggregationRequest
 */
export function getComparisonRequestMeasures(
  measureName: string,
): V1MetricsViewAggregationMeasure[] {
  return [
    {
      name: measureName + ComparisonDeltaPreviousSuffix,
      comparisonValue: {
        measure: measureName,
      },
    },
    {
      name: measureName + ComparisonDeltaAbsoluteSuffix,
      comparisonDelta: {
        measure: measureName,
      },
    },
    {
      name: measureName + ComparisonDeltaRelativeSuffix,
      comparisonRatio: {
        measure: measureName,
      },
    },
  ];
}
