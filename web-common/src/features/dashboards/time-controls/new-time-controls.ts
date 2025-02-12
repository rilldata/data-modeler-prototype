// WIP as of 04/19/2024

import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { writable, type Writable, get } from "svelte/store";
import {
  Interval,
  DateTime,
  type DurationObjectUnits,
  type DateTimeUnit,
  Duration,
  IANAZone,
} from "luxon";
import {
  queryServiceMetricsViewTimeRanges,
  type MetricsViewSpecAvailableTimeRange,
} from "@rilldata/web-common/runtime-client";

// CONSTANTS -> time-control-constants.ts

export const RILL_TO_UNIT: Record<
  RillPeriodToDate | RillPreviousPeriod,
  DateTimeUnit
> = {
  "rill-PDC": "day",
  "rill-PWC": "week",
  "rill-PMC": "month",
  "rill-PQC": "quarter",
  "rill-PYC": "year",
  "rill-TD": "day",
  "rill-WTD": "week",
  "rill-MTD": "month",
  "rill-QTD": "quarter",
  "rill-YTD": "year",
};

export const RILL_TO_LABEL: Record<
  RillPeriodToDate | RillPreviousPeriod | AllTime | CustomRange,
  string
> = {
  inf: "All Time",
  CUSTOM: "Custom",
  "rill-PDC": "Yesterday",
  "rill-PWC": "Previous week",
  "rill-PMC": "Previous month",
  "rill-PQC": "Previous quarter",
  "rill-PYC": "Previous year",
  "rill-TD": "Today",
  "rill-WTD": "Week to date",
  "rill-MTD": "Month to date",
  "rill-QTD": "Quarter to date",
  "rill-YTD": "Year to date",
};

export const RILL_PERIOD_TO_DATE = [
  "rill-TD",
  "rill-WTD",
  "rill-MTD",
  "rill-QTD",
  "rill-YTD",
] as const;

export const RILL_PREVIOUS_PERIOD = [
  "rill-PDC",
  "rill-PWC",
  "rill-PMC",
  "rill-PQC",
  "rill-PYC",
] as const;

export const RILL_LATEST = [
  "PT6H",
  "PT24H",
  "P7D",
  "P14D",
  "P3M",
  "P12M",
] as const;

// TYPES -> time-control-types.ts

type RillPeriodToDateTuple = typeof RILL_PERIOD_TO_DATE;
export type RillPeriodToDate = RillPeriodToDateTuple[number];

type RillPreviousPeriodTuple = typeof RILL_PREVIOUS_PERIOD;
export type RillPreviousPeriod = RillPreviousPeriodTuple[number];

type RillLatestTuple = typeof RILL_LATEST;
export type RillLatest = RillLatestTuple[number];

export const CUSTOM_TIME_RANGE_ALIAS = "CUSTOM";
export const ALL_TIME_RANGE_ALIAS = "inf";
export type AllTime = typeof ALL_TIME_RANGE_ALIAS;
export type CustomRange = typeof CUSTOM_TIME_RANGE_ALIAS;
export type ISODurationString = string;

export type NamedRange =
  | RillPeriodToDate
  | RillPreviousPeriod
  | AllTime
  | CustomRange;

// STORES -> time-control-stores.ts

class IntervalStore {
  private _interval: Writable<Interval> = writable(
    Interval.invalid("Uninitialized"),
  );

  subscribe = this._interval.subscribe;

  clear = () => {
    this._interval.set(Interval.invalid("Uninitialized"));
  };

  updateInterval(interval: Interval) {
    this._interval.set(interval);
  }
  Lea;
  updateEnd(end: DateTime) {
    this._interval.update((i) => i.set({ end }));
  }

  updateStart(start: DateTime) {
    this._interval.update((i) => i.set({ start }));
  }

  updateZone(zone: IANAZone, keepLocalTime = false) {
    this._interval.update((i) =>
      i.mapEndpoints((e) => e.setZone(zone, { keepLocalTime: keepLocalTime })),
    );
  }
}

class MetricsTimeControls {
  private _maxRange = new IntervalStore();
  private _zone: Writable<IANAZone> = writable(new IANAZone("UTC"));
  private _selected = writable<NamedRange | ISODurationString>(
    ALL_TIME_RANGE_ALIAS,
  );
  private _visibleRange = new IntervalStore();
  private _subrange = new IntervalStore();
  private _comparisonRange = new IntervalStore();
  private _showComparison: Writable<boolean> = writable(false);
  private _metricsViewName: string;

  constructor(maxStart: DateTime, maxEnd: DateTime, metricsViewName: string) {
    this._metricsViewName = metricsViewName;
    const maxInterval = Interval.fromDateTimes(
      maxStart.setZone("UTC"),
      maxEnd.setZone("UTC"),
    );
    this._maxRange.updateInterval(maxInterval);
    this._visibleRange.updateInterval(maxInterval);
  }

