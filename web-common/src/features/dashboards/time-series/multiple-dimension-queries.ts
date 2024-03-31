import { measureFilterResolutionsStore } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { selectedDimensionValues } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import {
  createAndExpression,
  createInExpression,
  filterExpressions,
  matchExpressionByName,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { Readable, derived, writable } from "svelte/store";

import { getDimensionFilterWithSearch } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-utils";
import {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { TimeSeriesDatum } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  V1Expression,
  createQueryServiceMetricsViewAggregation,
  createQueryServiceMetricsViewTimeSeries,
} from "@rilldata/web-common/runtime-client";
import { getFilterForComparedDimension, prepareTimeSeries } from "./utils";

export interface DimensionDataItem {
  dimension: string;
  value: string | null;
  total?: number;
  data: TimeSeriesDatum[];
  isFetching: boolean;
}

/***
 * Returns a list of dimension values which for which to fetch
 * timeseries data for a given dimension.
 *
 * For Overview Page -
 * Use the included values if present,
 * otherwise fetch the top values for the dimension
 *
 * For Time Dimension Detail Page -
 * Fetch all the top n values for the dimension
 * and further filter using search text if present
 */

export function getDimensionValuesForComparison(
  ctx: StateManagers,
  measures: string[],
  surface: "chart" | "table",
): Readable<{
  values: string[];
  filter: V1Expression;
  totals?: number[];
}> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      measureFilterResolutionsStore(ctx),
    ],
    (
      [runtime, name, dashboardStore, timeControls, measureFilterResolution],
      set,
    ) => {
      const isValidMeasureList =
        measures?.length > 0 && measures?.every((m) => m !== undefined);

      if (!isValidMeasureList) return;

      const dimensionName = dashboardStore?.selectedComparisonDimension;
      const isInTimeDimensionView = dashboardStore?.expandedMeasureName;

      // Values to be compared
      let comparisonValues: string[] = [];
      if (surface === "chart") {
        let dimensionValues = selectedDimensionValues({
          dashboard: dashboardStore,
        })(dimensionName);
        if (measureFilterResolution.filter) {
          // if there is a measure filter for this dimension. remove values not in that filter
          const dimVals = measureFilterResolution.filter.cond?.exprs?.find(
            (e) => matchExpressionByName(e, dimensionName),
          )?.cond?.exprs;
          if (dimVals?.length) {
            dimensionValues = dimensionValues.filter(
              (d) => dimVals.findIndex((dimVal) => dimVal.val === d) >= 0,
            );
          }
        }

        if (dimensionValues?.length) {
          // For TDD view max 11 allowed, for overview max 7 allowed
          comparisonValues = dimensionValues.slice(
            0,
            isInTimeDimensionView ? 11 : 7,
          );
        }
        return derived(
          [writable(comparisonValues), writable(dashboardStore?.whereFilter)],
          ([values, filter]) => {
            return {
              values,
              filter,
            };
          },
        ).subscribe(set);
      } else if (surface === "table") {
        let sortBy = isInTimeDimensionView
          ? dashboardStore.expandedMeasureName
          : dashboardStore.leaderboardMeasureName;
        if (dashboardStore?.dashboardSortType === SortType.DIMENSION) {
          sortBy = dimensionName;
        }

        return derived(
          createQueryServiceMetricsViewAggregation(
            runtime.instanceId,
            name,
            {
              measures: measures.map((measure) => ({ name: measure })),
              dimensions: [{ name: dimensionName }],
              where: sanitiseExpression(
                getDimensionFilterWithSearch(
                  dashboardStore?.whereFilter,
                  dashboardStore?.dimensionSearchText ?? "",
                  dimensionName,
                ),
                measureFilterResolution.filter,
              ),
              timeStart: timeControls.timeStart,
              timeEnd: timeControls.timeEnd,
              sort: [
                {
                  desc:
                    dashboardStore.sortDirection === SortDirection.DESCENDING,
                  name: sortBy,
                },
              ],
              limit: "250",
              offset: "0",
            },
            {
              query: {
                enabled:
                  timeControls.ready &&
                  !!dashboardStore?.selectedComparisonDimension &&
                  measureFilterResolution.ready,
                queryClient: ctx.queryClient,
              },
            },
          ),
          (topListData) => {
            if (topListData?.isFetching || !dimensionName)
              return {
                values: [],
                filter: dashboardStore?.whereFilter,
              };
            const columnName =
              topListData?.data?.schema?.fields?.[0]?.name || dimensionName;
            const totalValues = topListData?.data?.data?.map(
              (d) => d[measures[0]],
            ) as number[];
            const topListValues = topListData?.data?.data?.map(
              (d) => d[columnName],
            ) as string[];

            const computedFilter = getFilterForComparedDimension(
              dimensionName,
              dashboardStore?.whereFilter,
              topListValues,
            );

            return {
              totals: totalValues,
              values: computedFilter?.includedValues,
              filter: computedFilter?.updatedFilter,
            };
          },
        ).subscribe(set);
      }
    },
  );
}

/***
 * Fetches the timeseries data for a given dimension
 * for a infered set of dimension values and measures
 */
export function getDimensionValueTimeSeries(
  ctx: StateManagers,
  measures: string[],
  surface: "chart" | "table",
): Readable<DimensionDataItem[]> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      getDimensionValuesForComparison(ctx, measures, surface),
    ],
    (
      [runtime, metricViewName, dashboardStore, timeStore, dimensionValues],
      set,
    ) => {
      const dimensionName = dashboardStore?.selectedComparisonDimension;

      const start = timeStore?.adjustedStart;
      const end = timeStore?.adjustedEnd;
      const interval =
        timeStore?.selectedTimeRange?.interval ?? timeStore?.minTimeGrain;
      const zone = dashboardStore.selectedTimezone;

      const isValidMeasureList =
        measures?.length > 0 && measures?.every((m) => m !== undefined);

      if (!isValidMeasureList || !dimensionName) return;
      if (dashboardStore?.selectedScrubRange?.isScrubbing) return;

      return derived(
        (dimensionValues?.values ?? [])?.map((value, i) => {
          // create a copy
          const updatedFilter =
            filterExpressions(dimensionValues?.filter, () => true) ??
            createAndExpression([]);
          // add the value to "in" expression
          updatedFilter.cond?.exprs?.push(
            createInExpression(dimensionName, [value]),
          );

          return derived(
            [
              writable(value),
              createQueryServiceMetricsViewTimeSeries(
                runtime.instanceId,
                metricViewName,
                {
                  measureNames: measures,
                  where: sanitiseExpression(updatedFilter, undefined),
                  timeStart: start,
                  timeEnd: end,
                  timeGranularity: interval,
                  timeZone: zone,
                },
                {
                  query: {
                    enabled: !!timeStore.ready && !!ctx.dashboardStore,
                    queryClient: ctx.queryClient,
                  },
                },
              ),
            ],
            ([value, timeseries]) => {
              let prepData: TimeSeriesDatum[] = [];
              if (!timeseries?.isFetching && interval) {
                prepData = prepareTimeSeries(
                  timeseries?.data?.data || [],
                  undefined,
                  TIME_GRAIN[interval]?.duration,
                  zone,
                );
              }

              let total;
              if (surface === "table") {
                total = dimensionValues?.totals[i];
              }

              return {
                dimension: dimensionName,
                value,
                total,
                data: prepData,
                isFetching: timeseries.isFetching,
              };
            },
          );
        }),

        (combos) => {
          return combos;
        },
      ).subscribe(set);
    },
  );
}
