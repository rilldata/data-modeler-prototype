import type {
  QueryServiceMetricsViewComparisonToplistBody,
  MetricsViewDimension,
  MetricsViewMeasure,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import type { TimeControlState } from "./time-controls/time-control-store";
import { getQuerySortType } from "./leaderboard/leaderboard-utils";
import { SortType } from "./proto-state/derived-types";

export function isSummableMeasure(measure: MetricsViewMeasure): boolean {
  return (
    measure?.expression.toLowerCase()?.includes("count(") ||
    measure?.expression?.toLowerCase()?.includes("sum(")
  );
}

/**
 * Returns a sanitized column name appropriate for use in e.g. filters.
 *
 * Even though this is a one-liner, we externalize it as a function
 * becuase it is used in a few places and we want to make sure we
 * are consistent in how we handle this.
 */
export function getDimensionColumn(dimension: MetricsViewDimension) {
  return dimension?.column || dimension?.name;
}

export function prepareSortedQueryBody(
  dimensionName: string,
  measureNames: string[],
  timeControls: TimeControlState,
  sortMeasureName: string,

  sortType: SortType,
  sortAscending: boolean,
  filterForDimension: V1MetricsViewFilter
): QueryServiceMetricsViewComparisonToplistBody {
  let comparisonTimeRange = {
    start: timeControls.comparisonTimeStart,
    end: timeControls.comparisonTimeEnd,
  };

  // FIXME: As a temporary way of enabling sorting by dimension values,
  // Benjamin and Egor put in a patch that will allow us to use the
  // dimension name as the measure name. This will need to be updated
  // once they have stabilized the API.
  if (sortType === SortType.DIMENSION) {
    sortMeasureName = dimensionName;
    // note also that we need to remove the comparison time range
    // when sorting by dimension values, or the query errors
    comparisonTimeRange = undefined;
  }

  const querySortType = getQuerySortType(sortType);

  return {
    dimensionName,
    measureNames,
    baseTimeRange: {
      start: timeControls.timeStart,
      end: timeControls.timeEnd,
    },
    comparisonTimeRange,
    sort: [
      {
        ascending: sortAscending,
        measureName: sortMeasureName,
        type: querySortType,
      },
    ],
    filter: filterForDimension,
    limit: "250",
    offset: "0",
  };
}