  private applySubrange = () => {
    this._visibleRange.updateInterval(get(this._subrange));
    this._selected.set(CUSTOM_TIME_RANGE_ALIAS);
    this._subrange.clear();
  };

  private applyISODuration = async (iso: ISODurationString) => {
    const rightAnchor = get(this._maxRange).end;
    if (rightAnchor) {
      const interval = await deriveInterval(
        iso,
        rightAnchor,
        this._metricsViewName,
      );
      if (interval?.isValid) {
        this._visibleRange.updateInterval(interval);
        this._selected.set(iso);
      }
    }
  };

  private applyNamedRange = async (name: NamedRange) => {
    const rightAnchor = get(this._maxRange).end;
    if (rightAnchor) {
      const interval = await deriveInterval(
        name,
        rightAnchor,
        this._metricsViewName,
      );
      if (interval?.isValid) {
        this._visibleRange.updateInterval(interval);
        this._selected.set(name);
      }
    }
  };

  private applyCustomRange = (start: DateTime, end: DateTime) => {
    this._visibleRange.updateInterval(Interval.fromDateTimes(start, end));
    this._selected.set(CUSTOM_TIME_RANGE_ALIAS);
  };

  private applyRange = (string: NamedRange | ISODurationString | undefined) => {
    if (!string) return;

    if (string === ALL_TIME_RANGE_ALIAS) {
      this.applyAllTime();
    } else if (isRillPeriodToDate(string) || isRillPreviousPeriod(string)) {
      this.applyNamedRange(string);
    } else if (isValidISODuration(string)) {
      this.applyISODuration(string);
    } else if (string === CUSTOM_TIME_RANGE_ALIAS) {
      throw new Error("Custom time range requires start and end dates");
    } else {
      throw new Error("Invalid time range");
    }
  };

  private applyAllTime = () => {
    const maxInterval = get(this._maxRange);
    if (maxInterval) {
      this._visibleRange.updateInterval(maxInterval);
      this._selected.set(ALL_TIME_RANGE_ALIAS);
    }
  };

  updateZone = (zone: IANAZone) => {
    this._zone.set(zone);
    const rangeAlias = get(this._selected);
    this._maxRange.updateZone(zone);

    if (rangeAlias === CUSTOM_TIME_RANGE_ALIAS) {
      // If you've specified a custom range
      // We want to maintain the local time of the start and end
      this._visibleRange.updateZone(zone, true);
    } else {
      // Otherwise, we need to re-derive the interval based on the selection
      this.applyRange(rangeAlias);
    }
  };

  apply = {
    subrange: this.applySubrange,
    customRange: this.applyCustomRange,
    range: this.applyRange,
    allTime: this.applyAllTime,
  };

  switchComparison(bool?: boolean) {
    if (bool === undefined) {
      this._showComparison.update((b) => !b);
    } else {
      this._showComparison.set(bool);
    }
  }

  zone = this._zone;
  selected = this._selected;
  subrange = this._subrange;
  visibleRange = this._visibleRange;
  comparisonRange = this._comparisonRange;
}

class TimeControls {
  private _timeControls = new Map<string, MetricsTimeControls>();

  get(metricsViewName: string, maxStart?: DateTime, maxEnd?: DateTime) {
    let store = this._timeControls.get(metricsViewName);

    if (!store && maxStart && maxEnd) {
      store = new MetricsTimeControls(maxStart, maxEnd, metricsViewName);
      this._timeControls.set(metricsViewName, store);
    } else if (!store) {
      throw new Error("TimeControls.get() called without maxStart and maxEnd");
    }

    return store;
  }
}

export const timeControls = new TimeControls();

// UTILS -> time-control-utils.ts

export function isRillPreviousPeriod(
  value: string,
): value is RillPreviousPeriod {
  return RILL_PREVIOUS_PERIOD.includes(value as RillPreviousPeriod);
}

export function isRillPeriodToDate(value: string): value is RillPeriodToDate {
  return RILL_PERIOD_TO_DATE.includes(value as RillPeriodToDate);
}

import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

export async function deriveInterval(
  name: RillPeriodToDate | RillPreviousPeriod | ISODurationString,
  anchor: DateTime,
  metricsViewName: string,
): Promise<Interval> {
  if (name === ALL_TIME_RANGE_ALIAS || name === CUSTOM_TIME_RANGE_ALIAS) {
    throw new Error("Cannot derive interval for all time or custom range");
  }

  if (isRillPeriodToDate(name)) {
    const period = RILL_TO_UNIT[name];
    return getPeriodToDate(anchor, period);
  }

  if (isRillPreviousPeriod(name)) {
    const period = RILL_TO_UNIT[name];
    return getPreviousPeriodComplete(anchor, period, 1);
  }

  const duration = isValidISODuration(name);

  if (duration) {
    return getInterval(duration, anchor);
  } else {
    const response = await queryServiceMetricsViewTimeRanges(
      get(runtime).instanceId,
      metricsViewName,
      { expressions: [name] },
    );

    const timeRange = response.timeRanges?.pop();

    if (!timeRange?.start || !timeRange?.end) {
      return Interval.invalid("Invalid time range");
    }

    return Interval.fromDateTimes(
      DateTime.fromISO(timeRange.start),
      DateTime.fromISO(timeRange.end),
    );
  }
}

