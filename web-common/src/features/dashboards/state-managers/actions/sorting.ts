import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { DashboardMutables } from "./types";

export const toggleSort = (
  { dashboard, persistentDashboardStore }: DashboardMutables,
  sortType: SortType,
) => {
  if (sortType === undefined || dashboard.dashboardSortType === sortType) {
    // If it's already sorted by this type, cycle through: descending -> ascending -> unspecified
    if (dashboard.sortDirection === SortDirection.DESCENDING) {
      dashboard.sortDirection = SortDirection.ASCENDING;
    } else if (dashboard.sortDirection === SortDirection.ASCENDING) {
      dashboard.dashboardSortType = SortType.UNSPECIFIED;
      dashboard.sortDirection = SortDirection.UNSPECIFIED;
    } else {
      dashboard.dashboardSortType = sortType;
      dashboard.sortDirection = SortDirection.DESCENDING;
    }
  } else {
    // If it's a new sort type, start with descending
    dashboard.dashboardSortType = sortType;
    dashboard.sortDirection = SortDirection.DESCENDING;
  }

  persistentDashboardStore.updateDashboardSortType(dashboard.dashboardSortType);
  persistentDashboardStore.updateSortDirection(dashboard.sortDirection);
};

const contextColumnToSortType = {
  [LeaderboardContextColumn.DELTA_PERCENT]: SortType.DELTA_PERCENT,
  [LeaderboardContextColumn.DELTA_ABSOLUTE]: SortType.DELTA_ABSOLUTE,
  [LeaderboardContextColumn.PERCENT]: SortType.PERCENT,
};

export const toggleSortByActiveContextColumn = (args: DashboardMutables) => {
  const contextColumnSortType =
    contextColumnToSortType[args.dashboard.leaderboardContextColumn];
  toggleSort(args, contextColumnSortType);
};

export const sortActions = {
  /**
   * Sets the sort type for the dashboard (value, percent, delta, etc.)
   */
  toggleSort,

  /**
   * Toggles the sort type according to the active context column.
   */
  toggleSortByActiveContextColumn,

  /**
   * Sets the dashboard to be sorted by dimension value.
   * Note that this should only be used in the dimension table
   */
  sortByDimensionValue: (mutatorArgs: DashboardMutables) =>
    toggleSort(mutatorArgs, SortType.DIMENSION),

  /**
   * Sets the sort direction to descending.
   */
  setSortDescending: ({
    dashboard,
    persistentDashboardStore,
  }: DashboardMutables) => {
    dashboard.sortDirection = SortDirection.DESCENDING;
    persistentDashboardStore.updateSortDirection(dashboard.sortDirection);
  },
};
