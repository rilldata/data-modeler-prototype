import {
  PivotChipData,
  PivotChipType,
  PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  TDDChart,
  TDDState,
} from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { FromURLParamTimeDimensionMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { DashboardState_LeaderboardSortDirection } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

export function getMetricsExplorerFromUrl(
  searchParams: URLSearchParams,
  metricsView: V1MetricsViewSpec,
) {
  // TODO: replace this with V1ExplorePreset once it is available on main
  const entity: Partial<MetricsExplorerEntity> = {};
  const measures = getMapFromArray(metricsView.measures ?? [], (m) => m.name!);
  const dimensions = getMapFromArray(
    metricsView.dimensions ?? [],
    (d) => d.name!,
  );

  // TODO: filter

  if (searchParams.has("tr")) {
    entity.selectedTimeRange = fromTimeRangeUrlParam(
      searchParams.get("tr") as string,
    );
  }
  if (searchParams.has("tz")) {
    entity.selectedTimezone = searchParams.get("tz") as string;
  }
  if (searchParams.has("ctr")) {
    entity.selectedComparisonTimeRange = fromTimeRangeUrlParam(
      searchParams.get("ctr") as string,
    );
  }
  if (searchParams.has("cd")) {
    const cd = searchParams.get("cd") as string;
    if (dimensions.has(cd)) {
      entity.selectedComparisonDimension = cd;
    }
  }

  Object.assign(
    entity,
    fromOverviewUrlParams(searchParams, measures, dimensions),
  );

  entity.tdd = fromTimeDimensionUrlParams(searchParams, measures);

  entity.pivot = fromPivotUrlParams(searchParams, measures, dimensions);

  return entity;
}

function fromTimeRangeUrlParam(timeRange: string) {
  const selectedTimeRange: DashboardTimeControls = {
    name: timeRange,
  } as DashboardTimeControls;
  // backwards compatibility
  if (timeRange in TimeRangePreset) {
    selectedTimeRange.name = TimeRangePreset[timeRange];
  }

  return selectedTimeRange;
}

function fromOverviewUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
) {
  const entity: Partial<MetricsExplorerEntity> = {};

  if (searchParams.has("e.m")) {
    const mes = searchParams.get("e.m") as string;
    if (mes === "*") {
      entity.allMeasuresVisible = true;
      entity.visibleMeasureKeys = new Set(measures.keys());
    } else {
      entity.allMeasuresVisible = false;
      entity.visibleMeasureKeys = new Set(
        mes.split(",").filter((m) => measures.has(m)),
      );
    }
  }

  if (searchParams.has("e.d")) {
    const dims = searchParams.get("e.d") as string;
    if (dims === "*") {
      entity.allDimensionsVisible = true;
      entity.visibleDimensionKeys = new Set(dimensions.keys());
    } else {
      entity.allDimensionsVisible = false;
      entity.visibleDimensionKeys = new Set(
        dims.split(",").filter((d) => dimensions.has(d)),
      );
    }
  }

  if (searchParams.has("e.sb")) {
    const sortBy = searchParams.get("e.sb") as string;
    if (measures.has(sortBy)) {
      entity.leaderboardMeasureName = sortBy;
    }
  }
  if (searchParams.has("e.sd")) {
    const sortDir = searchParams.get("e.sd") as string;
    entity.sortDirection =
      sortDir === "ASC" ? SortDirection.ASCENDING : SortDirection.DESCENDING;
  }

  if (searchParams.has("e.ed")) {
    const dim = searchParams.get("e.ed") as string;
    if (dimensions.has(dim)) {
      entity.selectedDimensionName = dim;
    }
  }

  return entity;
}

function fromTimeDimensionUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasureV2>,
) {
  let ttdMeasure: string | undefined;
  let ttdChartType: TDDChart | undefined;
  let ttdPin: number | undefined;

  if (searchParams.has("tdd.m")) {
    const mes = searchParams.get("tdd.m") as string;
    if (measures.has(mes)) {
      ttdMeasure = mes;
    }
  }
  if (searchParams.has("tdd.ct")) {
    const ct = searchParams.get("tdd.ct") as string;
    if (ct in TDDChart) {
      ttdChartType = TDDChart[ct];
    }
  }
  if (searchParams.has("tdd.p")) {
    const pin = Number(searchParams.get("tdd.p") as string);
    if (!Number.isNaN(pin)) {
      ttdPin = pin;
    }
  }

  return <TDDState>{
    expandedMeasureName: ttdMeasure ?? "",
    chartType: ttdChartType ?? TDDChart.DEFAULT,
    pinIndex: ttdPin ?? -1,
  };
}

function fromPivotUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
): PivotState {
  const mapPivotEntry = (entry: string): PivotChipData | undefined => {
    if (entry in FromURLParamTimeDimensionMap) {
      const grain = FromURLParamTimeDimensionMap[entry];
      return {
        id: grain,
        title: TIME_GRAIN[grain]?.label,
        type: PivotChipType.Time,
      };
    }

    if (measures.has(entry)) {
      const m = measures.get(entry)!;
      return {
        id: entry,
        title: m.label || m.name || "Unknown",
        type: PivotChipType.Measure,
      };
    }

    if (dimensions.has(entry)) {
      const d = dimensions.get(entry)!;
      return {
        id: entry,
        title: d.label || d.name || "Unknown",
        type: PivotChipType.Dimension,
      };
    }

    return undefined;
  };

  const rowDimensions: PivotChipData[] = [];
  if (searchParams.has("p.r")) {
    const pivotRows = searchParams.get("p.r") as string;
    pivotRows.split(",").forEach((pivotRow) => {
      const chip = mapPivotEntry(pivotRow);
      if (!chip) return;
      rowDimensions.push(chip);
    });
  }
  const colMeasures: PivotChipData[] = [];
  const colDimensions: PivotChipData[] = [];
  if (searchParams.has("p.c")) {
    const pivotCols = searchParams.get("p.c") as string;
    pivotCols.split(",").forEach((pivotRow) => {
      const chip = mapPivotEntry(pivotRow);
      if (!chip) return;
      if (chip.type === PivotChipType.Measure) {
        colMeasures.push(chip);
      } else {
        colDimensions.push(chip);
      }
    });
  }

  return {
    active: searchParams.get("view") === "pivot",
    rows: {
      dimension: rowDimensions,
    },
    columns: {
      measure: colMeasures,
      dimension: colDimensions,
    },
    // TODO: other fields
    expanded: {},
    sorting: [],
    columnPage: 1,
    rowPage: 1,
    enableComparison: false,
    activeCell: null,
    rowJoinType: "nest",
  };
}
