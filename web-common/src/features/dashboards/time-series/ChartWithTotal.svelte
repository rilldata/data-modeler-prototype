<script lang="ts">
  import {
    createQueryServiceMetricsViewAggregation,
    type MetricsViewSpecDimensionV2,
    type MetricsViewSpecMeasureV2,
    type V1Expression,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import TDDAlternateChart from "../time-dimension-details/charts/TDDAlternateChart.svelte";
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { createQueryServiceMetricsViewTimeSeries } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAndExpression,
    filterExpressions,
    matchExpressionByName,
  } from "../stores/filter-utils";
  import { type TimeControlState } from "../time-controls/time-control-store";
  import {
    localToTimeZoneOffset,
    prepareTimeSeries,
    updateChartInteractionStore,
  } from "./utils";
  import { Period } from "@rilldata/web-common/lib/time/types";
  import { TDDChart } from "../time-dimension-details/types";
  import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
  import { onMount } from "svelte";
  import { LeaderboardContextColumn } from "../leaderboard-context-column";
  import TimeDimensionDisplay from "../time-dimension-details/TimeDimensionDisplay.svelte";
  import { getDimensionValueTimeSeries } from "./multiple-dimension-queries";
  import type { StateManagers } from "../state-managers/state-managers";
  import type { TimeSeriesDatum } from "./timeseries-data-store";
  import Chart from "@rilldata/web-common/components/time-series-chart/Chart.svelte";
  import { numberKindForMeasure } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type { Interval } from "luxon";
  import { MainLineColor } from "./chart-colors";

  export let measure: MetricsViewSpecMeasureV2;

  export let parentElement: HTMLElement;
  export let exploreName: string;
  export let metricsViewName: string;
  export let isComparison: boolean;
  export let mouseoverValue: DomainCoordinates<number | Date> | undefined;
  export let expandedMeasureName: string | undefined;
  export let isAllTime: boolean;
  export let scrubStart: Date | undefined;
  export let scrubEnd: Date | undefined;
  export let startValue: Date;
  export let endValue: Date;
  export let showComparison: boolean;
  // export let xExtents: [Date, Date];
  export let isScrubbing: boolean;
  export let timeGrain: V1TimeGrain;
  export let points: Interval[];
  export let interval: Interval;
  export let comparisonInterval: Interval | undefined;
  export let timeControls: TimeControlState;
  export let comparisonDimension: MetricsViewSpecDimensionV2 | undefined =
    undefined;
  export let filteredMeasures: string[];
  export let selectedTimeZone: string;
  export let leaderboardContextColumn: LeaderboardContextColumn;
  export let tddChartType: TDDChart;
  export let whereFilter: V1Expression | undefined;
  export let stateManagers: StateManagers | undefined = undefined;
  export let onExpandMeasure: () => void = () => {};
  export let onChartHover: (
    dimension: string,
    ts: number | Date | undefined,
    formattedData: TimeSeriesDatum[],
  ) => void = () => {};
  export let onChartBrush: (interval: {
    start: Date;
    end: Date;
  }) => void = () => {};
  export let onChartBrushEnd: (interval: {
    start: Date;
    end: Date;
  }) => void = () => {};
  export let onChartBrushClear: (interval: {
    start: Date;
    end: Date;
  }) => void = () => {};

  let visible = false;
  let container: HTMLElement;

  const observer = new IntersectionObserver(
    ([entry]) => {
      visible = entry.isIntersecting;
    },
    {
      root: parentElement,
      rootMargin: "120px",
      threshold: 0,
    },
  );

  onMount(() => {
    observer.observe(container);
  });

  $: ({ instanceId } = $runtime);

  $: isPercOfTotalAsContextColumn =
    leaderboardContextColumn === LeaderboardContextColumn.PERCENT;

  $: measureName = measure.name as string;

  $: updatedFilter = filterExpressions(
    whereFilter || createAndExpression([]),
    (e) => !matchExpressionByName(e, comparisonDimension?.name ?? ""),
  );

  $: primaryTimeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricsViewName,
    {
      measureNames: [measureName],
      where: whereFilter,
      timeStart: timeControls.adjustedStart,
      timeEnd: timeControls.adjustedEnd,
      timeGranularity: timeGrain,
      timeZone: selectedTimeZone,
    },
    {
      query: {
        enabled: visible && !!timeControls.ready,
        keepPreviousData: true,
      },
    },
  );

  $: comparisonTimeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricsViewName,
    {
      measureNames: [measureName],
      where: whereFilter,
      timeStart: timeControls.comparisonAdjustedStart,
      timeEnd: timeControls.comparisonAdjustedEnd,
      timeGranularity: timeGrain,
      timeZone: selectedTimeZone,
    },
    {
      query: {
        enabled:
          visible &&
          !!timeControls.ready &&
          isComparison &&
          !!timeControls.comparisonAdjustedStart,
        keepPreviousData: true,
      },
    },
  );

  $: console.log({
    timeStart: timeControls.comparisonAdjustedStart,
    timeEnd: timeControls.comparisonAdjustedEnd,
  });

  $: primaryTotalQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: [measure],
      where: whereFilter,
      timeRange: {
        start: timeControls.timeStart,
        end: timeControls.timeEnd,
      },
    },
    {
      query: {
        enabled: visible && !!timeControls.ready,
      },
    },
  );

  $: comparisonTotalQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: [measure],
      where: whereFilter,
      timeRange: {
        start: timeControls?.comparisonTimeStart,
        end: timeControls?.comparisonTimeEnd,
      },
    },
    {
      query: {
        enabled: visible && isComparison && !!timeControls.ready,
      },
    },
  );

  $: unfilteredTotalQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: [measure],
      where: updatedFilter,
      timeRange: {
        start: timeControls.timeStart,
        end: timeControls.timeEnd,
      },
    },
    {
      query: {
        enabled: visible && !!timeControls.ready,
      },
    },
  );

  $: ({ data: primaryData, error: primaryError } = $primaryTimeSeriesQuery);
  $: ({ data: comparisonData } = $comparisonTimeSeriesQuery);

  $: ({
    error: primaryTotalError,
    isFetching: primaryTotalIsFetching,
    data: primaryTotalData,
  } = $primaryTotalQuery);
  $: ({ data: comparisonTotalData } = $comparisonTotalQuery);
  $: ({ data: unfilteredTotalData } = $unfilteredTotalQuery);

  $: unfilteredTotal = unfilteredTotalData?.data?.[0]?.[measureName];
  $: primaryTotal = primaryTotalData?.data?.[0]?.[measureName];
  $: comparisonTotal = comparisonTotalData?.data?.[0]?.[measureName];

  $: intervalDuration = TIME_GRAIN[timeGrain]?.duration as Period;

  $: formattedData = prepareTimeSeries(
    primaryData?.data || [],
    comparisonData?.data || [],
    intervalDuration,
    selectedTimeZone,
  );

  $: dimensionDataQuery = getDimensionValueTimeSeries(
    stateManagers,
    filteredMeasures,
    "chart",
    visible && !!comparisonDimension,
  );

  $: dimensionData = $dimensionDataQuery;

  $: if (
    expandedMeasureName &&
    formattedData &&
    timeControls.selectedTimeRange &&
    !isScrubbing
  ) {
    updateChartInteractionStore(
      mouseoverValue?.x,
      undefined,
      isAllTime,
      formattedData,
    );
  }

  function updateScrub(start: Date, end: Date, isScrubbing: boolean) {
    const adjustedStart = start
      ? localToTimeZoneOffset(start, selectedTimeZone)
      : start;
    const adjustedEnd = end
      ? localToTimeZoneOffset(end, selectedTimeZone)
      : end;

    metricsExplorerStore.setSelectedScrubRange(exploreName, {
      start: adjustedStart,
      end: adjustedEnd,
      isScrubbing: isScrubbing,
    });
  }
