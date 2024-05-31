import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
  mapMeasureFilterToExpr,
  MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  V1Expression,
  V1MetricsViewAggregationRequest,
  V1Operation,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import * as yup from "yup";

export type AlertFormValues = {
  name: string;
  measure: string;
  splitByDimension: string;
  criteria: MeasureFilterEntry[];
  criteriaOperation: V1Operation;
  evaluationInterval: string;
  snooze: string;
  enableSlackNotification: boolean;
  slackChannels: { channel: string }[];
  slackUsers: { email: string }[];
  enableEmailNotification: boolean;
  emailRecipients: { email: string }[];
  // The following fields are not editable in the form, but they're state that's used throughout the form, so
  // it's helpful to have them here. Also, in the future they may be editable in the form.
  metricsViewName: string;
  whereFilter: V1Expression;
  timeRange: V1TimeRange;
  comparisonTimeRange: V1TimeRange | undefined;
};

export function getAlertQueryArgsFromFormValues(
  formValues: AlertFormValues,
): V1MetricsViewAggregationRequest {
  return {
    metricsView: formValues.metricsViewName,
    measures: [
      {
        name: formValues.measure,
      },
      ...(formValues.comparisonTimeRange
        ? [
            {
              name: formValues.measure + ComparisonDeltaAbsoluteSuffix,
              comparisonDelta: { measure: formValues.measure },
            },
            {
              name: formValues.measure + ComparisonDeltaRelativeSuffix,
              comparisonRatio: { measure: formValues.measure },
            },
          ]
        : []),
    ],
    dimensions: formValues.splitByDimension
      ? [{ name: formValues.splitByDimension }]
      : [],
    where: sanitiseExpression(formValues.whereFilter, undefined),
    having: sanitiseExpression(undefined, {
      cond: {
        op: formValues.criteriaOperation,
        exprs: formValues.criteria
          .map(mapMeasureFilterToExpr)
          .filter((e) => !!e) as V1Expression[],
      },
    }),
    timeRange: {
      isoDuration: formValues.timeRange.isoDuration,
      timeZone: formValues.timeRange.timeZone,
      roundToGrain: formValues.timeRange.roundToGrain,
    },
    sort: [
      {
        name: formValues.measure,
        desc: false,
      },
    ],
    ...(formValues.comparisonTimeRange
      ? {
          comparisonTimeRange: {
            isoDuration: formValues.comparisonTimeRange.isoDuration,
            isoOffset: formValues.comparisonTimeRange.isoOffset,
          },
        }
      : {}),
  };
}

export const alertFormValidationSchema = yup.object({
  name: yup.string().required("Required"),
  measure: yup.string().required("Required"),
  criteria: yup.array().of(
    yup.object().shape({
      measure: yup.string().required("Required"),
      operation: yup.string().required("Required"),
      value1: yup.number().required("Required"),
    }),
  ),
  criteriaOperation: yup.string().required("Required"),
  snooze: yup.string().required("Required"),
  slackUsers: yup.array().of(
    yup.object().shape({
      email: yup.string().email("Invalid email"),
    }),
  ),
  emailRecipients: yup.array().of(
    yup.object().shape({
      email: yup.string().email("Invalid email"),
    }),
  ),
});
export const FieldsByTab: (keyof AlertFormValues)[][] = [
  ["measure"],
  ["criteria", "criteriaOperation"],
  ["name", "snooze", "slackUsers", "emailRecipients"],
];

export function checkIsTabValid(
  tabIndex: number,
  formValues: AlertFormValues,
  errors: Record<string, string>,
): boolean {
  let hasRequiredFields: boolean;
  let hasErrors: boolean;

  if (tabIndex === 0) {
    hasRequiredFields = formValues.measure !== "";
    hasErrors = !!errors.measure;
  } else if (tabIndex === 1) {
    hasRequiredFields = true;
    formValues.criteria.forEach((criteria) => {
      if (
        criteria.measure === "" ||
        (criteria.operation as string) === "" ||
        criteria.value1 === ""
      ) {
        hasRequiredFields = false;
      }
    });
    hasErrors =
      typeof errors.criteria === "string"
        ? !!errors.criteria
        : (errors.criteria as Array<MeasureFilterEntry>).some(
            (c) =>
              c.measure !== "" ||
              (c.operation as string) !== "" ||
              c.measure !== "",
          );
  } else if (tabIndex === 2) {
    // TODO: do better for >1 recipients
    hasRequiredFields =
      formValues.name !== "" &&
      formValues.snooze !== "" &&
      formValues.emailRecipients[0].email !== "";

    // There's a bug in how `svelte-forms-lib` types the `$errors` store for arrays.
    // See: https://github.com/tjinauyeung/svelte-forms-lib/issues/154#issuecomment-1087331250
    const recipientErrors = errors.emailRecipients as unknown as {
      email: string;
    }[];

    hasErrors = !!errors.name || !!errors.snooze || !!recipientErrors[0].email;
  } else {
    throw new Error(`Unexpected tabIndex: ${tabIndex}`);
  }

  return hasRequiredFields && !hasErrors;
}
