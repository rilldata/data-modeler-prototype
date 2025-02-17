import { fromTimeRangeUrlParam } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { fromTimeRangesParams } from "@rilldata/web-common/features/dashboards/url-state/convertURLToExplorePreset";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { V1ExploreComparisonMode } from "@rilldata/web-common/runtime-client";

export function getLocalComparison(timeFilter: string | undefined) {
  let showLocalTimeComparison = false;
  let localComparisonTimeRange: DashboardTimeControls | undefined = undefined;
  if (!timeFilter) {
    return { showLocalTimeComparison, localComparisonTimeRange };
  }
  const urlParams = new URLSearchParams(timeFilter);
  const { preset, errors } = fromTimeRangesParams(urlParams, new Map());

  if (errors?.length) {
    return { showLocalTimeComparison, localComparisonTimeRange };
  }

  if (preset.compareTimeRange) {
    localComparisonTimeRange = fromTimeRangeUrlParam(preset.compareTimeRange);
    showLocalTimeComparison = true;
  } else if (
    preset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
  ) {
    showLocalTimeComparison = true;
  }

  return { showLocalTimeComparison, localComparisonTimeRange };
}
