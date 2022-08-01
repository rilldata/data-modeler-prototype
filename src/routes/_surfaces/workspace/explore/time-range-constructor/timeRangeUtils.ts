import {
  TimeGrain,
  TimeRangeName,
  TimeSeriesTimeRange,
} from "$common/database-service/DatabaseTimeSeriesActions";

// TODO: replace this with a call to the `/meta?metricsDefId={metricsDefId}` endpoint, once it's available
export const getSelectableTimeRangeNames = (
  allTimeRange: TimeSeriesTimeRange
): TimeRangeName[] => {
  if (!allTimeRange) return [];

  const allTimeRangeDuration = getTimeRangeDuration(
    TimeRangeName.AllTime,
    allTimeRange
  );

  const selectableTimeRangeNames: TimeRangeName[] = [];
  for (const timeRangeName in TimeRangeName) {
    const timeRangeDuration = getTimeRangeDuration(
      TimeRangeName[timeRangeName],
      allTimeRange
    );
    // only show a time range if it is within the time range of the data
    const showTimeRange = allTimeRangeDuration >= timeRangeDuration;
    if (showTimeRange) {
      selectableTimeRangeNames.push(TimeRangeName[timeRangeName]);
    }
  }

  return selectableTimeRangeNames;
};

// TODO: replace this with a call to the `/meta?metricsDefId={metricsDefId}` endpoint, once it's available
export const getDefaultTimeRangeName = (): TimeRangeName => {
  // Use AllTime for now. When we go to production real-time datasets, we'll want to change this.
  return TimeRangeName.AllTime;
};

// This is for pre-set relative time ranges – where the start/end dates are not yet deterimined.
// For custom time ranges, we'll need another function with "breakpoint" logic that analyzes the user-determined start/end dates.
export const getSelectableTimeGrains = (
  timeRangeName: TimeRangeName,
  allTimeRange: TimeSeriesTimeRange
): TimeGrain[] => {
  if (!allTimeRange) return [];
  const timeRangeDuration = getTimeRangeDuration(timeRangeName, allTimeRange);

  const timeGrains: TimeGrain[] = [];
  for (const timeGrain in TimeGrain) {
    // only show a time grain if it results in a reasonable number of points on the line chart
    const MINIMUM_POINTS_ON_LINE_CHART = 2;
    const MAXIMUM_POINTS_ON_LINE_CHART = 2500;
    const timeGrainDuration = getTimeGrainDuration(TimeGrain[timeGrain]);
    const pointsOnLineChart = timeRangeDuration / timeGrainDuration;
    const showTimeGrain =
      pointsOnLineChart >= MINIMUM_POINTS_ON_LINE_CHART &&
      pointsOnLineChart <= MAXIMUM_POINTS_ON_LINE_CHART;
    if (showTimeGrain) {
      timeGrains.push(TimeGrain[timeGrain]);
    }
  }
  if (timeGrains.length === 0) {
    throw new Error(`No time grains generated for time range ${timeRangeName}`);
  }
  return timeGrains;
};

export const getDefaultTimeGrain = (
  timeRangeName: TimeRangeName
): TimeGrain => {
  switch (timeRangeName) {
    case TimeRangeName.LastHour:
      return TimeGrain.FifteenMinutes;
    case TimeRangeName.Last6Hours:
      return TimeGrain.FifteenMinutes;
    case TimeRangeName.LastDay:
      return TimeGrain.OneHour;
    case TimeRangeName.Last2Days:
      return TimeGrain.OneHour;
    case TimeRangeName.Last5Days:
      return TimeGrain.OneHour;
    case TimeRangeName.LastWeek:
      return TimeGrain.OneHour;
    case TimeRangeName.Last2Weeks:
      return TimeGrain.OneDay;
    case TimeRangeName.Last30Days:
      return TimeGrain.OneDay;
    case TimeRangeName.Last60Days:
      return TimeGrain.OneDay;
    case TimeRangeName.AllTime:
      // TODO: this needs breakpoint logic using start/end time.
      return TimeGrain.OneDay;
    default:
      throw new Error(`No default time grain for time range ${timeRangeName}`);
  }
};

