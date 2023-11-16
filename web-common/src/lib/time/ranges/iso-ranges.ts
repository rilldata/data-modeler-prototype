import { PeriodAndUnits } from "@rilldata/web-common/lib/time/config";
import { convertTimeRangePreset } from "@rilldata/web-common/lib/time/ranges/index";
import {
  subtractFromPeriod,
  transformDate,
} from "@rilldata/web-common/lib/time/transforms";
import {
  RelativeTimeTransformation,
  TimeOffsetType,
  TimeRange,
  TimeRangePreset,
  TimeTruncationType,
} from "@rilldata/web-common/lib/time/types";
import { Duration } from "luxon";

/**
 * Converts an ISO duration to a time range.
 * Pass in the anchor to specify when the range should be from.
 * NOTE: This should only be used for default time range. UI presets have their own settings.
 */
export function isoDurationToTimeRange(
  isoDuration: string,
  anchor: Date,
  zone = "Etc/UTC"
) {
  const startTime = transformDate(
    anchor,
    getStartTimeTransformations(isoDuration),
    zone
  );
  const endTime = transformDate(
    anchor,
    getEndTimeTransformations(isoDuration),
    zone
  );
  return {
    startTime,
    endTime,
  };
}

export const ISODurationToTimeRangePreset: Partial<
  Record<TimeRangePreset, boolean>
> = {
  [TimeRangePreset.LAST_SIX_HOURS]: true,
  [TimeRangePreset.LAST_24_HOURS]: true,
  [TimeRangePreset.LAST_24_HOURS]: true,
  [TimeRangePreset.LAST_7_DAYS]: true,
  [TimeRangePreset.LAST_14_DAYS]: true,
  [TimeRangePreset.LAST_4_WEEKS]: true,
  [TimeRangePreset.LAST_12_MONTHS]: true,
  [TimeRangePreset.ALL_TIME]: true,
  [TimeRangePreset.TODAY]: true,
  [TimeRangePreset.WEEK_TO_DATE]: true,
  [TimeRangePreset.MONTH_TO_DATE]: true,
  [TimeRangePreset.QUARTER_TO_DATE]: true,
  [TimeRangePreset.YEAR_TO_DATE]: true,
};
export function isoDurationToFullTimeRange(
  isoDuration: string,
  start: Date,
  end: Date,
  zone = "Etc/UTC"
): TimeRange {
  if (!isoDuration) {
    return convertTimeRangePreset(TimeRangePreset.ALL_TIME, start, end, zone);
  }
  if (isoDuration in ISODurationToTimeRangePreset) {
    return convertTimeRangePreset(
      isoDuration as TimeRangePreset,
      start,
      end,
      zone
    );
  }

  const { startTime, endTime } = isoDurationToTimeRange(isoDuration, end, zone);
  return {
    name: isoDuration,
    start: startTime,
    end: endTime,
  };
}

export function humaniseISODuration(isoDuration: string): string {
  if (!isoDuration) return "";
  const duration = Duration.fromISO(isoDuration);
  let humanISO = duration.toHuman({
    listStyle: "long",
  });
  humanISO = humanISO.replace(/(\d) (\w)/g, (substring, n, c) => {
    return `${n} ${c.toUpperCase()}`;
  });
  humanISO = humanISO.replace(", and", " and");
  return humanISO;
}

export function getSmallestTimeGrain(isoDuration: string) {
  const duration = Duration.fromISO(isoDuration);
  for (const { grain, unit } of PeriodAndUnits) {
    if (duration[unit]) {
      return grain;
    }
  }

  return undefined;
}

function getStartTimeTransformations(
  isoDuration: string
): Array<RelativeTimeTransformation> {
  const duration = Duration.fromISO(isoDuration);
  const period = getSmallestUnit(duration);
  if (!period) return [];

  return [
    {
      period, // this is the offset alias for the given time range alias
      truncationType: TimeTruncationType.START_OF_PERIOD,
    }, // truncation
    // then offset that by -1 of smallest period
    {
      duration: subtractFromPeriod(duration, period).toISO() as string,
      operationType: TimeOffsetType.SUBTRACT,
    }, // operation
  ];
}

function getEndTimeTransformations(
  isoDuration: string
): Array<RelativeTimeTransformation> {
  const duration = Duration.fromISO(isoDuration);
  const period = getSmallestUnit(duration);
  if (!period) return [];

  return [
    {
      duration: period,
      operationType: TimeOffsetType.ADD,
    },
    {
      period,
      truncationType: TimeTruncationType.START_OF_PERIOD,
    },
  ];
}

function getSmallestUnit(duration: Duration) {
  for (const { period, unit } of PeriodAndUnits) {
    if (duration[unit]) {
      return period;
    }
  }

  return undefined;
}
