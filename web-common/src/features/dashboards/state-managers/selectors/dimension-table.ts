import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
import type {
  MetricsViewSpecDimensionV2,
  RpcStatus,
  V1MetricsViewAggregationResponse,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import { isSummableMeasure } from "../../dashboard-utils";
import type { DimensionTableRow } from "../../dimension-table/dimension-table-types";
import {
  prepareDimensionTableRows,
  prepareVirtualizedDimTableColumns,
} from "../../dimension-table/dimension-table-utils";
import { activeMeasureName, isValidPercentOfTotal } from "./active-measure";
import { selectedDimensionValues } from "./dimension-filters";
import { allMeasures, visibleMeasures } from "./measures";
import { isTimeComparisonActive } from "./time-range";
import type { DashboardDataSources } from "./types";

export const selectedDimensionValueNames = (
  dashData: DashboardDataSources,
): string[] => {
  const dimension = dashData.dashboard.selectedDimensionName;
  if (!dimension) return [];
  return selectedDimensionValues(dashData)(dimension);
};

export const primaryDimension = (
  dashData: DashboardDataSources,
): MetricsViewSpecDimensionV2 | undefined => {
  const dimName = dashData.dashboard.selectedDimensionName;
  return dashData.validMetricsView?.dimensions?.find(
    (dim) => dim.name === dimName,
  );
};

export const virtualizedTableColumns =
  (
    dashData: DashboardDataSources,
  ): ((
    totalsQuery: QueryObserverResult<
      V1MetricsViewAggregationResponse,
      RpcStatus
    >,
  ) => VirtualizedTableColumns[]) =>
  (totalsQuery) => {
    const dimension = primaryDimension(dashData);

    if (!dimension) return [];

    // temporary filter for advanced measures
    const measures = visibleMeasures(dashData).filter(
      (m) => !m.window && !m.requiredDimensions?.length,
    );

    const measureTotals: { [key: string]: number } = {};
    if (totalsQuery?.data?.data) {
      measures.map((m) => {
        if (m.name && isSummableMeasure(m)) {
          measureTotals[m.name] = totalsQuery.data?.data?.[0]?.[m.name];
        }
      });
    }

    return prepareVirtualizedDimTableColumns(
      dashData.dashboard,
      measures,
      measureTotals,
      dimension,
      isTimeComparisonActive(dashData),
      isValidPercentOfTotal(dashData),
    );
  };

export const prepareDimTableRows =
  (
    dashData: DashboardDataSources,
  ): ((
    sortedQuery: QueryObserverResult<
      V1MetricsViewAggregationResponse,
      RpcStatus
    >,
    unfilteredTotal: number,
  ) => DimensionTableRow[]) =>
  (sortedQuery, unfilteredTotal) => {
    const dimension = primaryDimension(dashData);

    if (!dimension) return [];

    const dimensionColumn = dimension.name ?? "";
    const leaderboardMeasureName = activeMeasureName(dashData);

    // FIXME: should this really be all measures, or just visible measures?
    const measures = allMeasures(dashData);

    return prepareDimensionTableRows(
      sortedQuery?.data?.data ?? [],
      measures,
      leaderboardMeasureName,
      dimensionColumn,
      isTimeComparisonActive(dashData),
      isValidPercentOfTotal(dashData),
      unfilteredTotal,
    );
  };

export const dimensionTableSelectors = {
  /**
   * gets the VirtualizedTableColumns array for the dimension table.
   */
  virtualizedTableColumns,

  /**
   * gets the MetricsViewSpecDimensionV2 for the dimension table's
   * primary dimension.
   */
  primaryDimension,

  /**
   * gets the names of the selected dimension values for the primary dimension.
   */
  selectedDimensionValueNames,

  /**
   * A readable containaing a function that will prepare
   * the dimension table rows for given a sorted query
   * and unfiltered total.
   */
  prepareDimTableRows,
};
