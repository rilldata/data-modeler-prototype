<script lang="ts">
  import type { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type { ScaleLinear, ScaleTime } from "d3-scale";
  import Point from "./Point.svelte";
  import type { Interval } from "luxon";

  export let hoveredPrimaryDataPoint: number | null;
  export let hoveredComparisonDataPoint: number | null;
  export let hoveredDimensionDataPoints: number[];
  export let colors: string[];
  export let flipLabel: boolean;
  export let yScale: ScaleLinear<number, number>;
  export let xScale: ScaleTime<number, number>;
  export let hoveredInterval: Interval<true>;
  export let formatterFunction: ReturnType<typeof createMeasureValueFormatter>;

  $: x = xScale(hoveredInterval.start.toJSDate());
</script>

{#each hoveredDimensionDataPoints as point, i (i)}
  <Point
    {flipLabel}
    {x}
    y={yScale(point)}
    color={colors[i]}
    label={formatterFunction(point)}
  />
{:else}
  {#if hoveredPrimaryDataPoint !== null}
    <line
      x1={x}
      x2={x}
      y1={yScale(hoveredPrimaryDataPoint)}
      y2="100%"
      class="stroke-primary-300 z-10 pointer-events-none"
      stroke-width="4"
      stroke-linecap="round"
      vector-effect="non-scaling-stroke"
    />

    <Point
      {flipLabel}
      y={yScale(hoveredPrimaryDataPoint)}
      {x}
      color="black"
      label={formatterFunction(hoveredPrimaryDataPoint)}
    />

    {#if hoveredComparisonDataPoint}
      <Point
        {flipLabel}
        y={yScale(hoveredComparisonDataPoint)}
        {x}
        color="black"
        label={formatterFunction(hoveredComparisonDataPoint)}
      />
    {/if}
  {/if}
{/each}

<style>
  * {
    pointer-events: none;
  }
</style>
