import { getComparisonRequestMeasures } from "@rilldata/web-common/features/dashboards/dashboard-utils";
import { getDimensionFilterWithSearch } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-utils";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  mapComparisonTimeRange,
  mapTimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationRequest,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, get, Readable } from "svelte/store";

export function getDimensionTableExportArgs(
  ctx: StateManagers,
): Readable<V1MetricsViewAggregationRequest | undefined> {
  return derived(
    [
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      useMetricsView(ctx),
    ],
    ([metricViewName, dashboardState, timeControlState, metricsView]) => {
      if (!metricsView.data || !timeControlState.ready) return undefined;

      const timeRange = mapTimeRange(
        timeControlState,
        metricsView.data.defaultTimeRange,
      );
      if (!timeRange) return undefined;

      const comparisonTimeRange = mapComparisonTimeRange(
        dashboardState,
        timeControlState,
        timeRange,
      );

      return getDimensionTableAggregationRequestForTime(
        metricViewName,
        dashboardState,
        timeRange,
        comparisonTimeRange,
      );
    },
  );
}

export function getDimensionTableAggregationRequestForTime(
  metricsView: string,
  dashboardState: MetricsExplorerEntity,
  timeRange: V1TimeRange,
  comparisonTimeRange: V1TimeRange | undefined,
): V1MetricsViewAggregationRequest {
  const measures: V1MetricsViewAggregationMeasure[] = [
    ...dashboardState.visibleMeasureKeys,
  ].map((name) => ({
    name: name,
  }));

  let apiSortName = dashboardState.leaderboardMeasureName;
  if (!dashboardState.visibleMeasureKeys.has(apiSortName)) {
    // if selected sort measure is not visible add it to list
    measures.push({ name: apiSortName });
  }
  if (comparisonTimeRange) {
    // insert beside the correct measure
    measures.splice(
      measures.findIndex((m) => m.name === apiSortName) + 1,
      0,
      ...getComparisonRequestMeasures(apiSortName),
    );
    switch (dashboardState.dashboardSortType) {
      case DashboardState_LeaderboardSortType.DELTA_ABSOLUTE:
        apiSortName += ComparisonDeltaAbsoluteSuffix;
        break;
      case DashboardState_LeaderboardSortType.DELTA_PERCENT:
        apiSortName += ComparisonDeltaRelativeSuffix;
        break;
    }
  }

  const where = sanitiseExpression(
    mergeMeasureFilters(
      dashboardState,
      getDimensionFilterWithSearch(
        dashboardState?.whereFilter,
        dashboardState?.dimensionSearchText ?? "",
        dashboardState.selectedDimensionName!,
      ),
    ),
    undefined,
  );

  return {
    instanceId: get(runtime).instanceId,
    metricsView,
    dimensions: [
      {
        name: dashboardState.selectedDimensionName,
      },
    ],
    measures,
    timeRange,
    ...(comparisonTimeRange ? { comparisonTimeRange } : {}),
    sort: [
      {
        name: apiSortName,
        desc: dashboardState.sortDirection === SortDirection.DESCENDING,
      },
    ],
    where,
    offset: "0",
  };
}
