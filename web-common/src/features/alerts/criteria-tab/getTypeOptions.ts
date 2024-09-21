import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import {
  MeasureFilterBaseTypeOptions,
  MeasureFilterComparisonTypeOptions,
  MeasureFilterPercentOfTotalOption,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { getComparisonLabel } from "@rilldata/web-common/lib/time/comparisons";
import { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";

export function getTypeOptions(
  formValues: AlertFormValues,
  selectedMeasure: MetricsViewSpecMeasureV2 | undefined,
) {
  const options: {
    value: string;
    label: string;
    disabled?: boolean;
    tooltip?: string;
  }[] = [...MeasureFilterBaseTypeOptions];

  if (
    formValues.comparisonTimeRange?.isoDuration ||
    formValues.comparisonTimeRange?.isoOffset
  ) {
    const comparisonLabel = getComparisonLabel(
      formValues.comparisonTimeRange,
    ).toLowerCase();
    options.push(
      ...MeasureFilterComparisonTypeOptions.map((o) => {
        return {
          ...o,
          label: `${o.label} ${comparisonLabel}`,
        };
      }),
    );
  } else {
    options.push(
      ...MeasureFilterComparisonTypeOptions.map((o) => {
        return {
          ...o,
          label: o.shortLabel,
          tooltip: "Available when comparing time periods.",
          disabled: true,
        };
      }),
    );
  }

  if (
    selectedMeasure?.validPercentOfTotal &&
    formValues.dimensions.length > 0
  ) {
    options.push(MeasureFilterPercentOfTotalOption);
  } else {
    options.push({
      ...MeasureFilterPercentOfTotalOption,
      tooltip: "Measure does not support percent-of-total.",
      disabled: true,
    });
  }

  return options;
}
