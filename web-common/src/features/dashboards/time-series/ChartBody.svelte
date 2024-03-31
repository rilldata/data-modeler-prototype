<script lang="ts">
  import {
    ChunkedLine,
    ClippedChunkedLine,
  } from "@rilldata/web-common/components/data-graphic/marks";
  import {
    AreaMutedColorGradientDark,
    AreaMutedColorGradientLight,
    LineMutedColor,
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
    MainLineColor,
    TimeComparisonLineColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
  import { previousValueStore } from "@rilldata/web-common/lib/store-utils";
  import { writable } from "svelte/store";

  export let xMin: Date | undefined = undefined;
  export let xMax: Date | undefined = undefined;
  export let yExtentMax: number | undefined = undefined;
  export let showComparison: boolean;
  export let dimensionValue: string | undefined | null;
  export let isHovering: boolean;
  export let data;
  export let dimensionData: DimensionDataItem[] = [];
  export let xAccessor: string;
  export let yAccessor: string;
  export let scrubStart;
  export let scrubEnd;
  export let colors: string[];

  $: hasSubrangeSelected = Boolean(scrubStart && scrubEnd);

  $: mainLineColor = hasSubrangeSelected ? LineMutedColor : MainLineColor;

  const focusedAreaGradient: [string, string] = [
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
  ];

  $: areaGradientColors = (
    hasSubrangeSelected
      ? [AreaMutedColorGradientDark, AreaMutedColorGradientLight]
      : focusedAreaGradient
  ) as [string, string];

  $: isDimValueHiglighted =
    dimensionValue !== undefined &&
    dimensionData.map((d) => d.value).includes(dimensionValue);

  // we delay the tween if previousYMax < yMax
  let yMaxStore = writable(yExtentMax);
  let previousYMax = previousValueStore(yMaxStore);

  $: if (typeof yExtentMax === "number") yMaxStore.set(yExtentMax);
  const timeRangeKey = writable(`${xMin}-${xMax}`);

  const previousTimeRangeKey = previousValueStore(timeRangeKey);

  // FIXME: move this function to utils.ts
  /** reset the keys to trigger animations on time range changes */
  let syncTimeRangeKey;
  $: {
    timeRangeKey.set(`${xMin}-${xMax}`);
    if ($previousTimeRangeKey !== $timeRangeKey) {
      if (syncTimeRangeKey) clearTimeout(syncTimeRangeKey);
      syncTimeRangeKey = setTimeout(() => {
        previousTimeRangeKey.set($timeRangeKey);
      }, 400);
    }
  }

  $: delay =
    $previousTimeRangeKey === $timeRangeKey &&
    yExtentMax &&
    $previousYMax < yExtentMax
      ? 100
      : 0;
</script>

<!-- key on the time range itself to prevent weird tweening animations.
    We'll need to migrate this to a more robust solution once we've figured out
    the right way to "tile" together a time series with multiple pages of data.
    -->
{#key $timeRangeKey}
  {#if dimensionData?.length}
    {#each dimensionData as d, i}
      {@const isHighlighted = d?.value === dimensionValue}

      <g
        class="transition-opacity"
        class:opacity-20={isDimValueHiglighted && !isHighlighted}
      >
        <ChunkedLine
          isComparingDimension
          lineColor="stroke-{colors[i]}"
          delay={$timeRangeKey !== $previousTimeRangeKey ? 0 : delay}
          duration={hasSubrangeSelected ||
          $timeRangeKey !== $previousTimeRangeKey
            ? 0
            : 200}
          lineClasses="stroke-{colors[i]}"
          data={d?.data || []}
          {xAccessor}
          {yAccessor}
        />
      </g>
    {/each}
  {:else}
    {#if showComparison}
      <g
        class="transition-opacity"
        class:opacity-80={isHovering}
        class:opacity-40={!isHovering}
      >
        <ChunkedLine
          lineColor={TimeComparisonLineColor}
          delay={$timeRangeKey !== $previousTimeRangeKey ? 0 : delay}
          duration={hasSubrangeSelected ||
          $timeRangeKey !== $previousTimeRangeKey
            ? 0
            : 200}
          {data}
          {xAccessor}
          yAccessor="comparison.{yAccessor}"
        />
      </g>
    {/if}
    <ChunkedLine
      lineColor={mainLineColor}
      {areaGradientColors}
      delay={$timeRangeKey !== $previousTimeRangeKey ? 0 : delay}
      duration={hasSubrangeSelected || $timeRangeKey !== $previousTimeRangeKey
        ? 0
        : 200}
      {data}
      {xAccessor}
      {yAccessor}
    />
    {#if hasSubrangeSelected}
      <ClippedChunkedLine
        start={Math.min(scrubStart, scrubEnd)}
        end={Math.max(scrubStart, scrubEnd)}
        lineColor={MainLineColor}
        areaGradientColors={focusedAreaGradient}
        delay={0}
        duration={0}
        {data}
        {xAccessor}
        {yAccessor}
      />
    {/if}
  {/if}
{/key}
