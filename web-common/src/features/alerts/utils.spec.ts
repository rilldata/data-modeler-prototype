import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import { generateAlertName } from "@rilldata/web-common/features/alerts/utils";
import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const MetricsView: V1MetricsViewSpec = {
  measures: [
    {
      name: "total_records",
      label: "Total records",
    },
  ],
};

describe("generateAlertName", () => {
  const TestCases: {
    title: string;
    formValues: Partial<AlertFormValues>;
    expected: string;
  }[] = [
    {
      title: "value criteria without comparison",
      formValues: {
        measures: ["total_records"],
        criteria: [
          {
            measure: "total_records",
            type: MeasureFilterType.Value,
            operation: MeasureFilterOperation.LessThan,
            value1: "10",
            value2: "",
          },
        ],
      },
      expected: "Total records value alert",
    },
    {
      title: "absolute change criteria with comparison",
      formValues: {
        measures: ["total_records"],
        criteria: [
          {
            measure: "total_records",
            type: MeasureFilterType.AbsoluteChange,
            operation: MeasureFilterOperation.GreaterThanOrEquals,
            value1: "10",
            value2: "",
          },
        ],
        comparisonTimeRange: {
          isoOffset: "rill-PW",
        },
      },
      expected: "Total records change vs previous week alert",
    },
    {
      title: "percent change criteria with comparison",
      formValues: {
        measures: ["total_records"],
        criteria: [
          {
            measure: "total_records",
            type: MeasureFilterType.PercentChange,
            operation: MeasureFilterOperation.GreaterThan,
            value1: "-10",
            value2: "",
          },
        ],
        comparisonTimeRange: {
          isoOffset: "rill-PM",
        },
      },
      expected: "Total records % change vs previous month alert",
    },
  ];

  for (const { title, formValues, expected } of TestCases) {
    it(title, () => {
      expect(
        generateAlertName(formValues as AlertFormValues, MetricsView),
      ).toEqual(expected);
    });
  }
});
