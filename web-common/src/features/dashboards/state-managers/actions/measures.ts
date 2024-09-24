import type { DashboardMutables } from "./types";

export const toggleMeasureVisibility = (
  { dashboard }: DashboardMutables,

  measureName: string,
) => {
  if (dashboard.visibleMeasureKeys.has(measureName)) {
    dashboard.visibleMeasureKeys.delete(measureName);
  } else {
    dashboard.visibleMeasureKeys.add(measureName);
  }
};

export const setVisibleMeasures = (
  { dashboard }: DashboardMutables,
  measureNames: string[],
) => {
  dashboard.visibleMeasureKeys = new Set(measureNames);
};

export const measureActions = {
  setVisibleMeasures,
  toggleMeasureVisibility,
};
