import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  V1MetricsViewAggregationDimension,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationRequest,
  type V1MetricsViewSpec,
  V1Operation,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";

export type AlertFormValuesSubset = Pick<
  AlertFormValues,
  | "metricsViewName"
  | "whereFilter"
  | "timeRange"
  | "measure"
  | "splitByDimension"
  | "criteria"
  | "criteriaOperation"
>;

export function extractAlertFormValues(
  queryArgs: V1MetricsViewAggregationRequest,
  metricsViewSpec: V1MetricsViewSpec,
): AlertFormValuesSubset {
  if (!queryArgs) return {} as AlertFormValuesSubset;

  const measures = queryArgs.measures as V1MetricsViewAggregationMeasure[];
  const dimensions =
    queryArgs.dimensions as V1MetricsViewAggregationDimension[];

  return {
    measure: measures[0]?.name ?? "",
    splitByDimension: dimensions[0]?.name ?? "",

    criteria:
      queryArgs.having?.cond?.exprs?.map((e) => ({
        field: e.cond?.exprs?.[0]?.ident as string,
        operation: e.cond?.op as string,
        value: String(e.cond?.exprs?.[1]?.val),
      })) ?? [],
    criteriaOperation: queryArgs.having?.cond?.op ?? V1Operation.OPERATION_AND,

    // These are not part of the form, but are used to track the state of the form
    metricsViewName: queryArgs.metricsView as string,
    whereFilter: queryArgs.where ?? createAndExpression([]),
    timeRange: (queryArgs.timeRange as V1TimeRange) ?? {
      isoDuration: metricsViewSpec.defaultTimeRange ?? TimeRangePreset.ALL_TIME,
    },
  };
}
