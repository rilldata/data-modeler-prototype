import { bisector } from "d3-array";
import type { ScaleLinear, ScaleTime } from "d3-scale";
import { area, curveLinear, curveStep, line } from "d3-shape";
import { timeFormat } from "d3-time-format";
import { getContext } from "svelte";
import { derived, writable } from "svelte/store";
import { contexts } from "./constants";
import { curveStepExtended } from "./marks/curveStepExtended";
import { ScaleType } from "./state";
import type {
  GraphicScale,
  ScaleStore,
  SimpleConfigurationStore,
} from "./state/types";

/**
 * Creates a string to be fed into the d attribute of a path,
 * producing a single path definition for one circle.
 * These completed, segmented arcs will not overlap in a way where
 * we can overplot if part of the same path.
 */
export function circlePath(cx: number, cy: number, r: number): string {
  return `
    M ${cx - r}, ${cy}
      a ${r},${r} 0 1,0 ${r * 2},0
      a ${r},${r} 0 1,0 ${-r * 2},0
    `;
}

const curves = {
  curveLinear,
  curveStep,
  curveStepExtended,
};

export function pathDoesNotDropToZero(yAccessor: string) {
  return (d, i: number, arr) => {
    return (
      !isNaN(d[yAccessor]) &&
      d[yAccessor] !== undefined &&
      // remove all zeroes where the previous or next value is also zero.
      // these do not add to our understanding.
      (!(i !== 0 && d[yAccessor] === 0 && arr[i - 1][yAccessor] === 0) ||
        !(
          i !== arr.length - 1 &&
          d[yAccessor] === 0 &&
          arr[i + 1][yAccessor] === 0
        ))
    );
  };
}

export interface PlotConfig {
  top: number;
  bottom: number;
  left: number;
  right: number;
  buffer: number;
  width: number;
  height: number;
  devicePixelRatio: number;
  plotTop: number;
  plotBottom: number;
  plotLeft: number;
  plotRight: number;
  fontSize: number;
  textGap: number;
  id: string;
}

interface LineGeneratorArguments {
  xAccessor: string;
  xScale:
    | ScaleLinear<number, number>
    | ScaleTime<Date, number>
    | ((d) => number);
  yScale:
    | ScaleLinear<number, number>
    | ScaleTime<Date, number>
    | ((d) => number);
  curve?: string;
  pathDefined?: (
    datum: object,
    i?: number,
    arr?: ArrayLike<unknown>,
  ) => boolean;
}

/**
 * A convenience function to generate a nice SVG path for a time series.
 * FIXME: rename to timeSeriesLineFactory.
 * FIXME: once we've gotten the data generics in place and threaded into components, let's make sure to type this.
 */
export function lineFactory(args: LineGeneratorArguments) {
  return (yAccessor: string) =>
    line()
      .x((d) => args.xScale(d[args.xAccessor]))
      .y((d) => args.yScale(d[yAccessor]))
      .curve(args.curve ? curves[args.curve] : curveLinear)
      .defined(args.pathDefined || pathDoesNotDropToZero(yAccessor));
}

/**
 * A convenience function to generate a nice SVG area path for a time series.
 * FIXME: rename to timeSeriesAreaFactory.
 * FIXME: once we've gotten the data generics in place and threaded into components, let's make sure to type this.
 */
export function areaFactory(args: LineGeneratorArguments) {
  return (yAccessor: string) =>
    area()
      .x((d) => args.xScale(d[args.xAccessor]))
      .y0(args.yScale(0))
      .y1((d) => args.yScale(d[yAccessor]))
      .curve(args.curve ? curves[args.curve] : curveLinear)
      .defined(args.pathDefined || pathDoesNotDropToZero(yAccessor));
}

/**
 * Return a list of ticks to be represented on the
 * axis or grid depending on axis-side, it's length and
 * the data type of field
 */
export function getTicks(
  xOrY: string,
  scale: GraphicScale,
  axisLength: number,
  scaleType: ScaleType,
) {
  const isDate = scaleType === ScaleType.DATE;
  const tickCount = Math.trunc(axisLength / (xOrY === "x" ? 100 : 50));

  let ticks = scale.ticks(tickCount);

  // Prevent overlapping ticks on X axis
  if (xOrY === "x" && axisLength / ticks.length < 60) {
    ticks = scale.ticks(tickCount - 1);
  }

  if (ticks.length <= 1) {
    if (isDate) ticks = scale.domain();
    else ticks = scale.nice().domain();
  }

  return ticks;
}

export function barplotPolyline(
  data,
  xLow,
  xHigh,
  yAccessor,
  X,
  Y,
  separator = 1,
  closeBottom = false,
  inflator = 1,
): string {
  if (!data?.length) return "";
  const path = data.reduce((pointsPathString, datum, i) => {
    const low = datum[xLow];
    const high = datum[xHigh];
    const count = datum[yAccessor];

    const x = X(low) + separator;

    const width = Math.max(0.5, X(high) - X(low) - separator * 2);
    const y = Y(0) * (1 - inflator) + Y(count) * inflator;

    const computedHeight = Math.min(
      Y(0),
      Y(0) * inflator - Y(count) * inflator,
    );
    const height = separator > 0 ? computedHeight : 0;

    // do not add zero values here
    if (count === 0) {
      return pointsPathString;
    }

    let p1 = "";

    const nextPointIsZero = i < data.length - 1 && data[i + 1][yAccessor] === 0;

    const lastPointWasZero = i > 0 && data[i - 1][yAccessor] === 0;

    if (separator === 0 && lastPointWasZero) {
      // we will need to start this thing at 0?
      p1 = `M${x},${y + computedHeight}`;
    } else if (separator > 0 || i === 0) {
      // standard case.
      p1 = `${i !== 0 ? "M" : ""}${x},${y + height}`;
    }

    const p2 = `${x},${y}`;
    const p3 = `${x + width},${y}`;

    const p4 =
      separator > 0 || nextPointIsZero
        ? `${x + width},${y + (separator > 0 ? height : computedHeight)}`
        : "";
    const closedBottom = closeBottom ? `${x},${y + height}` : "";

    return pointsPathString + `${p1} ${p2} ${p3} ${p4} ${closedBottom} `;
  }, " ");
  const lastElement = data.findLast((d) => d[yAccessor]);
  if (!lastElement) return "";
  return (
    `M${X(data[0][xLow]) + separator},${Y(0)} ` +
    path +
    ` ${X(lastElement[xHigh]) - separator},${Y(0)} `
  );
}

