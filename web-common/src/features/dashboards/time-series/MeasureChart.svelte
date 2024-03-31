<script lang="ts">
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
  import Body from "@rilldata/web-common/components/data-graphic/elements/Body.svelte";
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import WithBisector from "@rilldata/web-common/components/data-graphic/functional-components/WithBisector.svelte";
  import WithRoundToTimegrain from "@rilldata/web-common/components/data-graphic/functional-components/WithRoundToTimegrain.svelte";
  import {
    Axis,
    Grid,
  } from "@rilldata/web-common/components/data-graphic/guides";
  import { ScaleType } from "@rilldata/web-common/components/data-graphic/state";
  import type { ScaleStore } from "@rilldata/web-common/components/data-graphic/state/types";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { tableInteractionStore } from "@rilldata/web-common/features/dashboards/time-dimension-details/time-dimension-data-store";
  import DimensionValueMouseover from "@rilldata/web-common/features/dashboards/time-series/DimensionValueMouseover.svelte";
  import MeasurePan from "@rilldata/web-common/features/dashboards/time-series/MeasurePan.svelte";
  import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { numberKindForMeasure } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import type {
    MetricsViewSpecMeasureV2,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import { extent } from "d3-array";
  import { getContext } from "svelte";
  import { cubicOut } from "svelte/easing";
  import { fly } from "svelte/transition";
  import {
    DashboardTimeControls,
    TimeComparisonOption,
    TimeRangePreset,
    TimeRoundingStrategy,
  } from "../../../lib/time/types";
  import ChartBody from "./ChartBody.svelte";
  import MeasureScrub from "./MeasureScrub.svelte";
  import MeasureValueMouseover from "./MeasureValueMouseover.svelte";
  import {
    getOrderedStartEnd,
    localToTimeZoneOffset,
    niceMeasureExtents,
  } from "./utils";

  export let measure: MetricsViewSpecMeasureV2;
  export let metricViewName: string;
  export let width: number | undefined = undefined;
  export let height: number | undefined = undefined;
  export let xMin: Date | undefined = undefined;
  export let xMax: Date | undefined = undefined;
  export let yMin: number | undefined = undefined;
  export let yMax: number | undefined = undefined;

  export let timeGrain: V1TimeGrain;
  export let zone: string;

  export let showComparison = false;
  export let data;
  export let dimensionData: DimensionDataItem[] = [];
  export let xAccessor = "ts";
  export let labelAccessor = "label";
  export let yAccessor = "value";
  export let mouseoverValue: DomainCoordinates | undefined = undefined;
  export let validPercTotal: number | null = null;
  export let colors: string[];

  // control point for scrub functionality.
  export let isScrubbing = false;
  export let scrubStart;
  export let scrubEnd;

  export let mouseoverTimeFormat: (d: number | Date | string) => string = (v) =>
    v.toString();

  $: mouseoverFormat = createMeasureValueFormatter<null | undefined>(measure);
  $: numberKind = numberKindForMeasure(measure);

  const tweenProps = { duration: 400, easing: cubicOut };
  const xScale = getContext(contexts.scale("x")) as ScaleStore;

  let hovered: boolean = false;
  let scrub;
  let cursorClass;
  let preventScrubReset;

  $: hoveredTime =
    (mouseoverValue?.x instanceof Date && mouseoverValue?.x) ||
    $tableInteractionStore.time;

  $: hasSubrangeSelected = Boolean(scrubStart && scrubEnd);

  $: scrubStartCords = $xScale(scrubStart);
  $: scrubEndCords = $xScale(scrubEnd);
  $: mouseOverCords = $xScale(mouseoverValue?.x);

  let isOverStart = false;
  let isOverEnd = false;
  let isInsideScrub = false;

  $: if (mouseOverCords && scrubStartCords && scrubEndCords) {
    const min = Math.min(scrubStartCords, scrubEndCords);
    const max = Math.max(scrubStartCords, scrubEndCords);

    isOverStart = Math.abs(mouseOverCords - scrubStartCords) <= 5;
    isOverEnd = Math.abs(mouseOverCords - scrubEndCords) <= 5;

    isInsideScrub = Boolean(
      mouseOverCords > min + 5 && mouseOverCords < max - 5,
    );
  }

  $: isComparingDimension = Boolean(dimensionData?.length);

  /**
   * TODO: Optimize this such that we don't need to fetch main chart data
   * when comparing dimensions
   */
  $: [xExtentMin, xExtentMax] = extent(data, (d) => d[xAccessor]);
  $: [yExtentMin, yExtentMax] = extent(data, (d) => d[yAccessor]);
  let comparisonExtents;

  /** if we are making a comparison, factor this into the extents calculation.*/
  $: if (showComparison) {
    comparisonExtents = extent(data, (d) => d[`comparison.${yAccessor}`]);

    yExtentMin = Math.min(yExtentMin, comparisonExtents[0] || yExtentMin);
    yExtentMax = Math.max(yExtentMax, comparisonExtents[1] || yExtentMax);
  }

  /** if we have dimension data, factor that into the extents */
  let isFetchingDimensions = false;

  // Move to utils
  $: if (isComparingDimension) {
    let dimExtents = dimensionData.map((d) =>
      extent(d?.data || [], (datum) => datum[yAccessor]),
    );

    yExtentMin = dimExtents
      .map((e) => e[0])
      .reduce(
        (min, curr) => Math.min(min, isNaN(curr) ? Infinity : curr),
        Infinity,
      );
    yExtentMax = dimExtents
      .map((e) => e[1])
      .reduce(
        (max, curr) => Math.max(max, isNaN(curr) ? -Infinity : curr),
        -Infinity,
      );

    isFetchingDimensions = dimensionData.some((d) => d?.isFetching);
  }

  $: [internalYMin, internalYMax] = niceMeasureExtents(
    [
      yMin !== undefined ? yMin : yExtentMin,
      yMax !== undefined ? yMax : yExtentMax,
    ],
    6 / 5,
  );

  $: internalXMin = xMin || xExtentMin;
  $: internalXMax = xMax || xExtentMax;

  function inBounds(min, max, value) {
    return value >= min && value <= max;
  }

  function resetScrub() {
    metricsExplorerStore.setSelectedScrubRange(metricViewName, undefined);
  }

  function zoomScrub() {
    if (isScrubbing) return;

    const { start, end } = getOrderedStartEnd(scrubStart, scrubEnd);
    const adjustedStart = start ? localToTimeZoneOffset(start, zone) : start;
    const adjustedEnd = end ? localToTimeZoneOffset(end, zone) : end;

    metricsExplorerStore.setSelectedTimeRange(metricViewName, {
      name: TimeRangePreset.CUSTOM,
      start: adjustedStart,
      end: adjustedEnd,
    });
  }

  function updateScrub(start, end, isScrubbing) {
    const adjustedStart = start ? localToTimeZoneOffset(start, zone) : start;
    const adjustedEnd = end ? localToTimeZoneOffset(end, zone) : end;

    metricsExplorerStore.setSelectedScrubRange(metricViewName, {
      start: adjustedStart,
      end: adjustedEnd,
      isScrubbing: isScrubbing,
    });
  }

  function updateRange(start: Date, end: Date) {
    const timeRange = {
      name: TimeRangePreset.CUSTOM,
      start: start,
      end: end,
    };

    const comparisonTimeRange = showComparison
      ? ({
          name: TimeComparisonOption.CONTIGUOUS,
        } as DashboardTimeControls) // FIXME wrong typecasting across application
      : undefined;

    metricsExplorerStore.selectTimeRange(
      metricViewName,
      timeRange,
      timeGrain,
      comparisonTimeRange,
    );
  }

  function onMouseClick() {
    // skip if still scrubbing
    if (preventScrubReset) return;
    // skip if no scrub range selected
    if (!hasSubrangeSelected) return;

    const { start, end } = getOrderedStartEnd(scrubStart, scrubEnd);

    if (
      mouseoverValue?.x &&
      (mouseoverValue?.x < start || mouseoverValue?.x > end)
    ) {
      resetScrub();
    }
  }
</script>

<div class={`${cursorClass} select-none`}>
  <SimpleDataGraphic
    bind:hovered
    bind:mouseoverValue
    {height}
    left={0}
    let:config
    let:yScale
    on:click={() => onMouseClick()}
    on:scrub-end={() => scrub?.endScrub()}
    on:scrub-move={(e) => scrub?.moveScrub(e)}
    on:scrub-start={(e) => scrub?.startScrub(e)}
    overflowHidden={false}
    right={50}
    shareYScale={false}
    top={4}
    {width}
    xMaxTweenProps={tweenProps}
    xMinTweenProps={tweenProps}
    xType={ScaleType.DATE}
    yType={ScaleType.NUMBER}
    yMax={internalYMax}
    yMaxTweenProps={tweenProps}
    yMin={internalYMin}
    yMinTweenProps={tweenProps}
  >
    <Axis {numberKind} side="right" />
    <Grid />
    <MeasurePan
      on:pan={(e) => updateRange(e.detail.start, e.detail.end)}
      hovering={hovered}
    />
    <Body>
      <ChartBody
        {colors}
        {data}
        {dimensionData}
        dimensionValue={$tableInteractionStore?.dimensionValue}
        isHovering={Boolean(hoveredTime)}
        {scrubEnd}
        {scrubStart}
        {showComparison}
        {xAccessor}
        {xMax}
        {xMin}
        {yAccessor}
        {yExtentMax}
      />
      <line
        class="stroke-primary-200"
        x1={config.plotLeft}
        x2={config.plotLeft + config.plotRight}
        y1={yScale(0)}
        y2={yScale(0)}
      />
    </Body>
    {#if !isScrubbing && hoveredTime && !isFetchingDimensions}
      <WithRoundToTimegrain
        strategy={TimeRoundingStrategy.PREVIOUS}
        value={hoveredTime}
        {timeGrain}
        let:roundedValue
      >
        <WithBisector
          {data}
          callback={(d) => d[xAccessor]}
          value={roundedValue}
          let:point
        >
          {#if point && inBounds(internalXMin, internalXMax, point[xAccessor])}
            <g transition:fly={{ duration: 100, x: -4 }}>
              <text
                class="fill-gray-600"
                style:paint-order="stroke"
                stroke="white"
                stroke-width="3px"
                x={config.plotLeft + config.bodyBuffer + 6}
                y={config.plotTop + 10 + config.bodyBuffer}
              >
                {mouseoverTimeFormat(point[labelAccessor])}
              </text>
              {#if showComparison && point[`comparison.${labelAccessor}`]}
                <text
                  style:paint-order="stroke"
                  stroke="white"
                  stroke-width="3px"
                  class="fill-gray-400"
                  x={config.plotLeft + config.bodyBuffer + 6}
                  y={config.plotTop + 24 + config.bodyBuffer}
                >
                  {mouseoverTimeFormat(point[`comparison.${labelAccessor}`])} prev.
                </text>
              {/if}
            </g>
            <g transition:fly={{ duration: 100, x: -4 }}>
              {#if isComparingDimension}
                <DimensionValueMouseover
                  {point}
                  {xAccessor}
                  {yAccessor}
                  {colors}
                  {dimensionData}
                  dimensionValue={$tableInteractionStore?.dimensionValue}
                  {validPercTotal}
                  {mouseoverFormat}
                  {hovered}
                />
              {:else}
                <MeasureValueMouseover
                  {point}
                  {xAccessor}
                  {yAccessor}
                  {showComparison}
                  {mouseoverFormat}
                  {numberKind}
                />
              {/if}
            </g>
          {/if}
        </WithBisector>
      </WithRoundToTimegrain>
    {/if}

    <MeasureScrub
      bind:cursorClass
      bind:preventScrubReset
      bind:this={scrub}
      {data}
      {isInsideScrub}
      {isOverEnd}
      {isOverStart}
      {isScrubbing}
      {labelAccessor}
      {mouseoverTimeFormat}
      on:reset={() => resetScrub()}
      on:update={(e) =>
        updateScrub(e.detail.start, e.detail.stop, e.detail.isScrubbing)}
      on:zoom={() => zoomScrub()}
      start={scrubStart}
      stop={scrubEnd}
      timeGrainLabel={TIME_GRAIN[timeGrain].label}
    />
  </SimpleDataGraphic>
</div>
