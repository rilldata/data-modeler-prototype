import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type { DashboardMutables } from "./types";
import { getPersistentDashboardStore } from "@rilldata/web-common/features/dashboards/stores/persistent-dashboard-state";

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
  const deleted = dashboard.visibleDimensionKeys.delete(dimensionName);

  if (!deleted) {
    dashboard.visibleDimensionKeys.add(dimensionName);
  }
  const persistentDashboardStore = getPersistentDashboardStore();

  persistentDashboardStore.updateVisibleDimensions(
    Array.from(dashboard.visibleDimensionKeys),
  );
};

export const setVisibleDimensions = (
  { dashboard }: DashboardMutables,
  dimensions: string[],
) => {
  dashboard.visibleDimensionKeys = new Set(dimensions);

  const persistentDashboardStore = getPersistentDashboardStore();

  persistentDashboardStore.updateVisibleDimensions(
    Array.from(dashboard.visibleDimensionKeys),
  );
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
