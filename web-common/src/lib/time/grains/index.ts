/** Utility functions for using time grains within a Rill dashboard.
 * Most of these functions utilize the TIME_GRAIN object defined in config.ts
 * to generate either a subset of time grains or a single time grain.
 */

import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { Duration } from "luxon";
import { TIME_GRAIN } from "../config";
import { getTimeWidth } from "../transforms";
import type { AvailableTimeGrain, TimeGrain } from "../types";

export function unitToTimeGrain(unit: string): V1TimeGrain {
  return (
    Object.values(TIME_GRAIN).find((timeGrain) => timeGrain.label === unit)
      ?.grain || V1TimeGrain.TIME_GRAIN_UNSPECIFIED
  );
}

export function durationToMillis(duration: string): number {
  return Duration.fromISO(duration).toMillis();
}

// Get the default grain for a given time range.
export function getDefaultTimeGrain(start: Date, end: Date): TimeGrain {
  const timeRangeDurationMs = end.getTime() - start.getTime();
  if (
    timeRangeDurationMs <
    2 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration)
  ) {
    return TIME_GRAIN.TIME_GRAIN_MINUTE;
  } else if (
    timeRangeDurationMs < durationToMillis(TIME_GRAIN.TIME_GRAIN_WEEK.duration)
  ) {
    return TIME_GRAIN.TIME_GRAIN_HOUR;
  } else if (
    timeRangeDurationMs <
    durationToMillis(TIME_GRAIN.TIME_GRAIN_QUARTER.duration)
  ) {
    return TIME_GRAIN.TIME_GRAIN_DAY;
  } else if (
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration)
  ) {
    return TIME_GRAIN.TIME_GRAIN_WEEK;
  } else {
    return TIME_GRAIN.TIME_GRAIN_MONTH;
  }
}

// Return time grains that are allowed for a given time range.
export function getAllowedTimeGrains(start: Date, end: Date): TimeGrain[] {
  const timeRangeDurationMs = getTimeWidth(start, end);
  if (
    timeRangeDurationMs < durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration)
  ) {
    return [TIME_GRAIN.TIME_GRAIN_MINUTE];
  } else if (
    timeRangeDurationMs < durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    return [TIME_GRAIN.TIME_GRAIN_MINUTE, TIME_GRAIN.TIME_GRAIN_HOUR];
  } else if (
    timeRangeDurationMs < durationToMillis(TIME_GRAIN.TIME_GRAIN_WEEK.duration)
  ) {
    return [TIME_GRAIN.TIME_GRAIN_HOUR, TIME_GRAIN.TIME_GRAIN_DAY];
  } else if (
    timeRangeDurationMs < durationToMillis(TIME_GRAIN.TIME_GRAIN_MONTH.duration)
  ) {
    return [
      TIME_GRAIN.TIME_GRAIN_HOUR,
      TIME_GRAIN.TIME_GRAIN_DAY,
      TIME_GRAIN.TIME_GRAIN_WEEK,
    ];
  } else if (
    timeRangeDurationMs <
    durationToMillis(TIME_GRAIN.TIME_GRAIN_QUARTER.duration)
  ) {
    return [
      TIME_GRAIN.TIME_GRAIN_DAY,
      TIME_GRAIN.TIME_GRAIN_WEEK,
      TIME_GRAIN.TIME_GRAIN_MONTH,
    ];
  } else if (
    timeRangeDurationMs < durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration)
  ) {
    return [
      TIME_GRAIN.TIME_GRAIN_DAY,
      TIME_GRAIN.TIME_GRAIN_WEEK,
      TIME_GRAIN.TIME_GRAIN_MONTH,
      TIME_GRAIN.TIME_GRAIN_QUARTER,
    ];
  } else {
    return [
      TIME_GRAIN.TIME_GRAIN_WEEK,
      TIME_GRAIN.TIME_GRAIN_MONTH,
      TIME_GRAIN.TIME_GRAIN_QUARTER,
      TIME_GRAIN.TIME_GRAIN_YEAR,
    ];
  }
}

