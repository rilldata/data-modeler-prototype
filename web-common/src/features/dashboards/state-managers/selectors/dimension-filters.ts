import {
  createAndExpression,
  getValuesInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-generators";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import type { AtLeast } from "../types";

export const getFiltersForOtherDimensions = (
  dashData: AtLeast<DashboardDataSources, "dashboard">
): ((dimName: string) => V1Expression) => {
  return (dimName: string) => {
    const exprIdx = getWhereFilterExpressionIndex(dashData)(dimName);
    if (exprIdx === undefined || exprIdx === -1)
      return dashData.dashboard.whereFilter;

    return createAndExpression(
      dashData.dashboard.whereFilter.cond?.exprs?.filter(
        (e) => !matchExpressionByName(e, dimName)
      ) ?? []
    );
  };
};

export const selectedDimensionValues = (
  dashData: AtLeast<DashboardDataSources, "dashboard">
): ((dimName: string) => string[]) => {
  return (dimName: string) => {
    // FIXME: it is possible for this way of accessing the filters
    // to return the same value twice, which would seem to indicate
    // a bug in the way we're setting the filters / active values.
    // Need to investigate further to determine whether this is a
    // problem with the runtime or the client, but for now wrapping
    // it in a set dedupes the values.
    return [
      ...new Set(
        getValuesInExpression(
          getWhereFilterExpression(dashData)(dimName)
        ) as string[]
      ),
    ];
  };
};

export const atLeastOneSelection = (
  dashData: AtLeast<DashboardDataSources, "dashboard">
): ((dimName: string) => boolean) => {
  return (dimName: string) =>
    selectedDimensionValues(dashData)(dimName).length > 0;
};

export const isFilterExcludeMode = (
  dashData: AtLeast<DashboardDataSources, "dashboard">
): ((dimName: string) => boolean) => {
  return (dimName: string) =>
    dashData.dashboard.dimensionFilterExcludeMode.get(dimName) ?? false;
};

export const dimensionHasFilter = (
  dashData: AtLeast<DashboardDataSources, "dashboard">
) => {
  return (dimName: string) => {
    return getWhereFilterExpression(dashData)(dimName) !== undefined;
  };
};

export const matchExpressionByName = (e: V1Expression, name: string) => {
  return e.cond?.exprs?.[0].ident === name;
};

export const getWhereFilterExpression = (
  dashData: AtLeast<DashboardDataSources, "dashboard">
): ((name: string) => V1Expression | undefined) => {
  return (name: string) =>
    dashData.dashboard.whereFilter.cond?.exprs?.find((e) =>
      matchExpressionByName(e, name)
    );
};

export const getWhereFilterExpressionIndex = (
  dashData: AtLeast<DashboardDataSources, "dashboard">
): ((name: string) => number | undefined) => {
  return (name: string) =>
    dashData.dashboard.whereFilter?.cond?.exprs?.findIndex((e) =>
      matchExpressionByName(e, name)
    );
};

export const dimensionFilterSelectors = {
  /**
   * Returns a function that can be used to get
   * a copy of the dashboard's V1MetricsViewFilter that does not include
   * the filters for the specified dimension name.
   */
  getFiltersForOtherDimensions,

  /**
   * Returns a function that can be used to get the selected values
   * for the specified dimension name.
   */
  selectedDimensionValues,

  /**
   * Returns a function that can be used to get whether the specified
   * dimension has at least one selected value.
   */
  atLeastOneSelection,

  /**
   * Returns a function that can be used to get whether the specified
   * dimension is in exclude mode.
   */
  isFilterExcludeMode,

  dimensionHasFilter,
};
