import {
  createInExpression,
  getValueIndexInExpression,
  getValuesInExpression,
  negateExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import { getWhereFilterExpressionIndex } from "../selectors/dimension-filters";
import type { DashboardMutables } from "./types";

export function toggleDimensionValueSelection(
  { dashboard }: DashboardMutables,
  dimensionName: string,
  dimensionValue: string,
  keepPillVisible?: boolean,
  /**
   * This marks the value as being exclusive. All other selected values will be unselected.
   */
  isExclusiveFilter?: boolean,
) {
  if (dashboard.temporaryFilterName !== null) {
    dashboard.temporaryFilterName = null;
  }

  const isInclude = !dashboard.dimensionFilterExcludeMode.get(dimensionName);
  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(dimensionName);
  if (exprIdx === undefined || exprIdx === -1) {
    dashboard?.whereFilter?.cond?.exprs?.push(
      createInExpression(dimensionName, [dimensionValue], !isInclude),
    );
    return;
  }

  const expr = dashboard.whereFilter.cond?.exprs?.[exprIdx];
  if (!expr?.cond?.exprs) {
    // should never happen since getWhereFilterExpressionIndex runs a find
    return;
  }

  const inIdx = getValueIndexInExpression(expr, dimensionValue) as number;
  if (inIdx === -1) {
    if (isExclusiveFilter) {
      expr.cond.exprs.splice(1, expr.cond.exprs.length - 1, {
        val: dimensionValue,
      });
    } else {
      expr.cond.exprs.push({ val: dimensionValue });
    }
  } else {
    expr.cond.exprs.splice(inIdx, 1);
    // Only decrement pinIndex if the removed value was before the pinned value
    if (dashboard.pinIndex >= inIdx) {
      dashboard.pinIndex--;
    }
    // remove the dimension entry if all values are removed
    if (expr.cond.exprs.length === 1) {
      dashboard.whereFilter.cond?.exprs?.splice(exprIdx, 1);
      if (keepPillVisible) {
        dashboard.temporaryFilterName = dimensionName;
      }
    }
  }
}

export function toggleDimensionFilterMode(
  { dashboard }: DashboardMutables,
  dimensionName: string,
) {
  const exclude = dashboard.dimensionFilterExcludeMode.get(dimensionName);
  dashboard.dimensionFilterExcludeMode.set(dimensionName, !exclude);

  if (!dashboard.whereFilter?.cond?.exprs) {
    return;
  }

  const exprIdx = dashboard.whereFilter.cond.exprs.findIndex(
    (e) => e.cond?.exprs?.[0].ident === dimensionName,
  );
  if (exprIdx === -1) {
    return;
  }
  dashboard.whereFilter.cond.exprs[exprIdx] = negateExpression(
    dashboard.whereFilter.cond.exprs[exprIdx],
  );
}

export function removeDimensionFilter(
  { dashboard }: DashboardMutables,
  dimensionName: string,
) {
  if (dashboard.temporaryFilterName === dimensionName) {
    dashboard.temporaryFilterName = null;
    return;
  }

  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(dimensionName);
  if (exprIdx === undefined || exprIdx === -1) return;
  dashboard.whereFilter?.cond?.exprs?.splice(exprIdx, 1);
}

export function selectItemsInFilter(
  { dashboard }: DashboardMutables,
  dimensionName: string,
  values: (string | null)[],
) {
  const isInclude = !dashboard.dimensionFilterExcludeMode.get(dimensionName);
  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(dimensionName);
  if (exprIdx === undefined || exprIdx === -1) {
    dashboard.whereFilter.cond?.exprs?.push(
      createInExpression(dimensionName, values, !isInclude),
    );
    return;
  }

  const expr = dashboard.whereFilter.cond?.exprs?.[exprIdx];
  if (!expr?.cond?.exprs) {
    // should never happen since getWhereFilterExpressionIndex runs a find
    return;
  }

  // preserve old selections and add only new ones
  const oldValues = getValuesInExpression(expr);
  const newValues = values.filter((v) => !oldValues.includes(v));
  // newValuesSelected = newValues.length; // TODO
  expr.cond.exprs.push(...newValues.map((v): V1Expression => ({ val: v })));
}

export function deselectItemsInFilter(
  { dashboard }: DashboardMutables,
  dimensionName: string,
  values: (string | null)[],
) {
  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(dimensionName);
  if (exprIdx === undefined || exprIdx === -1) {
    return;
  }

  const expr = dashboard.whereFilter.cond?.exprs?.[exprIdx];
  if (!expr?.cond?.exprs) {
    // should never happen since getWhereFilterExpressionIndex runs a find
    return;
  }

  // remove only deselected values
  const oldValues = getValuesInExpression(expr);
  const newValues = oldValues.filter((v) => !values.includes(v));

  if (newValues.length) {
    expr.cond.exprs.splice(
      1,
      expr.cond.exprs.length - 1,
      ...newValues.map((v): V1Expression => ({ val: v })),
    );
  } else {
    dashboard.whereFilter.cond?.exprs?.splice(exprIdx, 1);
  }
}

export const dimensionFilterActions = {
  /**
   * Toggles whether the given dimension value is selected in the
   * dimension filter for the given dimension.
   *
   * Note that this is different than the include/exclude mode for
   * dimension filters. This is a toggle for a specific value, whereas
   * the include/exclude mode is a toggle for the entire dimension.
   */
  toggleDimensionValueSelection,
  toggleDimensionFilterMode,
  removeDimensionFilter,
  selectItemsInFilter,
  deselectItemsInFilter,
};
