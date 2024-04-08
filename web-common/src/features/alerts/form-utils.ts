import { mapAlertCriteriaToExpression } from "@rilldata/web-common/features/alerts/criteria-tab/map-alert-criteria";
import { type CompareWith } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
import {
  IsCompareMeasureFilterOperation,
  MeasureFilterOperation,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import {
  createAndExpression,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  type V1Expression,
  V1MetricsViewComparisonMeasureType,
  type V1MetricsViewComparisonRequest,
  type V1Operation,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import * as yup from "yup";

export type AlertCriteria = {
  field: string;
  operation: MeasureFilterOperation;
  compareWith: CompareWith;
  value: string;
};
export type AlertFormValues = {
  name: string;
  measure: string;
  splitByDimension: string;
  criteria: AlertCriteria[];
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
): V1MetricsViewComparisonRequest {
  const addComparison =
    !!formValues.comparisonTimeRange &&
    !!formValues.criteria.find(
      (c) => c.operation in IsCompareMeasureFilterOperation,
    );

  return {
    metricsViewName: formValues.metricsViewName,
    measures: [{ name: formValues.measure }],
    dimension: formValues.splitByDimension
      ? { name: formValues.splitByDimension }
      : undefined,
    where: sanitiseExpression(formValues.whereFilter, undefined),
    having: sanitiseExpression(
      undefined,
      createAndExpression(
        formValues.criteria
          .map(mapAlertCriteriaToExpression)
          .filter((e) => !!e) as V1Expression[],
      ),
    ),
    timeRange: {
      isoDuration: formValues.timeRange.isoDuration,
    },
    ...(addComparison
      ? {
          comparisonTimeRange: {
            // Use time range to add duration and offset in case comparison is not enabled
            isoDuration:
              formValues.comparisonTimeRange?.isoDuration ??
              formValues.timeRange.isoDuration,
            isoOffset:
              formValues.comparisonTimeRange?.isoOffset ??
              formValues.timeRange.isoDuration,
          },
        }
      : {}),
    sort: [
      {
        name: formValues.measure,
        sortType:
          V1MetricsViewComparisonMeasureType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
      },
    ],
    aliases: addComparison
      ? [
          {
            name: formValues.measure,
            type: V1MetricsViewComparisonMeasureType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
            alias: formValues.measure + "__delta_abs",
          },
          {
            name: formValues.measure,
            type: V1MetricsViewComparisonMeasureType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA,
            alias: formValues.measure + "__delta_rel",
          },
        ]
      : [],
  };
}

export const alertFormValidationSchema = yup.object({
  name: yup.string().required("Required"),
  measure: yup.string().required("Required"),
  criteria: yup.array().of(
    yup.object().shape({
      field: yup.string().required("Required"),
      operation: yup.string().required("Required"),
      value: yup.number().required("Required"),
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
  ["name", "measure"],
  ["criteria", "criteriaOperation"],
  ["snooze", "slackUsers", "emailRecipients"],
];

export function checkIsTabValid(
  tabIndex: number,
  formValues: AlertFormValues,
  errors: Record<string, string>,
): boolean {
  let hasRequiredFields: boolean;
  let hasErrors: boolean;

  if (tabIndex === 0) {
    hasRequiredFields = formValues.name !== "" && formValues.measure !== "";
    hasErrors = !!errors.name && !!errors.measure;
  } else if (tabIndex === 1) {
    hasRequiredFields = true;
    formValues.criteria.forEach((criteria) => {
      if (
        criteria.field === "" ||
        (criteria.operation as string) === "" ||
        criteria.value === ""
      ) {
        hasRequiredFields = false;
      }
    });
    hasErrors = !!errors.criteria;
  } else if (tabIndex === 2) {
    // TODO: do better for >1 recipients
    hasRequiredFields =
      formValues.snooze !== "" && formValues.emailRecipients[0].email !== "";

    // There's a bug in how `svelte-forms-lib` types the `$errors` store for arrays.
    // See: https://github.com/tjinauyeung/svelte-forms-lib/issues/154#issuecomment-1087331250
    const recipientErrors = errors.emailRecipients as unknown as {
      email: string;
    }[];

    hasErrors = !!errors.snooze || !!recipientErrors[0].email;
  } else {
    throw new Error(`Unexpected tabIndex: ${tabIndex}`);
  }

  return hasRequiredFields && !hasErrors;
}
