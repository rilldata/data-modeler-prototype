import {
  MeasureFilterOperation,
  MeasureFilterToProtoOperation,
  MeasureFilterType,
  ProtoToMeasureFilterOperations,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import {
  createBetweenExpression,
  createBinaryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { V1Expression, V1Operation } from "@rilldata/web-common/runtime-client";

export type MeasureFilterEntry = {
  measure: string;
  operation: MeasureFilterOperation;
  type: MeasureFilterType;
  value1: string;
  value2: string;
};

export function getEmptyMeasureFilterEntry(): MeasureFilterEntry {
  return {
    measure: "",
    operation: MeasureFilterOperation.GreaterThan,
    type: MeasureFilterType.Value,
    value1: "0",
    value2: "",
  };
}

export const ComparisonDeltaPreviousSuffix = "_prev";
export const ComparisonDeltaAbsoluteSuffix = "_delta";
export const ComparisonDeltaRelativeSuffix = "_delta_perc";
const HasSuffixRegex = /_delta(?:_perc)?/;

export function mapExprToMeasureFilter(
  expr: V1Expression | undefined,
): MeasureFilterEntry | undefined {
  if (!expr) return undefined;

  let value1 = 0;
  let value2: number | undefined;
  let field = "";
  let operation = MeasureFilterOperation.GreaterThan;
  let type = MeasureFilterType.Value;

  switch (expr.cond?.op) {
    case V1Operation.OPERATION_OR:
    case V1Operation.OPERATION_AND:
      // handle between and not-between
      field = expr.cond.exprs?.[0].cond?.exprs?.[0].ident ?? "";
      value1 = (expr.cond.exprs?.[0].cond?.exprs?.[1].val as number) ?? 0;
      value2 = (expr.cond.exprs?.[1].cond?.exprs?.[1].val as number) ?? 0;
      operation =
        expr.cond?.op === V1Operation.OPERATION_AND
          ? MeasureFilterOperation.Between
          : MeasureFilterOperation.NotBetween;
      break;

    case V1Operation.OPERATION_EQ:
    case V1Operation.OPERATION_NEQ:
    case V1Operation.OPERATION_GT:
    case V1Operation.OPERATION_GTE:
    case V1Operation.OPERATION_LT:
    case V1Operation.OPERATION_LTE:
      field = expr.cond.exprs?.[0].ident ?? "";
      value1 = (expr.cond.exprs?.[1].val as number) ?? 0;
      if (field.endsWith(ComparisonDeltaRelativeSuffix)) {
        // convert decimal to percent
        value1 *= 100;
      }
      operation =
        ProtoToMeasureFilterOperations[expr.cond?.op] ??
        MeasureFilterOperation.GreaterThan;
      break;
  }

  if (field.endsWith(ComparisonDeltaAbsoluteSuffix)) {
    type = MeasureFilterType.AbsoluteChange;
  } else if (field.endsWith(ComparisonDeltaRelativeSuffix)) {
    type = MeasureFilterType.PercentChange;
  }

  return {
    measure: field.replace(HasSuffixRegex, ""),
    operation,
    type,
    value1: value1.toString(),
    value2: value2?.toString() ?? "",
  };
}

export function mapMeasureFilterToExpr(
  measureFilter: MeasureFilterEntry,
): V1Expression | undefined {
  let value = Number(measureFilter.value1);
  if (Number.isNaN(value)) {
    return undefined;
  }

  let suffix = "";
  switch (measureFilter.type) {
    case MeasureFilterType.Value:
      break;
    case MeasureFilterType.AbsoluteChange:
      suffix = ComparisonDeltaAbsoluteSuffix;
      break;
    case MeasureFilterType.PercentChange:
      value /= 100;
      suffix = ComparisonDeltaRelativeSuffix;
      break;
    case MeasureFilterType.PercentOfTotal:
      // TODO
      return undefined;
  }

  switch (measureFilter.operation) {
    case MeasureFilterOperation.Equals:
    case MeasureFilterOperation.NotEquals:
    case MeasureFilterOperation.GreaterThan:
    case MeasureFilterOperation.GreaterThanOrEquals:
    case MeasureFilterOperation.LessThan:
    case MeasureFilterOperation.LessThanOrEquals:
      return createBinaryExpression(
        measureFilter.measure + suffix,
        MeasureFilterToProtoOperation[measureFilter.operation],
        value,
      );

    case MeasureFilterOperation.Between:
    case MeasureFilterOperation.NotBetween:
      // between is only for filter pills. so do not support non value filters here
      if (measureFilter.type !== MeasureFilterType.Value) return undefined;
      return createBetweenExpression(
        measureFilter.measure + suffix,
        value,
        Number(measureFilter.value2 ?? "0"),
        measureFilter.operation === MeasureFilterOperation.NotBetween,
      );
  }
}
