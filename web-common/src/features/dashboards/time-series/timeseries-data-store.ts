import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  type V1MetricsViewAggregationResponseDataItem,
  type V1MetricsViewTimeSeriesResponse,
  createQueryServiceMetricsViewTimeSeries,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

import { type DimensionDataItem } from "./multiple-dimension-queries";

export interface TimeSeriesDatum {
  ts?: Date;
  bin?: number;
  ts_comparison?: Date;
  ts_position?: Date;
  [key: string]: Date | string | number | undefined;
}

export type TimeSeriesDataState = {
  isFetching: boolean;
  isError: boolean;
  error: { [key: string]: string | undefined };

  // Computed prepared data for charts and table
  timeSeriesData?: TimeSeriesDatum[];
  total?: V1MetricsViewAggregationResponseDataItem;
  unfilteredTotal?: V1MetricsViewAggregationResponseDataItem;
  comparisonTotal?: V1MetricsViewAggregationResponseDataItem;
  dimensionChartData?: DimensionDataItem[];
};

export type TimeSeriesDataStore = Readable<TimeSeriesDataState>;

export function createMetricsViewTimeSeries(
  ctx: StateManagers,
  measures: string[],
  isComparison = false,
): CreateQueryResult<V1MetricsViewTimeSeriesResponse, HTTPError> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
    ],
    ([runtime, metricsViewName, dashboardStore, timeControls], set) =>
      createQueryServiceMetricsViewTimeSeries(
        runtime.instanceId,
        metricsViewName,
        {
          measureNames: measures,
          where: sanitiseExpression(
            mergeMeasureFilters(dashboardStore),
            undefined,
          ),
          timeStart: isComparison
            ? timeControls.comparisonAdjustedStart
            : timeControls.adjustedStart,
          timeEnd: isComparison
            ? timeControls.comparisonAdjustedEnd
            : timeControls.adjustedEnd,
          timeGranularity:
            timeControls.selectedTimeRange?.interval ??
            timeControls.minTimeGrain,
          timeZone: dashboardStore.selectedTimezone,
        },
        {
          query: {
            enabled:
              !!timeControls.ready &&
              !!ctx.dashboardStore &&
              // in case of comparison, we need to wait for the comparison start time to be available
              (!isComparison || !!timeControls.comparisonAdjustedStart),
            queryClient: ctx.queryClient,
            keepPreviousData: true,
          },
        },
      ).subscribe(set),
  );
}
