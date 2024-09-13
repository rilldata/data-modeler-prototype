import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createOrExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  convertFilterParamToExpression,
  stripParserError,
} from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";

describe("expression", () => {
  describe("positive cases", () => {
    const Cases = [
      {
        expr: "country IN ('US','IN') and state = 'ABC'",
        expected_expression: createAndExpression([
          createInExpression("country", ["US", "IN"]),
          createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
        ]),
      },
      {
        expr: "country IN ('US','IN') and state = 'ABC' and lat >= 12.56",
        expected_expression: createAndExpression([
          createInExpression("country", ["US", "IN"]),
          createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
          createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
        ]),
      },
      {
        expr: "country IN ('US','IN') AND state = 'ABC' OR lat >= 12.56",
        expected_expression: createAndExpression([
          createInExpression("country", ["US", "IN"]),
          createOrExpression([
            createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
            createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
          ]),
        ]),
      },
      {
        expr: "country not in ('US','IN') and (state = 'ABC' or lat >= 12.56)",
        expected_expression: createAndExpression([
          createInExpression("country", ["US", "IN"], true),
          createOrExpression([
            createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
            createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
          ]),
        ]),
      },
      {
        expr: "country NIN ('US','IN') and state having (lat >= 12.56)",
        expected_expression: createAndExpression([
          createInExpression("country", ["US", "IN"], true),
          createSubQueryExpression(
            "state",
            ["lat"],
            createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
          ),
        ]),
      },
    ];

    for (const { expr, expected_expression } of Cases) {
      it(expr, () => {
        expect(convertFilterParamToExpression(expr)).toEqual(
          expected_expression,
        );
      });
    }
  });

  describe("negative cases", () => {
    const Cases = [
      {
        expr: "country ('US','IN') and state = 'ABC'",
        err: `Syntax error at line 1 col 9:

1 country ('US','IN') and state = 'ABC'
          ^

Unexpected "(".`,
      },
      {
        expr: "country IN (US,'IN') and state = 'ABC'",
        err: `Syntax error at line 1 col 13:

1 country IN (US,'IN') and state = 'ABC'
              ^

Unexpected "U".`,
      },
    ];

    for (const { expr, err } of Cases) {
      it(expr, () => {
        let hasError = false;
        try {
          convertFilterParamToExpression(expr);
        } catch (e) {
          hasError = true;
          expect(stripParserError(e)).toEqual(err);
        }
        expect(hasError).to.be.true;
      });
    }
  });
});