const APITimeGrainOrder: V1TimeGrain[] = [
  V1TimeGrain.TIME_GRAIN_MILLISECOND,
  V1TimeGrain.TIME_GRAIN_SECOND,
  V1TimeGrain.TIME_GRAIN_MINUTE,
  V1TimeGrain.TIME_GRAIN_HOUR,
  V1TimeGrain.TIME_GRAIN_DAY,
  V1TimeGrain.TIME_GRAIN_WEEK,
  V1TimeGrain.TIME_GRAIN_MONTH,
  V1TimeGrain.TIME_GRAIN_QUARTER,
  V1TimeGrain.TIME_GRAIN_YEAR,
];

export function isGrainBigger(
  possiblyBiggerGrain: V1TimeGrain,
  possiblySmallerGrain: V1TimeGrain,
): boolean {
  const biggerIndex = APITimeGrainOrder.indexOf(possiblyBiggerGrain);
  const smallerIndex = APITimeGrainOrder.indexOf(possiblySmallerGrain);
  return biggerIndex > smallerIndex;
}
export function getMinGrain(...grains: V1TimeGrain[]) {
  let minGrain: V1TimeGrain = V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  let minGrainIndex = APITimeGrainOrder.length;

  for (const grain of grains) {
    if (grain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED) continue;
    const grainIndex = APITimeGrainOrder.indexOf(grain);
    if (grainIndex < minGrainIndex) {
      minGrain = grain;
      minGrainIndex = grainIndex;
    }
  }

  return minGrain;
}

export function checkValidTimeGrain(
  timeGrain: V1TimeGrain | undefined,
  timeGrainOptions: TimeGrain[],
  minTimeGrain: V1TimeGrain,
): boolean {
  if (!timeGrain) return false;
  if (!timeGrainOptions.find((t) => t.grain === timeGrain)) return false;

  // If minTimeGrain is not specified, then all available timeGrains are valid
  if (minTimeGrain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED) return true;

  const isGrainPossible = !isGrainBigger(minTimeGrain, timeGrain);
  return isGrainPossible;
}

export function findValidTimeGrain(
  timeGrain: V1TimeGrain,
  timeGrainOptions: TimeGrain[],
  minTimeGrain: V1TimeGrain,
): V1TimeGrain {
  const timeGrains = Object.values(TIME_GRAIN).map(
    (timeGrain) => timeGrain.grain,
  );

  const defaultIndex = timeGrains.indexOf(timeGrain);

  // Loop through the timeGrains starting from the default value
  for (let i = defaultIndex; i < timeGrains.length; i++) {
    const currentGrain = timeGrains[i];

    if (checkValidTimeGrain(currentGrain, timeGrainOptions, minTimeGrain)) {
      return currentGrain;
    }
  }
  // If no valid timeGrain is found, loop from the beginning of the array
  for (let i = 0; i < defaultIndex; i++) {
    const currentGrain = timeGrains[i];

    if (checkValidTimeGrain(currentGrain, timeGrainOptions, minTimeGrain)) {
      return currentGrain;
    }
  }

  // If no valid timeGrain is found, return the default timeGrain as fallback
  return timeGrain;
}

export function mapDurationToGrain(duration: string): V1TimeGrain {
  for (const g in TIME_GRAIN) {
    if (TIME_GRAIN[g].duration === duration) {
      return TIME_GRAIN[g].grain;
    }
  }
  return V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
}

export function timeGrainToDuration(timeGrain: V1TimeGrain): string {
  if (isAvailableTimeGrain(timeGrain)) {
    const grainConfig = TIME_GRAIN[timeGrain];
    return grainConfig.duration;
  } else {
    console.warn("Requested duration for invalid time grain: ", timeGrain);
    // Default to 1 day if the time grain is invalid to fail gracefully
    return "P1D";
  }
}

export function isAvailableTimeGrain(
  grain: V1TimeGrain,
): grain is AvailableTimeGrain {
  return (
    grain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED &&
    grain !== V1TimeGrain.TIME_GRAIN_MILLISECOND &&
    grain !== V1TimeGrain.TIME_GRAIN_SECOND
  );
}
