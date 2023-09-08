import {
  V1MetricsViewComparisonSortType,
  type V1MetricsViewComparisonRow,
  type V1MetricsViewComparisonValue,
} from "@rilldata/web-common/runtime-client";
import { PERC_DIFF } from "../../../components/data-types/type-utils";
import {
  FormatPreset,
  formatMeasurePercentageDifference,
  humanizeDataType,
} from "../humanize-numbers";
import { LeaderboardContextColumn } from "../leaderboard-context-column";
import { SortType } from "../proto-state/derived-types";
import type { NumberParts } from "@rilldata/web-common/lib/number-formatting/humanizer-types";

export function getFormatterValueForPercDiff(pct: number | null) {
  if (pct === null) return PERC_DIFF.PREV_VALUE_NO_DATA;
  return formatMeasurePercentageDifference(pct);
}

/**
 * A `V1MetricsViewComparisonRow` basically represents a row of data
 * in the *dimension detail table*, NOT in the leaderboard. Therefore,
 * to convert to rows of leaderboard data, we need to extract a single
 * measure from the dimension table shaped data (namely, the active
 * measure in the leaderboard).
 * @param params
 */
export function getLabeledComparisonFromComparisonRow(
  row: V1MetricsViewComparisonRow,
  measureName: string | number
): ComparisonValueWithLabel {
  const measure = row.measureValues?.find((v) => v.measureName === measureName);
  if (!measure) {
    throw new Error(
      `Could not find measure ${measureName} in row ${JSON.stringify(row)}`
    );
  }
  return {
    dimensionValue: row.dimensionValue as string | number,
    ...measure,
  };
}

export type LeaderboardItemData = {
  // The dimension value label to be shown in the leaderboard
  dimensionValue: string | number;

  // main value to be shown in the leaderboard
  value: number | null;

  // percent of total for summable measures; null if not summable
  pctOfTotal: number | null;

  // The value from the comparison period.
  // Techinally this might not be a "previous value" but
  // we use that name as a shorthand, since it's the most
  // common use case.
  prevValue: number | null;

  // the relative change from the previous value
  // note that this needs to be multiplied by 100 to get
  // the percentage change
  deltaRel: number | null;

  // the absolute change from the previous value
  deltaAbs: number | null;

  // selection is not enough to determine if the item is included
  // or excluded; for that we need to know the leaderboard's
  // include/exclude state
  selected: boolean;
};

function cleanUpComparisonValue(
  v: ComparisonValueWithLabel,
  total: number | null,
  selected: boolean
): LeaderboardItemData {
  if (!(Number.isFinite(v.baseValue) || v.baseValue === null)) {
    throw new Error(
      `Leaderboards only implemented for numeric baseValues or missing data (null). Got: ${JSON.stringify(
        v
      )}`
    );
  }
  const value = v.baseValue as number;

  return {
    dimensionValue: v.dimensionValue,
    value,
    pctOfTotal: total && value ? (value / total) * 100 : null,
    prevValue: Number.isFinite(v.comparisonValue)
      ? (v.comparisonValue as number)
      : null,
    deltaRel: Number.isFinite(v.deltaRel) ? (v.deltaRel as number) : null,
    deltaAbs: Number.isFinite(v.deltaAbs) ? (v.deltaAbs as number) : null,

    selected,
  };
}

/**
 * A `V1MetricsViewComparisonValue` augmented with the dimension
 * value that it corresponds to.
 */
type ComparisonValueWithLabel = V1MetricsViewComparisonValue & {
  dimensionValue: string | number;
};

/**
 *
 * @param values
 * @param selectedValues
 * @param total: the total of the measure for the current period,
 * or null if the measure is not valid_percent_of_total
 * @returns
 */
