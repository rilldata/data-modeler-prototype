import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges";
import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
import { Duration, Interval } from "luxon";
import { getTimeWidth, transformDate } from "../transforms";
import {
  type RelativeTimeTransformation,
  TimeComparisonOption,
  TimeOffsetType,
  TimeRangePreset,
} from "../types";

export function getComparisonTransform(
  start: Date,
  end: Date,
  comparison: TimeComparisonOption,
): RelativeTimeTransformation {
  if (
    comparison === TimeComparisonOption.CONTIGUOUS ||
    comparison === TimeComparisonOption.CUSTOM ||
    // @ts-expect-error Hack for `comparison` not being the correct type
    comparison === TimeRangePreset.CUSTOM
  ) {
    /** for custom & otherwise-contiguous comparisons,
     * we will calculate the width of the time range
     * and then subtract it from the start and end dates. */

    return {
      operationType: TimeOffsetType.SUBTRACT,
      duration: Interval.fromDateTimes(start, end).toDuration().toISO(),
    };
  } else {
    // map to a distinct Period-like TimeComparisonOption (e.g. "P1D")
    return {
      operationType: TimeOffsetType.SUBTRACT,
      duration: TIME_COMPARISON[comparison].offsetIso,
    };
  }
}

/** takes a start and end, and performs transformDate accordingly.
 * Contiguous periods (for instance "last 6 hours" or a custom range) is handled
 * a bit differently.
 */
export function getComparisonRange(
  start: Date,
  end: Date,
  comparison: TimeComparisonOption,
) {
  const transform = getComparisonTransform(start, end, comparison);
  return {
    start: transformDate(start, [transform]),
    end: transformDate(end, [transform]),
  };
}

/**
 * get the comparison range for a scrub such that
 * it aligns with the scrub start and end.
 */
export function getComparionRangeForScrub(
  start: Date,
  end: Date,
  comparisonStart: Date,
  comparisonEnd: Date,
  scrubStart: Date,
  scrubEnd: Date,
) {
  // validate if selected range and comparison range are of equal width
  if (
    getTimeWidth(start, end) !== getTimeWidth(comparisonStart, comparisonEnd)
  ) {
    // TODO: Have better handling on uneven widths caused by custom comparisons
    return { start: comparisonStart, end: comparisonEnd };
  } else {
    const startMOffset = getTimeWidth(start, scrubStart);
    const endMOffset = getTimeWidth(scrubEnd, end);

    return {
      start: new Date(comparisonStart.getTime() + startMOffset),
      end: new Date(comparisonEnd.getTime() - endMOffset),
    };
  }
}

// Returns true if the comparison range is inside the bounds.
export function isComparisonInsideBounds(
  // the earliest possible Date for this data set
  boundStart: Date,
  // the latest possible Date for this data set
  boundEnd: Date,
  start: Date,
  end: Date,
  comparison: TimeComparisonOption,
) {
  // compute comparison start and ends.
  const { start: comparisonStart, end: comparisonEnd } = getComparisonRange(
    start,
    end,
    comparison,
  );
  // check if comparison bounds are inside the bounds.
  return comparisonStart >= boundStart && comparisonEnd <= boundEnd;
}

export function isRangeLargerThanDuration(
  start: Date,
  end: Date,
  duration: string,
) {
  if (duration === TimeComparisonOption.CONTIGUOUS.toString()) {
    return false;
  }

  // To account for possible leap years
  if (duration === "P1Y") {
    return end.getFullYear() - start.getFullYear() > 1;
  }

  return (
    Interval.fromDateTimes(start, end).toDuration().toMillis() >
    Duration.fromISO(duration).toMillis()
  );
}

// Checks if last period is a duplicate comparison.
function isLastPeriodDuplicate(
  start: Date,
  end: Date,
  comparisonOptions: Array<TimeComparisonOption>,
) {
  const lastPeriod = getComparisonRange(
    start,
    end,
    TimeComparisonOption.CONTIGUOUS,
  );

  comparisonOptions = comparisonOptions.filter(
    (option) =>
      option !== TimeComparisonOption.CUSTOM &&
      option !== TimeComparisonOption.CONTIGUOUS,
  );

  return comparisonOptions.some((option) => {
    const { start: comparisonStart, end: comparisonEnd } = getComparisonRange(
      start,
      end,
      option,
    );
    return (
      comparisonStart.getTime() === lastPeriod.start.getTime() &&
      comparisonEnd.getTime() === lastPeriod.end.getTime()
    );
  });
}

/** get the available comparison options for a selected time range + the boundary range.
 * This is used to populate the comparison dropdown.
 * We need to check boundary conditions on all sides, but ultimately the two checks per comparison option are:
 * 1. is the comparison range inside the bounds?
 * 2. is the comparison range larger than the selected time range?
 */
export function getAvailableComparisonsForTimeRange(
  // the earliest and latest possible Dates for this data set
  boundStart: Date,
  boundEnd: Date,
  // start and end Dates for the currently focused time range
  start: Date,
  end: Date,
  comparisonOptions: TimeComparisonOption[],
) {
  let comparisons = comparisonOptions.filter((comparison) => {
    if (comparison === TimeComparisonOption.CUSTOM) {
      return false;
    }

    return (
      isComparisonInsideBounds(
        boundStart,
        boundEnd,
        start,
        end,
        // treat a custom comparison as contiguous.
        comparison,
      ) &&
      !isRangeLargerThanDuration(
        start,
        end,
        TIME_COMPARISON[comparison].offsetIso,
      )
    );
  });

  if (isLastPeriodDuplicate(start, end, comparisonOptions)) {
    comparisons = comparisons.filter(
      (comparison) => comparison !== TimeComparisonOption.CONTIGUOUS,
    );
  }

  return comparisons;
}

/** A convenience function that gets comparison range and states whether it is within bounds. */
export function getTimeComparisonParametersForComponent(
  comparisonOption: TimeComparisonOption | undefined,
  boundStart: Date | null | undefined,
  boundEnd: Date | null | undefined,
  currentStart: Date | null | undefined,
  currentEnd: Date | null | undefined,
) {
  if (
    !comparisonOption ||
    !boundStart ||
    !currentStart ||
    !currentEnd ||
    !boundEnd
  ) {
    return {
      start: undefined,
      end: undefined,
      isComparisonRangeAvailable: false,
    };
  }

  const { start, end } = getComparisonRange(
    currentStart,
    currentEnd,
    comparisonOption,
  );

  const isComparisonRangeAvailable = isComparisonInsideBounds(
    boundStart,
    boundEnd,
    start,
    end,
    comparisonOption,
  );

  return {
    start,
    end,
    isComparisonRangeAvailable,
  };
}

export function getComparisonLabel(comparisonTimeRange: V1TimeRange) {
  if (
    !comparisonTimeRange.isoOffset ||
    comparisonTimeRange.isoOffset === TimeRangePreset.CUSTOM
  ) {
    return prettyFormatTimeRange(
      new Date(comparisonTimeRange.start ?? ""),
      new Date(comparisonTimeRange.end ?? ""),
      TimeRangePreset.CUSTOM,
      "UTC", // TODO
    );
  }
  switch (true) {
    case comparisonTimeRange.isoOffset === TimeRangePreset.ALL_TIME:
      return "All time";
    case comparisonTimeRange.isoDuration === comparisonTimeRange.isoOffset:
      return "Previous period";
    case comparisonTimeRange.isoOffset in TIME_COMPARISON:
      return TIME_COMPARISON[comparisonTimeRange.isoOffset].label;
    default:
      return `Last ${humaniseISODuration(comparisonTimeRange.isoOffset)}`;
  }
}
