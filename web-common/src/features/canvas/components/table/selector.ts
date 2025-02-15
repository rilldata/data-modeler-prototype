import type { TableSpec } from "@rilldata/web-common/features/canvas/components/table";
import {
  validateDimensions,
  validateMeasures,
} from "@rilldata/web-common/features/canvas/components/validators";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { canEnablePivotComparison } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  type PivotDataStoreConfig,
  type PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { type Readable, derived } from "svelte/store";

export function getTableConfig(
  ctx: StateManagers,
  tableSpec: TableSpec,
  pivotState: PivotState,
  timeAndFilterStore: Readable<TimeAndFilterStore>,
): Readable<PivotDataStoreConfig> {
  const { metrics_view } = tableSpec;
  const {
    canvasEntity: {
      spec: { getMetricsViewFromName },
    },
  } = ctx;

  return derived(
    [getMetricsViewFromName(metrics_view), timeAndFilterStore],
    ([metricsView, $timeAndFilterStore]) => {
      const { timeRange, comparisonTimeRange, where } = $timeAndFilterStore;
      const enableComparison = canEnablePivotComparison(
        pivotState,
        comparisonTimeRange?.start,
      );

      const config: PivotDataStoreConfig = {
        measureNames: (tableSpec.measures || []).flatMap((name) => {
          const group = [name];
          if (enableComparison) {
            group.push(
              `${name}${COMPARISON_DELTA}`,
              `${name}${COMPARISON_PERCENT}`,
            );
          }
          return group;
        }),
        rowDimensionNames: tableSpec.row_dimensions || [],
        colDimensionNames: tableSpec.col_dimensions || [],
        allMeasures: metricsView?.measures || [],
        allDimensions: metricsView?.dimensions || [],
        whereFilter: where ?? createAndExpression([]),
        searchText: "",
        pivot: pivotState,
        enableComparison,
        comparisonTime: {
          start: comparisonTimeRange?.start,
          end: comparisonTimeRange?.end,
        },
        time: {
          timeStart: timeRange?.start,
          timeEnd: timeRange?.end,
          timeZone: timeRange.timeZone || "UTC",
          timeDimension: metricsView?.timeDimension || "",
        },
      };

      return config;
    },
  );
}

export function validateTableSchema(
  ctx: StateManagers,
  tableSpec: TableSpec,
): Readable<{
  isValid: boolean;
  error?: string;
}> {
  const { metrics_view } = tableSpec;
  return derived(
    ctx.canvasEntity.spec.getMetricsViewFromName(metrics_view),
    (metricsView) => {
      const measures = tableSpec.measures || [];
      const rowDimensions = tableSpec.row_dimensions || [];
      const colDimensions = tableSpec.col_dimensions || [];

      if (!metricsView) {
        return {
          isValid: false,
          error: `Metrics view ${metrics_view} not found`,
        };
      }

      if (!measures.length && !rowDimensions.length && !colDimensions.length) {
        return {
          isValid: false,
          error: "Select at least one measure or dimension for the table",
        };
      }
      const validateMeasuresRes = validateMeasures(metricsView, measures);
      if (!validateMeasuresRes.isValid) {
        const invalidMeasures = validateMeasuresRes.invalidMeasures.join(", ");
        return {
          isValid: false,
          error: `Invalid measure(s) "${invalidMeasures}" selected for the table`,
        };
      }

      const validateDimensionsRes = validateDimensions(
        metricsView,
        rowDimensions.concat(colDimensions),
      );

      if (!validateDimensionsRes.isValid) {
        const invalidDimensions =
          validateDimensionsRes.invalidDimensions.join(", ");

        return {
          isValid: false,
          error: `Invalid dimension(s) "${invalidDimensions}" selected for the table`,
        };
      }
      return {
        isValid: true,
        error: undefined,
      };
    },
  );
}
