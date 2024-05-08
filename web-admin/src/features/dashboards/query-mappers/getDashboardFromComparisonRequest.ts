import { getSelectedTimeRange } from "@rilldata/web-admin/features/dashboards/query-mappers/getSelectedTimeRange";
import type { QueryMapperArgs } from "@rilldata/web-admin/features/dashboards/query-mappers/types";
import { getSortType } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type { V1MetricsViewComparisonRequest } from "@rilldata/web-common/runtime-client";

export function getDashboardFromComparisonRequest({
  req,
  dashboard,
  metricsView,
  timeRangeSummary,
  executionTime,
}: QueryMapperArgs<V1MetricsViewComparisonRequest>) {
  if (req.where) dashboard.whereFilter = req.where;

  if (req.timeRange) {
    dashboard.selectedTimeRange = getSelectedTimeRange(
      req.timeRange,
      timeRangeSummary,
      req.timeRange.isoDuration,
      executionTime,
    );
  }

  if (req.timeRange?.timeZone) {
    dashboard.selectedTimezone = req.timeRange?.timeZone || "UTC";
  }

  if (req.comparisonTimeRange) {
    if (
      !req.comparisonTimeRange.isoOffset &&
      req.comparisonTimeRange.isoDuration
    ) {
      dashboard.selectedComparisonTimeRange = {
        name: TimeComparisonOption.CONTIGUOUS,
        start: undefined,
        end: undefined,
      };
    } else {
      dashboard.selectedComparisonTimeRange = getSelectedTimeRange(
        req.comparisonTimeRange,
        timeRangeSummary,
        req.comparisonTimeRange.isoOffset,
        executionTime,
      );
    }

    if (dashboard.selectedComparisonTimeRange) {
      dashboard.selectedComparisonTimeRange.interval =
        dashboard.selectedTimeRange?.interval;
    }
    dashboard.showTimeComparison = true;
  }

  dashboard.visibleMeasureKeys = new Set(req.measures.map((m) => m.name));

  // if the selected sort is a measure set it to leaderboardMeasureName
  if (
    req.sort?.length &&
    metricsView.measures.findIndex((m) => m.name === req.sort[0].name) >= 0
  ) {
    dashboard.leaderboardMeasureName = req.sort[0].name;
    dashboard.sortDirection = req.sort[0].desc
      ? SortDirection.DESCENDING
      : SortDirection.ASCENDING;
    dashboard.dashboardSortType = getSortType(req.sort[0].sortType);
  }

  dashboard.selectedDimensionName = req.dimension.name;
  dashboard.activePage = DashboardState_ActivePage.DIMENSION_TABLE;

  return dashboard;
}
