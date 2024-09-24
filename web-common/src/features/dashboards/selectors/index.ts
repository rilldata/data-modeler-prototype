import {
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  RpcStatus,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
  createQueryServiceMetricsViewComparison,
  createQueryServiceMetricsViewSchema,
  createQueryServiceMetricsViewTimeRange,
  type V1MetricsViewSchemaResponse,
} from "@rilldata/web-common/runtime-client";
import type {
  CreateQueryResult,
  QueryObserverResult,
} from "@tanstack/svelte-query";
import { Readable, derived } from "svelte/store";
import type { StateManagers } from "../state-managers/state-managers";

export const useMetricsView = <T = V1MetricsViewSpec>(
  ctx: StateManagers,
  selector?: (meta: V1MetricsViewSpec) => T,
): Readable<QueryObserverResult<T | V1MetricsViewSpec, RpcStatus>> => {
  return derived(
    [ctx.runtime, ctx.metricsViewName],
    ([runtime, metricViewName], set) => {
      return useResource(
        runtime.instanceId,
        metricViewName,
        ResourceKind.MetricsView,
        {
          select: (data) =>
            selector
              ? selector(data.resource?.metricsView?.state?.validSpec)
              : data.resource?.metricsView?.state?.validSpec,
          queryClient: ctx.queryClient,
        },
      ).subscribe(set);
    },
  );
};

export const useModelHasTimeSeries = (ctx: StateManagers) =>
  useMetricsView(
    ctx,
    (meta) => !!meta?.timeDimension,
  ) as CreateQueryResult<boolean>;

export function getFilterSearchListQuery(
  searchText: string,
  props: {
    instanceId: string;
    metricsViewName: string;
    dimensionName: string;
    leaderboardMeasureName: string;
    timeStart: string | undefined;
    timeEnd: string | undefined;
    addNull: boolean;
    enabled: boolean;
  },
) {
  return createQueryServiceMetricsViewComparison(
    props.instanceId,
    props.metricsViewName,
    {
      dimension: { name: props.dimensionName },
      measures: [{ name: props.leaderboardMeasureName }],
      timeRange: {
        start: props.timeStart,
        end: props.timeEnd,
      },
      limit: "100",
      offset: "0",
      sort: [{ name: props.dimensionName }],
      where: props.addNull
        ? createInExpression(props.dimensionName, [null])
        : createLikeExpression(props.dimensionName, `%${searchText}%`),
    },
    {
      query: {
        enabled: props.enabled,
        keepPreviousData: true,
      },
    },
  );
}

export function createTimeRangeSummary(
  ctx: StateManagers,
): CreateQueryResult<V1MetricsViewTimeRangeResponse> {
  return derived(
    [ctx.runtime, ctx.metricsViewName, useMetricsView(ctx)],
    ([runtime, metricsViewName, metricsView], set) =>
      createQueryServiceMetricsViewTimeRange(
        runtime.instanceId,
        metricsViewName,
        {},
        {
          query: {
            queryClient: ctx.queryClient,
            enabled: !metricsView.error && !!metricsView.data?.timeDimension,
          },
        },
      ).subscribe(set),
  );
}

export function createMetricsViewSchema(
  ctx: StateManagers,
): CreateQueryResult<V1MetricsViewSchemaResponse> {
  return derived(
    [ctx.runtime, ctx.metricsViewName],
    ([runtime, metricsViewName], set) =>
      createQueryServiceMetricsViewSchema(
        runtime.instanceId,
        metricsViewName,
        {},
        {
          query: {
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set),
  );
}