export function getPeriodToDate(date: DateTime, period: DateTimeUnit) {
  const periodStart = date.startOf(period);
  const exclusiveEnd = date.endOf("day").plus({ millisecond: 1 });

  return Interval.fromDateTimes(periodStart, exclusiveEnd);
}

export function getPreviousPeriodComplete(
  anchor: DateTime,
  period: DateTimeUnit,
  steps = 0,
) {
  const startOfCurrentPeriod = anchor.startOf(period);
  const shiftedStart = startOfCurrentPeriod.minus({ [period + "s"]: steps });
  const exclusiveEnd = shiftedStart.endOf(period).plus({ millisecond: 1 });

  return Interval.fromDateTimes(shiftedStart, exclusiveEnd);
}

export function getInterval(
  luxonDuration: Duration,
  endDate: DateTime,
  full = true,
) {
  const durationUnits = luxonDuration.toObject();
  const smallestUnit = getSmallestUnit(durationUnits);

  const end =
    smallestUnit && full
      ? endDate.endOf(smallestUnit).plus({ millisecond: 1 })
      : endDate;

  return Interval.before(end, durationUnits);
}

export function getSmallestUnit(
  units: DurationObjectUnits,
): DateTimeUnit | null {
  if (units.milliseconds) return "millisecond";
  if (units.seconds) return "second";
  if (units.minutes) return "minute";
  if (units.hours) return "hour";
  if (units.days) return "day";
  if (units.weeks) return "week";
  if (units.months) return "month";
  if (units.quarters) return "quarter";
  if (units.years) return "year";

  return null;
}

export function isValidISODuration(duration: string) {
  const luxonDuration = Duration.fromISO(duration);

  if (luxonDuration.isValid) return luxonDuration;
  return null;
}

export function getDurationLabel(isoDuration: string): string {
  if (!isValidISODuration(isoDuration)) {
    throw new Error("Invalid ISO duration");
  }

  return `Last ${humaniseISODuration(isoDuration)}`;
}

export function getRangeLabel(range: string): string {
  if (isRillPeriodToDate(range) || isRillPreviousPeriod(range)) {
    return RILL_TO_LABEL[range];
  }

  if (range === ALL_TIME_RANGE_ALIAS || range === CUSTOM_TIME_RANGE_ALIAS) {
    return RILL_TO_LABEL[range];
  }

  if (isValidISODuration(range)) {
    return getDurationLabel(range);
  }

  try {
    const rt = parseRillTime(range);
    const label = rt.getLabel();
    if (label.endsWith(", incomplete")) {
      return label.replace(", incomplete", "");
    } else {
      return label + ", complete";
    }
  } catch {
    // no-op
  }

  return range;
}

// BUCKETS FOR DISPLAYING IN DROPDOWN (yaml spec may make this unnecessary)

export type RangeBuckets = {
  latest: { label: string; range: ISODurationString }[];
  previous: { range: RillPreviousPeriod; label: string }[];
  periodToDate: { range: RillPeriodToDate; label: string }[];
};

const defaultBuckets = {
  previous: RILL_PREVIOUS_PERIOD.map((range) => ({
    range,
    label: RILL_TO_LABEL[range],
  })),
  latest: RILL_LATEST.map((range) => ({
    range,
    label: getDurationLabel(range),
  })),
  periodToDate: RILL_PERIOD_TO_DATE.map((range) => ({
    range,
    label: RILL_TO_LABEL[range],
  })),
};

export function bucketYamlRanges(
  availableRanges: MetricsViewSpecAvailableTimeRange[],
): RangeBuckets {
  const showDefaults = !availableRanges.length;

  if (showDefaults) {
    return defaultBuckets;
  }

  return availableRanges.reduce(
    (record, { range }) => {
      if (!range) return record;

      if (isRillPeriodToDate(range)) {
        record.periodToDate.push({ range, label: RILL_TO_LABEL[range] });
      } else if (isRillPreviousPeriod(range)) {
        record.previous.push({ range, label: RILL_TO_LABEL[range] });
      } else {
        record.latest.push({ range, label: getDurationLabel(range) });
      }

      return record;
    },
    <RangeBuckets>{
      previous: [],
      latest: [],
      periodToDate: [],
    },
  );
}
