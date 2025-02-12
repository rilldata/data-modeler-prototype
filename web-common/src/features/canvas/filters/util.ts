import {
  getAvailableComparisonsForTimeRange,
  getComparisonRange,
} from "@rilldata/web-common/lib/time/comparisons";
import { PREVIOUS_COMPLETE_DATE_RANGES } from "@rilldata/web-common/lib/time/config";
import {
  TimeComparisonOption,
  TimeRangePreset,
  type DashboardTimeControls,
} from "@rilldata/web-common/lib/time/types";

export function getComparisonOptionsForCanvas(
  selectedTimeRange: DashboardTimeControls | undefined,
) {
  if (!selectedTimeRange) {
    return [];
  }
  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(0),
    end: new Date(),
  };

  let allOptions = [...Object.values(TimeComparisonOption)];

  if (
    selectedTimeRange?.name &&
    selectedTimeRange?.name in PREVIOUS_COMPLETE_DATE_RANGES
  ) {
    // Previous complete ranges should only have previous period.
    // Other options dont make sense with our current wording of the comparison ranges.
    allOptions = [TimeComparisonOption.CONTIGUOUS, TimeComparisonOption.CUSTOM];
  }

  const timeComparisonOptions = getAvailableComparisonsForTimeRange(
    allTimeRange.start,
    allTimeRange.end,
    selectedTimeRange.start,
    selectedTimeRange.end,
    allOptions,
  );

  return timeComparisonOptions.map((co, i) => {
    const comparisonTimeRange = getComparisonRange(
      selectedTimeRange.start,
      selectedTimeRange.end,
      co,
    );
    return {
      name: co,
      key: i,
      start: comparisonTimeRange.start,
      end: comparisonTimeRange.end,
    };
  });
}
