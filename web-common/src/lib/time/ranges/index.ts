/**
 * Utility functinos around handling time ranges.
 *
 * FIXME:
 * - there's some legacy stuff that needs to get deprecated out of this.
 * - we need tests for this.
 */
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DEFAULT_TIME_RANGES, TIME_GRAIN } from "../config";
import {
  durationToMillis,
  getAllowedTimeGrains,
  isGrainBigger,
} from "../grains";
import {
  getDurationMultiple,
  getEndOfPeriod,
  getOffset,
  getStartOfPeriod,
  getTimeWidth,
  relativePointInTimeToAbsolute,
} from "../transforms";
import {
  RangePresetType,
  TimeOffsetType,
  TimeRange,
  TimeRangeMeta,
  TimeRangeOption,
  TimeRangePreset,
  TimeRangeType,
} from "../types";
import { removeTimezoneOffset } from "../../formatters";

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
      // End date is exclusive, so we need to add 1 millisecond to it
      const exclusiveEndDate = new Date(end.getTime() + 1);

      // All time is always an option
      timeRanges.push({
        name: timePreset,
        label: timeRange.label,
        start,
        end: exclusiveEndDate,
      });
    } else {
      const timeRangeDates = relativePointInTimeToAbsolute(
        end,
        timeRange.start,
        timeRange.end
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
      if (isGrainPossible && hasSomeGrainMatches) {
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
      end: new Date(end.getTime() + 1),
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
export const prettyFormatTimeRange = (
  start: Date,
  end: Date,
  timePreset: TimeRangeType
): string => {
  const isAllTime = timePreset === TimeRangePreset.ALL_TIME;
  const isCustom = timePreset === TimeRangePreset.CUSTOM;
  if (!start && end) {
    return `- ${end}`;
  }

  if (start && !end) {
    return `${start} -`;
  }

  if (!start && !end) {
    return "";
  }

  const TIMEZONE = "UTC";
  // const TIMEZONE = Intl.DateTimeFormat().resolvedOptions().timeZone; // the user's local timezone

  const startDate = start.getUTCDate(); // use start.getDate() for local timezone
  const startMonth = start.getUTCMonth();
  const startYear = start.getUTCFullYear();
  let endDate = end.getUTCDate();
  let endMonth = end.getUTCMonth();
  let endYear = end.getUTCFullYear();

  if (
    startDate === endDate &&
    startMonth === endMonth &&
    startYear === endYear
  ) {
    return `${start.toLocaleDateString(undefined, {
      month: "short",
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

  const timeRangeDurationMs = getTimeWidth(start, end);
  if (
    timeRangeDurationMs <= durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    if (isCustom) {
      // For custom time ranges, we want to show just the date
      return `${start.toLocaleDateString(undefined, {
        month: "short",
        timeZone: TIMEZONE,
      })} ${startDate}`;
    }

    return `${start.toLocaleDateString(undefined, {
      month: "short",
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

  let inclusiveEndDate;

  if (isAllTime) {
    inclusiveEndDate = new Date(end);
  } else {
    // beyond this point, we're dealing with time ranges that are full day periods
    // since time range is exclusive at the end, we need to subtract a day
    inclusiveEndDate = new Date(
      end.getTime() - durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
    );
    endDate = inclusiveEndDate.getUTCDate();
    endMonth = inclusiveEndDate.getUTCMonth();
    endYear = inclusiveEndDate.getUTCFullYear();
  }

  // month is the same
  if (startMonth === endMonth && startYear === endYear) {
    return `${start.toLocaleDateString(undefined, {
      month: "short",
      timeZone: TIMEZONE,
    })} ${startDate}-${endDate}, ${startYear}`;
  }

  // year is the same
  if (startYear === endYear) {
    return `${start.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      timeZone: TIMEZONE,
    })} - ${inclusiveEndDate.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      timeZone: TIMEZONE,
    })}, ${startYear}`;
  }
  // year is different
  const dateFormatOptions: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "short",
    day: "numeric",
    timeZone: TIMEZONE,
  };
  return `${start.toLocaleDateString(
    undefined,
    dateFormatOptions
  )} - ${inclusiveEndDate.toLocaleDateString(undefined, dateFormatOptions)}`;
};

/** Get extra data points for extrapolating the chart on both ends */
export function getAdjustedFetchTime(
  startTime: Date,
  endTime: Date,
  interval: V1TimeGrain
) {
  if (!startTime || !endTime)
    return { start: startTime?.toISOString(), end: endTime?.toISOString() };
  const offsetedStartTime = getOffset(
    startTime,
    TIME_GRAIN[interval].duration,
    TimeOffsetType.SUBTRACT
  );

  // the data point previous to the first date inside the chart.
  const fetchStartTime = getStartOfPeriod(
    offsetedStartTime,
    TIME_GRAIN[interval].duration
  );

  const offsetedEndTime = getOffset(
    endTime,
    TIME_GRAIN[interval].duration,
    TimeOffsetType.ADD
  );

  // the data point after the last complete date.
  const fetchEndTime = getStartOfPeriod(
    offsetedEndTime,
    TIME_GRAIN[interval].duration
  );

  return {
    start: fetchStartTime.toISOString(),
    end: fetchEndTime.toISOString(),
  };
}

export function getAdjustedChartTime(
  start: Date,
  end: Date,
  interval: V1TimeGrain,
  timePreset: TimeRangeType
) {
  if (!start || !end)
    return {
      start,
      end,
    };

  const grainDuration = TIME_GRAIN[interval].duration;

  let adjustedEnd = new Date(end);

  if (timePreset === TimeRangePreset.ALL_TIME) {
    // No offset has been applied to All time range so far
    // Adjust end according to the interval
    adjustedEnd = getEndOfPeriod(adjustedEnd, grainDuration);
  }

  const offsetDuration = getDurationMultiple(grainDuration, 0.45);
  adjustedEnd = getOffset(adjustedEnd, offsetDuration, TimeOffsetType.SUBTRACT);

  adjustedEnd = removeTimezoneOffset(adjustedEnd);

  return {
    start: removeTimezoneOffset(new Date(start)),
    end: adjustedEnd,
  };
}
