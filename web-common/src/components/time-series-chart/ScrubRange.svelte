<script lang="ts">
  import {
    ScrubArea0Color,
    ScrubArea1Color,
    ScrubArea2Color,
    ScrubBoxColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";

  export let xStart: number;
  export let xEnd: number;
  export let scrubbing: boolean;
  export let onGrabHandle: (anchor: "start" | "end") => void;
  export let handleScrubGrab;

  let grabbing = false;
</script>

<line
  role="presentation"
  x1={xStart}
  x2={xStart}
  y1="0%"
  y2="100%"
  stroke={ScrubBoxColor}
  stroke-width="1"
  class="cursor-ew-resize"
  vector-effect="non-scaling-stroke"
  on:mousedown|stopPropagation={() => {
    onGrabHandle("start");
  }}
/>
<line
  role="presentation"
  x1={xEnd}
  x2={xEnd}
  y1="0%"
  y2="100%"
  stroke={ScrubBoxColor}
  class="cursor-ew-resize"
  stroke-width="1"
  vector-effect="non-scaling-stroke"
  on:mousedown|stopPropagation={() => {
    onGrabHandle("end");
  }}
/>

{#if xStart !== null && xEnd !== null && xStart !== xEnd}
  <g style:mix-blend-mode="hue">
    <rect
      x={0}
      y={0}
      width="100%"
      height="100%"
      fill="white"
      class="pointer-events-none"
    />
    <rect
      class="pointer-events-none"
      role="presentation"
      on:mousedown|stopPropagation={handleScrubGrab}
      x={xStart}
      width={xEnd - xStart}
      y={0}
      height="100%"
      fill="url(#scrubbing-gradient)"
      opacity={scrubbing ? 0.4 : 0.2}
    />
  </g>
{/if}

<defs>
  <linearGradient gradientUnits="userSpaceOnUse" id="scrubbing-gradient">
    <stop stop-color={ScrubArea0Color} />
    <stop offset="0.36" stop-color={ScrubArea1Color} />
    <stop offset="1" stop-color={ScrubArea2Color} />
  </linearGradient>
</defs>

{#if grabbing}
  <filter id="shadow">
    <feDropShadow
      dx="0"
      dy="0"
      stdDeviation="3"
      flood-color="rgba(0, 0, 0, 1)"
    />
  </filter>
{/if}
