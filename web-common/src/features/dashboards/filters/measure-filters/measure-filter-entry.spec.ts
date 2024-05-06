import {
  mapExprToMeasureFilter,
  mapMeasureFilterToExpr,
  MeasureFilterComparisonType,
  MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { MeasureFilterOperation } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import {
  createAndExpression,
  createBinaryExpression,
  createOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { V1Expression, V1Operation } from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";

const TestCases: [
  title: string,
  measureFilter: MeasureFilterEntry,
  expr: V1Expression | undefined,
][] = [
  [
    "greater than",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.GreaterThan,
      comparison: MeasureFilterComparisonType.None,
      not: false,
    },
    createBinaryExpression("imp", V1Operation.OPERATION_GT, 10),
  ],
  [
    "greater than or equals",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.GreaterThanOrEquals,
      comparison: MeasureFilterComparisonType.None,
      not: false,
    },
    createBinaryExpression("imp", V1Operation.OPERATION_GTE, 10),
  ],
  [
    "less than",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.LessThan,
      comparison: MeasureFilterComparisonType.None,
      not: false,
    },
    createBinaryExpression("imp", V1Operation.OPERATION_LT, 10),
  ],
  [
    "less than or equals",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.LessThanOrEquals,
      comparison: MeasureFilterComparisonType.None,
      not: false,
    },
    createBinaryExpression("imp", V1Operation.OPERATION_LTE, 10),
  ],
  [
    "between",
    {
      value1: "10",
      value2: "20",
      measure: "imp",
      operation: MeasureFilterOperation.Between,
      comparison: MeasureFilterComparisonType.None,
      not: false,
    },
    createAndExpression([
      createBinaryExpression("imp", V1Operation.OPERATION_GT, 10),
      createBinaryExpression("imp", V1Operation.OPERATION_LT, 20),
    ]),
  ],
  [
    "not between",
    {
      value1: "10",
      value2: "20",
      measure: "imp",
      operation: MeasureFilterOperation.NotBetween,
      comparison: MeasureFilterComparisonType.None,
      not: false,
    },
    createOrExpression([
      createBinaryExpression("imp", V1Operation.OPERATION_LTE, 10),
      createBinaryExpression("imp", V1Operation.OPERATION_GTE, 20),
    ]),
  ],
  [
    "invalid greater than",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.GreaterThan,
      comparison: MeasureFilterComparisonType.PercentageComparison,
      not: false,
    },
    undefined,
  ],
  [
    "increases by value",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.IncreasesBy,
      comparison: MeasureFilterComparisonType.AbsoluteComparison,
      not: false,
    },
    createBinaryExpression("imp__delta_abs", V1Operation.OPERATION_GT, 10),
  ],
  [
    "decreases by percent",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.DecreasesBy,
      comparison: MeasureFilterComparisonType.PercentageComparison,
      not: false,
    },
    createBinaryExpression("imp__delta_rel", V1Operation.OPERATION_LT, -0.1),
  ],
  [
    "changes by percent",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.ChangesBy,
      comparison: MeasureFilterComparisonType.PercentageComparison,
      not: false,
    },
    createOrExpression([
      createBinaryExpression("imp__delta_rel", V1Operation.OPERATION_LT, -0.1),
      createBinaryExpression("imp__delta_rel", V1Operation.OPERATION_GT, 0.1),
    ]),
  ],
  [
    // TODO
    "share of totals",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.ShareOfTotalsGreaterThan,
      comparison: MeasureFilterComparisonType.PercentageComparison,
      not: false,
    },
    undefined,
  ],
];

describe("mapMeasureFilterToExpr", () => {
  TestCases.forEach(([title, criteria, expr]) => {
    it(title, () => {
      expect(mapMeasureFilterToExpr(criteria)).toEqual(expr);
    });
  });
});

describe("mapMeasureFilterToExpr with NOT", () => {
  TestCases.forEach(([title, criteria, expr]) => {
    if (!expr) return;
    it(title, () => {
      expect(
        mapMeasureFilterToExpr({
          ...criteria,
          not: true,
        }),
      ).toEqual({
        cond: {
          op: V1Operation.OPERATION_NOT,
          exprs: [expr],
        },
      });
    });
  });
});

describe("mapExprToMeasureFilter", () => {
  TestCases.forEach(([title, criteria, expr]) => {
    if (!expr) return;
    it(title, () => {
      expect(mapExprToMeasureFilter(expr)).toEqual(criteria);
    });
  });
});

describe("mapExprToMeasureFilter with NOT", () => {
  TestCases.forEach(([title, criteria, expr]) => {
    if (!expr) return;
    it(title, () => {
      expect(
        mapExprToMeasureFilter({
          cond: {
            op: V1Operation.OPERATION_NOT,
            exprs: [expr],
          },
        }),
      ).toEqual({ ...criteria, not: true });
    });
  });
});
