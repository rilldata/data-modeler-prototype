import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { setLeaderboardMeasureName } from "@rilldata/web-common/features/dashboards/state-managers/actions/core-actions";
import {
  removeDimensionFilter,
  toggleDimensionValueSelection,
} from "@rilldata/web-common/features/dashboards/state-managers/actions/dimension-filters";
import {
  setPrimaryDimension,
  toggleDimensionVisibility,
} from "@rilldata/web-common/features/dashboards/state-managers/actions/dimensions";
import {
  removeMeasureFilter,
  setMeasureFilter,
} from "@rilldata/web-common/features/dashboards/state-managers/actions/measure-filters";
import { toggleMeasureVisibility } from "@rilldata/web-common/features/dashboards/state-managers/actions/measures";
import {
  setSortDescending,
  toggleSort,
} from "@rilldata/web-common/features/dashboards/state-managers/actions/sorting";
import type { DashboardMutables } from "@rilldata/web-common/features/dashboards/state-managers/actions/types";
import {
  metricsExplorerStore,
  updateMetricsExplorerByName,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { createPersistentDashboardStore } from "@rilldata/web-common/features/dashboards/stores/persistent-dashboard-state";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

export type TestDashboardMutation = (mut: DashboardMutables) => void;
export const AD_BIDS_APPLY_PUB_DIMENSION_FILTER: TestDashboardMutation = (
  mut,
) => toggleDimensionValueSelection(mut, AD_BIDS_PUBLISHER_DIMENSION, "Google");
export const AD_BIDS_REMOVE_PUB_DIMENSION_FILTER: TestDashboardMutation = (
  mut,
) => removeDimensionFilter(mut, AD_BIDS_PUBLISHER_DIMENSION);
export const AD_BIDS_APPLY_DOM_DIMENSION_FILTER: TestDashboardMutation = (
  mut,
) => toggleDimensionValueSelection(mut, AD_BIDS_DOMAIN_DIMENSION, "google.com");

export const AD_BIDS_APPLY_IMP_MEASURE_FILTER: TestDashboardMutation = (mut) =>
  setMeasureFilter(mut, AD_BIDS_PUBLISHER_DIMENSION, {
    measure: AD_BIDS_IMPRESSIONS_MEASURE,
    type: MeasureFilterType.Value,
    operation: MeasureFilterOperation.GreaterThan,
    value1: "10",
    value2: "",
  });
export const AD_BIDS_REMOVE_IMP_MEASURE_FILTER: TestDashboardMutation = (mut) =>
  removeMeasureFilter(
    mut,
    AD_BIDS_PUBLISHER_DIMENSION,
    AD_BIDS_IMPRESSIONS_MEASURE,
  );
export const AD_BIDS_APPLY_BP_MEASURE_FILTER: TestDashboardMutation = (mut) =>
  setMeasureFilter(mut, AD_BIDS_DOMAIN_DIMENSION, {
    measure: AD_BIDS_BID_PRICE_MEASURE,
    type: MeasureFilterType.Value,
    operation: MeasureFilterOperation.GreaterThan,
    value1: "10",
    value2: "",
  });

export const AD_BIDS_SET_P7D_TIME_RANGE_FILTER: TestDashboardMutation = () =>
  metricsExplorerStore.selectTimeRange(
    AD_BIDS_EXPLORE_NAME,
    { name: TimeRangePreset.LAST_7_DAYS } as any,
    V1TimeGrain.TIME_GRAIN_DAY,
    undefined,
    AD_BIDS_METRICS_INIT,
  );
export const AD_BIDS_SET_P4W_TIME_RANGE_FILTER: TestDashboardMutation = () =>
  metricsExplorerStore.selectTimeRange(
    AD_BIDS_EXPLORE_NAME,
    { name: TimeRangePreset.LAST_4_WEEKS } as any,
    V1TimeGrain.TIME_GRAIN_WEEK,
    undefined,
    AD_BIDS_METRICS_INIT,
  );
export const AD_BIDS_SET_KATHMANDU_TIMEZONE: TestDashboardMutation = () =>
  metricsExplorerStore.setTimeZone(AD_BIDS_EXPLORE_NAME, "Asia/Kathmandu");
export const AD_BIDS_SET_LA_TIMEZONE: TestDashboardMutation = () =>
  metricsExplorerStore.setTimeZone(AD_BIDS_EXPLORE_NAME, "America/Los_Angeles");
export const AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER: TestDashboardMutation =
  () => {
    metricsExplorerStore.displayTimeComparison(AD_BIDS_EXPLORE_NAME, true);
    metricsExplorerStore.setSelectedComparisonRange(
      AD_BIDS_EXPLORE_NAME,
      { name: "rill-PP" } as any,
      AD_BIDS_METRICS_INIT,
    );
  };
export const AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER: TestDashboardMutation =
  () => {
    metricsExplorerStore.displayTimeComparison(AD_BIDS_EXPLORE_NAME, true);
    metricsExplorerStore.setSelectedComparisonRange(
      AD_BIDS_EXPLORE_NAME,
      { name: "rill-PWC" } as any,
      AD_BIDS_METRICS_INIT,
    );
  };
export const AD_BIDS_DISABLE_COMPARE_TIME_RANGE_FILTER: TestDashboardMutation =
  () => metricsExplorerStore.displayTimeComparison(AD_BIDS_EXPLORE_NAME, false);

export const AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY: TestDashboardMutation =
  (mut) => {
    toggleMeasureVisibility(
      mut,
      AD_BIDS_EXPLORE_INIT.measures!,
      AD_BIDS_BID_PRICE_MEASURE,
    );
  };
export const AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY: TestDashboardMutation =
  (mut) => {
    toggleDimensionVisibility(
      mut,
      AD_BIDS_EXPLORE_INIT.dimensions!,
      AD_BIDS_DOMAIN_DIMENSION,
    );
  };

export const AD_BIDS_SORT_DESC_BY_IMPRESSIONS: TestDashboardMutation = (
  mut,
) => {
  setLeaderboardMeasureName(mut, AD_BIDS_IMPRESSIONS_MEASURE);
  setSortDescending(mut);
};
export const AD_BIDS_SORT_ASC_BY_BID_PRICE: TestDashboardMutation = (mut) => {
  setLeaderboardMeasureName(mut, AD_BIDS_BID_PRICE_MEASURE);
  setSortDescending(mut);
  toggleSort(mut, mut.dashboard.dashboardSortType);
};

export const AD_BIDS_OPEN_PUB_DIMENSION_TABLE: TestDashboardMutation = (mut) =>
  setPrimaryDimension(mut, AD_BIDS_PUBLISHER_DIMENSION);
export const AD_BIDS_OPEN_DOM_DIMENSION_TABLE: TestDashboardMutation = (mut) =>
  setPrimaryDimension(mut, AD_BIDS_DOMAIN_DIMENSION);
export const AD_BIDS_CLOSE_DIMENSION_TABLE: TestDashboardMutation = (mut) =>
  setPrimaryDimension(mut, "");

export const AD_BIDS_OPEN_IMP_TDD: TestDashboardMutation = () =>
  metricsExplorerStore.setExpandedMeasureName(
    AD_BIDS_EXPLORE_NAME,
    AD_BIDS_IMPRESSIONS_MEASURE,
  );
export const AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD: TestDashboardMutation = () =>
  metricsExplorerStore.setTDDChartType(
    AD_BIDS_EXPLORE_NAME,
    TDDChart.STACKED_BAR,
  );

export const AD_BIDS_OPEN_BP_TDD: TestDashboardMutation = () =>
  metricsExplorerStore.setExpandedMeasureName(
    AD_BIDS_EXPLORE_NAME,
    AD_BIDS_BID_PRICE_MEASURE,
  );
export const AD_BIDS_CLOSE_TDD: TestDashboardMutation = () =>
  metricsExplorerStore.setExpandedMeasureName(AD_BIDS_EXPLORE_NAME, "");

export const AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS: TestDashboardMutation = () =>
  metricsExplorerStore.createPivot(
    AD_BIDS_EXPLORE_NAME,
    {
      dimension: [
        {
          id: AD_BIDS_PUBLISHER_DIMENSION,
          title: AD_BIDS_PUBLISHER_DIMENSION,
          type: PivotChipType.Dimension,
        },
        {
          id: V1TimeGrain.TIME_GRAIN_HOUR,
          title: "hour",
          type: PivotChipType.Time,
        },
      ],
    },
    {
      dimension: [
        {
          id: AD_BIDS_DOMAIN_DIMENSION,
          title: AD_BIDS_DOMAIN_DIMENSION,
          type: PivotChipType.Dimension,
        },
        {
          id: V1TimeGrain.TIME_GRAIN_DAY,
          title: "day",
          type: PivotChipType.Time,
        },
      ],
      measure: [
        {
          id: AD_BIDS_IMPRESSIONS_MEASURE,
          title: AD_BIDS_IMPRESSIONS_MEASURE,
          type: PivotChipType.Measure,
        },
      ],
    },
  );
export const AD_BIDS_OPEN_DOMAIN_BID_PRICE_PIVOT: TestDashboardMutation = () =>
  metricsExplorerStore.createPivot(
    AD_BIDS_EXPLORE_NAME,
    {
      dimension: [
        {
          id: AD_BIDS_DOMAIN_DIMENSION,
          title: AD_BIDS_DOMAIN_DIMENSION,
          type: PivotChipType.Dimension,
        },
        {
          id: V1TimeGrain.TIME_GRAIN_DAY,
          title: "day",
          type: PivotChipType.Time,
        },
      ],
    },
    {
      dimension: [],
      measure: [
        {
          id: AD_BIDS_IMPRESSIONS_MEASURE,
          title: AD_BIDS_IMPRESSIONS_MEASURE,
          type: PivotChipType.Measure,
        },
      ],
    },
  );
export const AD_BIDS_TOGGLE_PIVOT: TestDashboardMutation = () =>
  metricsExplorerStore.setPivotMode(AD_BIDS_EXPLORE_NAME, false);

export const AD_BIDS_OPEN_PUB_IMP_PIVOT: TestDashboardMutation = () =>
  metricsExplorerStore.createPivot(
    AD_BIDS_EXPLORE_NAME,
    {
      dimension: [
        {
          id: V1TimeGrain.TIME_GRAIN_HOUR,
          title: "hour",
          type: PivotChipType.Time,
        },
        {
          id: AD_BIDS_PUBLISHER_DIMENSION,
          title: AD_BIDS_PUBLISHER_DIMENSION,
          type: PivotChipType.Dimension,
        },
      ],
    },
    {
      dimension: [],
      measure: [
        {
          id: AD_BIDS_IMPRESSIONS_MEASURE,
          title: AD_BIDS_IMPRESSIONS_MEASURE,
          type: PivotChipType.Measure,
        },
      ],
    },
  );
export const AD_BIDS_OPEN_DOM_BP_PIVOT: TestDashboardMutation = () =>
  metricsExplorerStore.createPivot(
    AD_BIDS_EXPLORE_NAME,
    {
      dimension: [
        {
          id: AD_BIDS_DOMAIN_DIMENSION,
          title: AD_BIDS_DOMAIN_DIMENSION,
          type: PivotChipType.Dimension,
        },
      ],
    },
    {
      dimension: [],
      measure: [
        {
          id: AD_BIDS_BID_PRICE_MEASURE,
          title: AD_BIDS_IMPRESSIONS_MEASURE,
          type: PivotChipType.Measure,
        },
      ],
    },
  );

export function applyMutationsToDashboard(
  name: string,
  mutations: TestDashboardMutation[],
) {
  updateMetricsExplorerByName(name, (dashboard) => {
    const dashboardMutables = {
      dashboard,
      persistentDashboardStore: createPersistentDashboardStore("dummy"),
    } as DashboardMutables;
    mutations.forEach((mutation) => mutation(dashboardMutables));
  });
}