export function prepareLeaderboardItemData(
  values: ComparisonValueWithLabel[],
  numberAboveTheFold: number,
  selectedValues: (string | number)[],
  total: number | null
): {
  aboveTheFold: LeaderboardItemData[];
  selectedBelowTheFold: LeaderboardItemData[];
  noAvailableValues: boolean;
  showExpandTable: boolean;
} {
  const aboveTheFold: LeaderboardItemData[] = [];
  const selectedBelowTheFold: LeaderboardItemData[] = [];
  let selectedValuesCopy = [...selectedValues];
  // console.log({ values, len: values.length, selectedValues });
  values.forEach((v, i) => {
    // console.log({ dimval: v.dimensionValue, selectedValues });
    const selected =
      selectedValuesCopy.findIndex((value) => value === v.dimensionValue) >= 0;
    // drop the value from the selectedValues array so that we'll
    // have any left over values that were selected but not included
    // in the results returned by the API
    if (selected)
      selectedValuesCopy = selectedValuesCopy.filter(
        (value) => value !== v.dimensionValue
      );
    if (i < numberAboveTheFold) {
      aboveTheFold.push(cleanUpComparisonValue(v, total, selected));
    } else if (selected) {
      selectedBelowTheFold.push(cleanUpComparisonValue(v, total, selected));
    }
  });

  // FIXME: note that it is possible for some values to be selected
  // but not included in the results returned by the API, for example
  // if a dimension value is selected and then a filter is applied
  // that pushes it out of the top N. In that case, we will follow
  // the previous strategy, and just push a dummy value with only
  // the dimension value and nulls for all measure values.
  selectedValuesCopy.forEach((v) => {
    selectedBelowTheFold.push({
      dimensionValue: v,
      selected: true,
      value: null,
      pctOfTotal: null,
      prevValue: null,
      deltaRel: null,
      deltaAbs: null,
    });
  });

  const noAvailableValues = values.length === 0;
  const showExpandTable = values.length > numberAboveTheFold;

  return {
    aboveTheFold,
    selectedBelowTheFold,
    noAvailableValues,
    showExpandTable,
  };
}

/**
 * Returns the formatted value for the context column
 * accounting for the context column type.
 */
export function formatContextColumnValue(
  itemData: LeaderboardItemData,
  contextType: LeaderboardContextColumn,
  formatPreset: FormatPreset
): string | NumberParts | PERC_DIFF.PREV_VALUE_NO_DATA {
  if (contextType === LeaderboardContextColumn.DELTA_PERCENT) {
    return getFormatterValueForPercDiff(itemData.deltaRel);
  } else if (contextType === LeaderboardContextColumn.PERCENT) {
    return getFormatterValueForPercDiff(itemData.pctOfTotal);
  } else if (contextType === LeaderboardContextColumn.DELTA_ABSOLUTE) {
    return humanizeDataType(itemData.deltaAbs, formatPreset);
  } else {
    return "";
  }
}
export const contextColumnWidth = (
  contextType: LeaderboardContextColumn
): string => {
  if (contextType === LeaderboardContextColumn.DELTA_PERCENT) {
    return "44px";
  } else if (contextType === LeaderboardContextColumn.PERCENT) {
    return "44px";
  } else if (contextType === LeaderboardContextColumn.DELTA_ABSOLUTE) {
    return "56px";
  } else {
    return "0px";
  }
};

export function getQuerySortType(sortType: SortType) {
  return (
    {
      [SortType.VALUE]:
        V1MetricsViewComparisonSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE,

      [SortType.DELTA_ABSOLUTE]:
        V1MetricsViewComparisonSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA,

      [SortType.DELTA_PERCENT]:
        V1MetricsViewComparisonSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_REL_DELTA,

      // NOTE: sorting by percent-of-total has the same effect
      // as sorting by base value
      [SortType.PERCENT]:
        V1MetricsViewComparisonSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE,

      // NOTE: UNSPECIFIED is not actually a valid sort type,
      // but it is required by protobuf serialization
      [SortType.UNSPECIFIED]:
        V1MetricsViewComparisonSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE,

      // FIXME: sort by dimension value is not yet implemented,
      // for now fall back to sorting by base value
      [SortType.DIMENSION]:
        V1MetricsViewComparisonSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE,
    }[sortType] ||
    V1MetricsViewComparisonSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE
  );
}
