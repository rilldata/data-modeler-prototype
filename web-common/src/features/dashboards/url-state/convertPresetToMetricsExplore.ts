import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { convertURLToExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/convertURLToExplorePreset";
import {
  getMultiFieldError,
  getSingleFieldError,
} from "@rilldata/web-common/features/dashboards/url-state/error-message-helpers";
import {
  FromURLParamTDDChartMap,
  FromURLParamTimeDimensionMap,
  FromURLParamTimeGrainMap,
  ToActivePageViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import {
  getMapFromArray,
  getMissingValues,
} from "@rilldata/web-common/lib/arrayUtils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type MetricsViewSpecDimensionV2,
  type MetricsViewSpecMeasureV2,
  V1ExploreComparisonMode,
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

export function convertURLToMetricsExplore(
  searchParams: URLSearchParams,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  basePreset: V1ExplorePreset,
) {
  const errors: Error[] = [];
  const { preset, errors: errorsFromPreset } = convertURLToExplorePreset(
    searchParams,
    metricsView,
    explore,
    basePreset,
  );
  errors.push(...errorsFromPreset);
  const { partialExploreState, errors: errorsFromEntity } =
    convertPresetToMetricsExplore(metricsView, explore, preset);
  errors.push(...errorsFromEntity);
  return { partialExploreState, errors };
}

/**
 * Converts a V1ExplorePreset to our internal metrics explore state.
 * V1ExplorePreset could come from url state, bookmark, alert or report.
 */
export function convertPresetToMetricsExplore(
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  const partialExploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];

  const measures = getMapFromArray(
    metricsView.measures?.filter((m) => explore.measures?.includes(m.name!)) ??
      [],
    (m) => m.name!,
  );
  const dimensions = getMapFromArray(
    metricsView.dimensions?.filter((d) =>
      explore.dimensions?.includes(d.name!),
    ) ?? [],
    (d) => d.name!,
  );

  if (preset.view) {
    partialExploreState.activePage = Number(
      ToActivePageViewMap[preset.view] ?? "0",
    );
  }

  if (preset.where) {
    const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
      preset.where,
    );
    partialExploreState.whereFilter = dimensionFilters;
    partialExploreState.dimensionThresholdFilters = dimensionThresholdFilters;
  }

  const { partialExploreState: trPartialState, errors: trErrors } =
    fromTimeRangesParams(preset, dimensions);
  Object.assign(partialExploreState, trPartialState);
  errors.push(...trErrors);

  const { partialExploreState: ovPartialState, errors: ovErrors } =
    fromOverviewUrlParams(measures, dimensions, explore, preset);
  Object.assign(partialExploreState, ovPartialState);
  errors.push(...ovErrors);

  const { partialExploreState: tddPartialState, errors: tddErrors } =
    fromTimeDimensionUrlParams(measures, preset);
  Object.assign(partialExploreState, tddPartialState);
  errors.push(...tddErrors);

  const { partialExploreState: pivotPartialState, errors: pivotErrors } =
    fromPivotUrlParams(measures, dimensions, preset);
  Object.assign(partialExploreState, pivotPartialState);
  errors.push(...pivotErrors);

  return { partialExploreState, errors };
}

function fromTimeRangesParams(
  preset: V1ExplorePreset,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
) {
  const partialExploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];

  if (preset.timeRange) {
    const { timeRange, error } = fromTimeRangeUrlParam(preset.timeRange);
    if (error) errors.push(error);
    partialExploreState.selectedTimeRange = timeRange;
  }

  if (preset.timeGrain && partialExploreState.selectedTimeRange) {
    partialExploreState.selectedTimeRange.interval =
      FromURLParamTimeGrainMap[preset.timeGrain];
  }

  if (preset.timezone) {
    partialExploreState.selectedTimezone = preset.timezone;
  }

  if (preset.compareTimeRange) {
    const { timeRange, error } = fromTimeRangeUrlParam(preset.compareTimeRange);
    if (error) errors.push(error);
    partialExploreState.selectedComparisonTimeRange = timeRange;
    partialExploreState.showTimeComparison = true;
    // unset compare dimension
    partialExploreState.selectedComparisonDimension = "";
  }

  if (preset.comparisonDimension) {
    if (dimensions.has(preset.comparisonDimension)) {
      partialExploreState.selectedComparisonDimension =
        preset.comparisonDimension;
      // unset compare time ranges
      partialExploreState.selectedComparisonTimeRange = undefined;
      partialExploreState.showTimeComparison = false;
    } else {
      errors.push(
        getSingleFieldError("compare dimension", preset.comparisonDimension),
      );
    }
  }

  if (
    preset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE
  ) {
    // unset all comparison setting if mode is none
    partialExploreState.selectedComparisonTimeRange = undefined;
    partialExploreState.selectedComparisonDimension = "";
    partialExploreState.showTimeComparison = false;
  }

  return { partialExploreState, errors };
}

function fromTimeRangeUrlParam(tr: string): {
  timeRange?: DashboardTimeControls;
  error?: Error;
} {
  // TODO: validation
  return {
    timeRange: {
      name: tr,
    } as DashboardTimeControls,
  };

  // return {
  //   error: new Error(`unknown time range: ${tr}`),
  // };
}

