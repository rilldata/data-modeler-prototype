import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { getValidComparisonOption } from "@rilldata/web-common/features/dashboards/time-controls/time-range-store";
import { getDefaultTimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { ExploreStateDefaultTimezone } from "@rilldata/web-common/features/dashboards/url-state/defaults";
import {
  ToURLParamTDDChartMap,
  ToURLParamTimeGrainMapMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ISODurationToTimePreset } from "@rilldata/web-common/lib/time/ranges";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { getLocalIANA } from "@rilldata/web-common/lib/time/timezone";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  V1ExploreComparisonMode,
  V1ExploreSortType,
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
  type V1MetricsViewTimeRangeResponse,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";

export function getDefaultExplorePreset(
  explore: V1ExploreSpec,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  const defaultExplorePreset: V1ExplorePreset = {
    view: V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
    where: createAndExpression([]),

    measures: explore.measures,
    dimensions: explore.dimensions,

    timeRange: fullTimeRange ? "inf" : "",
    timezone: explore.defaultPreset?.timezone ?? getLocalIANA(),
    timeGrain: "",
    comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE,
    compareTimeRange: "",
    comparisonDimension: "",

    exploreSortBy:
      explore.defaultPreset?.measures?.[0] ?? explore.measures?.[0],
    exploreSortAsc: false,
    exploreSortType: V1ExploreSortType.EXPLORE_SORT_TYPE_VALUE,
    exploreExpandedDimension: "",

    timeDimensionMeasure: "",
    timeDimensionChartType: ToURLParamTDDChartMap[TDDChart.DEFAULT],
    timeDimensionPin: false,

    pivotCols: [],
    pivotRows: [],
    pivotSortBy: "",
    pivotSortAsc: false,

    ...(explore.defaultPreset ?? {}),
  };

  if (!explore.timeZones?.length) {
    // this is the old behaviour. if no timezones are configures for the explore, default it to UTC and not local IANA
    defaultExplorePreset.timezone = ExploreStateDefaultTimezone;
  } else if (!explore.timeZones?.includes(defaultExplorePreset.timezone!)) {
    // else if the default is not in the list of timezones
    if (explore.timeZones?.includes(ExploreStateDefaultTimezone)) {
      defaultExplorePreset.timezone = ExploreStateDefaultTimezone;
    } else {
      defaultExplorePreset.timezone = explore.timeZones[0];
    }
  }

  if (!defaultExplorePreset.timeGrain) {
    defaultExplorePreset.timeGrain = getDefaultPresetTimeGrain(
      defaultExplorePreset,
      fullTimeRange,
    );
  }

  if (defaultExplorePreset.comparisonMode) {
    Object.assign(
      defaultExplorePreset,
      getDefaultComparisonFields(
        defaultExplorePreset,
        explore,
        fullTimeRange?.timeRangeSummary,
      ),
    );
  }

  return defaultExplorePreset;
}

function getDefaultPresetTimeGrain(
  defaultExplorePreset: V1ExplorePreset,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  if (
    !defaultExplorePreset.timeRange ||
    !fullTimeRange?.timeRangeSummary?.min ||
    !fullTimeRange?.timeRangeSummary?.max
  )
    return "";

  const fullTimeStart = new Date(fullTimeRange.timeRangeSummary.min);
  const fullTimeEnd = new Date(fullTimeRange.timeRangeSummary.max);
  const timeRange = isoDurationToFullTimeRange(
    defaultExplorePreset.timeRange,
    fullTimeStart,
    fullTimeEnd,
    defaultExplorePreset.timezone,
  );

  return (
    ToURLParamTimeGrainMapMap[
      getDefaultTimeGrain(timeRange.start, timeRange.end)
    ] ?? ""
  );
}

function getDefaultComparisonFields(
  defaultExplorePreset: V1ExplorePreset,
  explore: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
): V1ExplorePreset {
  if (
    defaultExplorePreset.comparisonMode ===
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_UNSPECIFIED ||
    defaultExplorePreset.comparisonMode ===
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE
  ) {
    return {};
  }

  if (
    defaultExplorePreset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION
  ) {
    return {
      comparisonDimension:
        defaultExplorePreset.comparisonDimension || explore.dimensions?.[0],
    };
  }

  if (
    !defaultExplorePreset.timeRange ||
    defaultExplorePreset.timeRange === "inf" ||
    !timeRangeSummary?.min ||
    !timeRangeSummary?.max
  ) {
    return {};
  }

  let comparisonOption = defaultExplorePreset.compareTimeRange;

  if (!comparisonOption) {
    const preset = ISODurationToTimePreset(
      defaultExplorePreset.timeRange,
      true,
    );
    if (!preset) return {};

    const allTimeRange = {
      name: TimeRangePreset.ALL_TIME,
      start: new Date(timeRangeSummary.min),
      end: new Date(timeRangeSummary.max),
    };

    const timeRange = isoDurationToFullTimeRange(
      preset,
      allTimeRange.start,
      allTimeRange.end,
      defaultExplorePreset.timezone,
    );

    comparisonOption = getValidComparisonOption(
      explore.timeRanges,
      timeRange,
      undefined,
      allTimeRange,
    );
  }

  return {
    compareTimeRange: comparisonOption,
    exploreSortType:
      defaultExplorePreset.exploreSortType ??
      V1ExploreSortType.EXPLORE_SORT_TYPE_DELTA_PERCENT,
  };
}
