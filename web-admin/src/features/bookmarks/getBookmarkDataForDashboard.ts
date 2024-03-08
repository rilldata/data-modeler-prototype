import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";

/**
 * Returns the bookmark data to be stored.
 * Converts the dashboard to base64 protobuf string using {@link getProtoFromDashboardState}
 *
 * @param dashboard
 * @param filtersOnly Only dimension/measure filters and the selected time range is stored.
 * @param absoluteTimeRange Time ranges is treated as absolute.
 */
export function getBookmarkDataForDashboard(
  dashboard: MetricsExplorerEntity,
  filtersOnly: boolean,
  absoluteTimeRange: boolean,
): string {
  if (absoluteTimeRange) {
    dashboard = {
      ...dashboard,
    };
    if (
      dashboard.selectedTimeRange?.start &&
      dashboard.selectedTimeRange?.end
    ) {
      dashboard.selectedTimeRange = {
        name: TimeRangePreset.CUSTOM,
        interval: dashboard.selectedTimeRange.interval,
        start: dashboard.selectedTimeRange.start,
        end: dashboard.selectedTimeRange.end,
      };
    }
    if (
      dashboard.selectedComparisonTimeRange?.start &&
      dashboard.selectedComparisonTimeRange?.end
    ) {
      dashboard.selectedComparisonTimeRange = {
        name: TimeRangePreset.CUSTOM,
        interval: dashboard.selectedComparisonTimeRange.interval,
        start: dashboard.selectedComparisonTimeRange.start,
        end: dashboard.selectedComparisonTimeRange.end,
      };
    }
  }

  if (filtersOnly) {
    return getProtoFromDashboardState({
      whereFilter: dashboard.whereFilter,
      dimensionThresholdFilters: dashboard.dimensionThresholdFilters,
      selectedTimeRange: dashboard.selectedTimeRange,
    } as MetricsExplorerEntity);
  }

  return getProtoFromDashboardState(dashboard);
}
