import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { additionalMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
import { getIndependentMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { QueryServiceMetricsViewAggregationBody } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import { prepareSortedQueryBody } from "../../dashboard-utils";
import {
  activeMeasureName,
  isAnyMeasureSelected,
  selectedMeasureNames,
} from "./active-measure";
import { sortingSelectors } from "./sorting";
import { isTimeControlReady, timeControlsState } from "./time-range";
import { getFiltersForOtherDimensions } from "./dimension-filters";
import { updateFilterOnSearch } from "../../dimension-table/dimension-table-utils";
import { dimensionTableSearchString } from "./dimension-table";

/**
 * Returns the sorted query body for the dimension table for the
 * active dimension.
 *
 * Safety: this readable should ONLY be used when the active dimension
 * is not undefined, ie then the dimension table is visible.
 */
export function dimensionTableSortedQueryBody(
  dashData: DashboardDataSources,
): QueryServiceMetricsViewAggregationBody {
  const dimensionName = dashData.dashboard.selectedDimensionName;
  if (!dimensionName) {
    return {};
  }
  let filters = getFiltersForOtherDimensions(dashData)(dimensionName);
  const searchString = dimensionTableSearchString(dashData);
  if (searchString !== undefined) {
    filters = updateFilterOnSearch(filters, searchString, dimensionName);
  }

  return prepareSortedQueryBody(
    dimensionName,
    measuresForDimensionTable(dashData),
    timeControlsState(dashData),
    sortingSelectors.sortMeasure(dashData),
    sortingSelectors.sortType(dashData),
    sortingSelectors.sortedAscending(dashData),
    mergeMeasureFilters(dashData.dashboard, filters),
    250,
  );
}

function measuresForDimensionTable(dashData: DashboardDataSources) {
  const allMeasures = new Set([
    ...selectedMeasureNames(dashData),
    ...additionalMeasures(dashData),
  ]);
  const validSpec = dashData.metricsSpecQueryResult.data;
  if (!validSpec) {
    throw new Error("No valid metrics spec");
  }
  return getIndependentMeasures(validSpec, [...allMeasures]);
}

export function dimensionTableTotalQueryBody(
  dashData: DashboardDataSources,
): QueryServiceMetricsViewAggregationBody {
  const dimensionName = dashData.dashboard.selectedDimensionName;
  if (!dimensionName) {
    return {};
  }
  return leaderboardDimensionTotalQueryBody(dashData)(dimensionName);
}

/**
 * Returns a function that can be used to get the sorted query body
 * for a leaderboard for the given dimension.
 */
export function leaderboardSortedQueryBody(
  dashData: DashboardDataSources,
): (dimensionName: string) => QueryServiceMetricsViewAggregationBody {
  const validSpec = dashData.metricsSpecQueryResult.data;
  if (!validSpec) {
    throw new Error("No valid metrics spec");
  }

  return (dimensionName: string) =>
    prepareSortedQueryBody(
      dimensionName,
      getIndependentMeasures(validSpec, additionalMeasures(dashData)),
      timeControlsState(dashData),
      sortingSelectors.sortMeasure(dashData),
      sortingSelectors.sortType(dashData),
      sortingSelectors.sortedAscending(dashData),
      mergeMeasureFilters(
        dashData.dashboard,
        getFiltersForOtherDimensions(dashData)(dimensionName),
      ),
      8,
    );
}

export function leaderboardSortedQueryOptions(
  dashData: DashboardDataSources,
): (
  dimensionName: string,
  enabled: boolean,
) => { query: { enabled: boolean } } {
  return (dimensionName: string, enabled: boolean) => {
    const sortedQueryEnabled =
      timeControlsState(dashData).ready === true &&
      !!getFiltersForOtherDimensions(dashData)(dimensionName);
    return {
      query: {
        enabled: enabled && sortedQueryEnabled,
      },
    };
  };
}

export function leaderboardDimensionTotalQueryBody(
  dashData: DashboardDataSources,
): (dimensionName: string) => QueryServiceMetricsViewAggregationBody {
  return (dimensionName: string) => ({
    measures: [{ name: activeMeasureName(dashData) }],
    where: sanitiseExpression(
      mergeMeasureFilters(
        dashData.dashboard,
        getFiltersForOtherDimensions(dashData)(dimensionName),
      ),
      undefined,
    ),
    timeStart: timeControlsState(dashData).timeStart,
    timeEnd: timeControlsState(dashData).timeEnd,
  });
}

export function leaderboardDimensionTotalQueryOptions(
  dashData: DashboardDataSources,
): (dimensionName: string) => { query: { enabled: boolean } } {
  return (dimensionName: string) => {
    return {
      query: {
        enabled:
          isAnyMeasureSelected(dashData) &&
          isTimeControlReady(dashData) &&
          !!getFiltersForOtherDimensions(dashData)(dimensionName),
      },
    };
  };
}

export const leaderboardQuerySelectors = {
  /**
   * Readable containing the sorted query body for the dimension table.
   */
  dimensionTableSortedQueryBody,

  /**
   * Readable containing the totals query body for the dimension table.
   */
  dimensionTableTotalQueryBody,

  /**
   * Readable containg a function that will return
   * the sorted query body for a leaderboard for the given dimension.
   */
  leaderboardSortedQueryBody,

  /**
   * Readable containg a function that will return
   * the sorted query options for a leaderboard for the given dimension.
   */
  leaderboardSortedQueryOptions,

  /**
   * Readable containg a function that will return
   * the totals query body for a leaderboard for the given dimension.
   */
  leaderboardDimensionTotalQueryBody,

  /**
   * Readable containg a function that will return
   * the totals query options for a leaderboard for the given dimension.
   */
  leaderboardDimensionTotalQueryOptions,
};
