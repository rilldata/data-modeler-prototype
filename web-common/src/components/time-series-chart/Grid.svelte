<script lang="ts">
  import type { ScaleLinear, ScaleTime } from "d3-scale";
  import { DateTime } from "luxon";

  export let xScale: ScaleTime<number, number>;
  export let yScale: ScaleLinear<number, number>;

  export let timeZone: string;

  let showX = true;
  let showY = true;

  $: xTicks = xScale.ticks(3).map((tick) => {
    // return tick;
    return DateTime.fromJSDate(tick).setZone(timeZone, { keepLocalTime: true });
  });
</script>

<g shape-rendering="crispEdges">
  {#if showX}
    {#each xTicks as tick, i (i)}
      <line
        x1={xScale(tick)}
        x2={xScale(tick)}
        y1="0%"
        y2="100%"
        class="stroke-gray-300"
        stroke-width={1}
        stroke-dasharray="1,1"
        vector-effect="non-scaling-stroke"
      />
    {/each}
  {/if}

  {#if showY}
    {#each yScale.ticks(2) as tick, i (i)}
      <line
        y1={yScale(tick)}
        y2={yScale(tick)}
        x1="0%"
        x2="100%"
        class="stroke-gray-300"
        stroke-width={1}
        stroke-dasharray="1,1"
        vector-effect="non-scaling-stroke"
      />
    {/each}
  {/if}
</g>
