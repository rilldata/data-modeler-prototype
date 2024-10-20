<script lang="ts">
  import { line, curveLinear, area } from "d3-shape";
  import type { ScaleTime, ScaleLinear } from "d3-scale";
  import { MainLineColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import {
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import type { V1TimeSeriesValue } from "@rilldata/web-common/runtime-client";

  export let data: V1TimeSeriesValue[];
  export let xKey: string;
  export let yKey: string;
  export let xScale: ScaleTime<number, number>;
  export let yScale: ScaleLinear<number, number>;
  export let color = MainLineColor;
  export let strokeWidth = 1;
  export let fill: boolean | undefined;

  $: lineFunction = line<V1TimeSeriesValue>()
    .defined(
      (d) => d.records?.[yKey] !== null && d.records?.[yKey] !== undefined,
    )
    .x((data) => {
      return xScale(new Date(data[xKey] as string));
    })
    .y((data) => {
      let value = data.records?.[yKey];

      // if (value === null) value = data[i - 1]?.records?.[yKey];

      return yScale(value);
    })
    .curve(curveLinear);

  $: areaFunction = area<V1TimeSeriesValue>()
    .defined(
      (d) => d.records?.[yKey] !== null && d.records?.[yKey] !== undefined,
    )
    .x((data) => xScale(new Date(data[xKey])))
    .y0((data) => yScale(data.records?.[yKey] as number))
    .y1(yScale.range()[0])
    .curve(curveLinear);

  $: areaPath = areaFunction(data);

  $: path = lineFunction(data);
</script>

{#if fill}
  <path d={areaPath} fill="url(#chart-gradient)" class="pointer-events-none" />

  <defs>
    <linearGradient id="chart-gradient" x1="0" x2="0" y1="0" y2="1">
      <stop
        offset="5%"
        stop-color={MainAreaColorGradientDark}
        stop-opacity={0.3}
      />
      <stop
        offset="95%"
        stop-color={MainAreaColorGradientLight}
        stop-opacity={0.3}
      />
    </linearGradient>
  </defs>
{/if}

<path
  d={path}
  fill="none"
  stroke={color}
  stroke-width={strokeWidth}
  vector-effect="non-scaling-stroke"
  class="pointer-events-none"
/>
