import type { QueryMapperArgs } from "@rilldata/web-admin/features/dashboards/query-mappers/types";
import {
  convertExprToToplist,
  fillTimeRange,
} from "@rilldata/web-admin/features/dashboards/query-mappers/utils";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
  ComparisonPercentOfTotal,
  mapExprToMeasureFilter,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import {
  createAndExpression,
  forEachIdentifier,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  getQueryServiceMetricsViewSchemaQueryKey,
  queryServiceMetricsViewSchema,
  type V1Expression,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

export async function getDashboardFromAggregationRequest({
  queryClient,
  instanceId,
  req,
  dashboard,
  timeRangeSummary,
  executionTime,
  metricsView,
  annotations,
}: QueryMapperArgs<V1MetricsViewAggregationRequest>) {
  let loadedFromState = false;
  if (annotations["web_open_state"]) {
    await mergeDashboardFromUrlState(
      queryClient,
      instanceId,
      dashboard,
      metricsView,
      annotations["web_open_state"],
    );
    loadedFromState = true;
  }

  fillTimeRange(
    dashboard,
    req.timeRange,
    req.comparisonTimeRange,
    timeRangeSummary,
    executionTime,
  );

  if (req.where) {
    const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
      req.where,
    );
    dashboard.whereFilter = dimensionFilters;
    dashboard.dimensionThresholdFilters = dimensionThresholdFilters;
  }
  if (req.having?.cond?.exprs?.length && req.dimensions?.[0]?.name) {
    const dimension = req.dimensions[0].name;
    if (
      // we only support 1-1 map from measure filter to dimension
      req.having.cond.exprs.length > 1 ||
      // we do not support comparison based measure filter
      exprHasComparison(req.having) ||
      // there is already measure filter from dashboard during alert creation
      dashboard.dimensionThresholdFilters.length > 0
    ) {
      const expr = await convertExprToToplist(
        queryClient,
        instanceId,
        dashboard.name,
        dimension,
        req.measures?.[0]?.name ?? "",
        req.timeRange,
        req.comparisonTimeRange,
        executionTime,
        req.where,
        req.having,
      );
      if (expr) {
        dashboard.whereFilter =
          mergeFilters(
            dashboard.whereFilter ?? createAndExpression([]),
            createAndExpression([expr]),
          ) ?? createAndExpression([]);
      }
    } else {
      dashboard.dimensionThresholdFilters = [
        {
          name: dimension,
          filters:
            req.having.cond?.exprs
              ?.map(mapExprToMeasureFilter)
              .filter(Boolean) ?? [],
        },
      ];
    }
  }

  // everything after this can be loaded from the dashboard state if present
  if (loadedFromState) return dashboard;

  if (req.timeRange?.timeZone) {
    dashboard.selectedTimezone = req.timeRange?.timeZone || "UTC";
  }

  dashboard.visibleMeasureKeys = new Set(
    req.measures?.map((m) => m.name ?? "") ?? [],
  );

  // if the selected sort is a measure set it to leaderboardMeasureName
  if (
    req.sort?.[0] &&
    (metricsView.measures?.findIndex((m) => m.name === req.sort?.[0]?.name) ??
      -1) >= 0
  ) {
    dashboard.leaderboardMeasureName = req.sort[0].name ?? "";
    dashboard.sortDirection = req.sort[0].desc
      ? SortDirection.DESCENDING
      : SortDirection.ASCENDING;
    dashboard.dashboardSortType = SortType.VALUE;
  }

  if (req.dimensions?.length) {
    dashboard.selectedDimensionName = req.dimensions[0].name;
    dashboard.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
  } else {
    dashboard.tdd = {
      chartType: TDDChart.DEFAULT,
      expandedMeasureName: req.measures?.[0]?.name ?? "",
      pinIndex: -1,
    };
    dashboard.activePage = DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL;
  }

  return dashboard;
}

function exprHasComparison(expr: V1Expression) {
  let hasComparison = false;
  forEachIdentifier(expr, (e, ident) => {
    if (
      ident.endsWith(ComparisonDeltaAbsoluteSuffix) ||
      ident.endsWith(ComparisonDeltaRelativeSuffix) ||
      ident.endsWith(ComparisonPercentOfTotal)
    ) {
      hasComparison = true;
    }
  });
  return hasComparison;
}

async function mergeDashboardFromUrlState(
  queryClient: QueryClient,
  instanceId: string,
  dashboard: MetricsExplorerEntity,
  metricsViewSpec: V1MetricsViewSpec,
  urlState: string,
) {
  const schemaResp = await queryClient.fetchQuery({
    queryKey: getQueryServiceMetricsViewSchemaQueryKey(
      instanceId,
      dashboard.name,
    ),
    queryFn: () => queryServiceMetricsViewSchema(instanceId, dashboard.name),
  });
  if (!schemaResp.schema) return;

  const parsedDashboard = getDashboardStateFromUrl(
    urlState,
    metricsViewSpec,
    schemaResp.schema,
  );
  for (const k in parsedDashboard) {
    dashboard[k] = parsedDashboard[k];
  }
}
