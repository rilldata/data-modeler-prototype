import { PERC_DIFF } from "../../../components/data-types/type-utils";
import { formatMeasurePercentageDifference } from "../humanize-numbers";

export function getFormatterValueForPercDiff(numerator, denominator) {
  if (denominator === 0) return PERC_DIFF.PREV_VALUE_ZERO;
  if (!denominator) return PERC_DIFF.PREV_VALUE_NO_DATA;
  if (numerator === null || numerator === undefined)
    return PERC_DIFF.CURRENT_VALUE_NO_DATA;

  const percDiff = numerator / denominator;
  return formatMeasurePercentageDifference(percDiff);
}

/**
 * @typedef {Object} LeaderboardItemData
 * This is the data that is passed to each individual leaderboard item component.
 * Along with some shared data that is passed to all items in a leaderboard,
 * this should be enough to render the leaderboard.
 */
export type LeaderboardItemData = {
  label: string | number;
  value: number;
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
