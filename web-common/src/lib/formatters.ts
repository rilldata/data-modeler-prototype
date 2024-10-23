import { removeLocalTimezoneOffset } from "@rilldata/web-common/lib/time/timezone";
import { format } from "d3-format";
import { timeFormat } from "d3-time-format";
import {
  CATEGORICALS,
  DATES,
  FLOATS,
  INTEGERS,
  type Interval,
  INTERVALS,
  isList,
  isNested,
  isStruct,
  PreviewRollupInterval,
  TIMESTAMPS,
} from "./duckdb-data-types";
import { formatDuckdbIntervalLossless } from "./number-formatting/strategies/intervals";

/** This heuristic is courtesy Dominik Moritz.
 * Best used in cases where (1) you have no context for the number, and (2) you
 * want have "enough resolution to distinguish numbers when they should be distinguishable."
 */
export function justEnoughPrecision(n: number) {
  if (typeof n !== "number") throw Error("argument must be a number");
  // return only integer in this case.
  if (n >= 10 ** 4) return n.toFixed();
  const str = n.toString();
  // if there are no floating point digits, return the string
  if (n === Math.round(n)) return str;

  // otherwise, proceed.
  const [left, right] = str.split(".");

  // count the integer side
  const leftSideDigits = left
    .split("")
    .filter((l) => l !== "-") // remove the negative sign
    .join("").length;

  // calculate the remaining available precision
  const remainingPrecision = Math.max(0, 5 - leftSideDigits);
  // take the remaining precision from the floating point side.
  const remainingFloatingPoints = right.slice(0, remainingPrecision);
  // format a new string
  return `${left}${remainingFloatingPoints.length ? "." : ""}${
    remainingFloatingPoints || ""
  }`;
}

const zeroPad = format("02d");
const msPad = format("03d");
export const formatInteger = format(",");
const formatRate = format(".1f");

/**  */
export const singleDigitPercentage = format(".0%");

/**
 * changes precision depending on the
 */
export function formatBigNumberPercentage(v) {
  if (v < 0.0001) {
    const f = format(".4%")(v);
    if (f === "0.0000%") {
      return "~ 0%";
    } else {
      return f;
    }
  } else {
    return format(".2%")(v);
  }
}

export const standardTimestampFormat = (v, type = "TIMESTAMP") => {
  let fmt = timeFormat("%Y-%m-%d %H:%M:%S Z");
  if (type === "DATE") {
    fmt = timeFormat("%Y-%m-%d");
  }
  return fmt(removeLocalTimezoneOffset(new Date(v)));
};

export const fullTimestampFormat = (v) => {
  const fmt = timeFormat("%Y-%m-%d %H:%M:%S.%L");
  return fmt(removeLocalTimezoneOffset(new Date(v)));
};

export const datePortion = timeFormat("%Y-%m-%d");
export const timePortion = timeFormat("%H:%M:%S");

export function microsToTimestring(microseconds: number) {
  // to format micros, we need to translate this to hh:mm:ss.
  // start with hours/
  const sign = Math.sign(microseconds);
  const micros = Math.abs(microseconds);
  const hours = Math.trunc(micros / 1000 / 1000 / 60 / 60);
  let remaining = micros - hours * 1000 * 1000 * 60 * 60;
  const minutes = Math.trunc(remaining / 1000 / 1000 / 60);
  //const seconds = (remaining - (minutes * 1000 * 1000 * 60)) / 1000 / 1000;
  remaining -= minutes * 1000 * 1000 * 60;
  const seconds = Math.trunc(remaining / 1000 / 1000);
  remaining -= seconds * 1000 * 1000;
  const ms = Math.trunc(remaining / 1000);
  if (hours === 0 && minutes === 0 && seconds === 0 && ms > 0) {
    return `${sign == 1 ? "" : "-"}${ms}ms`;
  }
  return `${sign == 1 ? "" : "-"}${zeroPad(hours)}:${zeroPad(
    minutes,
  )}:${zeroPad(seconds)}.${msPad(ms)}`;
}

/** convert a start and end date
 * to a human readable time range (e.g. 6 months, 24 years, etc)
 */
export function datesToFormattedTimeRange(start: Date, end: Date) {
  const interval = (end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24);
  return intervalToTimestring(interval);
}

