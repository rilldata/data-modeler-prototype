/**
 * Utility functinos around handling time ranges.
 *
 * FIXME:
 * - there's some legacy stuff that needs to get deprecated out of this.
 * - we need tests for this.
 */
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DEFAULT_TIME_RANGES } from "../config";
import {
  durationToMillis,
  getAllowedTimeGrains,
  isGrainBigger,
} from "../grains";
import { getTimeWidth, relativePointInTimeToAbsolute } from "../transforms";
import {
  RangePresetType,
  TimeRange,
  TimeRangeMeta,
  TimeRangeOption,
  TimeRangePreset,
  TimeRangeType,
} from "../types";

/**
 * Returns true if the range defined by start and end is completely
 * inside the range defined by otherStart and otherEnd.
 */
export function isRangeInsideOther(
  start: Date,
  end: Date,
  otherStart: Date,
  otherEnd: Date
) {
  return otherStart >= start && otherEnd <= end;
}

// Loop through all presets to check if they can be a part of subset of given start and end date
export function getChildTimeRanges(
  start: Date,
  end: Date,
  ranges: Record<string, TimeRangeMeta>,
  minTimeGrain: V1TimeGrain
): TimeRangeOption[] {
  const timeRanges: TimeRangeOption[] = [];

  const allowedTimeGrains = getAllowedTimeGrains(start, end);
  const allowedMaxGrain = allowedTimeGrains[allowedTimeGrains.length - 1];
  for (const timePreset in ranges) {
    const timeRange = ranges[timePreset];
    if (timeRange.rangePreset == RangePresetType.ALL_TIME) {
      // All time is always an option
      timeRanges.push({
        name: timePreset,
        label: timeRange.label,
        start,
        end,
      });
    } else {
      const timeRangeDates = relativePointInTimeToAbsolute(
        end,
        timeRange.start,
        timeRange.end
      );

      const timeRangeIsSmallerThanAllTime = isRangeInsideOther(
        start,
        end,
        timeRangeDates.startDate,
        timeRangeDates.endDate
      );

      // check if time range is possible with given minTimeGrain
      const thisRangeAllowedGrains = getAllowedTimeGrains(
        timeRangeDates.startDate,
        timeRangeDates.endDate
      );

      const hasSomeGrainMatches = thisRangeAllowedGrains.some((grain) => {
        return (
          !isGrainBigger(minTimeGrain, grain.grain) &&
          durationToMillis(grain.duration) <=
            getTimeWidth(timeRangeDates.startDate, timeRangeDates.endDate)
        );
      });

      const isGrainPossible = !isGrainBigger(
        minTimeGrain,
        allowedMaxGrain.grain
      );
      if (
        timeRangeIsSmallerThanAllTime &&
        isGrainPossible &&
        hasSomeGrainMatches
      ) {
        timeRanges.push({
          name: timePreset,
          label: timeRange.label,
          start: timeRangeDates.startDate,
          end: timeRangeDates.endDate,
        });
      }
    }
  }

  return timeRanges;
}

// TODO: investigate whether we need this after we've removed the need
// for the config's default_time_Range to be an ISO duration.
export function ISODurationToTimePreset(
  isoDuration: string,
  defaultToAllTime = true
): TimeRangeType {
  switch (isoDuration) {
    case "PT6H":
      return TimeRangePreset.LAST_SIX_HOURS;
    case "P1D":
      return TimeRangePreset.LAST_24_HOURS;
    case "P7D":
      return TimeRangePreset.LAST_7_DAYS;
    case "P4W":
      return TimeRangePreset.LAST_4_WEEKS;
    case "inf":
      return TimeRangePreset.ALL_TIME;
    default:
      return defaultToAllTime ? TimeRangePreset.ALL_TIME : undefined;
  }
}

/* Converts a Time Range preset to a TimeRange object */
export function convertTimeRangePreset(
  timeRangePreset: TimeRangeType,
  start: Date,
  end: Date
): TimeRange {
  if (timeRangePreset === TimeRangePreset.ALL_TIME) {
    return {
      name: timeRangePreset,
      start,
      end,
    };
  }
  const timeRange = DEFAULT_TIME_RANGES[timeRangePreset];
  const timeRangeDates = relativePointInTimeToAbsolute(
    end,
    timeRange.start,
    timeRange.end
  );

  return {
    name: timeRangePreset,
    start: timeRangeDates.startDate,
    end: timeRangeDates.endDate,
  };
}

/**
 * Formats a start and end for usage in the application.
 * NOTE: this is primarily used for the time range picker. We might want to
 * colocate the code w/ the component.
 */
export const prettyFormatTimeRange = (start: Date, end: Date): string => {
  if (!start && end) {
    return `- ${end}`;
  }

  if (start && !end) {
    return `${start} -`;
  }

  if (!start && !end) {
    return "";
  }

  // TODO: Do we still need to do this?
  // timeRange.end is exclusive. We subtract one ms to render the last inclusive value.
  end = new Date(end.getTime() - 1);

  const TIMEZONE = "UTC";
  // const TIMEZONE = Intl.DateTimeFormat().resolvedOptions().timeZone; // the user's local timezone

  const startDate = start.getUTCDate(); // use start.getDate() for local timezone
  const startMonth = start.getUTCMonth();
  const startYear = start.getUTCFullYear();
  const endDate = end.getUTCDate();
  const endMonth = end.getUTCMonth();
  const endYear = end.getUTCFullYear();

  // day is the same
  if (
    startDate === endDate &&
    startMonth === endMonth &&
    startYear === endYear
  ) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
      timeZone: TIMEZONE,
    })} ${startDate}, ${startYear} (${start
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone: TIMEZONE,
      })
      .replace(/\s/g, "")}-${end
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone: TIMEZONE,
      })
      .replace(/\s/g, "")})`;
  }

  // month is the same
  if (startMonth === endMonth && startYear === endYear) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
      timeZone: TIMEZONE,
    })} ${startDate}-${endDate}, ${startYear} (${start
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone: TIMEZONE,
      })
      .replace(/\s/g, "")}-${end
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone: TIMEZONE,
      })
      .replace(/\s/g, "")})`;
  }
  // year is the same
  if (startYear === endYear) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
      day: "numeric",
      timeZone: TIMEZONE,
    })} - ${end.toLocaleDateString(undefined, {
      month: "long",
      day: "numeric",
      timeZone: TIMEZONE,
    })}, ${startYear}`;
  }
  // year is different
  const dateFormatOptions: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "long",
    day: "numeric",
    timeZone: TIMEZONE,
  };
  return `${start.toLocaleDateString(
    undefined,
    dateFormatOptions
  )} - ${end.toLocaleDateString(undefined, dateFormatOptions)}`;
};
