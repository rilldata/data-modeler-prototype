import { getMinGrain } from "@rilldata/web-common/lib/time/grains";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

export enum RillTimeType {
  Unknown = "Unknown",
  Latest = "Latest",
  PreviousPeriod = "Previous period",
  PeriodToDate = "Period To Date",
}

export class RillTime {
  public timeRange: string;
  public readonly isComplete: boolean;
  public readonly end: RillTimeAnchor;
  public readonly type: RillTimeType;

  public constructor(
    public readonly start: RillTimeAnchor,
    end: RillTimeAnchor,
    public readonly timeRangeGrain: RillTimeRangeGrain | undefined,
    public modifier: RillTimeRangeModifier | undefined,
  ) {
    this.type = start.getType();

    this.end = end;
    this.isComplete =
      this.end &&
      (this.end.type === RillTimeAnchorType.Relative ||
        this.end.truncate !== undefined);
  }

  public getLabel() {
    if (this.type === RillTimeType.Unknown || !!this.modifier) {
      return this.timeRange;
    }

    const hasNonStandardStart =
      this.start.type === RillTimeAnchorType.Custom || !!this.start.offset;
    const hasNonStandardEnd =
      this.end &&
      ((this.end.type === RillTimeAnchorType.Relative &&
        this.end.grain &&
        this.end.grain.count !== 0) ||
        this.end.type === RillTimeAnchorType.Custom ||
        !!this.end.offset);
    if (hasNonStandardStart || hasNonStandardEnd) {
      return this.timeRange;
    }

    const start = capitalizeFirstChar(this.start.getLabel());
    if (this.isComplete) return start;
    return `${start}, incomplete`;
  }

  public getRangeGrain() {
    return getMinGrain(
      this.start.getRangeGrain(),
      this.end?.getRangeGrain() ?? V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
    );
  }

  public getBucketGrain() {
    if (!this.timeRangeGrain) {
      return V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
    }
    return (
      ToAPIGrain[this.timeRangeGrain.grain] ??
      V1TimeGrain.TIME_GRAIN_UNSPECIFIED
    );
  }

  public addTimezone(timezone: string) {
    this.modifier ??= <RillTimeRangeModifier>{};
    this.modifier.timeZone = timezone;
  }

  public toString() {
    let rillTime = this.start.toString();
    if (this.end) {
      rillTime += "," + this.end.toString();
    }

    if (this.timeRangeGrain) {
      rillTime += ":" + rangeGrainToString(this.timeRangeGrain);
    }

    if (this.modifier) {
      const modifierPart = rangeModifierToString(this.modifier);
      if (modifierPart) rillTime += "@" + modifierPart;
    }

    return rillTime;
  }
}

export enum RillTimeAnchorType {
  Now = "Now",
  Earliest = "Earliest",
  Latest = "Latest",
  Relative = "Relative",
  Custom = "Custom",
}

const GrainToUnit = {
  s: "second",
  S: "second",
  m: "minute",
  h: "hour",
  H: "hour",
  d: "day",
  D: "day",
  w: "week",
  W: "week",
  M: "month",
  q: "Quarter",
  Q: "Quarter",
  y: "year",
  Y: "year",
};
const ToAPIGrain: Record<string, V1TimeGrain> = {
  s: V1TimeGrain.TIME_GRAIN_SECOND,
  S: V1TimeGrain.TIME_GRAIN_SECOND,
  m: V1TimeGrain.TIME_GRAIN_MINUTE,
  h: V1TimeGrain.TIME_GRAIN_HOUR,
  H: V1TimeGrain.TIME_GRAIN_HOUR,
  d: V1TimeGrain.TIME_GRAIN_DAY,
  D: V1TimeGrain.TIME_GRAIN_DAY,
  w: V1TimeGrain.TIME_GRAIN_WEEK,
  W: V1TimeGrain.TIME_GRAIN_WEEK,
  M: V1TimeGrain.TIME_GRAIN_MONTH,
  q: V1TimeGrain.TIME_GRAIN_QUARTER,
  Q: V1TimeGrain.TIME_GRAIN_QUARTER,
  y: V1TimeGrain.TIME_GRAIN_YEAR,
  Y: V1TimeGrain.TIME_GRAIN_YEAR,
};
export const InvalidTime = "Invalid";
export class RillTimeAnchor {
  public truncate: RillTimeGrain | undefined = undefined;
  public absolute: string | undefined = undefined;
  public grain: RillTimeGrain | undefined = undefined;
  public offset: RillTimeAnchor | undefined = undefined;

  public constructor(public readonly type: RillTimeAnchorType) {}