export function intervalToTimestring(inputInterval: Interval | number) {
  const interval =
    typeof inputInterval === "number"
      ? { months: 0, days: inputInterval, micros: 0 }
      : inputInterval;
  const months = interval.months
    ? `${formatInteger(interval.months)} month${
        interval.months > 1 ? "s" : ""
      } `
    : "";
  const days = interval.days
    ? `${justEnoughPrecision(interval.days)} day${
        interval.days > 1 ? "s" : ""
      } `
    : "";
  const time =
    interval.months > 0 || interval.days > 1
      ? ""
      : microsToTimestring(interval.micros);
  // if only days && days > 365, convert to years?
  if (interval.months === 0 && interval.days > 0 && interval.days > 365)
    return `${formatRate(interval.days / 365)} years`;
  return `${months}${days}${time}`;
}

export function formatCompactInteger(n: number) {
  let fmt: (number) => string;
  if (n <= 1000) {
    fmt = formatInteger;
    return fmt(Math.trunc(n));
  } else {
    fmt = format(".3s");
    return fmt(n);
  }
}

export function formatDataType(value: unknown, type: string) {
  if (value === undefined) return "";
  if (INTEGERS.has(type) || type?.startsWith("DECIMAL")) {
    return value;
  } else if (FLOATS.has(type)) {
    return value;
  } else if (CATEGORICALS.has(type)) {
    return value;
  } else if (TIMESTAMPS.has(type)) {
    return standardTimestampFormat(value, type);
  } else if (INTERVALS.has(type)) {
    return intervalToTimestring(value as Interval);
  } else if (isStruct(type)) {
    return JSON.stringify(value).replace(/"/g, "'");
  } else if (isList(type)) {
    return value !== undefined && value !== null
      ? `[${(value as Array<any>)
          ?.map((entry) => (+entry ? +entry : `'${entry}'`))
          ?.join(", ")}]`
      : `null`;
  } else if (isNested(type)) {
    return JSON.stringify(value).replace(/"/g, "'");
  }
  return JSON.stringify(value).replace(/"/g, "'");
}

/**
 * Formats a value as a string that can be used in a duckdb query.
 * This is not intended for display purposes, but useful for
 * situations like the shift-click copy action where a string is
 * needed for use in further queries. Ideally, this string should
 * parse to _exactly_ the same value as the original value that is
 * passed in (as of 2023-08, this is aspirational, and this
 * function cannot be relied upon to do that for all data types)
 *
 * TODO: make sure this is used everywhere the shift-click action
 * returns a string that is likely to be used in a query.
 * TODO: make sure this returns the correct string for all data types.
 * As of the initial implementation in 2023-08, this provides parity
 * with (and slightly improves) the existing shift-click action, but
 * it has not been fully thought out for all datatypes.
 *
 * @param value
 * @param type
 */
export function formatDataTypeAsDuckDbQueryString(
  value: unknown,
  type: string,
): string {
  if (value === undefined) return "undefined";
  if (value === null) return "null";
  if (
    INTEGERS.has(type) ||
    type?.startsWith("DECIMAL") ||
    CATEGORICALS.has(type) ||
    FLOATS.has(type)
  ) {
    return value.toString();
  } else if (DATES.has(type)) {
    // NOTE: `DATE` must come before `TIMESTAMP` in this list
    // because `DATE` is a subset of `TIMESTAMP`.
    return `DATE '${standardTimestampFormat(value, type)}'`;
  } else if (TIMESTAMPS.has(type)) {
    return `TIMESTAMP '${standardTimestampFormat(value, type)}'`;
  } else if (INTERVALS.has(type)) {
    return `INTERVAL '${formatDuckdbIntervalLossless(value as Interval)}'`;
  } else if (isStruct(type)) {
    return JSON.stringify(value).replace(/"/g, "'");
  } else if (isList(type)) {
    return value !== undefined && value !== null
      ? `[${(value as Array<any>)
          ?.map((entry) => (+entry ? +entry : `'${entry}'`))
          ?.join(", ")}]`
      : `null`;
  } else if (isNested(type)) {
    return JSON.stringify(value).replace(/"/g, "'");
  }
  return JSON.stringify(value).replace(/"/g, "'");
}

const what = "2";

console.log(+what);

/** These will be used in the string */
export const PreviewRollupIntervalFormatter = {
  [PreviewRollupInterval.ms]:
    "millisecond-level" /** showing rows binned by ms */,
  [PreviewRollupInterval.second]:
    "second-level" /** showing rows binned by second */,
  [PreviewRollupInterval.minute]:
    "minute-level" /** showing rows binned by minute */,
  [PreviewRollupInterval.hour]: "hourly" /** showing hourly counts */,
  [PreviewRollupInterval.day]: "daily" /** showing daily counts */,
  [PreviewRollupInterval.month]: "monthly" /** showing monthly counts */,
  [PreviewRollupInterval.year]: "yearly" /** showing yearly counts */,
};
