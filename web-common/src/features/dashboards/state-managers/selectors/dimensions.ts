import type { MetricsViewSpecDimensionV2 } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";

export const allDimensions = ({
  metricsSpecQueryResult,
}: DashboardDataSources) => {
  return metricsSpecQueryResult.data?.dimensions ?? [];
};

export const visibleDimensions = ({
  metricsSpecQueryResult,
  dashboard,
}: DashboardDataSources): MetricsViewSpecDimensionV2[] => {
  const dimensions = metricsSpecQueryResult.data?.dimensions ?? [];

  if (dashboard.visibleDimensionKeys === null) {
    return dimensions;
  } else {
    return dimensions.filter(
      (d) => d.name && dashboard.visibleDimensionKeys.has(d.name),
    );
  }
};

export const dimensionTableDimName = ({
  dashboard,
}: DashboardDataSources): string | undefined => {
  return dashboard.selectedDimensionName;
};

export const dimensionTableColumnName = (
  dashData: DashboardDataSources,
): ((name: string) => string) => {
  return (name: string) => {
    const dim = getDimensionByName(dashData)(name);
    return dim?.column || name;
  };
};

export const getDimensionByName = (
  dashData: DashboardDataSources,
): ((name: string) => MetricsViewSpecDimensionV2 | undefined) => {
  return (name: string) => {
    return allDimensions(dashData)?.find(
      (dimension) => dimension.name === name,
    );
  };
};

export const getDimensionDisplayName = (
  dashData: DashboardDataSources,
): ((name: string) => string) => {
  return (name: string) => {
    const dim = getDimensionByName(dashData)(name);
    return (dim?.label?.length ? dim?.label : dim?.name) ?? name;
  };
};

export const getDimensionDescription = (
  dashData: DashboardDataSources,
): ((name: string) => string) => {
  return (name: string) => {
    const dim = getDimensionByName(dashData)(name);
    return dim?.description || "";
  };
};

export const comparisonDimension = (dashData: DashboardDataSources) => {
  const dimName = dashData.dashboard.selectedComparisonDimension;
  if (!dimName) return undefined;
  return getDimensionByName(dashData)(dimName);
};

export const dimensionSelectors = {
  /**
   * Gets all dimensions for the dashboard, or undefined if there are none.
   */
  allDimensions,

  /**
   * Gets all visible dimensions in the dashboard.
   */
  visibleDimensions,

  /**
   * Returns a function that can be used to get a MetricsViewSpecDimensionV2
   * by name; this fn returns undefined if the dashboard has no dimension with that name.
   */
  getDimensionByName,
  /**
   * Returns a function that can be used to get a dimension's display
   * given its "key" name.
   */
  getDimensionDisplayName,
  /**
   * Returns a function that can be used to get a dimension's description
   * given its "key" name. Returns an empty string if the dimension has no description.
   */
  getDimensionDescription,

  /**
   * Gets the dimension that is currently being compared.
   * Returns undefined otherwise.
   */
  comparisonDimension,

  /**
   * Gets the name of the dimension that is currently selected in the dimension table.
   * Returns undefined if no dimension is selected, in which case the dimension table
   * is not shown.
   */
  dimensionTableDimName,

  /**
   * Gets the name of the column that is currently selected in the dimension table.
   */
  dimensionTableColumnName,
};
