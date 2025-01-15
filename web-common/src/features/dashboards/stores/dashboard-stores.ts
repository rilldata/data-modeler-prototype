import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getWhereFilterExpressionIndex } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import { AdvancedMeasureCorrector } from "@rilldata/web-common/features/dashboards/stores/AdvancedMeasureCorrector";
import { getFullInitExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import {
  createAndExpression,
  filterExpressions,
  forEachIdentifier,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { type MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import {
  TimeRangePreset,
  type DashboardTimeControls,
  type ScrubRange,
  type TimeRange,
} from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1ExploreSpec,
  V1Expression,
  V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import {
  V1Operation,
  type V1StructType,
} from "@rilldata/web-common/runtime-client";
import type { ExpandedState, SortingState } from "@tanstack/svelte-table";
import { derived, writable, type Readable } from "svelte/store";
import { SortType } from "web-common/src/features/dashboards/proto-state/derived-types";
import type { PivotColumns, PivotRows } from "../pivot/types";
import { PivotChipType, type PivotChipData } from "../pivot/types";

export interface MetricsExplorerStoreType {
  entities: Record<string, MetricsExplorerEntity>;
}
const { update, subscribe } = writable({
  entities: {},
} as MetricsExplorerStoreType);

function updateMetricsExplorerProto(metricsExplorer: MetricsExplorerEntity) {
  metricsExplorer.proto = getProtoFromDashboardState(metricsExplorer);
}

export const updateMetricsExplorerByName = (
  name: string,
  callback: (metricsExplorer: MetricsExplorerEntity) => void,
) => {
  update((state) => {
    if (!state.entities[name]) {
      return state;
    }

    callback(state.entities[name]);
    // every change triggers a proto update
    updateMetricsExplorerProto(state.entities[name]);
    return state;
  });
};

function includeExcludeModeFromFilters(filters: V1Expression | undefined) {
  const map = new Map<string, boolean>();
  if (!filters) return map;
  forEachIdentifier(filters, (e, ident) => {
    if (e.cond?.op === V1Operation.OPERATION_NIN) {
      map.set(ident, true);
    }
  });
  return map;
}

function syncMeasures(
  explore: V1ExploreSpec,
  metricsExplorer: MetricsExplorerEntity,
) {
  const measuresSet = new Set(explore.measures ?? []);

  // sync measures with selected leaderboard measure.
  if (
    explore.measures?.length &&
    !measuresSet.has(metricsExplorer.leaderboardMeasureName)
  ) {
    metricsExplorer.leaderboardMeasureName = explore.measures[0];
  }

  if (
    metricsExplorer.tdd.expandedMeasureName &&
    !measuresSet.has(metricsExplorer.tdd.expandedMeasureName)
  ) {
    metricsExplorer.tdd.expandedMeasureName = undefined;
  }

  metricsExplorer.pivot.columns.measure =
    metricsExplorer.pivot.columns.measure.filter((measure) =>
      measuresSet.has(measure.id),
    );

  if (metricsExplorer.allMeasuresVisible) {
    // this makes sure that the visible keys is in sync with list of measures
    metricsExplorer.visibleMeasureKeys = measuresSet;
  } else {
    // remove any keys from visible measure if it doesn't exist anymore
    for (const measureKey of metricsExplorer.visibleMeasureKeys) {
      if (!measuresSet.has(measureKey)) {
        metricsExplorer.visibleMeasureKeys.delete(measureKey);
      }
    }
    // If there are no visible measures, make the first measure visible
    if (
      explore.measures?.length &&
      metricsExplorer.visibleMeasureKeys.size === 0
    ) {
      metricsExplorer.visibleMeasureKeys = new Set([explore.measures[0]]);
    }
  }
}

function syncDimensions(
  explore: V1ExploreSpec,
  metricsExplorer: MetricsExplorerEntity,
) {
  // Having a map here improves the lookup for existing dimension name
  const dimensionsSet = new Set(explore.dimensions ?? []);
  metricsExplorer.whereFilter =
    filterExpressions(metricsExplorer.whereFilter, (e) => {
      if (!e.cond?.exprs?.length) return true;
      return dimensionsSet.has(e.cond.exprs[0].ident!);
    }) ?? createAndExpression([]);

  if (
    metricsExplorer.selectedDimensionName &&
    !dimensionsSet.has(metricsExplorer.selectedDimensionName)
  ) {
    metricsExplorer.selectedDimensionName = undefined;
    metricsExplorer.activePage = DashboardState_ActivePage.DEFAULT;
  }

  metricsExplorer.pivot.rows.dimension =
    metricsExplorer.pivot.rows.dimension.filter(
      (dimension) =>
        dimensionsSet.has(dimension.id) ||
        dimension.type === PivotChipType.Time,
    );

  metricsExplorer.pivot.columns.dimension =
    metricsExplorer.pivot.columns.dimension.filter(
      (dimension) =>
        dimensionsSet.has(dimension.id) ||
        dimension.type === PivotChipType.Time,
    );

  if (metricsExplorer.allDimensionsVisible) {
    // this makes sure that the visible keys is in sync with list of dimensions
    metricsExplorer.visibleDimensionKeys = dimensionsSet;
  } else {
    // remove any keys from visible dimension if it doesn't exist anymore
    for (const dimensionKey of metricsExplorer.visibleDimensionKeys) {
      if (!dimensionsSet.has(dimensionKey)) {
        metricsExplorer.visibleDimensionKeys.delete(dimensionKey);
      }
    }
  }
}

const metricsViewReducers = {
  init(name: string, initState: Partial<MetricsExplorerEntity> = {}) {
    update((state) => {
      state.entities[name] = getFullInitExploreState(name, initState);

      updateMetricsExplorerProto(state.entities[name]);

      return state;
    });
  },

  syncFromUrl(
    name: string,
    urlState: string,
    metricsView: V1MetricsViewSpec,
    explore: V1ExploreSpec,
    schema: V1StructType,
  ) {
    if (!urlState || !metricsView) return;
    // not all data for MetricsExplorerEntity will be filled out here.
    // Hence, it is a Partial<MetricsExplorerEntity>
    const partial = getDashboardStateFromUrl(
      urlState,
      metricsView,
      explore,
      schema,
    );
    if (!partial) return;

    updateMetricsExplorerByName(name, (metricsExplorer) => {
      for (const key in partial) {
        metricsExplorer[key] = partial[key];
      }
      // this hack is needed since what is shown for comparison is not a single source
      // TODO: use an enum and get rid of this
      if (!partial.showTimeComparison) {
        metricsExplorer.showTimeComparison = false;
      }
      metricsExplorer.dimensionFilterExcludeMode =
        includeExcludeModeFromFilters(partial.whereFilter);
      AdvancedMeasureCorrector.correct(metricsExplorer, metricsView);
    });
  },

  mergePartialExplorerEntity(
    name: string,
    partialExploreState: Partial<MetricsExplorerEntity>,
    metricsView: V1MetricsViewSpec,
  ) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      for (const key in partialExploreState) {
        metricsExplorer[key] = partialExploreState[key];
      }
      // this hack is needed since what is shown for comparison is not a single source
      // TODO: use an enum and get rid of this
      if (!partialExploreState.showTimeComparison) {
        metricsExplorer.showTimeComparison = false;
      }
      metricsExplorer.dimensionFilterExcludeMode =
        includeExcludeModeFromFilters(partialExploreState.whereFilter);
      AdvancedMeasureCorrector.correct(metricsExplorer, metricsView);
    });
  },

  sync(name: string, explore: V1ExploreSpec) {
    if (!name || !explore || !explore.measures) return;
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      // remove references to non existent measures
      syncMeasures(explore, metricsExplorer);

      // remove references to non existent dimensions
      syncDimensions(explore, metricsExplorer);
    });
  },

  setPivotMode(name: string, mode: boolean) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.pivot = { ...metricsExplorer.pivot, active: mode };
      if (mode) {
        metricsExplorer.activePage = DashboardState_ActivePage.PIVOT;
      } else if (metricsExplorer.selectedDimensionName) {
        metricsExplorer.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
      } else if (metricsExplorer.tdd.expandedMeasureName) {
        metricsExplorer.activePage =
          DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL;
      } else {
        metricsExplorer.activePage = DashboardState_ActivePage.DEFAULT;
      }
    });
  },

  setPivotRows(name: string, value: PivotChipData[]) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.pivot.rowPage = 1;
      metricsExplorer.pivot.activeCell = null;

      const dimensions: PivotChipData[] = [];

      value.forEach((val) => {
        if (val.type !== PivotChipType.Measure) {
          dimensions.push(val);
        }
      });

      if (metricsExplorer.pivot.sorting.length) {
        const accessor = metricsExplorer.pivot.sorting[0].id;
        const anchorDimension = dimensions?.[0]?.id;
        if (accessor !== anchorDimension) {
          metricsExplorer.pivot.sorting = [];
        }
      }

      metricsExplorer.pivot.rows = {
        dimension: dimensions,
      };
    });
  },

  setPivotColumns(name: string, value: PivotChipData[]) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.pivot.rowPage = 1;
      metricsExplorer.pivot.activeCell = null;
      metricsExplorer.pivot.expanded = {};

      const dimensions: PivotChipData[] = [];
      const measures: PivotChipData[] = [];

      value.forEach((val) => {
        if (val.type === PivotChipType.Measure) {
          measures.push(val);
        } else {
          dimensions.push(val);
        }
      });

      // Reset sorting if the sorting field is not in the pivot columns
      if (metricsExplorer.pivot.sorting.length) {
        const accessor = metricsExplorer.pivot.sorting[0].id;
        const anchorDimension = metricsExplorer.pivot.rows.dimension?.[0].id;
        if (accessor !== anchorDimension) {
          metricsExplorer.pivot.sorting = [];
        }
      }
      metricsExplorer.pivot.columns = {
        dimension: dimensions,
        measure: measures,
      };
    });
  },

  addPivotField(name: string, value: PivotChipData, rows: boolean) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.pivot.rowPage = 1;
      metricsExplorer.pivot.activeCell = null;
      if (value.type === PivotChipType.Measure) {
        metricsExplorer.pivot.columns.measure.push(value);
      } else {
        if (rows) {
          metricsExplorer.pivot.rows.dimension.push(value);
        } else {
          metricsExplorer.pivot.columns.dimension.push(value);
        }
      }
    });
  },

  setPivotExpanded(name: string, expanded: ExpandedState) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.pivot = { ...metricsExplorer.pivot, expanded };
    });
  },

  setPivotComparison(name: string, enableComparison: boolean) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.pivot = { ...metricsExplorer.pivot, enableComparison };
    });
  },

  setPivotSort(name: string, sorting: SortingState) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.pivot = {
        ...metricsExplorer.pivot,
        sorting,
        rowPage: 1,
        expanded: {},
        activeCell: null,
      };
    });
  },

  setPivotColumnPage(name: string, pageNumber: number) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.pivot = {
        ...metricsExplorer.pivot,
        columnPage: pageNumber,
      };
    });
  },

  setPivotRowPage(name: string, pageNumber: number) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.pivot = {
        ...metricsExplorer.pivot,
        rowPage: pageNumber,
      };
    });
  },

  setPivotActiveCell(name: string, rowId: string, columnId: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.pivot.activeCell = { rowId, columnId };
    });
  },

  removePivotActiveCell(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.pivot.activeCell = null;
    });
  },

  createPivot(name: string, rows: PivotRows, columns: PivotColumns) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.activePage = DashboardState_ActivePage.PIVOT;
      metricsExplorer.pivot = {
        ...metricsExplorer.pivot,
        active: true,
        rows,
        columns,
        expanded: {},
        sorting: [],
        columnPage: 1,
        rowPage: 1,
        activeCell: null,
      };
    });
  },

  setExpandedMeasureName(name: string, measureName: string | undefined) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.tdd.expandedMeasureName = measureName;
      if (measureName) {
        metricsExplorer.activePage =
          DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL;
      } else {
        metricsExplorer.activePage = DashboardState_ActivePage.DEFAULT;
      }

      // If going into TDD view and already having a comparison dimension,
      // then set the pinIndex
      if (metricsExplorer.selectedComparisonDimension) {
        metricsExplorer.tdd.pinIndex = getPinIndexForDimension(
          metricsExplorer,
          metricsExplorer.selectedComparisonDimension,
        );
      }
    });
  },

  setPinIndex(name: string, index: number) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.tdd.pinIndex = index;
    });
  },

  setTDDChartType(name: string, type: TDDChart) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.tdd.chartType = type;
    });
  },

  setSelectedTimeRange(name: string, timeRange: DashboardTimeControls) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      setSelectedScrubRange(metricsExplorer, undefined);
      metricsExplorer.selectedTimeRange = timeRange;
    });
  },

  setSelectedScrubRange(name: string, scrubRange: ScrubRange | undefined) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      setSelectedScrubRange(metricsExplorer, scrubRange);
    });
  },

  setMetricDimensionName(name: string, dimensionName: string | null) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedDimensionName = dimensionName ?? undefined;
      if (dimensionName) {
        metricsExplorer.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
      } else {
        metricsExplorer.activePage = DashboardState_ActivePage.DEFAULT;
      }
    });
  },

  setComparisonDimension(name: string, dimensionName: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedComparisonDimension = dimensionName;
      metricsExplorer.tdd.pinIndex = getPinIndexForDimension(
        metricsExplorer,
        dimensionName,
      );
    });
  },

  disableAllComparisons(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedComparisonDimension = undefined;
    });
  },

  setSelectedComparisonRange(
    name: string,
    comparisonTimeRange: DashboardTimeControls,
    metricsViewSpec: V1MetricsViewSpec,
  ) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      if (comparisonTimeRange) {
        metricsExplorer.showTimeComparison = true;
      }
      metricsExplorer.selectedComparisonTimeRange = comparisonTimeRange;
      AdvancedMeasureCorrector.correct(metricsExplorer, metricsViewSpec);
    });
  },

  setTimeZone(name: string, zoneIANA: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      // Reset scrub when timezone changes
      setSelectedScrubRange(metricsExplorer, undefined);

      metricsExplorer.selectedTimezone = zoneIANA;
    });
  },

  displayTimeComparison(name: string, showTimeComparison: boolean) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.showTimeComparison = showTimeComparison;
    });
  },

  selectTimeRange(
    name: string,
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    comparisonTimeRange: DashboardTimeControls | undefined,
    metricsViewSpec: V1MetricsViewSpec,
  ) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      if (!timeRange.name) return;

      // Reset scrub when range changes
      setSelectedScrubRange(metricsExplorer, undefined);

      if (timeRange.name === TimeRangePreset.ALL_TIME) {
        metricsExplorer.showTimeComparison = false;
      }

      metricsExplorer.selectedTimeRange = {
        ...timeRange,
        interval: timeGrain,
      };

      metricsExplorer.selectedComparisonTimeRange = comparisonTimeRange;

      AdvancedMeasureCorrector.correct(metricsExplorer, metricsViewSpec);
    });
  },

  // Temporary method to use this method outside stateManagers availability
  clearAllFilters(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.whereFilter = createAndExpression([]);
      metricsExplorer.dimensionThresholdFilters = [];
    });
  },

  remove(name: string) {
    update((state) => {
      delete state.entities[name];
      return state;
    });
  },
};

