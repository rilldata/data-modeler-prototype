import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
import {
  getFilteredMeasuresAndDimensions,
  getIndependentMeasures,
} from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  createTotalsForMeasure,
  createUnfilteredTotalsForMeasure,
} from "@rilldata/web-common/features/dashboards/time-series/totals-data-store";
import { prepareTimeSeries } from "@rilldata/web-common/features/dashboards/time-series/utils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { Period } from "@rilldata/web-common/lib/time/types";
import {
  V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationResponseDataItem,
  V1MetricsViewTimeSeriesResponse,
  createQueryServiceMetricsViewTimeSeries,
} from "@rilldata/web-common/runtime-client";
import { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { Writable, derived, writable, type Readable } from "svelte/store";
import { memoizeMetricsStore } from "../state-managers/memoize-metrics-store";
import {
  DimensionDataItem,
  getDimensionValueTimeSeries,
} from "./multiple-dimension-queries";

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
    ([runtime, metricViewName, dashboardStore, timeControls], set) =>
      createQueryServiceMetricsViewTimeSeries(
        runtime.instanceId,
        metricViewName,
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

export function createTimeSeriesDataStore(
  ctx: StateManagers,
): TimeSeriesDataStore {
  return derived(
    [useMetricsView(ctx), useTimeControlStore(ctx), ctx.dashboardStore],
    ([metricsView, timeControls, dashboardStore], set) => {
      const validSpec = metricsView.data;
      if (!timeControls.ready || timeControls.isFetching || !validSpec) {
        set({
          isFetching: true,
          isError: false,
          error: {},
        });
        return;
      }

      const showComparison = timeControls.showTimeComparison;
      const interval =
        timeControls.selectedTimeRange?.interval ?? timeControls.minTimeGrain;

      const allMeasures =
        metricsView.data?.measures?.map((measure) => measure.name as string) ||
        [];
      let measures = allMeasures;
      const expandedMeasuerName = dashboardStore?.tdd?.expandedMeasureName;
      if (expandedMeasuerName) {
        measures = allMeasures.filter(
          (measure) => measure === expandedMeasuerName,
        );
      } else {
        measures = dashboardStore?.visibleMeasureKeys
          ? [...dashboardStore.visibleMeasureKeys]
          : [];
      }

      const { measures: filteredMeasures } = getFilteredMeasuresAndDimensions({
        dashboard: dashboardStore,
      })(validSpec, measures);
      const independentMeasures = getIndependentMeasures(validSpec, measures);

      const primaryTimeSeries =
        measures.length > 0
          ? createMetricsViewTimeSeries(ctx, filteredMeasures, false)
          : writable({
              isFetching: false,
              isError: false,
              data: null,
              error: {},
            });

      const primaryTotals =
        measures.length > 0
          ? createTotalsForMeasure(ctx, independentMeasures, false)
          : writable({
              isFetching: false,
              isError: false,
              data: null,
              error: {},
            });

      let unfilteredTotals:
        | CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError>
        | Writable<null> = writable(null);

      if (dashboardStore?.selectedComparisonDimension) {
        unfilteredTotals = createUnfilteredTotalsForMeasure(
          ctx,
          independentMeasures,
          dashboardStore?.selectedComparisonDimension,
        );
      }
      let comparisonTimeSeries:
        | CreateQueryResult<V1MetricsViewTimeSeriesResponse, HTTPError>
        | Writable<null> = writable(null);
      let comparisonTotals:
        | CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError>
        | Writable<null> = writable(null);
      if (showComparison) {
        comparisonTimeSeries = createMetricsViewTimeSeries(
          ctx,
          filteredMeasures,
          true,
        );
        comparisonTotals = createTotalsForMeasure(
          ctx,
          independentMeasures,
          true,
        );
      }

      let dimensionTimeSeriesCharts:
        | Readable<DimensionDataItem[]>
        | Writable<null> = writable(null);
      if (dashboardStore?.selectedComparisonDimension) {
        dimensionTimeSeriesCharts = getDimensionValueTimeSeries(
          ctx,
          filteredMeasures,
          "chart",
        );
      }

      return derived(
        [
          primaryTimeSeries,
          comparisonTimeSeries,
          primaryTotals,
          unfilteredTotals,
          comparisonTotals,
          dimensionTimeSeriesCharts,
        ],
        ([
          primary,
          comparison,
          primaryTotal,
          unfilteredTotal,
          comparisonTotal,
          dimensionChart,
        ]) => {
          let preparedTimeSeriesData: TimeSeriesDatum[] = [];

          if (!primary.isFetching && interval) {
            const intervalDuration = TIME_GRAIN[interval]?.duration as Period;
            preparedTimeSeriesData = prepareTimeSeries(
              primary?.data?.data || [],
              comparison?.data?.data || [],
              intervalDuration,
              dashboardStore.selectedTimezone,
            );
          }

          let isError = false;

          const error = {};
          if (primary.error) {
            isError = true;
            error["timeseries"] = (
              primary.error as HTTPError
            ).response?.data?.message;
          }
          if (primaryTotal.error) {
            isError = true;
            error["totals"] = (
              primaryTotal.error as HTTPError
            ).response?.data?.message;
          }

          return {
            isFetching: primary?.isFetching || primaryTotal?.isFetching,
            isError,
            error,
            timeSeriesData: preparedTimeSeriesData,
            total: primaryTotal?.data?.data?.[0],
            unfilteredTotal: unfilteredTotal?.data?.data?.[0],
            comparisonTotal: comparisonTotal?.data?.data?.[0],
            dimensionChartData: (dimensionChart as DimensionDataItem[]) || [],
          };
        },
      ).subscribe(set);
    },
  );
}

/**
 * Memoized version of the store. Currently, memoized by metrics view name.
 */
export const useTimeSeriesDataStore = memoizeMetricsStore<TimeSeriesDataStore>(
  (ctx: StateManagers) => createTimeSeriesDataStore(ctx),
);
