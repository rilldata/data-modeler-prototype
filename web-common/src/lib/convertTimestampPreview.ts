import {
  addZoneOffset,
  removeLocalTimezoneOffset as remove,
} from "@rilldata/web-common/lib/time/timezone";

export function convertTimestampPreviewFcn(
  ts,
  removeLocalTimezoneOffset = false,
) {
  return removeLocalTimezoneOffset ? remove(new Date(ts)) : new Date(ts);
}

/** used to convert a timestamp preview from the server for a sparkline. */
export function convertTimestampPreview(d, removeLocalTimezoneOffset = false) {
  return d.map((di) => {
    const pi = { ...di };
    pi.ts = convertTimestampPreviewFcn(di.ts, removeLocalTimezoneOffset);
    return pi;
  });
}

/** used to remove local timezone offset and add dashboard selected zone offset */
export function adjustOffsetForZone(
  ts: Date | string | undefined,
  zone: string,
): Date | undefined {
  if (!ts) return undefined;
  return addZoneOffset(remove(new Date(ts)), zone);
}
