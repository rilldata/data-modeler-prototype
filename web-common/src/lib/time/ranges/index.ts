/**
 * Utility functinos around handling time ranges.
 *
 * FIXME:
 * - there's some legacy stuff that needs to get deprecated out of this.
 * - we need tests for this.
 */
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { getSmallestTimeGrain } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  addZoneOffset,
  getDateMonthYearForTimezone,
  removeLocalTimezoneOffset,
} from "@rilldata/web-common/lib/time/timezone";
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
  TimeComparisonOption,
  TimeOffsetType,
  type TimeRange,
  type TimeRangeMeta,
  type TimeRangeOption,
  TimeRangePreset,
} from "../types";

// Loop through all presets to check if they can be a part of subset of given start and end date
export function getChildTimeRanges(
  start: Date,
  end: Date,
  ranges: Record<string, TimeRangeMeta>,
  minTimeGrain: V1TimeGrain,
  zone: string,
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
        name: timePreset as TimeRangePreset,
        label: timeRange.label,
        start,
        end: exclusiveEndDate,
      });
    } else {
      const timeRangeDates = relativePointInTimeToAbsolute(
        end,
        timeRange.start,
        timeRange.end,
        zone,
      );

      // check if time range is possible with given minTimeGrain
      const thisRangeAllowedGrains = getAllowedTimeGrains(
        timeRangeDates.startDate,
        timeRangeDates.endDate,
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
        allowedMaxGrain.grain,
      );

      const isRangeValid =
        isWithinRange(timeRangeDates.startDate, start, end) ||
        isWithinRange(timeRangeDates.endDate, start, end);

      if (isRangeValid && isGrainPossible && hasSomeGrainMatches) {
        timeRanges.push({
          name: timePreset as TimeRangePreset,
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
  isoDuration: string | undefined,
  defaultToAllTime = true,
): TimeRangePreset | undefined {
  switch (isoDuration) {
    case "PT6H":
      return TimeRangePreset.LAST_SIX_HOURS;
    case "PT24H":
      return TimeRangePreset.LAST_24_HOURS;
    case "P1D":
      return TimeRangePreset.LAST_24_HOURS;
    case "P7D":
      return TimeRangePreset.LAST_7_DAYS;
    case "P14D":
      return TimeRangePreset.LAST_14_DAYS;
    case "P4W":
      return TimeRangePreset.LAST_4_WEEKS;
    case "P2W":
      return TimeRangePreset.LAST_14_DAYS;
    case "inf":
      return TimeRangePreset.ALL_TIME;
    default:
      return defaultToAllTime ? TimeRangePreset.ALL_TIME : undefined;
  }
}

/* Converts a Time Range preset to a TimeRange object */
export function convertTimeRangePreset(
  timeRangePreset: TimeRangePreset | TimeComparisonOption,
  start: Date,
  end: Date,
  zone: string,
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
    timeRange.end,
    zone,
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
  start: Date | undefined,
  end: Date | undefined,
  timePreset: TimeRangePreset | TimeComparisonOption | undefined,
  timeZone: string,
): string => {
  const isAllTime = timePreset === TimeRangePreset.ALL_TIME;

  if (!start) {
    return end ? `- ${end}` : "";
  }
  if (!end) {
    return `${start} -`;
  }

  const {
    day: startDate,
    month: startMonth,
    year: startYear,
  } = getDateMonthYearForTimezone(start, timeZone);

  let {
    day: endDate,
    month: endMonth,
    year: endYear,
  } = getDateMonthYearForTimezone(end, timeZone);

  if (
    startDate === endDate &&
    startMonth === endMonth &&
    startYear === endYear
  ) {
    return `${start.toLocaleDateString(undefined, {
      month: "short",
      timeZone,
    })} ${startDate}, ${startYear} (${start
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone,
      })
      .replace(/\s/g, "")}-${end
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone,
      })
      .replace(/\s/g, "")})`;
  }

  const timeRangeDurationMs = getTimeWidth(start, end);
  if (
    timeRangeDurationMs <= durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    return `${start.toLocaleDateString(undefined, {
      month: "short",
      timeZone,
    })} ${startDate}-${endDate}, ${startYear} (${start
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone,
      })
      .replace(/\s/g, "")}-${end
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone,
      })
      .replace(/\s/g, "")})`;
  }

  let inclusiveEndDate;

  let timeString = "";

  const startTime = start.toLocaleTimeString(undefined, { timeZone });
  const endTime = end.toLocaleTimeString(undefined, { timeZone });

  if (isAllTime) {
    inclusiveEndDate = new Date(end);
  } else if (startTime === "12:00:00 am" && endTime === "12:00:00 am") {
    // beyond this point, we're dealing with time ranges that are full day periods
    // since time range is exclusive at the end, we need to subtract a day
    inclusiveEndDate = new Date(
      end.getTime() - durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration),
    );

    const inclusiveEndDateWithTimeZone = getDateMonthYearForTimezone(
      inclusiveEndDate,
      timeZone,
    );

    endDate = inclusiveEndDateWithTimeZone.day;
    endMonth = inclusiveEndDateWithTimeZone.month;
    endYear = inclusiveEndDateWithTimeZone.year;
  } else {
    // display full time when the hours are not at 00:00
    inclusiveEndDate = end;

    timeString = `(${start
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone,
      })
      .replace(/\s/g, "")}-${end
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone,
      })
      .replace(/\s/g, "")})`;
  }

  // month is the same
  if (startMonth === endMonth && startYear === endYear) {
    return `${start.toLocaleDateString(undefined, {
      month: "short",
      timeZone,
    })} ${startDate}-${endDate}, ${startYear} ${timeString}`;
  }

  // year is the same
  if (startYear === endYear) {
    return `${start.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      timeZone,
    })} - ${inclusiveEndDate.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      timeZone,
    })}, ${startYear} ${timeString}`;
  }
  // year is different
  const dateFormatOptions: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "short",
    day: "numeric",
    timeZone,
  };
  return `${start.toLocaleDateString(
    undefined,
    dateFormatOptions,
  )} - ${inclusiveEndDate.toLocaleDateString(undefined, dateFormatOptions)}`;
};

/**
 * Return start and end date such that the results include
 * extra data points for extrapolating the chart on both ends
 */
export function getAdjustedFetchTime(
  startTime: Date,
  endTime: Date,
  zone: string,
  interval: V1TimeGrain | undefined,
) {
  if (!startTime || !endTime || !interval)
    return { start: startTime?.toISOString(), end: endTime?.toISOString() };
  // const offsetedStartTime = getOffset(
  //   startTime,
  //   TIME_GRAIN[interval].duration,
  //   TimeOffsetType.SUBTRACT,
  // );

  // the data point previous to the first date inside the chart.
  const fetchStartTime = getStartOfPeriod(
    startTime,
    TIME_GRAIN[interval].duration,
    zone,
  );

  const endOfPeriod = getEndOfPeriod(
    endTime,
    TIME_GRAIN[interval].duration,
    zone,
  );

  const fetchEndTime = getOffset(
    endOfPeriod,
    TIME_GRAIN[interval].duration,
    TimeOffsetType.ADD,
  );

  // the data point after the last complete date.

  return {
    start: fetchStartTime.toISOString(),
    end: fetchEndTime.toISOString(),
  };
}

/**
 * Return start and end date to be used as extents of the
 * time series charts
 */
export function getAdjustedChartTime(
  start: Date | undefined,
  end: Date | undefined,
  zone: string,
  interval: V1TimeGrain | undefined,
  timePreset: TimeRangePreset | TimeComparisonOption | undefined,
  defaultTimeRange: string | undefined,
  chartType: TDDChart,
) {
  if (!start || !end || !interval)
    return {
      start,
      end,
    };

  const grainDuration = TIME_GRAIN[interval].duration;
  const offsetDuration = getDurationMultiple(grainDuration, 0.45);

  let adjustedStart = new Date(start);
  let adjustedEnd = new Date(end);

  if (timePreset === TimeRangePreset.ALL_TIME) {
    // No offset has been applied to All time range so far
    // Adjust end according to the interval
    adjustedStart = getStartOfPeriod(adjustedStart, grainDuration, zone);
    adjustedStart = getOffset(
      adjustedStart,
      offsetDuration,
      TimeOffsetType.ADD,
    );
    adjustedEnd = getEndOfPeriod(adjustedEnd, grainDuration, zone);
  } else if (timePreset && timePreset === TimeRangePreset.DEFAULT) {
    // For default presets the iso range can be mixed. There the offset added will be the smallest unit in the range.
    // But for the graph we need the offset based on selected grain.
    const smallestTimeGrain = getSmallestTimeGrain(defaultTimeRange);
    // Only add this if the selected grain is greater than the smallest unit in the iso range
    if (smallestTimeGrain && isGrainBigger(interval, smallestTimeGrain)) {
      adjustedEnd = getEndOfPeriod(adjustedEnd, grainDuration, zone);
    }
  } else {
    // Make sure end is always at the end of the period
    adjustedEnd = getEndOfPeriod(
      new Date(adjustedEnd.getTime() - 1),
      grainDuration,
      zone,
    );
  }

  if (
    chartType === TDDChart.GROUPED_BAR ||
    chartType === TDDChart.STACKED_BAR
  ) {
    adjustedStart = getOffset(
      adjustedStart,
      offsetDuration,
      TimeOffsetType.SUBTRACT,
    );
  }

  /**
   * We need to remove the offset from the end to remove whitespace
   */
  adjustedEnd = getOffset(adjustedEnd, offsetDuration, TimeOffsetType.SUBTRACT);

  return {
    /**
     * Values in the charts are displayed in the local time zone.
     * To get the correct values, we need to remove the local time zone offset
     * and add the offset of the selected time zone.
     */
    start: addZoneOffset(removeLocalTimezoneOffset(adjustedStart), zone),
    end: addZoneOffset(removeLocalTimezoneOffset(adjustedEnd), zone),
  };
}

function isWithinRange(time: Date, start: Date, end: Date) {
  return time.getTime() >= start.getTime() && time.getTime() <= end.getTime();
}