function fromOverviewUrlParams(
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
  explore: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  const partialExploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];

  if (preset.measures?.length) {
    const selectedMeasures = preset.measures.filter((m) => measures.has(m));
    const missingMeasures = getMissingValues(selectedMeasures, preset.measures);
    if (missingMeasures.length) {
      errors.push(getMultiFieldError("measure", missingMeasures));
    }

    partialExploreState.allMeasuresVisible =
      selectedMeasures.length === explore.measures?.length;
    partialExploreState.visibleMeasureKeys = new Set(selectedMeasures);
  }

  if (preset.dimensions?.length) {
    const selectedDimensions = preset.dimensions.filter((d) =>
      dimensions.has(d),
    );
    const missingDimensions = getMissingValues(
      selectedDimensions,
      preset.dimensions,
    );
    if (missingDimensions.length) {
      errors.push(getMultiFieldError("dimension", missingDimensions));
    }

    partialExploreState.allDimensionsVisible =
      selectedDimensions.length === explore.dimensions?.length;
    partialExploreState.visibleDimensionKeys = new Set(selectedDimensions);
  }

  if (preset.overviewSortBy) {
    if (measures.has(preset.overviewSortBy)) {
      partialExploreState.leaderboardMeasureName = preset.overviewSortBy;
    } else {
      errors.push(
        getSingleFieldError("sort by measure", preset.overviewSortBy),
      );
    }
  }

  if (preset.overviewSortAsc !== undefined) {
    partialExploreState.sortDirection = preset.overviewSortAsc
      ? SortDirection.ASCENDING
      : SortDirection.DESCENDING;
  }

  if (preset.overviewExpandedDimension !== undefined) {
    if (preset.overviewExpandedDimension === "") {
      partialExploreState.selectedDimensionName = "";
      // if preset didnt have a view then this is a dimension table unset.
      if (
        preset.view === V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED ||
        preset.view === undefined
      ) {
        partialExploreState.activePage = DashboardState_ActivePage.DEFAULT;
      }
    } else if (dimensions.has(preset.overviewExpandedDimension)) {
      partialExploreState.selectedDimensionName =
        preset.overviewExpandedDimension;
      if (
        preset.view === V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW ||
        preset.view === V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED ||
        preset.view === undefined
      ) {
        partialExploreState.activePage =
          DashboardState_ActivePage.DIMENSION_TABLE;
      }
    } else {
      errors.push(
        getSingleFieldError(
          "expanded dimension",
          preset.overviewExpandedDimension,
        ),
      );
    }
  }

  return { partialExploreState, errors };
}

function fromTimeDimensionUrlParams(
  measures: Map<string, MetricsViewSpecMeasureV2>,
  preset: V1ExplorePreset,
): {
  partialExploreState: Partial<MetricsExplorerEntity>;
  errors: Error[];
} {
  if (preset.timeDimensionMeasure === undefined) {
    return {
      partialExploreState: {},
      errors: [],
    };
  }

  const errors: Error[] = [];

  let expandedMeasureName = preset.timeDimensionMeasure;
  if (expandedMeasureName && !measures.has(expandedMeasureName)) {
    expandedMeasureName = "";
    errors.push(getSingleFieldError("expanded measure", expandedMeasureName));
  }

  // unset
  if (expandedMeasureName === "") {
    return {
      partialExploreState: {
        tdd: {
          expandedMeasureName: "",
          chartType: TDDChart.DEFAULT,
          pinIndex: -1,
        },
      },
      errors,
    };
  }

  const partialExploreState: Partial<MetricsExplorerEntity> = {
    tdd: {
      expandedMeasureName,
      chartType: preset.timeDimensionChartType
        ? FromURLParamTDDChartMap[preset.timeDimensionChartType]
        : TDDChart.DEFAULT,
      pinIndex: preset.timeDimensionPin ? Number(preset.timeDimensionPin) : -1,
    },
  };

  return {
    partialExploreState,
    errors,
  };
}

function fromPivotUrlParams(
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
  preset: V1ExplorePreset,
): {
  partialExploreState: Partial<MetricsExplorerEntity>;
  errors: Error[];
} {
  const errors: Error[] = [];

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
        title: m.displayName || m.name || "Unknown",
        type: PivotChipType.Measure,
      };
    }

    if (dimensions.has(entry)) {
      const d = dimensions.get(entry)!;
      return {
        id: entry,
        title: d.displayName || d.name || "Unknown",
        type: PivotChipType.Dimension,
      };
    }

    errors.push(getSingleFieldError("pivot entry", entry));

    return undefined;
  };

  let hasSomePivotFields = false;

  const rowDimensions: PivotChipData[] = [];
  if (preset.pivotRows) {
    preset.pivotRows.forEach((pivotRow) => {
      const chip = mapPivotEntry(pivotRow);
      if (!chip) return;
      rowDimensions.push(chip);
    });
    hasSomePivotFields = true;
  }

  const colMeasures: PivotChipData[] = [];
  const colDimensions: PivotChipData[] = [];
  if (preset.pivotCols) {
    preset.pivotCols.forEach((pivotRow) => {
      const chip = mapPivotEntry(pivotRow);
      if (!chip) return;
      if (chip.type === PivotChipType.Measure) {
        colMeasures.push(chip);
      } else {
        colDimensions.push(chip);
      }
    });
    hasSomePivotFields = true;
  }

  if (!hasSomePivotFields) {
    return {
      partialExploreState: {},
      errors,
    };
  }

  return {
    partialExploreState: {
      pivot: {
        active: preset.view === V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
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
        enableComparison: true,
        activeCell: null,
        rowJoinType: "nest",
      },
    },
    errors,
  };
}
