<script lang="ts">
  import { scaleLinear, scaleTime } from "d3-scale";

  export let xExtents: [Date, Date];
  export let yExtents: [number, number];
  export let xScale = scaleTime();
  export let yScale = scaleLinear();
  export let divisor: number;

  $: yScale = scaleLinear().domain(yExtents).range(yExtents);

  $: xScale = scaleTime()
    .domain(xExtents)
    .range([xExtents[0].getTime() / divisor, xExtents[1].getTime() / divisor]);
</script>

<slot {xScale} {yScale} />