/** utilizes the provided scales to calculate the line thinness in a way
 * that enables higher-density "overplotted lines".
 */

export function createAdaptiveLineThicknessStore(yAccessor) {
  let data;

  // get xScale, yScale, and config from contexts
  const xScale = getContext<ScaleStore>(contexts.scale("x"));
  const yScale = getContext<ScaleStore>(contexts.scale("y"));
  const config = getContext<SimpleConfigurationStore>(contexts.config);

  // capture data state.
  const dataStore = writable(data);

  const store = derived(
    [xScale, yScale, config, dataStore],
    ([$xScale, $yScale, $config, $data]) => {
      if (!$data) {
        return 1;
      }
      const totalTravelDistance = $data
        .filter((di) => di[yAccessor] !== null)
        .map((di, i) => {
          if (i === $data.length - 1) {
            return 0;
          }
          const max = Math.max(
            $yScale($data[i + 1][yAccessor]),
            $yScale($data[i][yAccessor]),
          );
          const min = Math.min(
            $yScale($data[i + 1][yAccessor]),
            $yScale($data[i][yAccessor]),
          );
          if (isNaN(min) || isNaN(max)) return 1 / $data.length;
          return Math.abs(max - min);
        })
        .reduce((acc, v) => acc + v, 0);

      const yIshDistanceTravelled =
        2 /
        (totalTravelDistance /
          (($xScale.range()[1] - $xScale.range()[0]) *
            ($config.devicePixelRatio || 3)));

      const xIshDistanceTravellled =
        (($xScale.range()[1] - $xScale.range()[0]) *
          ($config.devicePixelRatio || 3) *
          0.7) /
        $data.length /
        1.5;

      const value = Math.min(
        1,
        /** to determine the stroke width of the path, let's look at
         * the bigger of two values:
         * 1. the "y-ish" distance travelled
         * the inverse of "total travel distance", which is the Y
         * gap size b/t successive points divided by the zoom window size;
         * 2. time series length / available X pixels
         * the time series divided by the total number of pixels in the existing
         * zoom window.
         *
         * These heuristics could be refined, but this seems to provide a reasonable approximation for
         * the stroke width. (1) excels when lots of successive points are close together in the Y direction,
         * whereas (2) excels` when a line is very, very noisy (and thus the X direction is the main constraint).
         */
        Math.max(
          // the y-ish distance travelled
          yIshDistanceTravelled,
          // the time series length / available X pixels
          xIshDistanceTravellled,
        ),
      );

      return value;
    },
  );

  return {
    subscribe: store.subscribe,
    /** trigger an update when the data changes */
    setData(d) {
      dataStore.set(d);
    },
  };
}

// This is function equivalent of WithBisector
export function bisectData<T>(
  value: Date,
  direction: "left" | "right" | "center",
  accessor: keyof T,
  data: ArrayLike<T>,
): { position: number; entry: T } {
  const bisect = bisector<T, unknown>((d) => d[accessor])[direction];
  const position = bisect(data, value);

  return {
    position,
    entry: data[position],
  };
}

/** For a scale domain returns a formatter for axis label and super label */
export function createTimeFormat(
  scaleDomain: [Date, Date],
  numberOfValues: number,
): [(d: Date) => string, ((d: Date) => string) | undefined] {
  const diff =
    Math.abs(scaleDomain[1]?.getTime() - scaleDomain[0]?.getTime()) / 1000;
  if (!diff) return [timeFormat("%d %b"), timeFormat("%Y")];
  const gap = diff / (numberOfValues - 1); // time gap between two consecutive values

  // If the gap is less than a second, format in milliseconds
  if (gap < 1) {
    return [timeFormat("%M:%S.%L"), timeFormat("%H %d %b %Y")];
  }
  // If the gap is less than a minute, format in seconds
  else if (gap < 60) {
    return [timeFormat("%M:%S"), timeFormat("%H %d %b %Y")];
  }
  // If the gap is less than 24 hours, format in hours and minutes
  else if (gap < 60 * 60 * 24) {
    return [timeFormat("%H:%M"), timeFormat("%d %b %Y")];
  }
  // If the gap is less than 30 days, format in days
  else if (gap < 60 * 60 * 24 * 30) {
    return [timeFormat("%b %d"), timeFormat("%Y")];
  }
  // If the gap is less than a year, format in months
  else if (gap < 60 * 60 * 24 * 365) {
    return [timeFormat("%b"), timeFormat("%Y")];
  }
  // Else format in years
  else {
    return [timeFormat("%Y"), undefined];
  }
}
