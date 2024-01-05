import { V1Operation } from "@rilldata/web-common/runtime-client";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

export function createLikeExpression(
  ident: string,
  like: string,
  negate: boolean
) {
  return {
    cond: {
      op: negate ? V1Operation.OPERATION_NLIKE : V1Operation.OPERATION_LIKE,
      exprs: [{ ident }, { val: like }],
    },
  } as V1Expression;
}

export function createInExpression(
  ident: string,
  vals: Array<any>,
  negate: boolean
) {
  return {
    cond: {
      op: negate ? V1Operation.OPERATION_NIN : V1Operation.OPERATION_IN,
      exprs: [{ ident }, ...vals.map((val) => ({ val }))],
    },
  } as V1Expression;
}

export function createAndExpression(exprs: Array<V1Expression>) {
  return {
    cond: {
      op: V1Operation.OPERATION_AND,
      exprs,
    },
  } as V1Expression;
}

export function createOrExpression(exprs: Array<V1Expression>) {
  return {
    cond: {
      op: V1Operation.OPERATION_OR,
      exprs,
    },
  } as V1Expression;
}

export function createBinaryExpression(
  ident: string,
  op: V1Operation,
  val: number
) {
  return {
    cond: {
      op,
      exprs: [{ ident }, { val }],
    },
  } as V1Expression;
}

export function createBetweenExpression(
  ident: string,
  val1: number,
  val2: number,
  negate: boolean
) {
  const exprs = [
    {
      cond: {
        op: negate ? V1Operation.OPERATION_LTE : V1Operation.OPERATION_GT,
        exprs: [{ ident }, { val: val1 }],
      },
    },
    {
      cond: {
        op: negate ? V1Operation.OPERATION_GTE : V1Operation.OPERATION_LT,
        exprs: [{ ident }, { val: val2 }],
      },
    },
  ] as Array<V1Expression>;
  if (negate) {
    return createOrExpression(exprs);
  } else {
    return createAndExpression(exprs);
  }
}

const conditionOperationComplement: Partial<Record<V1Operation, V1Operation>> =
  {
    [V1Operation.OPERATION_EQ]: V1Operation.OPERATION_NEQ,
    [V1Operation.OPERATION_LT]: V1Operation.OPERATION_GTE,
    [V1Operation.OPERATION_LTE]: V1Operation.OPERATION_GT,
    [V1Operation.OPERATION_IN]: V1Operation.OPERATION_NIN,
    [V1Operation.OPERATION_LIKE]: V1Operation.OPERATION_NLIKE,
    [V1Operation.OPERATION_AND]: V1Operation.OPERATION_OR,
  };
// add inverse of existing values above
for (const c in conditionOperationComplement) {
  conditionOperationComplement[conditionOperationComplement[c]] = c;
}

export function negateExpression(expr: V1Expression): V1Expression {
  if ("ident" in expr || "val" in expr || !expr.cond) return expr;
  return {
    cond: {
      op:
        conditionOperationComplement[expr.cond.op as V1Operation] ??
        V1Operation.OPERATION_EQ,
      exprs: expr.cond.exprs,
    },
  };
}

export function forEachExpression(
  expr: V1Expression,
  cb: (e: V1Expression, depth?: number) => void,
  depth = 0
) {
  if (!expr.cond?.exprs) {
    cb(expr, depth);
    return;
  }

  for (const subExpr of expr.cond.exprs) {
    cb(subExpr, depth);
    forEachExpression(subExpr, cb, depth + 1);
  }
}

/**
 * Creates a copy of the expression with sub expressions removed based on {@link checker}
 */
export function removeExpressions(
  expr: V1Expression,
  checker: (e: V1Expression) => boolean
): V1Expression | undefined {
  if (!expr.cond?.exprs) {
    return {
      ...expr,
    };
  }

  const newExpr: V1Expression = {
    cond: {
      op: expr.cond.op,
      exprs: expr.cond.exprs
        .map((e) => removeExpressions(e, checker))
        .filter((e) => e !== undefined && checker(e)) as Array<V1Expression>,
    },
  };

  switch (expr.cond.op) {
    // and/or will have only sub expressions
    case V1Operation.OPERATION_AND:
    case V1Operation.OPERATION_OR:
      if (newExpr.cond?.exprs?.length === 0) return undefined;
      break;

    // other types will have identifier as 1st expression
    default:
      if (!newExpr.cond?.exprs?.length || newExpr.cond.exprs.length > 1)
        return undefined;
      break;
  }

  return newExpr;
}

export function getValueIndexInExpression(expr: V1Expression, value: any) {
  return expr.cond?.exprs?.findIndex((e, i) => i > 0 && e.val === value);
}

export function getValuesInExpression(expr?: V1Expression) {
  return expr ? (expr.cond?.exprs?.slice(1).map((e) => e.val) as any[]) : [];
}
