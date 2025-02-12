import { stripMeasureSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { base64ToProto } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import {
  createAndExpression,
  filterIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { convertLegacyStateToExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/convertLegacyStateToExplorePreset";
import { CustomTimeRangeRegex } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import {
  getMultiFieldError,
  getSingleFieldError,
} from "@rilldata/web-common/features/dashboards/url-state/error-message-helpers";
import {
  convertFilterParamToExpression,
  stripParserError,
} from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import {
  FromURLParamsSortTypeMap,
  FromURLParamTimeDimensionMap,
  FromURLParamTimeGrainMap,
  FromURLParamTimeRangePresetMap,
  FromURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { validateRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import {
  getMapFromArray,
  getMissingValues,
} from "@rilldata/web-common/lib/arrayUtils";
import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
import { validateISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { DashboardState } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type MetricsViewSpecDimensionV2,
  type MetricsViewSpecMeasureV2,
  V1ExploreComparisonMode,
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
  type V1Expression,
  type V1MetricsViewSpec,
  V1Operation,
} from "@rilldata/web-common/runtime-client";

export function convertURLToExplorePreset(
  searchParams: URLSearchParams,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  defaultExplorePreset: V1ExplorePreset,
) {
  const preset: V1ExplorePreset = {
    ...defaultExplorePreset,
  };
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

  // Support legacy dashboard param.
  // This will be applied 1st so that any newer params added can be applied as well.
  if (searchParams.has(ExploreStateURLParams.LegacyProtoState)) {
    const legacyState = searchParams.get(
      ExploreStateURLParams.LegacyProtoState,
    ) as string;
    const { preset: presetFromLegacyState, errors: errorsFromLegacyState } =
      fromLegacyStateUrlParam(
        legacyState,
        metricsView,
        explore,
        defaultExplorePreset,
      );
    Object.assign(preset, presetFromLegacyState);
    errors.push(...errorsFromLegacyState);
  }

  if (searchParams.has(ExploreStateURLParams.WebView)) {
    const view = searchParams.get(ExploreStateURLParams.WebView) as string;
    if (view in FromURLParamViewMap) {
      preset.view = FromURLParamViewMap[view];
    } else {
      errors.push(getSingleFieldError("view", view));
    }
  }

  if (searchParams.has(ExploreStateURLParams.Filters)) {
    const { expr, errors: filterErrors } = fromFilterUrlParam(
      searchParams.get(ExploreStateURLParams.Filters) as string,
      measures,
      dimensions,
    );
    if (filterErrors) errors.push(...filterErrors);
    if (expr) preset.where = expr;
  }

  const { preset: trPreset, errors: trErrors } = fromTimeRangesParams(
    searchParams,
    dimensions,
  );
  Object.assign(preset, trPreset);
  errors.push(...trErrors);

  // only extract params if the view is explicitly set to the relevant one
  switch (preset.view) {
    case V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE:
    case V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED: {
      const { preset: ovPreset, errors: ovErrors } = fromExploreUrlParams(
        searchParams,
        measures,
        dimensions,
        explore,
      );
      Object.assign(preset, ovPreset);
      errors.push(...ovErrors);
      break;
    }

    case V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION: {
      const { preset: tddPreset, errors: tddErrors } =
        fromTimeDimensionUrlParams(searchParams, measures);
      Object.assign(preset, tddPreset);
      errors.push(...tddErrors);
      break;
    }

    case V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT: {
      const { preset: pivotPreset, errors: pivotErrors } = fromPivotUrlParams(
        searchParams,
        measures,
        dimensions,
      );
      Object.assign(preset, pivotPreset);
      errors.push(...pivotErrors);
      break;
    }
  }

  return { preset, errors };
}

function fromLegacyStateUrlParam(
  legacyState: string,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  defaultExplorePreset: V1ExplorePreset,
) {
  try {
    legacyState = legacyState.includes("%")
      ? decodeURIComponent(legacyState)
      : legacyState;
    const legacyDashboardState = DashboardState.fromBinary(
      base64ToProto(legacyState),
    );

    return convertLegacyStateToExplorePreset(
      legacyDashboardState,
      metricsView,
      explore,
      defaultExplorePreset,
    );
  } catch (e) {
    return {
      preset: {},
      errors: [e], // TODO: parse and show meaningful error
    };
  }
}

function fromFilterUrlParam(
  filter: string,
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
): {
  expr?: V1Expression;
  errors?: Error[];
} {
  try {
    let expr = convertFilterParamToExpression(filter);
    if (!expr) {
      return {
        expr: createAndExpression([]),
        errors: [new Error("Failed to parse filter: " + filter)],
      };
    }

    // if root is not AND/OR then add AND
    if (
      expr?.cond?.op !== V1Operation.OPERATION_AND &&
      expr?.cond?.op !== V1Operation.OPERATION_OR
    ) {
      expr = createAndExpression([expr]);
    }
    const errors: Error[] = [];
    const missingDims: string[] = [];
    const missingFields: string[] = [];
    expr =
      filterIdentifiers(expr, (e, ident) => {
        if (
          // these we are sure are dimensions so add errors as "missing dimension"
          e.cond?.op === V1Operation.OPERATION_IN ||
          e.cond?.op === V1Operation.OPERATION_NIN ||
          !!e.subquery
        ) {
          if (dimensions.has(ident)) {
            return true;
          }
          missingDims.push(ident);
          return false;
        }

        if (
          measures.has(ident) ||
          measures.has(stripMeasureSuffix(ident)) ||
          dimensions.has(ident)
        ) {
          return true;
        }
        missingFields.push(ident);

        return false;
      }) ?? createAndExpression([]);
    if (missingDims.length) {
      errors.push(getMultiFieldError("filter dimension", missingDims));
    }
    if (missingFields.length) {
      errors.push(getMultiFieldError("filter field", missingFields));
    }
    return { expr, errors };
  } catch (e) {
    return {
      errors: [new Error("Selected filter is invalid: " + stripParserError(e))],
    };
  }
}

function fromTimeRangesParams(
  searchParams: URLSearchParams,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  if (searchParams.has(ExploreStateURLParams.TimeRange)) {
    const tr = searchParams.get(ExploreStateURLParams.TimeRange) as string;
    if (
      tr in FromURLParamTimeRangePresetMap ||
      validateISODuration(tr) ||
      CustomTimeRangeRegex.test(tr)
    ) {
      preset.timeRange = tr;
    } else {
      const rillTimeError = validateRillTime(tr);
      if (rillTimeError) {
        errors.push(getSingleFieldError("time range", tr));
      } else {
        preset.timeRange = tr;
      }
    }
  }

  if (searchParams.has(ExploreStateURLParams.TimeZone)) {
    preset.timezone = searchParams.get(
      ExploreStateURLParams.TimeZone,
    ) as string;
  }

  if (searchParams.has(ExploreStateURLParams.ComparisonTimeRange)) {
    const ctr = searchParams.get(
      ExploreStateURLParams.ComparisonTimeRange,
    ) as string;
    if (ctr in TIME_COMPARISON || CustomTimeRangeRegex.test(ctr)) {
      preset.compareTimeRange = ctr;
      preset.comparisonMode ??=
        V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME;
    } else if (ctr == "") {
      preset.compareTimeRange = "";
    } else {
      errors.push(getSingleFieldError("compare time range", ctr));
    }
  }

  if (searchParams.has(ExploreStateURLParams.TimeGrain)) {
    const tg = searchParams.get(ExploreStateURLParams.TimeGrain) as string;
    if (tg in FromURLParamTimeGrainMap) {
      preset.timeGrain = tg;
    } else {
      errors.push(getSingleFieldError("time grain", tg));
    }
  }

  if (searchParams.has(ExploreStateURLParams.ComparisonDimension)) {
    const comparisonDimension = searchParams.get(
      ExploreStateURLParams.ComparisonDimension,
    ) as string;
    // unsetting a default from url
    if (comparisonDimension === "") {
      preset.comparisonDimension = "";
      preset.comparisonMode ??=
        V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE;
    } else if (dimensions.has(comparisonDimension)) {
      preset.comparisonDimension = comparisonDimension;
      preset.comparisonMode ??=
        V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION;
    } else {
      errors.push(
        getSingleFieldError("compare dimension", comparisonDimension),
      );
    }
  }

  if (searchParams.has(ExploreStateURLParams.HighlightedTimeRange)) {
    const selectTr = searchParams.get(
      ExploreStateURLParams.HighlightedTimeRange,
    ) as string;
    if (CustomTimeRangeRegex.test(selectTr) || selectTr === "") {
      preset.selectTimeRange = selectTr;
    } else {
      errors.push(getSingleFieldError("highlighted time range", selectTr));
    }
  }
  return { preset, errors };
}

function fromExploreUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
  explore: V1ExploreSpec,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  if (searchParams.has(ExploreStateURLParams.VisibleMeasures)) {
    const mes = searchParams.get(
      ExploreStateURLParams.VisibleMeasures,
    ) as string;
    if (mes === "*") {
      preset.measures = explore.measures ?? [];
    } else {
      const selectedMeasures = mes.split(",").filter((m) => measures.has(m));
      preset.measures = selectedMeasures;
      const missingMeasures = getMissingValues(
        selectedMeasures,
        mes.split(","),
      );
      if (missingMeasures.length) {
        errors.push(getMultiFieldError("measure", missingMeasures));
      }
    }
  }

  if (searchParams.has(ExploreStateURLParams.VisibleDimensions)) {
    const dims = searchParams.get(
      ExploreStateURLParams.VisibleDimensions,
    ) as string;
    if (dims === "*") {
      preset.dimensions = explore.dimensions ?? [];
    } else {
      const selectedDimensions = dims
        .split(",")
        .filter((d) => dimensions.has(d));
      preset.dimensions = selectedDimensions;
      const missingDimensions = getMissingValues(
        selectedDimensions,
        dims.split(","),
      );
      if (missingDimensions.length) {
        errors.push(getMultiFieldError("dimension", missingDimensions));
      }
    }
  }

  if (searchParams.has(ExploreStateURLParams.ExpandedDimension)) {
    const dim = searchParams.get(
      ExploreStateURLParams.ExpandedDimension,
    ) as string;
    if (
      dimensions.has(dim) ||
      // we are unsetting from a default preset
      dim === ""
    ) {
      preset.exploreExpandedDimension = dim;
    } else {
      errors.push(getSingleFieldError("expanded dimension", dim));
    }
  }

  if (searchParams.has(ExploreStateURLParams.SortBy)) {
    const sortBy = searchParams.get(ExploreStateURLParams.SortBy) as string;
    if (measures.has(sortBy)) {
      if (
        (preset.measures && preset.measures.includes(sortBy)) ||
        !preset.measures
      ) {
        preset.exploreSortBy = sortBy;
      } else {
        errors.push(
          getSingleFieldError("sort by measure", sortBy, "It is hidden."),
        );
      }
    } else {
      errors.push(getSingleFieldError("sort by measure", sortBy));
    }
  }

  if (searchParams.has(ExploreStateURLParams.SortDirection)) {
    preset.exploreSortAsc =
      (searchParams.get(ExploreStateURLParams.SortDirection) as string) ===
      "ASC";
  }

  if (searchParams.has(ExploreStateURLParams.SortType)) {
    const sortType = searchParams.get(ExploreStateURLParams.SortType) as string;
    if (sortType in FromURLParamsSortTypeMap) {
      preset.exploreSortType = FromURLParamsSortTypeMap[sortType];
    } else {
      errors.push(getSingleFieldError("sort type", sortType));
    }
  }

  return { preset, errors };
}

function fromTimeDimensionUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasureV2>,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  if (searchParams.has(ExploreStateURLParams.ExpandedMeasure)) {
    const mes = searchParams.get(
      ExploreStateURLParams.ExpandedMeasure,
    ) as string;
    if (
      measures.has(mes) ||
      // we are unsetting from a default preset
      mes === ""
    ) {
      preset.timeDimensionMeasure = mes;
    } else {
      errors.push(getSingleFieldError("expanded measure", mes));
    }
  }
  if (searchParams.has(ExploreStateURLParams.ChartType)) {
    preset.timeDimensionChartType = searchParams.get(
      ExploreStateURLParams.ChartType,
    ) as string;
  }
  if (searchParams.has(ExploreStateURLParams.Pin)) {
    preset.timeDimensionPin = true;
  }

  return {
    preset,
    errors,
  };
}

function fromPivotUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  if (searchParams.has(ExploreStateURLParams.PivotRows)) {
    const rows = (
      searchParams.get(ExploreStateURLParams.PivotRows) as string
    ).split(",");
    const validRows = rows.filter(
      (r) => dimensions.has(r) || r in FromURLParamTimeDimensionMap,
    );
    preset.pivotRows = validRows;
    const missingRows = getMissingValues(validRows, rows);
    if (missingRows.length) {
      errors.push(getMultiFieldError("pivot row", missingRows));
    }
  }

  if (searchParams.has(ExploreStateURLParams.PivotColumns)) {
    const cols = (
      searchParams.get(ExploreStateURLParams.PivotColumns) as string
    ).split(",");
    const validCols = cols.filter(
      (c) =>
        dimensions.has(c) ||
        measures.has(c) ||
        c in FromURLParamTimeDimensionMap,
    );
    preset.pivotCols = validCols;
    const missingCols = getMissingValues(validCols, cols);
    if (missingCols.length) {
      errors.push(getMultiFieldError("pivot column", missingCols));
    }
  }

  if (searchParams.has(ExploreStateURLParams.SortBy)) {
    const sortBy = searchParams.get(ExploreStateURLParams.SortBy) as string;
    preset.pivotSortBy = sortBy;
  }

  if (searchParams.has(ExploreStateURLParams.SortDirection)) {
    preset.pivotSortAsc =
      (searchParams.get(ExploreStateURLParams.SortDirection) as string) ===
      "ASC";
  }

  // TODO: other fields like expanded state and pin are not supported right now
  return { preset, errors };
}
