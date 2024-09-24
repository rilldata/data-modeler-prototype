import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type { DashboardMutables } from "./types";

export const setPrimaryDimension = (
  { dashboard }: DashboardMutables,

  dimensionName: string | undefined,
) => {
  dashboard.selectedDimensionName = dimensionName;
  if (dimensionName) {
    dashboard.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
  } else {
    dashboard.activePage = DashboardState_ActivePage.DEFAULT;
  }
};

export const toggleDimensionVisibility = (
  { dashboard }: DashboardMutables,

  dimensionName: string,
) => {
  if (dashboard.visibleDimensionKeys.has(dimensionName)) {
    dashboard.visibleDimensionKeys.delete(dimensionName);
  } else {
    dashboard.visibleDimensionKeys.add(dimensionName);
  }
};

export const setVisibleDimensions = (
  { dashboard }: DashboardMutables,
  dimensions: string[],
) => {
  dashboard.visibleDimensionKeys = new Set(dimensions);
};

export const dimensionActions = {
  /**
   * Sets the primary dimension for the dashboard, which
   * activates the dimension table. Setting the primary dimension
   * to undefined closes the dimension table.
   */
  setPrimaryDimension,
  toggleDimensionVisibility,
  setVisibleDimensions,
};