export const metricsExplorerStore: Readable<MetricsExplorerStoreType> &
  typeof metricsViewReducers = {
  subscribe,
  ...metricsViewReducers,
};

export function useExploreState(name: string): Readable<MetricsExplorerEntity> {
  return derived(metricsExplorerStore, ($store) => {
    return $store.entities[name];
  });
}

export function sortTypeForContextColumnType(
  contextCol: LeaderboardContextColumn,
): SortType {
  const sortType = {
    [LeaderboardContextColumn.DELTA_PERCENT]: SortType.DELTA_PERCENT,
    [LeaderboardContextColumn.DELTA_ABSOLUTE]: SortType.DELTA_ABSOLUTE,
    [LeaderboardContextColumn.PERCENT]: SortType.PERCENT,
    [LeaderboardContextColumn.HIDDEN]: SortType.VALUE,
  }[contextCol];

  // Note: the above map needs to be EXHAUSTIVE over
  // LeaderboardContextColumn variants. If we ever add a new
  // context column type, we need to add it to the map above.
  // Otherwise, we will throw an error here.
  if (!sortType) {
    throw new Error(`Invalid context column type: ${contextCol}`);
  }
  return sortType;
}

function setSelectedScrubRange(
  metricsExplorer: MetricsExplorerEntity,
  scrubRange: ScrubRange | undefined,
) {
  if (scrubRange === undefined) {
    metricsExplorer.lastDefinedScrubRange = undefined;
  } else if (!scrubRange.isScrubbing && scrubRange?.start && scrubRange?.end) {
    metricsExplorer.lastDefinedScrubRange = scrubRange;
  }

  metricsExplorer.selectedScrubRange = scrubRange;
}

function getPinIndexForDimension(
  metricsExplorer: MetricsExplorerEntity,
  dimensionName: string,
) {
  const dimensionEntryIndex = getWhereFilterExpressionIndex({
    dashboard: metricsExplorer,
  })(dimensionName);
  if (dimensionEntryIndex === undefined || dimensionEntryIndex === -1)
    return -1;

  const dimExpr =
    metricsExplorer.whereFilter.cond?.exprs?.[dimensionEntryIndex];
  if (!dimExpr?.cond?.exprs?.length) return -1;

  // 1st entry in the expression is the identifier. hence the -2 here.
  return dimExpr.cond.exprs.length - 2;
}

export const dimensionSearchText = writable("");