export const makeTimeRange = (
  timeRangeName: TimeRangeName,
  timeGrain: TimeGrain,
  allTimeRange: TimeSeriesTimeRange
): TimeSeriesTimeRange => {
  switch (timeRangeName) {
    case TimeRangeName.LastHour: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last6Hours: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 6 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.LastDay: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last2Days: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 2 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last5Days: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 5 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.LastWeek: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 7 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last2Weeks: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(
        endDate.getTime() - 2 * 7 * 24 * 60 * 60 * 1000
      );
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last30Days: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 30 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last60Days: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 60 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.AllTime: {
      const startDate = roundDateDown(new Date(allTimeRange?.start), timeGrain);
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    default:
      throw new Error(`Unknown time range name: ${timeRangeName}`);
  }
};

export const makeTimeRanges = (
  timeRangeNames: TimeRangeName[],
  allTimeRangeInDataset: TimeSeriesTimeRange
): TimeSeriesTimeRange[] => {
  if (!timeRangeNames || !allTimeRangeInDataset) return [];

  const timeRanges: TimeSeriesTimeRange[] = [];
  for (const timeRangeName of timeRangeNames) {
    const defaultTimeGrain = getDefaultTimeGrain(timeRangeName);
    const timeRange = makeTimeRange(
      timeRangeName,
      defaultTimeGrain,
      allTimeRangeInDataset
    );
    timeRanges.push(timeRange);
  }
  return timeRanges;
};

export const prettyFormatTimeRange = (
  timeRange: TimeSeriesTimeRange
): string => {
  if (!timeRange?.start && timeRange?.end) {
    return `- ${timeRange.end}`;
  }

  if (timeRange?.start && !timeRange?.end) {
    return `${timeRange.start} -`;
  }

  if (!timeRange?.start && !timeRange?.end) {
    return "";
  }

  const start = new Date(timeRange.start);
  const end = new Date(timeRange.end);

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

export const prettyTimeGrain = (timeGrain: TimeGrain): string => {
  if (!timeGrain) return "";
  switch (timeGrain) {
    case TimeGrain.FiveMinutes:
      return "5 minute";
    case TimeGrain.FifteenMinutes:
      return "15 minute";
    case TimeGrain.OneHour:
      return "hourly";
    case TimeGrain.OneDay:
      return "daily";
    case TimeGrain.OneWeek:
      return "weekly";
    case TimeGrain.OneMonth:
      return "monthly";
    case TimeGrain.OneYear:
      return "yearly";
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

const getTimeRangeDuration = (
  timeRangeName: TimeRangeName,
  allTimeRange: TimeSeriesTimeRange
): number => {
  switch (timeRangeName) {
    case TimeRangeName.LastHour:
      return 60 * 60 * 1000;
    case TimeRangeName.Last6Hours:
      return 6 * 60 * 60 * 1000;
    case TimeRangeName.LastDay:
      return 24 * 60 * 60 * 1000;
    case TimeRangeName.Last2Days:
      return 2 * 24 * 60 * 60 * 1000;
    case TimeRangeName.Last5Days:
      return 5 * 24 * 60 * 60 * 1000;
    case TimeRangeName.LastWeek:
      return 7 * 24 * 60 * 60 * 1000;
    case TimeRangeName.Last2Weeks:
      return 2 * 7 * 24 * 60 * 60 * 1000;
    case TimeRangeName.Last30Days:
      return 30 * 24 * 60 * 60 * 1000;
    case TimeRangeName.Last60Days:
      return 60 * 24 * 60 * 60 * 1000;
    case TimeRangeName.AllTime:
      return (
        new Date(allTimeRange.end).getTime() -
        new Date(allTimeRange.start).getTime()
      );
    default:
      throw new Error(`Unknown time range name: ${timeRangeName}`);
  }
};

const getTimeGrainDuration = (timeGrain: TimeGrain): number => {
  switch (timeGrain) {
    case TimeGrain.FiveMinutes:
      return 5 * 60 * 1000;
    case TimeGrain.FifteenMinutes:
      return 15 * 60 * 1000;
    case TimeGrain.OneHour:
      return 60 * 60 * 1000;
    case TimeGrain.OneDay:
      return 24 * 60 * 60 * 1000;
    case TimeGrain.OneWeek:
      return 7 * 24 * 60 * 60 * 1000;
    case TimeGrain.OneMonth:
      return 30 * 24 * 60 * 60 * 1000;
    case TimeGrain.OneYear:
      return 365 * 24 * 60 * 60 * 1000;
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

const roundDateDown = (date: Date | undefined, timeGrain: TimeGrain): Date => {
  if (!date) return new Date();
  switch (timeGrain) {
    case TimeGrain.FiveMinutes: {
      const interval = 5 * 60 * 1000;
      return new Date(Math.round(date.getTime() / interval) * interval);
    }
    case TimeGrain.FifteenMinutes: {
      const interval = 15 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneHour: {
      const interval = 60 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneDay: {
      const interval = 24 * 60 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneWeek: {
      // rounds to the most recent Monday
      const day = date.getUTCDay();
      const dateRoundedDownByDay = roundDateDown(date, TimeGrain.OneDay);
      const timeFromMonday = (day === 0 ? 6 : day - 1) * 24 * 60 * 60 * 1000;
      return new Date(dateRoundedDownByDay.getTime() - timeFromMonday);
    }
    case TimeGrain.OneMonth: {
      // rounds to the 1st of the current month
      return new Date(date.getUTCFullYear(), date.getUTCMonth(), 1);
    }
    case TimeGrain.OneYear: {
      // rounds to January 1st of the current year
      return new Date(date.getUTCFullYear(), 1, 1);
    }
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

const roundDateUp = (date: Date | undefined, timeGrain: TimeGrain): Date => {
  if (!date) return new Date();
  switch (timeGrain) {
    case TimeGrain.FiveMinutes: {
      const interval = 5 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    case TimeGrain.FifteenMinutes: {
      const interval = 15 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneHour: {
      const interval = 60 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneDay: {
      const interval = 24 * 60 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneWeek: {
      // rounds to the next Monday
      const day = date.getUTCDay();
      const dateRoundedDownByDay = roundDateDown(date, TimeGrain.OneDay);
      const timeUntilMonday = (day === 0 ? 1 : 8 - day) * 24 * 60 * 60 * 1000;
      return new Date(dateRoundedDownByDay.getTime() + timeUntilMonday);
    }
    case TimeGrain.OneMonth: {
      // rounds to the 1st of the next month
      return new Date(date.getUTCFullYear(), date.getUTCMonth() + 1, 1);
    }
    case TimeGrain.OneYear: {
      // rounds to Jan 1st of the next year
      return new Date(date.getUTCFullYear() + 1, 1, 1);
    }
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};
