<script lang="ts" context="module">
  import { derived, writable } from "svelte/store";

  export const hoverIndexStore = (() => {
    const { subscribe, set } = writable<number | null>(null);
    return {
      subscribe,
      set,
      reset: () => set(null),
    };
  })();

  export const scrubRangeStore = (() => {
    const startStore = writable<Date | null>(null);
    const endStore = writable<Date | null>(null);

    return {
      validRange: derived([startStore, endStore], ([start, end]) => {
        if (start === null || end === null) return null;
        return {
          start,
          end,
        };
      }),
      reset() {
        startStore.set(null);
        endStore.set(null);
      },

      set: {
        start(date: Date) {
          startStore.set(date);
        },
        end(date: Date) {
          endStore.set(date);
        },
      },
      start: startStore,
      end: endStore,
    };
  })();
</script>

<script lang="ts">
  import { min, max, extent } from "d3-array";
  import Line from "./Line.svelte";
  import { scaleLinear, scaleTime } from "d3-scale";
  import {
    LineMutedColor,
    MainLineColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import Grid from "./Grid.svelte";
  import {
    V1TimeGrain,
    type V1TimeSeriesValue,
  } from "@rilldata/web-common/runtime-client";
  import HoverLine from "./HoverLine.svelte";
  import { type Interval } from "luxon";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import ScrubRange from "./ScrubRange.svelte";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";

  export let primaryData: V1TimeSeriesValue[];
  export let comparisonData: V1TimeSeriesValue[] | null = [];
  export let dimensionData: { data: V1TimeSeriesValue[]; color: string }[] = [];
  export let comparisonInterval: Interval<boolean> | undefined;
  export let timeGrain: V1TimeGrain;
  export let xAccessor = "ts";
  export let yAccessor: string;
  export let showGrid = true;
  export let showComparison = false;
  export let showAxis = true;
  export let showBorder = true;
  export let points: Interval[];
  export let interval: Interval<true>;
  export let selectedTimeZone: string;
  export let formatterFunction: ReturnType<typeof createMeasureValueFormatter>;

  let initialIndex = -1;
  let element: HTMLDivElement;
  let scrubbing: "start" | "end" | null = null;
  let yScale = scaleLinear();
  let contentRect = new DOMRect(0, 0, 0, 0);

  $: xExtents = [interval.start.toJSDate(), interval.end.toJSDate()];

  $: xScale = scaleTime().domain(xExtents).range([0, width]);

  $: comparisonScale =
    comparisonInterval?.start && comparisonInterval?.end
      ? scaleTime()
          .domain([
            comparisonInterval.start?.toJSDate(),
            comparisonInterval.end?.toJSDate(),
          ])
          .range([0, width])
      : scaleTime();

  $: primaryExtents = extent(
    primaryData,
    (datum) => datum?.records?.[yAccessor] as number,
  );

  $: comparisonExtents = extent(
    comparisonData ?? [],
    (datum) => datum?.records?.[yAccessor] as number | null,
  );

  $: dimensionExtents = [
    min(dimensionData, (line) =>
      min(line.data, (d) => d?.records?.[yAccessor] as number | null),
    ),

    max(dimensionData, (line) =>
      max(line.data, (d) => d?.records?.[yAccessor] as number | null),
    ),
  ];

  $: mins = [
    primaryExtents[0],
    comparisonExtents[0],
    dimensionExtents[0],
  ].filter(isNumber);
  $: maxes = [
    primaryExtents[1],
    comparisonExtents[1],
    dimensionExtents[1],
  ].filter(isNumber);

  function isNumber(value: unknown): value is number {
    return value !== undefined && value !== null;
  }

  $: yExtents = [Math.min(0, min(mins) ?? 0), max(maxes) ?? 0];

  $: console.log(primaryExtents, comparisonExtents, dimensionExtents);

  $: yScale = yScale.domain(yExtents).range([height, 0 + height * 0.2]);

  function handleMouseUp() {
    scrubbing = null;

    if (!$validRange) scrubRangeStore.reset();

    // scrubRange?.end = mouseOverDate;
    // window.removeEventListener("mouseup", handleMouseUp);
  }

  const { validRange } = scrubRangeStore;

  $: yTicks = yScale.ticks(2);

  $: ({ height, width } = contentRect);

  $: hoverIndex = $hoverIndexStore;

  $: hoveredPrimaryDataPoint =
    hoverIndex !== null
      ? (primaryData[hoverIndex]?.records?.[yAccessor] as number)
      : null;
  $: hoveredComparisonDataPoint =
    hoverIndex !== null
      ? (comparisonData?.[hoverIndex]?.records?.[yAccessor] as number)
      : null;

  $: hoveredDimensionDataPoints =
    hoverIndex !== null
      ? dimensionData
          .map((line) => {
            if (hoverIndex === null) {
              return undefined;
            }
            return line.data[hoverIndex]?.records?.[yAccessor];
          })
          .filter((d) => d !== undefined && d !== null)
      : [];

  $: hoveredInterval = hoverIndex !== null ? points[hoverIndex] : null;

  $: grainWidth = width / points.length;
</script>

<div class="wrapper" class:cursor-ew-resize={Boolean(scrubbing)}>
  <div class="flex flex-col size-full" bind:this={element} bind:contentRect>
    {#if hoveredInterval}
      <span class="absolute">
        {hoveredInterval?.start?.toLocaleString(
          TIME_GRAIN[timeGrain].formatDate,
        )}
      </span>
    {/if}
    <!-- {#if showAxis}
      <div>
        <svg class="w-full h-8 overflow-visible">
          {#each xScale.ticks(3) as tick, i (i)}
            <text
              x="{((tick.getTime() / DIVISOR - numberXExtents[0]) / dateWidth) *
                100}%"
              y="50%"
              text-anchor="middle"
              dominant-baseline="middle"
              font-size="0.65rem"
              fill="#6B7280"
            >
              {DateTime.fromJSDate(tick).toLocaleString({
                // month: "short",
                // day: "numeric",
                hour: "numeric",
              })}
            </text>
          {/each}
        </svg>
      </div>
    {/if} -->
    <svg
      role="presentation"
      class:border-b={showBorder}
      class="border-1 border-gray-300 cursor-pointer"
      preserveAspectRatio="xMinYMin meet"
      viewBox="{0} {0} {width} {height}"
      on:mouseup={handleMouseUp}
      on:mouseleave={() => {
        hoverIndexStore.reset();
        // grabbing = false;
      }}
    >
      {#if showGrid}
        <Grid {xScale} {yScale} timeZone={selectedTimeZone} />
      {/if}

      {#each points as point, i (i)}
        {@const date = point?.start?.toJSDate()}

        {#if date}
          <rect
            x={xScale(date)}
            y={0}
            width={grainWidth}
            height="100%"
            role="presentation"
            class:opacity-80={hoverIndex === i}
            class="opacity-0 fill-primary-50"
            on:mousedown|preventDefault={() => {
              if ($validRange) scrubRangeStore.reset();
              scrubbing = "end";
              initialIndex = i;

              scrubRangeStore.set.start(date);

              window.addEventListener("mouseup", handleMouseUp);
            }}
            on:mouseenter={() => {
              console.log(date);
              hoverIndexStore.set(i);

              if (scrubbing) {
                if (i > initialIndex) {
                  scrubbing = "end";
                } else if (i < initialIndex) {
                  scrubbing = "start";
                }

                scrubRangeStore.set[scrubbing](date);
              }
            }}
          />
        {/if}
      {/each}

      {#each dimensionData as { data, color }, i (i)}
        <Line
          {data}
          {xScale}
          {color}
          {yScale}
          xKey={xAccessor}
          yKey={yAccessor}
          fill={false}
          strokeWidth={1.5}
        />
      {:else}
        <Line
          data={primaryData}
          {xScale}
          color={MainLineColor}
          {yScale}
          xKey={xAccessor}
          yKey={yAccessor}
          fill={true}
          strokeWidth={1}
        />

        {#if showComparison && comparisonData?.length}
          <Line
            data={comparisonData}
            xScale={comparisonScale}
            color={LineMutedColor}
            {yScale}
            xKey={xAccessor}
            yKey={yAccessor}
            fill={false}
            strokeWidth={1}
          />
        {/if}
      {/each}

      {#if $validRange}
        {@const { start, end } = $validRange}
        <ScrubRange
          xStart={xScale(start)}
          xEnd={xScale(end)}
          scrubbing={!!scrubbing}
          handleScrubGrab={() => {}}
          onGrabHandle={() => {}}
        />
      {/if}

      {#if hoveredInterval && hoveredInterval.isValid}
        <HoverLine
          flipLabel={false}
          {hoveredInterval}
          {hoveredPrimaryDataPoint}
          {hoveredComparisonDataPoint}
          {hoveredDimensionDataPoints}
          {xScale}
          {yScale}
          {formatterFunction}
          colors={dimensionData.map((line) => line.color)}
        />
      {/if}
    </svg>
  </div>

  {#if showAxis}
    <svg class="w-8 h-full overflow-visible">
      {#each yTicks as tick, i (i)}
        <text
          x="0"
          y={yScale(tick)}
          font-size="0.65rem"
          dominant-baseline="middle"
          fill="#6B7280"
        >
          {formatterFunction(tick)}
        </text>
      {/each}
    </svg>
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply size-full relative flex-1 flex gap-2;
  }
</style>
