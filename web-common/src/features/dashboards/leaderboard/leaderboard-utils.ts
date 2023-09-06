import type {
  V1MetricsViewComparisonRow,
  V1MetricsViewComparisonValue,
} from "@rilldata/web-common/runtime-client";
import { PERC_DIFF } from "../../../components/data-types/type-utils";
import {
  FormatPreset,
  formatMeasurePercentageDifference,
  humanizeDataType,
} from "../humanize-numbers";
import { LeaderboardContextColumn } from "../leaderboard-context-column";

export function getFormatterValueForPercDiff(numerator, denominator) {
  if (denominator === 0) return PERC_DIFF.PREV_VALUE_ZERO;
  if (!denominator) return PERC_DIFF.PREV_VALUE_NO_DATA;
  if (numerator === null || numerator === undefined)
    return PERC_DIFF.CURRENT_VALUE_NO_DATA;

  const percDiff = numerator / denominator;
  return formatMeasurePercentageDifference(percDiff);
}

export type LeaderboardItemData = {
  label: string | number;
  // main value to be shown in the leaderboard
  value: number;
  // the comparison value, which may be either the previous value
  // (used to calculate the absolute or percentage change) or
  // the measure total (used to calculate the percentage of total)
  comparisonValue: number;
  // selection is not enough to determine if the item is included
  // or excluded; for that we need to know the leaderboard's
  // include/exclude state
  selected: boolean;
};

export function prepareLeaderboardItemData(
  values: { value: number; label: string | number }[],
  selectedValues: (string | number)[],
  comparisonMap: Map<string | number, number>
): LeaderboardItemData[] {
  return values.map((v) => {
    const selected =
      selectedValues.findIndex((value) => value === v.label) >= 0;
    const comparisonValue = comparisonMap.get(v.label);

    return {
      ...v,
      selected,
      comparisonValue,
    };
  });
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

export type LeaderboardItemData2 = {
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

  // the % change from the previous value
  deltaPct: number | null;

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
): LeaderboardItemData2 {
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
    deltaPct: Number.isFinite(v.deltaRel) ? (v.deltaRel as number) * 100 : null,
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
export function prepareLeaderboardItemData2(
  values: ComparisonValueWithLabel[],
  numberAboveTheFold: number,
  selectedValues: (string | number)[],
  total: number | null
): {
  aboveTheFold: LeaderboardItemData2[];
  selectedBelowTheFold: LeaderboardItemData2[];
} {
  const aboveTheFold: LeaderboardItemData2[] = [];
  const selectedBelowTheFold: LeaderboardItemData2[] = [];
  // console.log({ values, len: values.length, selectedValues });
  values.forEach((v, i) => {
    // console.log({ dimval: v.dimensionValue, selectedValues });
    const selected =
      selectedValues.findIndex((value) => value === v.dimensionValue) >= 0;
    // drop the value from the selectedValues array so that we'll
    // have any left over values that were selected but not included
    // in the results returned by the API
    if (selected)
      selectedValues = selectedValues.filter(
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
  selectedValues.forEach((v) => {
    selectedBelowTheFold.push({
      dimensionValue: v,
      selected: true,
      value: null,
      pctOfTotal: null,
      prevValue: null,
      deltaPct: null,
      deltaAbs: null,
    });
  });

  return { aboveTheFold, selectedBelowTheFold };
}

/**
 * Returns the formatted value for the context column
 * given the
 * accounting for the context column type.
 */
export function formatContextColumnValue(
  itemData: LeaderboardItemData,
  unfilteredTotal: number,
  contextType: LeaderboardContextColumn,
  formatPreset: FormatPreset
): string {
  const { value, comparisonValue } = itemData;
  let formattedValue = "";

  if (contextType === LeaderboardContextColumn.DELTA_PERCENT) {
    formattedValue = getFormatterValueForPercDiff(
      value && comparisonValue ? value - comparisonValue : null,
      comparisonValue
    );
  } else if (contextType === LeaderboardContextColumn.PERCENT) {
    formattedValue = getFormatterValueForPercDiff(value, unfilteredTotal);
  } else if (contextType === LeaderboardContextColumn.DELTA_ABSOLUTE) {
    formattedValue = humanizeDataType(
      value && comparisonValue ? value - comparisonValue : null,
      formatPreset
    );
  } else {
    formattedValue = "";
  }
  return formattedValue;
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
