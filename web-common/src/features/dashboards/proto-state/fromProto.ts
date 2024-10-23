import { protoBase64, type Timestamp } from "@bufbuild/protobuf";
import {
  type MeasureFilterEntry,
  mapExprToMeasureFilter,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import {
  PivotChipType,
  type PivotChipData,
  type PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  FromProtoOperationMap,
  FromProtoPivotRowJoinTypeMap,
  FromProtoTimeGrainMap,
} from "@rilldata/web-common/features/dashboards/proto-state/enum-maps";
import { convertFilterToExpression } from "@rilldata/web-common/features/dashboards/proto-state/filter-converter";
import {
  createAndExpression,
  filterIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { getMapFromArray } from "@rilldata/utils";
import {
  BOOLEANS,
  INTEGERS,
  isFloat,
} from "@rilldata/web-common/lib/duckdb-data-types";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import type {
  DashboardTimeControls,
  ScrubRange,
} from "@rilldata/web-common/lib/time/types";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import type { Expression } from "@rilldata/web-common/proto/gen/rill/runtime/v1/expression_pb";
import type { TimeGrain } from "@rilldata/web-common/proto/gen/rill/runtime/v1/time_grain_pb";
import {
  DashboardState,
  DashboardState_ActivePage,
  DashboardState_LeaderboardContextColumn,
  DashboardState_PivotRowJoinType,
  DashboardTimeRange,
  PivotElement,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  MetricsViewSpecDimensionV2,
  StructTypeField,
  V1ExploreSpec,
  V1Expression,
  V1MetricsViewSpec,
  V1StructType,
} from "@rilldata/web-common/runtime-client";

// TODO: make a follow up PR to use the one from the proto directly
const LeaderboardContextColumnReverseMap: Record<
  DashboardState_LeaderboardContextColumn,
  LeaderboardContextColumn
> = {
  [DashboardState_LeaderboardContextColumn.UNSPECIFIED]:
    LeaderboardContextColumn.HIDDEN,
  [DashboardState_LeaderboardContextColumn.PERCENT]:
    LeaderboardContextColumn.PERCENT,
  [DashboardState_LeaderboardContextColumn.DELTA_PERCENT]:
    LeaderboardContextColumn.DELTA_PERCENT,
  [DashboardState_LeaderboardContextColumn.DELTA_ABSOLUTE]:
    LeaderboardContextColumn.DELTA_ABSOLUTE,
  [DashboardState_LeaderboardContextColumn.HIDDEN]:
    LeaderboardContextColumn.HIDDEN,
};

const TDDChartTypeReverseMap: Record<string, TDDChart> = {
  default: TDDChart.DEFAULT,
  stacked_bar: TDDChart.STACKED_BAR,
  grouped_bar: TDDChart.GROUPED_BAR,
  stacked_area: TDDChart.STACKED_AREA,
};

export function getDashboardStateFromUrl(
  urlState: string,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  schema: V1StructType,
): Partial<MetricsExplorerEntity> {
  // backwards compatibility for older urls that had encoded state
  urlState = urlState.includes("%") ? decodeURIComponent(urlState) : urlState;
  return getDashboardStateFromProto(
    base64ToProto(urlState),
    metricsView,
    explore,
    schema,
  );
}

export function getDashboardStateFromProto(
  binary: Uint8Array,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  schema: V1StructType,
): Partial<MetricsExplorerEntity> {
  const dashboard = DashboardState.fromBinary(binary);
  const entity: Partial<MetricsExplorerEntity> = {};

  if (dashboard.filters) {
    // backwards compatibility for our older filter format
    entity.whereFilter = convertFilterToExpression(dashboard.filters);
    // older values could be strings for non-string values,
    // so we correct them using metrics view schema
    entity.whereFilter =
      correctFilterValues(
        entity.whereFilter,
        metricsView.dimensions ?? [],
        schema,
      ) ?? createAndExpression([]);
  } else if (dashboard.where) {
    entity.whereFilter = fromExpressionProto(dashboard.where);
  }
  if (dashboard.having) {
    entity.dimensionThresholdFilters = dashboard.having.map((h) => {
      const expr = fromExpressionProto(h.filter as Expression);
      return {
        name: h.name,
        filters: expr?.cond?.exprs
          ?.map(mapExprToMeasureFilter)
          .filter(Boolean) as MeasureFilterEntry[],
      };
    });
  }
  if (dashboard.compareTimeRange) {
    entity.selectedComparisonTimeRange = fromTimeRangeProto(
      dashboard.compareTimeRange,
    );
    // backwards compatibility
    correctComparisonTimeRange(entity.selectedComparisonTimeRange);
  }
  if (dashboard.showTimeComparison !== undefined) {
    entity.showTimeComparison = Boolean(dashboard.showTimeComparison);
  }

  if (dashboard.timeRange) {
    entity.selectedTimeRange = fromTimeRangeProto(dashboard.timeRange);
    if (dashboard.timeGrain) {
      entity.selectedTimeRange.interval =
        FromProtoTimeGrainMap[dashboard.timeGrain];
    }
  }

  if (dashboard.scrubRange) {
    entity.selectedScrubRange = fromTimeRangeProto(
      dashboard.scrubRange,
    ) as ScrubRange;
    entity.lastDefinedScrubRange = fromTimeRangeProto(
      dashboard.scrubRange,
    ) as ScrubRange;
  } else {
    entity.selectedScrubRange = undefined;
    entity.lastDefinedScrubRange = undefined;
  }

  if (dashboard.leaderboardMeasure) {
    entity.leaderboardMeasureName = dashboard.leaderboardMeasure;
  }
  if (dashboard.comparisonDimension) {
    entity.selectedComparisonDimension = dashboard.comparisonDimension;
  } else {
    entity.selectedComparisonDimension = undefined;
  }
  if (dashboard.expandedMeasure) {
    entity.tdd = {
      pinIndex: dashboard.pinIndex ?? -1,
      chartType: chartTypeMap(dashboard.chartType),
      expandedMeasureName: dashboard.expandedMeasure,
    };
  } else if (dashboard.activePage !== undefined) {
    entity.tdd = {
      pinIndex: -1,
      chartType: TDDChart.DEFAULT,
      expandedMeasureName: "",
    };
  }

  entity.selectedTimezone = dashboard.selectedTimezone ?? "UTC";

  if (dashboard.allMeasuresVisible) {
    entity.allMeasuresVisible = true;
    entity.visibleMeasureKeys = new Set(explore.measures);
  } else if (dashboard.visibleMeasures?.length) {
    entity.allMeasuresVisible = false;
    entity.visibleMeasureKeys = new Set(dashboard.visibleMeasures);
  }

  if (dashboard.allDimensionsVisible) {
    entity.allDimensionsVisible = true;
    entity.visibleDimensionKeys = new Set(explore.dimensions);
  } else if (dashboard.visibleDimensions?.length) {
    entity.allDimensionsVisible = false;
    entity.visibleDimensionKeys = new Set(dashboard.visibleDimensions);
  }

  if (dashboard.leaderboardContextColumn !== undefined) {
    entity.leaderboardContextColumn =
      LeaderboardContextColumnReverseMap[dashboard.leaderboardContextColumn];
  }

  if (dashboard.leaderboardSortDirection) {
    entity.sortDirection = dashboard.leaderboardSortDirection;
  }
  if (dashboard.leaderboardSortType) {
    entity.dashboardSortType = dashboard.leaderboardSortType;
  }

  entity.pivot = fromPivotProto(dashboard, metricsView);

  Object.assign(entity, fromActivePageProto(dashboard));

  return entity;
}

export function base64ToProto(message: string) {
  return protoBase64.dec(message);
}

function fromExpressionProto(expression: Expression): V1Expression | undefined {
  switch (expression.expression.case) {
    case "ident":
      return {
        ident: expression.expression.value,
      };

    case "val":
      return {
        val:
          expression.expression.value.kind.case === "nullValue"
            ? null
            : expression.expression.value.kind.value,
      };

    case "cond":
      return {
        cond: {
          op: FromProtoOperationMap[expression.expression.value.op],
          exprs: expression.expression.value.exprs
            .map((e) => fromExpressionProto(e))
            .filter((e): e is V1Expression => e !== undefined),
        },
      };
  }
}

function correctFilterValues(
  filter: V1Expression,
  dimensions: MetricsViewSpecDimensionV2[],
  schema: V1StructType,
) {
  return filterIdentifiers(filter, (e, ident) => {
    const dim = dimensions?.find((d) => d.name === ident);
    // ignore if dimension is not present anymore
    if (!dim) return false;
    const field = schema.fields?.find((f) => f.name === ident);
    // ignore if field is not found
    if (!field) return false;

    if (e.cond?.exprs) {
      e.cond.exprs =
        e.cond.exprs.map((e, i) => {
          if (i === 0) return e; // 1st expr is always the identifier
          return correctFilterValue(e, field);
        }) ?? [];
    }
    return true;
  });
}

function correctFilterValue(
  valueExpr: V1Expression,
  field: StructTypeField,
): V1Expression {
  if (valueExpr.val === null) {
    return valueExpr;
  }
  if (typeof valueExpr.val === "string") {
    // older filters were storing everything as strings
    return {
      val: correctStringFilterValue(valueExpr.val, field),
    };
  }
  return valueExpr;
}

function correctStringFilterValue(val: string, field: StructTypeField) {
  if (!field.type?.code) return val;

  if (INTEGERS.has(field.type?.code)) {
    return Number.parseInt(val);
  } else if (isFloat(field.type?.code)) {
    return Number.parseFloat(val);
  } else if (BOOLEANS.has(field.type?.code)) {
    return val === "true";
  } else {
    // TODO: other types
    return val;
  }
}

function fromTimeRangeProto(timeRange: DashboardTimeRange) {
  const selectedTimeRange: DashboardTimeControls = {
    name: timeRange.name,
  } as DashboardTimeControls;
  // backwards compatibility
  if (timeRange.name && timeRange.name in TimeRangePreset) {
    selectedTimeRange.name = TimeRangePreset[timeRange.name];
  }

  if (timeRange.timeStart) {
    selectedTimeRange.start = fromTimeProto(timeRange.timeStart);
  }
  if (timeRange.timeEnd) {
    selectedTimeRange.end = fromTimeProto(timeRange.timeEnd);
  }

  return selectedTimeRange;
}

function fromPivotProto(
  dashboard: DashboardState,
  metricsView: V1MetricsViewSpec,
): PivotState {
  const dimensionsMap = getMapFromArray(
    metricsView.dimensions ?? [],
    (d) => d.name,
  );
  const mapDimension: (name: string) => PivotChipData = (name: string) => {
    const dim = dimensionsMap.get(name);
    return {
      id: name,
      title: dim?.label || dim?.name || "Unknown",
      type: PivotChipType.Dimension,
    };
  };
  const mapTimeDimension: (grain: TimeGrain) => PivotChipData = (
    grain: TimeGrain,
  ) => ({
    id: FromProtoTimeGrainMap[grain],
    title: TIME_GRAIN[FromProtoTimeGrainMap[grain]].label,
    type: PivotChipType.Time,
  });
  const mapAllDimension: (dimension: PivotElement) => PivotChipData = (
    dimension: PivotElement,
  ) => {
    if (dimension?.element.case === "pivotTimeDimension") {
      const grain = dimension?.element.value;
      return {
        id: FromProtoTimeGrainMap[grain],
        title: TIME_GRAIN[FromProtoTimeGrainMap[grain]].label,
        type: PivotChipType.Time,
      };
    } else {
      return mapDimension(dimension?.element.value as string);
    }
  };

  const measuresMap = getMapFromArray(
    metricsView.measures ?? [],
    (m) => m.name,
  );
  const mapMeasure: (name: string) => PivotChipData = (name: string) => {
    const mes = measuresMap.get(name);
    return {
      id: name,
      title: mes?.label || mes?.name || "Unknown",
      type: PivotChipType.Measure,
    };
  };

  let rowDimensions: PivotChipData[] = [];
  let colDimensions: PivotChipData[] = [];
  if (
    dashboard.pivotRowAllDimensions?.length ||
    dashboard.pivotColumnAllDimensions?.length
  ) {
    rowDimensions = dashboard.pivotRowAllDimensions.map(mapAllDimension);
    colDimensions = dashboard.pivotColumnAllDimensions.map(mapAllDimension);
  } else if (
    // backwards compatibility for old URLs
    dashboard.pivotRowDimensions?.length ||
    dashboard.pivotRowTimeDimensions?.length ||
    dashboard.pivotColumnDimensions?.length ||
    dashboard.pivotColumnTimeDimensions?.length
  ) {
    rowDimensions = [
      ...dashboard.pivotRowTimeDimensions.map(mapTimeDimension),
      ...dashboard.pivotRowDimensions.map(mapDimension),
    ];
    colDimensions = [
      ...dashboard.pivotColumnTimeDimensions.map(mapTimeDimension),
      ...dashboard.pivotColumnDimensions.map(mapDimension),
    ];
  }

  return {
    active: dashboard.pivotIsActive ?? false,
    rows: {
      dimension: rowDimensions,
    },
    columns: {
      dimension: colDimensions,
      measure: dashboard.pivotColumnMeasures.map(mapMeasure),
    },
    expanded: dashboard.pivotExpanded,
    sorting: dashboard.pivotSort ?? [],
    columnPage: dashboard.pivotColumnPage ?? 1,
    rowPage: 1,
    enableComparison: dashboard.pivotEnableComparison ?? true,
    activeCell: null,
    rowJoinType:
      FromProtoPivotRowJoinTypeMap[
        dashboard.pivotRowJoinType || DashboardState_PivotRowJoinType.NEST
      ],
  };
}

function correctComparisonTimeRange(
  comparisonTimeRange: DashboardTimeControls,
) {
  switch (comparisonTimeRange.name as string) {
    case "CONTIGUOUS":
      comparisonTimeRange.name = TimeComparisonOption.CONTIGUOUS;
      break;
    case "P1D":
      comparisonTimeRange.name = TimeComparisonOption.DAY;
      break;
    case "P1W":
      comparisonTimeRange.name = TimeComparisonOption.WEEK;
      break;
    case "P1M":
      comparisonTimeRange.name = TimeComparisonOption.MONTH;
      break;
    case "P3M":
      comparisonTimeRange.name = TimeComparisonOption.QUARTER;
      break;
    case "P1Y":
      comparisonTimeRange.name = TimeComparisonOption.YEAR;
      break;
  }
}

function chartTypeMap(chartType: string | undefined): TDDChart {
  if (!chartType || !TDDChartTypeReverseMap[chartType]) {
    return TDDChart.DEFAULT;
  }
  return TDDChartTypeReverseMap[chartType];
}

function fromTimeProto(timestamp: Timestamp) {
  return new Date(Number(timestamp.seconds));
}

function fromActivePageProto(
  dashboard: DashboardState,
): Partial<
  Pick<MetricsExplorerEntity, "activePage" | "selectedDimensionName">
> {
  switch (dashboard.activePage) {
    case DashboardState_ActivePage.UNSPECIFIED:
      // backwards compatibility
      if (dashboard.selectedDimension) {
        return {
          activePage: DashboardState_ActivePage.DIMENSION_TABLE,
          selectedDimensionName: dashboard.selectedDimension,
        };
      } else if (dashboard.expandedMeasure) {
        return {
          activePage: DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
          selectedDimensionName: undefined,
        };
      }
      // return empty so that nothing is overridden
      // this is used to store partial data in the proto, like filters only, which should not override selected values
      return {};

    case DashboardState_ActivePage.DEFAULT:
    case DashboardState_ActivePage.PIVOT:
    case DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL:
      return {
        activePage: dashboard.activePage,
        selectedDimensionName: undefined,
      };

    case DashboardState_ActivePage.DIMENSION_TABLE:
      return {
        activePage: dashboard.activePage,
        selectedDimensionName: dashboard.selectedDimension,
      };
  }
}
