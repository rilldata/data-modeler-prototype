import type { ChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import type {
  ChartConfig,
  ChartSortDirection,
} from "@rilldata/web-common/features/canvas/components/charts/types";
import { timeGrainToVegaTimeUnitMap } from "@rilldata/web-common/features/canvas/components/charts/util";
import type { ComponentFilterProperties } from "@rilldata/web-common/features/canvas/components/types";
import {
  validateDimensions,
  validateMeasures,
} from "@rilldata/web-common/features/canvas/components/validators";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  createQueryServiceMetricsViewAggregation,
  type MetricsViewSpecDimensionV2,
  type MetricsViewSpecMeasureV2,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationResponse,
  type V1MetricsViewAggregationResponseDataItem,
  type V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export type ChartDataResult = {
  data: V1MetricsViewAggregationResponseDataItem[];
  isFetching: boolean;
  fields: Record<
    string,
    | MetricsViewSpecMeasureV2
    | MetricsViewSpecDimensionV2
    | TimeDimensionDefinition
    | undefined
  >;
  error?: HTTPError | null;
};

export interface TimeDimensionDefinition {
  field: string;
  displayName: string;
  timeUnit?: string;
  format?: string;
}

export function getChartData(
  ctx: StateManagers,
  config: ChartConfig,
): Readable<ChartDataResult> {
  const chartDataQuery = createChartDataQuery(ctx, config);
  const { spec } = ctx.canvasEntity;

  const fields: { name: string; type: "measure" | "dimension" | "time" }[] = [];
  if (config.y?.field) fields.push({ name: config.y.field, type: "measure" });
  if (config.x?.field)
    fields.push({
      name: config.x.field,
      type: config.x.type === "temporal" ? "time" : "dimension",
    });
  if (typeof config.color === "object" && config.color?.field) {
    fields.push({ name: config.color.field, type: "dimension" });
  }

  // Match each field to its corresponding measure or dimension spec.
  const fieldReadableMap = fields.map((field) => {
    if (field.type === "measure") {
      return spec.getMeasureForMetricView(field.name, config.metrics_view);
    } else if (field.type === "dimension") {
      return spec.getDimensionForMetricView(field.name, config.metrics_view);
    } else {
      return getTimeDimensionDefinition(ctx, field.name);
    }
  });

  return derived(
    [chartDataQuery, ...fieldReadableMap],
    ([chartData, ...fieldMap]) => {
      const fieldSpecMap = fields.reduce(
        (acc, field, index) => {
          acc[field.name] = fieldMap?.[index];
          return acc;
        },
        {} as Record<
          string,
          | MetricsViewSpecMeasureV2
          | MetricsViewSpecDimensionV2
          | TimeDimensionDefinition
          | undefined
        >,
      );
      return {
        data: chartData?.data?.data || [],
        isFetching: chartData.isFetching,
        error: chartData.error,
        fields: fieldSpecMap,
      };
    },
  );
}

export function createChartDataQuery(
  ctx: StateManagers,
  config: ChartConfig & ComponentFilterProperties,
  limit = "5000",
  offset = "0",
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  const { canvasEntity } = ctx;

  const timeAndFilterStore = canvasEntity.createTimeAndFilterStore(
    config.metrics_view,
    {
      componentTimeRange: config.time_range,
      componentFilter: config.dimension_filters,
    },
  );

  let measures: V1MetricsViewAggregationMeasure[] = [];
  let dimensions: V1MetricsViewAggregationDimension[] = [];

  if (config.y?.type === "quantitative" && config.y?.field) {
    measures = [{ name: config.y?.field }];
  }

  let sort: V1MetricsViewAggregationSort | undefined;

  return derived(
    [ctx.runtime, timeAndFilterStore],
    ([runtime, $timeAndFilterStore], set) => {
      const { timeRange, where, timeGrain } = $timeAndFilterStore;

      if (config.x?.type === "nominal" && config.x?.field) {
        sort = vegaSortToAggregationSort(config.x?.sort, config);
        dimensions = [{ name: config.x?.field }];
      } else if (config.x?.type === "temporal" && timeGrain) {
        dimensions = [{ name: config.x?.field, timeGrain }];
      }

      if (typeof config.color === "object" && config.color?.field) {
        dimensions = [...dimensions, { name: config.color.field }];
      }

      return createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        config.metrics_view,
        {
          measures,
          dimensions,
          sort: sort ? [sort] : undefined,
          where,
          timeRange,
          limit,
          offset,
        },
        {
          query: {
            enabled:
              !!config.time_range || (!!timeRange?.start && !!timeRange?.end),
            queryClient: ctx.queryClient,
            keepPreviousData: true,
          },
        },
      ).subscribe(set);
    },
  );
}

export function getTimeDimensionDefinition(
  ctx: StateManagers,
  field: string,
): Readable<TimeDimensionDefinition> {
  const {
    timeControls: { selectedTimeRange },
  } = ctx.canvasEntity;
  return derived([selectedTimeRange], ([$selectedTimeRange]) => {
    const grain = $selectedTimeRange?.interval;
    const displayName = "Time";

    if (grain) {
      const timeUnit = timeGrainToVegaTimeUnitMap[grain];
      const format = TIME_GRAIN[grain]?.d3format as string;
      return {
        field,
        timeUnit,
        displayName,
        format,
      };
    }
    return {
      field,
      displayName,
    };
  });
}

export function validateChartSchema(
  ctx: StateManagers,
  chartSpec: ChartSpec,
): Readable<{
  isValid: boolean;
  error?: string;
}> {
  const { metrics_view, x, y, color } = chartSpec;
  let measures: string[] = [];
  let dimensions: string[] = [];

  if (y?.field) measures = [y.field];
  if (typeof color === "object" && color?.field)
    dimensions = [...dimensions, color.field];

  return derived(
    ctx.canvasEntity.spec.getMetricsViewFromName(metrics_view),
    (metricsView) => {
      if (!metricsView) {
        return {
          isValid: false,
          error: `Metrics view ${metrics_view} not found`,
        };
      }

      const timeDimension = metricsView.timeDimension;
      if (x?.field && x.field !== timeDimension) dimensions = [x.field];

      const validateMeasuresRes = validateMeasures(metricsView, measures);
      if (!validateMeasuresRes.isValid) {
        const invalidMeasures = validateMeasuresRes.invalidMeasures.join(", ");
        return {
          isValid: false,
          error: `Invalid measure ${invalidMeasures} selected`,
        };
      }

      const validateDimensionsRes = validateDimensions(metricsView, dimensions);

      if (!validateDimensionsRes.isValid) {
        const invalidDimensions =
          validateDimensionsRes.invalidDimensions.join(", ");

        return {
          isValid: false,
          error: `Invalid dimension(s) ${invalidDimensions} selected`,
        };
      }
      return {
        isValid: true,
        error: undefined,
      };
    },
  );
}

function vegaSortToAggregationSort(
  sort: ChartSortDirection | undefined,
  config: ChartConfig,
): V1MetricsViewAggregationSort | undefined {
  if (!sort) return undefined;
  const field =
    sort === "x" || sort === "-x" ? config.x?.field : config.y?.field;
  if (!field) return undefined;

  return {
    name: field,
    desc: sort === "-x" || sort === "-y",
  };
}
