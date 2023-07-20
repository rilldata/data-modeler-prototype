import { DateTime } from "luxon";
import { timeZoneNameToAbbreviationMap } from "@rilldata/web-common/lib/time/timezone/abbreviationMap";

export function getLocalIANA(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

export function getTimeZoneNameFromIANA(now: Date, iana: string): string {
  return DateTime.fromJSDate(now).setZone(iana).toFormat("ZZZZZ");
}

export function getAbbreviationForIANA(now: Date, iana: string): string {
  const zoneName = getTimeZoneNameFromIANA(now, iana);

  if (zoneName in timeZoneNameToAbbreviationMap)
    return timeZoneNameToAbbreviationMap[zoneName];

  // fallback to the offset
  return DateTime.fromJSDate(now).setZone(iana).toFormat("ZZZZ");
}

export function getOffsetForIANA(now: Date, iana: string): string {
  return DateTime.fromJSDate(now).setZone(iana).toFormat("ZZ");
}

export function getLabelForIANA(now: Date, iana: string) {
  const abbreviation = getAbbreviationForIANA(now, iana);
  const offset = getOffsetForIANA(now, iana);

  return {
    abbreviation,
    offset: `GMT ${offset}`,
    iana,
  };
}

export function getDateMonthYearForTimezone(date: Date, timezone: string) {
  const timeZoneDate = DateTime.fromJSDate(date).setZone(timezone);
  const day = timeZoneDate.day;
  const month = timeZoneDate.month;
  const year = timeZoneDate.year;
  return { day, month, year };
}

// FIX ME
export function getDateStringForZone(
  value,
  timeZone: string,
  dateFormatOptions
) {
  const options = {
    timeZone,
    ...dateFormatOptions,
  };

  console.log(
    value.toISOString(),
    timeZone,
    new Date(value).toLocaleDateString(undefined, options)
  );
  return new Date(value).toLocaleDateString(undefined, options);
}
