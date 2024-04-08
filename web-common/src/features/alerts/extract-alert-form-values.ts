import { mapExpressionToAlertCriteria } from "@rilldata/web-common/features/alerts/criteria-tab/map-alert-criteria";
import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  V1AlertSpec,
  V1MetricsViewAggregationRequest,
  type V1MetricsViewComparisonRequest,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  V1Operation,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";

export type AlertFormValuesSubset = Pick<
  AlertFormValues,
  | "metricsViewName"
  | "whereFilter"
  | "timeRange"
  | "comparisonTimeRange"
  | "measure"
  | "splitByDimension"
  | "criteria"
  | "criteriaOperation"
>;

export function extractAlertFormValues(
  queryArgs: V1MetricsViewAggregationRequest,
  metricsViewSpec: V1MetricsViewSpec,
  allTimeRange: V1MetricsViewTimeRangeResponse,
): AlertFormValuesSubset {
  if (!queryArgs) return {} as AlertFormValuesSubset;

  const timeRange = (queryArgs.timeRange as V1TimeRange) ?? {
    isoDuration: metricsViewSpec.defaultTimeRange ?? TimeRangePreset.ALL_TIME,
  };
  if (!timeRange.end && allTimeRange.timeRangeSummary?.max) {
    timeRange.end = allTimeRange.timeRangeSummary?.max;
  }

  return {
    measure: queryArgs.measures?.[0]?.name ?? "",
    splitByDimension: queryArgs.dimensions?.[0]?.name ?? "",

    criteria:
      queryArgs.having?.cond?.exprs?.map(mapExpressionToAlertCriteria) ?? [],
    criteriaOperation: queryArgs.having?.cond?.op ?? V1Operation.OPERATION_AND,

    // These are not part of the form, but are used to track the state of the form
    metricsViewName: queryArgs.metricsView as string,
    whereFilter: queryArgs.where ?? createAndExpression([]),
    timeRange,
    comparisonTimeRange: undefined,
  };
}

export function extractAlertFormValueFromComparison(
  queryArgs: V1MetricsViewComparisonRequest,
  metricsViewSpec: V1MetricsViewSpec,
  allTimeRange: V1MetricsViewTimeRangeResponse,
) {
  if (!queryArgs) return {} as AlertFormValuesSubset;

  const timeRange = queryArgs.timeRange ?? {
    isoDuration: metricsViewSpec.defaultTimeRange ?? TimeRangePreset.ALL_TIME,
  };
  if (!timeRange.end && allTimeRange.timeRangeSummary?.max) {
    timeRange.end = allTimeRange.timeRangeSummary?.max;
  }

  return {
    measure: queryArgs.measures?.[0]?.name ?? "",
    splitByDimension: queryArgs.dimension?.name ?? "",

    criteria:
      queryArgs.having?.cond?.exprs?.map(mapExpressionToAlertCriteria) ?? [],
    criteriaOperation: queryArgs.having?.cond?.op ?? V1Operation.OPERATION_AND,

    // These are not part of the form, but are used to track the state of the form
    metricsViewName: queryArgs.metricsViewName as string,
    whereFilter: queryArgs.where ?? createAndExpression([]),
    timeRange,
    comparisonTimeRange: queryArgs.comparisonTimeRange,
  };
}

export type AlertNotificationValues = Pick<
  AlertFormValues,
  | "enableSlackNotification"
  | "slackChannels"
  | "slackUsers"
  | "enableEmailNotification"
  | "emailRecipients"
>;

export function extractAlertNotification(
  alertSpec: V1AlertSpec,
): AlertNotificationValues {
  const slackNotifier = alertSpec.notifiers?.find(
    (n) => n.connector === "slack",
  );
  const slackChannels: string[] | undefined =
    slackNotifier?.properties?.channels;
  const slackUsers: string[] | undefined = slackNotifier?.properties?.channels;

  const emailNotifier = alertSpec.notifiers?.find(
    (n) => n.connector === "email",
  );
  const emailRecipients: string[] | undefined =
    emailNotifier?.properties?.recipients;

  return {
    enableSlackNotification: !!slackNotifier,
    slackChannels: mapAndAddEmptyEntry(slackChannels, "channel"),
    slackUsers: mapAndAddEmptyEntry(slackUsers, "email"),

    enableEmailNotification: !!emailNotifier,
    emailRecipients: mapAndAddEmptyEntry(emailRecipients, "email"),
  };
}

function mapAndAddEmptyEntry<R>(entries: string[] | undefined, key: string): R {
  const mappedEntries = entries?.map((e) => ({ [key]: e })) ?? [];
  mappedEntries.push({ [key]: "" });
  return mappedEntries as R;
}
