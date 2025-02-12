import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { DateTimeUnit } from "luxon";

// Used for luxon's time units
export enum TimeUnit {
  PT1M = "minute",
  PT1H = "hour",
  P1D = "day",
  P1W = "week",
  P1M = "month",
  P3M = "quarter",
  P1Y = "year",
}

/** a Period is a natural duration of time that maps nicely to calendar time.
 * For instance, when we say a day period, we understand this means a 24-hour period
 * that starts at 00:00:00 and ends at 23:59:59.999. These periods are used for
 * time truncation functions.
 */
export enum Period {
  MINUTE = "PT1M",
  HOUR = "PT1H",
  DAY = "P1D",
  WEEK = "P1W",
  MONTH = "P1M",
  QUARTER = "P3M",
  YEAR = "P1Y",
}

export enum RangePresetType {
  PERIOD_ANCHORED = "PERIOD_ANCHORED",
  OFFSET_ANCHORED = "OFFSET_ANCHORED",
  ALL_TIME = "ALL_TIME",
  FIXED_RANGE = "FIXED_RANGE",
}

export enum ReferencePoint {
  LATEST_DATA = "LatestData",
  NOW = "Now",
  MIN_OF_LATEST_DATA_AND_NOW = "MinOfLatestDataAndNow",
}

/** An offset defines an operation on a point in time, primarily used to map from a
 * datetime to something we can pass into the dashboard APIs.
 * An offset on its own is just one operation; in a configuration, you create an array of operations,
 * and those map to indiidual sequential operations.
 *
 * Why are we defining these offsets as an interface / object rather than a function?
 * This will enable us to define wholly-custom time ranges in the configuration. Given that
 * there are really only four operations and one input – a duration – this is a fairly tractable and
 * elegant way to handle almost all of the basic time functions of interest in Rill.
 *
 */

export enum TimeOffsetType {
  /** Add the associated duration to this datetime.
   * @example 2020-05-02 12:22:53 -> ADD PT1H -> 2020-05-02 13:22:53
   */
  ADD = "ADD",
  /** Subtract the associted duration to this datetime.
   * @example 2020-05-02 12:22:53 -> SUBTRACT PT1H -> 2020-05-02 11:22:53
   */
  SUBTRACT = "SUBTRACT",
}

interface TimeOffset {
  duration: Period | string;
  operationType: TimeOffsetType;
}

interface TimeTruncation {
  period: Period;
  truncationType: TimeTruncationType;
}

/**
 * These types tell Rill to take the supplied duration, and map it to the beginning
 * or end of the period in which the datetime object is currently in. We utilize the ISO8601 duration
 * to specify when this duration should technically start; we will likely drastically limit the complexity
 * to a small subset of available values. For now, we'll be capitalizing on the Period enum to keep the set
 * of available periods to a normal amount.
 */
export enum TimeTruncationType {
  /**
   * @example 2020-05-02 12:23:53 -> START_OF_PERIOD PT1H -> 2020-05-02 12:00:00.000
   */
  START_OF_PERIOD = "START_OF_PERIOD",
  /**
   * @example 2020-05-02 12:22:53 -> END_OF_PERIOD PT1H -> 2020-05-02 12:59:59.999
   */
  END_OF_PERIOD = "END_OF_PERIOD",
}

/** An offset defines a transformation of an existing datetime into something more usable by our APIs
 * (and more coherent to humans). We only need to specify two types of offsets:
 * - operation, like subtracting or adding a time duration.
 * - truncation, which enables us to get the beginning of end of a period of interest.
 */
export type RelativeTimeTransformation = TimeOffset | TimeTruncation;

export interface RelativePointInTime {
  reference: ReferencePoint;
  transformation: RelativeTimeTransformation[];
}

export interface TimeRangeMeta {
  label: string;
  rillSyntax?: string;
  defaultGrain?: V1TimeGrain; // Affordance for future use
  rangePreset?: RangePresetType | string;
  defaultComparison?: TimeComparisonOption | string;
  start?: string | RelativePointInTime;
  end?: string | RelativePointInTime;
}

// Enum for ease of access to the default presets
export enum TimeRangePreset {
  ALL_TIME = "inf",
  LAST_SIX_HOURS = "PT6H",
  LAST_24_HOURS = "PT24H",
  LAST_7_DAYS = "P7D",
  LAST_14_DAYS = "P14D",
  LAST_4_WEEKS = "P4W",
  LAST_12_MONTHS = "P12M",
  TODAY = "rill-TD",
  WEEK_TO_DATE = "rill-WTD",
  MONTH_TO_DATE = "rill-MTD",
  QUARTER_TO_DATE = "rill-QTD",
  YEAR_TO_DATE = "rill-YTD",
  YESTERDAY_COMPLETE = "rill-PDC",
  PREVIOUS_WEEK_COMPLETE = "rill-PWC",
  PREVIOUS_MONTH_COMPLETE = "rill-PMC",
  PREVIOUS_QUARTER_COMPLETE = "rill-PQC",
  PREVIOUS_YEAR_COMPLETE = "rill-PYC",
  CUSTOM = "CUSTOM",
  DEFAULT = "DEFAULT",
}

export interface TimeRange {
  name?: string;
  start: Date;
  end: Date;
}

export interface TimeRangeString {
  start: string | undefined;
  end: string | undefined;
}

export interface TimeRangeOption extends TimeRange {
  label: string;
}

export interface ScrubRange extends TimeRange {
  isScrubbing: boolean;
}

export interface DashboardTimeControls extends TimeRange {
  interval?: V1TimeGrain;
  label?: string;
}

/** defines configuration for a time grain object in Rill. */
export interface TimeGrain {
  /** the grain defined by the runtime */
  grain: V1TimeGrain;
  /** a human-readable name, e.g. minute, second, etc. */
  label: DateTimeUnit;
  /** the ISO8601 duration, e.g. P1D, PT6H */
  duration: Period;
  /** the d3 time format string that outputs a human-readable
   * representation. Currently used for TDD column headers */
  d3format: string;
  /** the DateTimeFormatOptions of the Intl API that outputs
   * a human-readable representation of a timestamp based on
   * the time grain. This preserves locale-based formatting.
   */
  formatDate: Intl.DateTimeFormatOptions;
}

// limit the set of available time grains to those supported
// by th dashboard.
export type AvailableTimeGrain = Exclude<
  V1TimeGrain,
  "TIME_GRAIN_UNSPECIFIED" | "TIME_GRAIN_MILLISECOND" | "TIME_GRAIN_SECOND"
>;

export enum TimeComparisonOption {
  CONTIGUOUS = "rill-PP",
  CUSTOM = "CUSTOM_COMPARISON_RANGE",
  DAY = "rill-PD",
  WEEK = "rill-PW",
  MONTH = "rill-PM",
  QUARTER = "rill-PQ",
  YEAR = "rill-PY",
}

export enum TimeRoundingStrategy {
  NEAREST = "NEAREST",
  PREVIOUS = "PREVIOUS",
}
