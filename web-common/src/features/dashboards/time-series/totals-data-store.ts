import type { QueryObserverResult } from "@rilldata/svelte-query";
import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  createAndExpression,
  filterExpressions,
  matchExpressionByName,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { timeControlStateSelector } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  createQueryServiceMetricsViewAggregation,
  type V1MetricsViewAggregationResponse,
} from "@rilldata/web-common/runtime-client";
import { type HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function createTotalsForMeasure(
  ctx: StateManagers,
  measures: string[],
  isComparison = false,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.validSpecStore,
      ctx.timeRangeSummaryStore,
      ctx.dashboardStore,
    ],
    (
      [runtime, metricsViewName, validSpec, timeRangeSummary, dashboard],
      set,
    ) => {
      if (
        !validSpec?.data?.metricsView ||
        !validSpec?.data?.explore ||
        timeRangeSummary.isFetching ||
        !dashboard
      ) {
        set({
          isFetching: true,
          isError: false,
        } as QueryObserverResult<V1MetricsViewAggregationResponse, HTTPError>);
        return;
      }

      const { metricsView, explore } = validSpec.data;
      // This indirection makes sure only one update of dashboard store triggers this
      const timeControls = timeControlStateSelector([
        metricsView,
        explore,
        timeRangeSummary,
        dashboard,
      ]);

      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricsViewName,
        {
          measures: measures.map((measure) => ({ name: measure })),
          where: sanitiseExpression(mergeMeasureFilters(dashboard), undefined),
          timeRange: {
            start: isComparison
              ? timeControls?.comparisonTimeStart
              : timeControls.timeStart,
            end: isComparison
              ? timeControls?.comparisonTimeEnd
              : timeControls.timeEnd,
          },
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set);
    },
  );
}

export function createUnfilteredTotalsForMeasure(
  ctx: StateManagers,
  measures: string[],
  dimensionName: string,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.validSpecStore,
      ctx.timeRangeSummaryStore,
      ctx.dashboardStore,
    ],
    (
      [runtime, metricsViewName, validSpec, timeRangeSummary, dashboard],
      set,
    ) => {
      if (
        !validSpec?.data?.metricsView ||
        !validSpec?.data?.explore ||
        timeRangeSummary.isFetching ||
        !dashboard
      ) {
        set({
          isFetching: true,
          isError: false,
        } as QueryObserverResult<V1MetricsViewAggregationResponse, HTTPError>);
        return;
      }

      const { metricsView, explore } = validSpec.data;
      // This indirection makes sure only one update of dashboard store triggers this
      const timeControls = timeControlStateSelector([
        metricsView,
        explore,
        timeRangeSummary,
        dashboard,
      ]);

      const filter = sanitiseExpression(
        mergeMeasureFilters(dashboard),
        undefined,
      );

      const updatedFilter = filterExpressions(
        filter || createAndExpression([]),
        (e) => !matchExpressionByName(e, dimensionName),
      );

      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricsViewName,
        {
          measures: measures.map((measure) => ({ name: measure })),
          where: updatedFilter,
          timeStart: timeControls.timeStart,
          timeEnd: timeControls.timeEnd,
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set);
    },
  );
}