  public static now() {
    return new RillTimeAnchor(RillTimeAnchorType.Now);
  }
  public static earliest() {
    return new RillTimeAnchor(RillTimeAnchorType.Earliest);
  }
  public static latest() {
    return new RillTimeAnchor(RillTimeAnchorType.Latest);
  }
  public static relative(grain: RillTimeGrain) {
    return new RillTimeAnchor(RillTimeAnchorType.Relative).withGrain(grain);
  }
  public static absolute(time: string) {
    return new RillTimeAnchor(RillTimeAnchorType.Custom).withAbsolute(time);
  }

  public withGrain(grain: RillTimeGrain) {
    this.grain = grain;
    return this;
  }

  public withOffset(anchor: RillTimeAnchor) {
    this.offset = anchor;
    return this;
  }

  public withAbsolute(time: string) {
    this.absolute = time;
    return this;
  }

  public withTruncate(truncate: RillTimeGrain) {
    this.truncate = truncate;
    return this;
  }

  public getLabel() {
    const grain = this.grain ?? this.truncate;
    if (!grain) {
      return RillTimeAnchorType.Earliest.toString();
    }

    const unit = GrainToUnit[grain.grain];
    if (!unit) return InvalidTime;

    if (grain.count === 0) {
      if (unit === "day") return "today";
      return `${unit} to date`;
    }

    if (grain.count > 0) return InvalidTime;

    if (grain.count === -1) {
      return `previous ${unit}`;
    }
    return `last ${-grain.count} ${unit}s`;
  }

  public getType() {
    const grain = this.grain ?? this.truncate;
    if (!grain || grain.count > 0) {
      return RillTimeType.Unknown;
    }

    if (grain.count === 0) {
      return RillTimeType.PeriodToDate;
    }
    if (grain.count === -1) {
      return RillTimeType.PreviousPeriod;
    }
    return RillTimeType.Latest;
  }

  public getRangeGrain() {
    return getMinGrain(
      this.grain?.grain
        ? ToAPIGrain[this.grain.grain]
        : V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      this.truncate?.grain
        ? ToAPIGrain[this.truncate.grain]
        : V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      this.offset?.grain?.grain
        ? ToAPIGrain[this.offset.grain.grain]
        : V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
    );
  }

  public toString() {
    let anchor = "";
    switch (this.type) {
      case RillTimeAnchorType.Now:
        anchor = "now";
        break;
      case RillTimeAnchorType.Earliest:
        anchor = "earliest";
        break;
      case RillTimeAnchorType.Latest:
        anchor = "latest";
        break;
      case RillTimeAnchorType.Relative:
        anchor = this.grain ? grainToString(this.grain, true) : "";
        break;
      case RillTimeAnchorType.Custom:
        anchor = this.absolute ?? "";
        break;
    }

    if (this.truncate) {
      anchor += "/" + grainToString(this.truncate, false);
    }

    if (this.offset) {
      anchor += this.offset.toString();
    }

    return anchor;
  }
}

export type RillTimeGrain = {
  grain: string;
  count: number;
};

export function newSignedRillTimeGrain(
  grain: string,
  sign: "+" | "-",
  count: number,
) {
  if (sign === "-") {
    count = -count;
  }
  return <RillTimeGrain>{
    grain,
    count,
  };
}

function grainToString(grain: RillTimeGrain, includeZero: boolean) {
  let grainPart = grain.grain;
  if (grainPart !== "m" && grainPart !== "M") {
    grainPart = grainPart.toLowerCase();
  }
  const countSign = grain.count > 0 ? "+" : "";
  const countPart =
    includeZero || !!grain.count ? `${countSign}${grain.count}` : "";
  return `${countPart}${grainPart}`;
}

export type RillTimeRangeGrain = {
  grain: string;
  isComplete: boolean;
};

function rangeGrainToString(grain: RillTimeRangeGrain) {
  let grainPart = grain.grain;
  if (grainPart !== "m" && grainPart !== "M") {
    grainPart = grainPart.toLowerCase();
  }
  if (!grain.isComplete) return grainPart;
  return `|${grainPart}|`;
}

export type RillTimeRangeModifier = {
  timeZone: string | undefined;
  at: RillTimeAnchor | undefined;
};

function rangeModifierToString(grain: RillTimeRangeModifier) {
  let str = "";
  if (grain.at) {
    str += grain.at.toString();
  }

  if (grain.timeZone) {
    str += `${str ? " " : ""}{${grain.timeZone}}`;
  }

  return str;
}

function capitalizeFirstChar(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}
