import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  QueryServiceMetricsViewComparisonBody,
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  V1MetricsViewAggregationMeasure,
  V1Expression,
} from "@rilldata/web-common/runtime-client";
import type { TimeControlState } from "./time-controls/time-control-store";
import { getQuerySortType } from "./leaderboard/leaderboard-utils";
import { SortType } from "./proto-state/derived-types";

const countRegex = /count(?=[^(]*\()/i;
const sumRegex = /sum(?=[^(]*\()/i;

export function isSummableMeasure(measure: MetricsViewSpecMeasureV2): boolean {
  const expression = measure.expression?.toLowerCase();
  return !!(expression?.match(countRegex) || expression?.match(sumRegex));
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
  return dimension?.column || dimension?.name;
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
  havingFilterForDimension: V1Expression | undefined,
  limit: number,
): QueryServiceMetricsViewComparisonBody {
  let comparisonTimeRange = {
    start: timeControls.comparisonTimeStart,
    end: timeControls.comparisonTimeEnd,
  };

  // api now expects measure names for which comparison are calculated
  // to keep current behaviour add sort measure name to comparison measures
  let comparisonMeasures: string[] = [];
  if (comparisonTimeRange.start && comparisonTimeRange.end && sortMeasureName) {
    comparisonMeasures = [sortMeasureName];
  }

  // FIXME: As a temporary way of enabling sorting by dimension values,
  // Benjamin and Egor put in a patch that will allow us to use the
  // dimension name as the measure name. This will need to be updated
  // once they have stabilized the API.
  if (sortType === SortType.DIMENSION || sortMeasureName === null) {
    sortMeasureName = dimensionName;
    // note also that we need to remove the comparison time range
    // when sorting by dimension values, or the query errors
    comparisonTimeRange = undefined;
    // and we need to remove the comparison measures
    comparisonMeasures = [];
  }

  const querySortType = getQuerySortType(sortType);

  return {
    dimension: {
      name: dimensionName,
    },
    measures: measureNames.map(
      (n) =>
        <V1MetricsViewAggregationMeasure>{
          name: n,
        },
    ),
    comparisonMeasures: comparisonMeasures,
    timeRange: {
      start: timeControls.timeStart,
      end: timeControls.timeEnd,
    },
    comparisonTimeRange,
    sort: [
      {
        desc: !sortAscending,
        name: sortMeasureName,
        sortType: querySortType,
      },
    ],
    where: sanitiseExpression(
      whereFilterForDimension,
      havingFilterForDimension,
    ),
    limit: limit.toString(),
    offset: "0",
  };
}
