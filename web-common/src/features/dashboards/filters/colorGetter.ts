import { CHECKMARK_COLORS } from "../config";

type DashboardName = string;
type DimensionName = string;
type DimensionValue = string;

function createColorLookup() {
  const map = new Map<
    DashboardName,
    Map<DimensionName, Map<DimensionValue, number>>
  >();

  return {
    get: (
      dashboardName: DashboardName,
      dimensionName: DimensionName,
      dimensionValue: DimensionValue,
    ) => {
      let dashboardMap = map.get(dashboardName);

      if (!dashboardMap) {
        dashboardMap = new Map<DimensionName, Map<DimensionValue, number>>();
        map.set(dashboardName, dashboardMap);
      }

      let dimensionMap = dashboardMap.get(dimensionName);

      if (!dimensionMap) {
        dimensionMap = new Map<DimensionValue, number>();
        dashboardMap.set(dimensionName, dimensionMap);
      }

      let colorIndex = dimensionMap.get(dimensionValue);

      if (colorIndex === undefined) {
        colorIndex = firstMissingPositive([...dimensionMap.values()]);
        dimensionMap.set(dimensionValue, colorIndex);
      }

      return CHECKMARK_COLORS[colorIndex] ?? "gray-300";
    },
    remove: (
      dashboardName: DashboardName,
      dimensionName: DimensionName,
      dimensionValue: DimensionValue,
    ) => {
      map.get(dashboardName)?.get(dimensionName)?.delete(dimensionValue);
    },
    removeDimension: (
      dashboardName: DashboardName,
      dimensionName: DimensionName,
    ) => {
      map.get(dashboardName)?.delete(dimensionName);
    },
  };
}

function firstMissingPositive(nums: number[]) {
  const numSet = new Set(nums);
  let smallest = 0;
  while (numSet.has(smallest)) {
    smallest++;
  }
  return smallest;
}

export const colorGetter = createColorLookup();