</script>

<div
  class="flex flex-row gap-x-4 w-full h-24"
  bind:this={container}
  role="presentation"
  aria-label="{measure.label ?? measure.name} chart"
>
  <MeasureBigNumber
    {measure}
    value={primaryTotal}
    isMeasureExpanded={!!expandedMeasureName}
    showComparison={isComparison}
    comparisonValue={comparisonTotal}
    errorMessage={primaryTotalError?.message}
    status={primaryTotalError
      ? EntityStatus.Error
      : primaryTotalIsFetching
        ? EntityStatus.Running
        : EntityStatus.Idle}
    {onExpandMeasure}
  />

  {#if primaryError}
    <div
      class="flex flex-col p-5 items-center justify-center text-xs ui-copy-muted"
    >
      {#if primaryError.response.data?.message}
        <span>
          Error: {primaryError.response.data.message}
        </span>
      {:else}
        <span>Unable to fetch data from the API</span>
      {/if}
    </div>
  {:else if expandedMeasureName && tddChartType != TDDChart.DEFAULT}
    <TDDAlternateChart
      {timeGrain}
      chartType={tddChartType}
      {expandedMeasureName}
      totalsData={formattedData}
      {dimensionData}
      xMin={startValue}
      xMax={endValue}
      isTimeComparison={isComparison}
      isScrubbing={Boolean(isScrubbing)}
      on:chart-hover={(e) => {
        const { dimension, ts } = e.detail;

        onChartHover(ts, dimension, formattedData);
      }}
      on:chart-brush={(e) => {
        const { interval } = e.detail;

        onChartBrush(interval);
      }}
      on:chart-brush-end={(e) => {
        const { interval } = e.detail;
        onChartBrushEnd(interval);
      }}
      on:chart-brush-clear={(e) => {
        const { start, end } = e.detail;

        onChartBrushClear({ start, end });
      }}
    />
  {:else if primaryData?.data && timeGrain && interval.isValid}
    <Chart
      {interval}
      {points}
      {selectedTimeZone}
      formatterFunction={createMeasureValueFormatter(measure)}
      yAccessor={measure.name ?? ""}
      {timeGrain}
      dimensionData={[]}
      {showComparison}
      {comparisonInterval}
      primaryData={primaryData.data}
      comparisonData={comparisonData?.data ?? []}
      onUpdateScrub={updateScrub}
      showBorder={false}
      scrubRange={{ start: scrubStart, end: scrubEnd }}
    />
  {:else}
    <div class="flex items-center justify-center w-24">
      <Spinner status={EntityStatus.Running} />
    </div>
  {/if}
</div>

{#if expandedMeasureName}
  <hr class="border-t border-gray-200 -ml-4" />
  <TimeDimensionDisplay
    error={primaryError}
    {measure}
    {exploreName}
    formattedTimeSeriesData={formattedData}
    {primaryTotal}
    {comparisonTotal}
    {unfilteredTotal}
    {comparisonDimension}
    showTimeComparison={isComparison}
  />
{/if}
