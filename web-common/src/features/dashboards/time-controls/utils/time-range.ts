import type { V1TimeGrain } from "../../../../runtime-client";
import {
  isRangeInsideOther,
  relativePointInTimeToAbsolute,
} from "./time-anchors";
import {
  getAllowedTimeGrains,
  getTimeGrainFromRuntimeGrain,
  isMinGrainBigger,
} from "./time-grain";
import {
  Period,
  RangePreset,
  ReferencePoint,
  TimeOffsetType,
  TimeRange,
  TimeRangeOption,
  TimeTruncationType,
} from "./time-types";

export const TIME_RANGES: TimeRange[] = [
  {
    label: "Last 6 Hours",
    rangePreset: RangePreset.OFFSET_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "PT6H", operationType: TimeOffsetType.SUBTRACT }, // operation
        {
          period: Period.HOUR, //TODO: How to handle user selected timegrains?
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "PT1H", operationType: TimeOffsetType.SUBTRACT },
        {
          period: Period.HOUR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  {
    label: "Last day",
    rangePreset: RangePreset.OFFSET_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "P1D", operationType: TimeOffsetType.SUBTRACT }, // operation
        {
          period: Period.HOUR, //TODO: How to handle user selected timegrains?
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "PT1H", operationType: TimeOffsetType.SUBTRACT },
        {
          period: Period.HOUR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  {
    label: "All time data",
    rangePreset: RangePreset.ALL_TIME,
  },
];

// Loop through all presets to check if they can be a part of subset of given start and end date
export function getChildTimeRanges(
  start: Date,
  end: Date,
  minTimeGrain: V1TimeGrain
): TimeRangeOption[] {
  const timeRanges: TimeRangeOption[] = [];

  for (const timeRange of TIME_RANGES) {
    if (timeRange.rangePreset == RangePreset.ALL_TIME) {
      // All time is always an option
      timeRanges.push({
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
      const isValidTimeRange = isRangeInsideOther(
        timeRangeDates.startDate,
        timeRangeDates.endDate,
        start,
        end
      );

      const allowedTimeGrains = getAllowedTimeGrains(start, end);
      const allowedMaxGrain = allowedTimeGrains[allowedTimeGrains.length - 1];
      const isGrainPossible = !isMinGrainBigger(minTimeGrain, allowedMaxGrain);

      if (isValidTimeRange && isGrainPossible) {
        timeRanges.push({
          label: timeRange.label,
          start: timeRangeDates.startDate,
          end: timeRangeDates.endDate,
        });
      }
    }
  }

  return timeRanges;
}

export function isTimeRangeValidForMinTimeGrain(
  start: Date,
  end: Date,
  minTimeGrain: V1TimeGrain
): boolean {
  const timeGrain = getTimeGrainFromRuntimeGrain(minTimeGrain);
  if (!timeGrain) return true;
  if (!start || !end) return true;

  const allowedTimeGrains = getAllowedTimeGrains(start, end);
  const maxAllowedTimeGrain = allowedTimeGrains[allowedTimeGrains.length - 1];
  return !isMinGrainBigger(minTimeGrain, maxAllowedTimeGrain);
}

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

export function getDateFromISOString(isoString: string): string {
  return isoString.split("T")[0];
}

export function getISOStringFromDate(date: string): string {
  return date + "T00:00:00.000Z";
}
