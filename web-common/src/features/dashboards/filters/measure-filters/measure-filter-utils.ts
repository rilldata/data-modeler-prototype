import {
  mapExprToMeasureFilter,
  mapMeasureFilterToExpr,
  MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  createAndExpression,
  createSubQueryExpression,
  filterExpressions,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  DimensionThresholdFilter,
  MetricsExplorerEntity,
} from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { V1Expression, V1Operation } from "@rilldata/web-common/runtime-client";

export function mergeMeasureFilters(
  dashboard: MetricsExplorerEntity,
  whereFilter = dashboard.whereFilter,
) {
  return mergeDimensionAndMeasureFilter(
    whereFilter,
    dashboard.dimensionThresholdFilters,
  );
}

export function mergeDimensionAndMeasureFilter(
  whereFilter: V1Expression,
  dimensionThresholdFilters: DimensionThresholdFilter[],
) {
  const where =
    filterExpressions(whereFilter, () => true) ?? createAndExpression([]);
  where.cond?.exprs?.push(
    ...dimensionThresholdFilters.map(convertDimensionThresholdFilter),
  );
  return where;
}

/**
 * Splits where filter into dimension and measure filters.
 * Measure filters will be sub-queries
 */
export function splitWhereFilter(whereFilter: V1Expression | undefined) {
  const dimensionFilters = createAndExpression([]);
  const dimensionThresholdFilters: DimensionThresholdFilter[] = [];
  whereFilter?.cond?.exprs?.filter((e) => {
    const subqueryExpr = e.cond?.exprs?.[1];
    if (subqueryExpr?.subquery) {
      dimensionThresholdFilters.push({
        name: subqueryExpr.subquery.dimension ?? "",
        filters:
          (subqueryExpr.subquery.having?.cond?.exprs
            ?.map(mapExprToMeasureFilter)
            .filter(Boolean) as MeasureFilterEntry[]) ?? [],
      });
      return;
    }

    dimensionFilters.cond?.exprs?.push(e);
  });

  return { dimensionFilters, dimensionThresholdFilters };
}

function convertDimensionThresholdFilter(
  dtf: DimensionThresholdFilter,
): V1Expression {
  return createSubQueryExpression(
    dtf.name,
    dtf.filters.map((f) => f.measure),
    createAndExpression(
      dtf.filters.map(mapMeasureFilterToExpr).filter(Boolean) as V1Expression[],
    ),
  );
}
