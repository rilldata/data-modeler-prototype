import { getComparisonRequestMeasures } from "@rilldata/web-common/features/dashboards/dashboard-utils";
import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  TimeControlState,
  useTimeControlStore,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  mapComparisonTimeRange,
  mapTimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import {
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationRequest,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, get, Readable } from "svelte/store";

export function getTDDExportArgs(
  ctx: StateManagers,
): Readable<V1MetricsViewAggregationRequest | undefined> {
  return derived(
    [
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      useMetricsView(ctx),
    ],
    ([metricsViewName, dashboardState, timeControlState, metricsView]) =>
      getTDDAggregationRequest(
        metricsViewName,
        dashboardState,
        timeControlState,
        metricsView.data,
      ),
  );
}

export function getTDDAggregationRequest(
  metricsViewName: string,
  dashboardState: MetricsExplorerEntity,
  timeControlState: TimeControlState,
  metricsView: V1MetricsViewSpec | undefined,
): undefined | V1MetricsViewAggregationRequest {
  if (
    !metricsView ||
    !timeControlState.ready ||
    !dashboardState.tdd.expandedMeasureName
  )
    return undefined;

  const timeRange = mapTimeRange(
    timeControlState,
    metricsView.defaultTimeRange,
  );
  if (!timeRange) return undefined;

  const comparisonTimeRange = mapComparisonTimeRange(
    dashboardState,
    timeControlState,
    timeRange,
  );

  const measures: V1MetricsViewAggregationMeasure[] = [
    { name: dashboardState.tdd.expandedMeasureName },
  ];
  if (
    !!comparisonTimeRange?.start &&
    !!comparisonTimeRange?.end &&
    !!timeControlState.selectedComparisonTimeRange
  ) {
    measures.push(
      ...getComparisonRequestMeasures(dashboardState.tdd.expandedMeasureName),
    );
  }

  // CAST SAFETY: exports are only available in TDD when a comparison dimension is selected
  const dimensionName = dashboardState.selectedComparisonDimension as string;
  const timeDimension = metricsView.timeDimension ?? "";

  return {
    instanceId: get(runtime).instanceId,
    metricsView: metricsViewName,
    dimensions: [
      { name: dimensionName },
      {
        name: metricsView.timeDimension ?? "",
        timeGrain: dashboardState.selectedTimeRange?.interval,
        timeZone: dashboardState.selectedTimezone,
      },
    ],
    measures,
    timeRange,
    ...(comparisonTimeRange ? { comparisonTimeRange } : {}),
    pivotOn: [timeDimension],
    sort: [
      {
        name: dimensionName,
        desc: dashboardState.sortDirection === SortDirection.DESCENDING,
      },
    ],
    where: sanitiseExpression(mergeMeasureFilters(dashboardState), undefined),
    offset: "0",
  };
}
