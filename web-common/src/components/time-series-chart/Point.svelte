<script lang="ts">
  export let showLabel = true;
  export let label: string | null | undefined;
  export let y: number;
  export let x: number;
  export let flipLabel = false;
  export let color: string;
</script>

{#if showLabel && label}
  <text
    class="name-text"
    x={x + (flipLabel ? -10 : 10)}
    {y}
    font-size="12px"
    text-anchor={flipLabel ? "end" : "start"}
    dominant-baseline="middle"
    stroke="white"
  >
    {label}
  </text>
{/if}

<path
  d="M {x} {y} l 0.0001 0"
  stroke-linecap="round"
  stroke-width="8.5"
  stroke="white"
  vector-effect="non-scaling-stroke"
/>

<path
  d="M {x} {y} l 0.0001 0"
  stroke-linecap="round"
  stroke-width="6"
  stroke={color}
  vector-effect="non-scaling-stroke"
/>

<style>
  * {
    pointer-events: none;
  }
  text {
    font-size: 11px;
    paint-order: stroke;
    stroke: #ffffff;
    stroke-width: 3px;
    stroke-linecap: butt;
    stroke-linejoin: miter;
    font-weight: 500;

    /* Make all characters and numbers of equal width for easy scanibility */
    font-feature-settings:
      "case" 0,
      "cpsp" 0,
      "dlig" 0,
      "frac" 0,
      "dnom" 0,
      "numr" 0,
      "salt" 0,
      "subs" 0,
      "sups" 0,
      "tnum",
      "zero" 0,
      "ss01",
      "ss02" 0,
      "ss03" 0,
      "ss04" 0,
      "cv01" 0,
      "cv02" 0,
      "cv03" 0,
      "cv04" 0,
      "cv05" 0,
      "cv06" 0,
      "cv07" 0,
      "cv08" 0,
      "cv09" 0,
      "cv10" 0,
      "cv11" 0,
      "calt",
      "ccmp",
      "kern";
  }
</style>
