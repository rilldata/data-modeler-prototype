/**
 * This module defines configured presets for time ranges & time grains.
 * We define them as JSON objects primarily users will eventually be able to
 * manually define these in the dashboard configuration.
 */

import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { Duration, DurationUnit } from "luxon";
import {
  type AvailableTimeGrain,
  Period,
  RangePresetType,
  ReferencePoint,
  TimeComparisonOption,
  type TimeGrain,
  TimeOffsetType,
  type TimeRangeMeta,
  TimeRangePreset,
  TimeTruncationType,
} from "./types";

export type TimeRangeMetaSet = Partial<Record<TimeRangePreset, TimeRangeMeta>>;

/**
 * The "latest" window time ranges are defined as a set of time ranges that are
 * anchored to the latest data point in the dataset with a conceptually-fixed
 * lookback window. For example, the "Last 6 Hours" time range is anchored to
 * the latest data point in the dataset, and then looks back 6 hours from that
 * point.
 *
 * This description is not 100% accurate, of course, since the latest data point
 * may be during an incomplete period. For now, we are truncating to a reasonable
 * periodicity (e.g. to the start of the hour) and then applying the offset.
 */
export const LATEST_WINDOW_TIME_RANGES: TimeRangeMetaSet = {
  [TimeRangePreset.LAST_SIX_HOURS]: {
    label: "Last 6 Hours",
    rillSyntax: "-6h",
    rangePreset: RangePresetType.OFFSET_ANCHORED,
    defaultComparison: TimeComparisonOption.CONTIGUOUS,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.HOUR, // this is the offset alias for the given time range alias
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
        // then offset that by 5 hours
        { duration: "PT5H", operationType: TimeOffsetType.SUBTRACT }, // operation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "PT1H", operationType: TimeOffsetType.ADD },
        {
          period: Period.HOUR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },

  [TimeRangePreset.LAST_24_HOURS]: {
    label: "Last 24 Hours",
    rillSyntax: "-24h",
    rangePreset: RangePresetType.OFFSET_ANCHORED,
    defaultComparison: TimeComparisonOption.DAY,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.HOUR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
        { duration: "PT23H", operationType: TimeOffsetType.SUBTRACT }, // operation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "PT1H", operationType: TimeOffsetType.ADD },
        {
          period: Period.HOUR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },

  [TimeRangePreset.LAST_7_DAYS]: {
    label: "Last 7 Days",
    rillSyntax: "-7D",
    rangePreset: RangePresetType.OFFSET_ANCHORED,
    defaultComparison: TimeComparisonOption.WEEK,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
        { duration: "P6D", operationType: TimeOffsetType.SUBTRACT }, // operation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "P1D", operationType: TimeOffsetType.ADD },
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  [TimeRangePreset.LAST_14_DAYS]: {
    label: "Last 14 Days",
    rillSyntax: "-14D",
    rangePreset: RangePresetType.OFFSET_ANCHORED,
    defaultComparison: TimeComparisonOption.CONTIGUOUS,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
        { duration: "P13D", operationType: TimeOffsetType.SUBTRACT },
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "P1D", operationType: TimeOffsetType.ADD },
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  [TimeRangePreset.LAST_4_WEEKS]: {
    label: "Last 4 Weeks",
    rillSyntax: "-4W",
    rangePreset: RangePresetType.OFFSET_ANCHORED,
    defaultComparison: TimeComparisonOption.CONTIGUOUS,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.WEEK,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
        { duration: "P3W", operationType: TimeOffsetType.SUBTRACT },
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "P1W", operationType: TimeOffsetType.ADD },
        {
          period: Period.WEEK,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  [TimeRangePreset.LAST_12_MONTHS]: {
    label: "Last 12 Months",
    rillSyntax: "-12M",
    rangePreset: RangePresetType.OFFSET_ANCHORED,
    defaultComparison: TimeComparisonOption.YEAR,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.MONTH,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
        { duration: "P11M", operationType: TimeOffsetType.SUBTRACT },
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "P1M", operationType: TimeOffsetType.ADD },
        {
          period: Period.MONTH,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
};

/**
 * The "period to date" time ranges are defined as a set of time ranges that are
 * anchored to the latest data point in the dataset, with a start datetime
 * that's anchored to the beginning of the period that the latest data point is in.
 * For example, the "Today" time range is anchored to the latest data point in
 * the dataset, and then looks back to the start of that day.
 *
 * Like the latest window ranges, wetruncate the latest data point datetime to the
 * start of a reasonable period for now.
 */
export const PERIOD_TO_DATE_RANGES: TimeRangeMetaSet = {
  [TimeRangePreset.TODAY]: {
    label: "Today",
    rillSyntax: "0D",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    defaultComparison: TimeComparisonOption.DAY,
    start: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
    end: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        { duration: "P1D", operationType: TimeOffsetType.ADD },
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  [TimeRangePreset.WEEK_TO_DATE]: {
    label: "Week to Date",
    rillSyntax: "0W",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    defaultComparison: TimeComparisonOption.WEEK,
    start: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        {
          period: Period.WEEK,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
    end: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        { duration: "P1D", operationType: TimeOffsetType.ADD },
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  [TimeRangePreset.MONTH_TO_DATE]: {
    label: "Month to Date",
    rillSyntax: "0M",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    defaultComparison: TimeComparisonOption.MONTH,
    start: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        {
          period: Period.MONTH,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
    end: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        { duration: "P1D", operationType: TimeOffsetType.ADD },
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  [TimeRangePreset.QUARTER_TO_DATE]: {
    label: "Quarter to Date",
    rillSyntax: "0Q",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    defaultComparison: TimeComparisonOption.CONTIGUOUS,
    start: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        {
          period: Period.QUARTER,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
    end: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        { duration: "P1D", operationType: TimeOffsetType.ADD },
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  [TimeRangePreset.YEAR_TO_DATE]: {
    label: "Year to Date",
    rillSyntax: "0Y",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    defaultComparison: TimeComparisonOption.YEAR,
    start: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        {
          period: Period.YEAR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
    end: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        { duration: "P1D", operationType: TimeOffsetType.ADD },
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
};

export const PREVIOUS_COMPLETE_DATE_RANGES: TimeRangeMetaSet = {
  [TimeRangePreset.YESTERDAY_COMPLETE]: {
    label: "Yesterday",
    rillSyntax: "-1D, 0D",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    defaultComparison: TimeComparisonOption.CONTIGUOUS,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
        {
          duration: "P1D",
          operationType: TimeOffsetType.SUBTRACT,
        },
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  [TimeRangePreset.PREVIOUS_WEEK_COMPLETE]: {
    label: "Previous week",
    rillSyntax: "-1W, 0W",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    defaultComparison: TimeComparisonOption.CONTIGUOUS,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.WEEK,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
        {
          duration: "P1W",
          operationType: TimeOffsetType.SUBTRACT,
        },
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.WEEK,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  [TimeRangePreset.PREVIOUS_MONTH_COMPLETE]: {
    label: "Previous month",
    rillSyntax: "-1M, OM",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    defaultComparison: TimeComparisonOption.CONTIGUOUS,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.MONTH,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
        {
          duration: "P1M",
          operationType: TimeOffsetType.SUBTRACT,
        },
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.MONTH,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  [TimeRangePreset.PREVIOUS_QUARTER_COMPLETE]: {
    label: "Previous quarter",
    rillSyntax: "-1Q, OQ",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    defaultComparison: TimeComparisonOption.CONTIGUOUS,
    start: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        {
          period: Period.QUARTER,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
        {
          duration: "P3M",
          operationType: TimeOffsetType.SUBTRACT,
        },
      ],
    },
    end: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        {
          period: Period.QUARTER,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  [TimeRangePreset.PREVIOUS_YEAR_COMPLETE]: {
    label: "Previous year",
    rillSyntax: "-1Y, OY",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    defaultComparison: TimeComparisonOption.CONTIGUOUS,
    start: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        {
          period: Period.YEAR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
        {
          duration: "P1Y",
          operationType: TimeOffsetType.SUBTRACT,
        },
      ],
    },
    end: {
      reference: ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW,
      transformation: [
        {
          period: Period.YEAR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
};

export const ALL_TIME = {
  label: "All Time",
  rangePreset: RangePresetType.ALL_TIME,
  // this comparison period is a no-op
  defaultComparison: TimeComparisonOption.CONTIGUOUS,
};

export const CUSTOM = {
  label: "Custom",
  rangePreset: RangePresetType.FIXED_RANGE,
  defaultComparison: TimeComparisonOption.CONTIGUOUS,
};

export const DEFAULT = {
  label: "Default",
  rangePreset: RangePresetType.FIXED_RANGE,
  defaultComparison: TimeComparisonOption.CONTIGUOUS,
};

// TODO: get rid of Partial here
export const DEFAULT_TIME_RANGES: TimeRangeMetaSet = {
  ...LATEST_WINDOW_TIME_RANGES,
  ...PERIOD_TO_DATE_RANGES,
  ...PREVIOUS_COMPLETE_DATE_RANGES,
  [TimeRangePreset.ALL_TIME]: ALL_TIME,
  CUSTOM,
  DEFAULT,
};

/**
 * A time grain is a unit of time that is used to group data points,
 * e.g. "hour" or "day". The time grain is used to aggregate records
 * for the purposes of time series visualization and analysis.
 */

export const TIME_GRAIN: Record<AvailableTimeGrain, TimeGrain> = {
  TIME_GRAIN_MINUTE: {
    grain: V1TimeGrain.TIME_GRAIN_MINUTE,
    label: "minute",
    duration: Period.MINUTE,
    d3format: "%M:%S",
    formatDate: {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    },
  },
  TIME_GRAIN_HOUR: {
    grain: V1TimeGrain.TIME_GRAIN_HOUR,
    label: "hour",
    duration: Period.HOUR,
    d3format: "%H:%M",
    formatDate: {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
    },
  },
  TIME_GRAIN_DAY: {
    grain: V1TimeGrain.TIME_GRAIN_DAY,
    label: "day",
    duration: Period.DAY,
    d3format: "%b %d",
    formatDate: {
      year: "numeric",
      month: "short",
      day: "numeric",
    },
  },
  TIME_GRAIN_WEEK: {
    grain: V1TimeGrain.TIME_GRAIN_WEEK,
    label: "week",
    duration: Period.WEEK,
    d3format: "%b %d",
    formatDate: {
      year: "numeric",
      month: "short",
      day: "numeric",
    },
  },
  TIME_GRAIN_MONTH: {
    grain: V1TimeGrain.TIME_GRAIN_MONTH,
    label: "month",
    duration: Period.MONTH,
    d3format: "%b %Y",
    formatDate: {
      year: "numeric",
      month: "short",
    },
  },
  TIME_GRAIN_QUARTER: {
    grain: V1TimeGrain.TIME_GRAIN_QUARTER,
    label: "quarter",
    duration: Period.QUARTER,
    d3format: "Q%q %Y",
    formatDate: {
      year: "numeric",
      month: "short",
    },
  },
  TIME_GRAIN_YEAR: {
    grain: V1TimeGrain.TIME_GRAIN_YEAR,
    label: "year",
    duration: Period.YEAR,
    d3format: "%Y",
    formatDate: {
      year: "numeric",
    },
  },
};

/** The default configurations for time comparisons. */
export const TIME_COMPARISON = {
  [TimeComparisonOption.CONTIGUOUS]: {
    label: "Previous period",
    shorthand: "prev. period",
    description: "Compare the current time range to the previous time range",
    comparisonType: TimeComparisonOption.CONTIGUOUS,
    offsetIso: "",
  },
  [TimeComparisonOption.CUSTOM]: {
    label: "Custom range",
    shorthand: "comparing",
    description: "Compare the current time range to a custom time range",
    comparisonType: TimeComparisonOption.CUSTOM,
    offsetIso: "",
  },
  [TimeComparisonOption.DAY]: {
    label: "Previous day",
    shorthand: "prev. day",
    description:
      "Compare the current time range to the same time range the day before",
    comparisonType: TimeComparisonOption.DAY,
    offsetIso: "P1D",
  },
  [TimeComparisonOption.WEEK]: {
    label: "Previous week",
    shorthand: "prev. wk",
    description:
      "Compare the current time range to the same time range the week before",
    comparisonType: TimeComparisonOption.WEEK,
    offsetIso: "P1W",
  },
  [TimeComparisonOption.MONTH]: {
    label: "Previous month",
    shorthand: "prev. month",
    description:
      "Compare the current time range to the same time range the month before",
    comparisonType: TimeComparisonOption.MONTH,
    offsetIso: "P1M",
  },
  [TimeComparisonOption.QUARTER]: {
    label: "Previous quarter",
    shorthand: "prev. qtr",
    description:
      "Compare the current time range to the same time range the quarter before",
    comparisonType: TimeComparisonOption.QUARTER,
    offsetIso: "P3M",
  },

  [TimeComparisonOption.YEAR]: {
    label: "Previous year",
    shorthand: "prev. yr",
    description:
      "Compare the current time range to the same time range the year before",
    comparisonType: TimeComparisonOption.YEAR,
    offsetIso: "P1Y",
  },
};

export const NO_COMPARISON_LABEL = "No comparison dimension";

export const DEFAULT_TIMEZONES = [
  "America/Los_Angeles",
  "America/Chicago",
  "America/New_York",
  "Europe/London",
  "Europe/Paris",
  "Asia/Jerusalem",
  "Europe/Moscow",
  "Asia/Kolkata",
  "Asia/Shanghai",
  "Asia/Tokyo",
  "Australia/Sydney",
];

/**
 * Mapping of {@link Period} to the unit in {@link Duration}
 */
export const PeriodAndUnits: Array<{
  period: Period;
  unit: DurationUnit;
  grain: V1TimeGrain;
}> = [
  {
    period: Period.MINUTE,
    unit: "minutes",
    grain: V1TimeGrain.TIME_GRAIN_MINUTE,
  },
  {
    period: Period.HOUR,
    unit: "hours",
    grain: V1TimeGrain.TIME_GRAIN_HOUR,
  },
  {
    period: Period.DAY,
    unit: "days",
    grain: V1TimeGrain.TIME_GRAIN_DAY,
  },
  {
    period: Period.WEEK,
    unit: "weeks",
    grain: V1TimeGrain.TIME_GRAIN_WEEK,
  },
  {
    period: Period.MONTH,
    unit: "months",
    grain: V1TimeGrain.TIME_GRAIN_MONTH,
  },
  {
    period: Period.YEAR,
    unit: "years",
    grain: V1TimeGrain.TIME_GRAIN_YEAR,
  },
];
export const PeriodToUnitsMap: Partial<Record<Period, keyof Duration>> = {};
PeriodAndUnits.forEach(({ period, unit }) => (PeriodToUnitsMap[period] = unit));
