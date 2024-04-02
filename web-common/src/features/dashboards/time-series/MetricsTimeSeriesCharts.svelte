<script lang="ts">
  import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import SearchableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterButton.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
  import { createShowHideMeasuresStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { chartInteractionColumn } from "@rilldata/web-common/features/dashboards/time-dimension-details/time-dimension-data-store";
  import BackToOverview from "@rilldata/web-common/features/dashboards/time-series/BackToOverview.svelte";
  import { useTimeSeriesDataStore } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
  import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
  import { getAdjustedChartTime } from "@rilldata/web-common/lib/time/ranges";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
  import { TIME_GRAIN } from "../../../lib/time/config";
  import { runtime } from "../../../runtime-client/runtime-store";
  import Spinner from "../../entity-management/Spinner.svelte";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import ChartInteractions from "./ChartInteractions.svelte";
  import MeasureChart from "./MeasureChart.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";
  import type { DimensionDataItem } from "./multiple-dimension-queries";
  import { V1MetricsViewAggregationResponse } from "@rilldata/web-common/runtime-client";
  import { page } from "$app/stores";
  import { prepareTimeSeries } from "@rilldata/web-common/features/dashboards/time-series/utils";

  export let metricViewName: string;
  export let workspaceWidth: number;
  export let totals: void | V1MetricsViewAggregationResponse;

  $: timeSeries = $page.data.timeSeries;
  $: timeZone = $page.data.timeZone;
  $: timeGrain = $page.data.timeGrain;

  $: formattedData = prepareTimeSeries(
    timeSeries.data,
    undefined,
    TIME_GRAIN[timeGrain].duration,
    timeZone,
  );

  // const {
  //   selectors: {
  //     measures: { isMeasureValidPercentOfTotal },
  //     dimensionFilters: { includedDimensionValues },
  //   },
  // } = getStateManagers();

  const timeControlsStore = useTimeControlStore(getStateManagers());
  const timeSeriesDataStore = useTimeSeriesDataStore(getStateManagers());

  let scrubStart;
  let scrubEnd;

  let mouseoverValue: DomainCoordinates | undefined = undefined;
  let startValue: Date;
  let endValue: Date;

  // let dataCopy: TimeSeriesDatum[];
  let dimensionDataCopy: DimensionDataItem[] = [];

  $: dashboardStore = useDashboardStore(metricViewName);
  $: instanceId = $runtime.instanceId;

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metricsView = useMetricsView(instanceId, metricViewName);

  $: showHideMeasures = createShowHideMeasuresStore(
    metricViewName,
    metricsView,
  );

  $: expandedMeasureName = $dashboardStore?.expandedMeasureName;
  $: comparisonDimension = $dashboardStore?.selectedComparisonDimension;
  $: showComparison = !comparisonDimension && $timeControlsStore.showComparison;
  $: interval =
    $timeControlsStore.selectedTimeRange?.interval ??
    $timeControlsStore.minTimeGrain;
  $: isScrubbing = $dashboardStore?.selectedScrubRange?.isScrubbing;

  $: isPercOfTotalAsContextColumn =
    $dashboardStore?.leaderboardContextColumn ===
    LeaderboardContextColumn.PERCENT;
  $: includedValuesForDimension =
    $page.url.searchParams.get(comparisonDimension ?? "")?.split(",") ?? [];

  // List of measures which will be shown on the dashboard
  $: renderedMeasures = $metricsView.data?.measures?.filter(
    expandedMeasureName
      ? (measure) => measure.name === expandedMeasureName
      : (_, i) => $showHideMeasures.selectedItems[i],
  );

  // $: totals = $timeSeriesDataStore.total;
  $: totalsComparisons = $timeSeriesDataStore.comparisonTotal;

  $: if (
    $timeSeriesDataStore?.dimensionChartData?.length ||
    !comparisonDimension ||
    includedValuesForDimension.length === 0
  ) {
    dimensionDataCopy = $timeSeriesDataStore.dimensionChartData || [];
  }
  $: dimensionData = dimensionDataCopy;

  // FIXME: move this logic to a function + write tests.
  $: if ($timeControlsStore.ready) {
    // adjust scrub values for Javascript's timezone changes
    scrubStart = adjustOffsetForZone(
      $dashboardStore?.selectedScrubRange?.start,
      $dashboardStore.selectedTimezone,
    );
    scrubEnd = adjustOffsetForZone(
      $dashboardStore?.selectedScrubRange?.end,
      $dashboardStore.selectedTimezone,
    );

    const slicedData =
      $timeControlsStore.selectedTimeRange?.name === TimeRangePreset.ALL_TIME
        ? formattedData?.slice(1)
        : formattedData?.slice(1, -1);
    chartInteractionColumn.update((state) => {
      const { start, end } = getOrderedStartEnd(scrubStart, scrubEnd);

      const { position: startPos } = bisectData(
        start,
        "center",
        "ts_position",
        slicedData,
      );
      const { position: endPos } = bisectData(
        end,
        "center",
        "ts_position",
        slicedData,
      );

      return {
        hover: isScrubbing ? undefined : state.hover,
        scrubStart: startPos,
        scrubEnd: endPos,
      };
    });

    const adjustedChartValue = getAdjustedChartTime(
      $timeControlsStore.selectedTimeRange?.start,
      $timeControlsStore.selectedTimeRange?.end,
      $dashboardStore?.selectedTimezone,
      interval,
      $timeControlsStore.selectedTimeRange?.name,
      $metricsView?.data?.defaultTimeRange,
    );

    if (adjustedChartValue?.start) {
      startValue = adjustedChartValue?.start;
    }
    if (adjustedChartValue?.end) {
      endValue = adjustedChartValue?.end;
    }
  }

  $: if (
    expandedMeasureName &&
    formattedData &&
    $timeControlsStore.selectedTimeRange &&
    !isScrubbing
  ) {
    if (!mouseoverValue?.x || !(mouseoverValue.x instanceof Date)) {
      chartInteractionColumn.update((state) => ({
        ...state,
        hover: undefined,
      }));
    } else {
      const { position: columnNum } = bisectData(
        mouseoverValue.x,
        "center",
        "ts_position",
        $timeControlsStore.selectedTimeRange?.name === TimeRangePreset.ALL_TIME
          ? formattedData?.slice(1)
          : formattedData?.slice(1, -1),
      );

      if ($chartInteractionColumn?.hover !== columnNum)
        chartInteractionColumn.update((state) => ({
          ...state,
          hover: columnNum,
        }));
    }
  }

  const toggleMeasureVisibility = (e) => {
    showHideMeasures.toggleVisibility(e.detail.name);
  };
  const setAllMeasuresNotVisible = () => {
    showHideMeasures.setAllToNotVisible();
  };
  const setAllMeasuresVisible = () => {
    showHideMeasures.setAllToVisible();
  };
</script>

<TimeSeriesChartContainer
  enableFullWidth={Boolean(expandedMeasureName)}
  end={endValue}
  start={startValue}
  {workspaceWidth}
>
  <div class="flex pl-1">
    {#if expandedMeasureName}
      <BackToOverview {metricViewName} />
    {:else}
      <SearchableFilterButton
        label="Measures"
        on:deselect-all={setAllMeasuresNotVisible}
        on:item-clicked={toggleMeasureVisibility}
        on:select-all={setAllMeasuresVisible}
        selectableItems={$showHideMeasures.selectableItems}
        selectedItems={$showHideMeasures.selectedItems}
        tooltipText="Choose measures to display"
      />
    {/if}
  </div>

  <div class="z-10 gap-x-9 flex flex-row pt-4" style:padding-left="118px">
    <div class="relative w-full">
      <ChartInteractions
        {metricViewName}
        {showComparison}
        timeGrain={interval}
      />
      <div class="translate-x-5">
        {#if $dashboardStore?.selectedTimeRange && startValue && endValue}
          <SimpleDataGraphic
            height={26}
            overflowHidden={false}
            top={29}
            bottom={0}
            xMin={startValue}
            xMax={endValue}
          >
            <Axis superlabel side="top" placement="start" />
          </SimpleDataGraphic>
        {/if}
      </div>
    </div>
  </div>

  <!-- bignumbers and line charts -->
  {#if renderedMeasures}
    <div class="flex flex-col gap-y-2 overflow-y-scroll h-full max-h-fit pb-4">
      <!-- FIXME: this is pending the remaining state work for show/hide measures and dimensions -->
      {#each renderedMeasures as measure (measure.name)}
        <!-- FIXME: I can't select the big number by the measure id. -->
        <!-- for bigNum, catch nulls and convert to undefined.  -->
        <!-- {@const bigNum = totals.data?.[0]?.[measure?.name ?? ""] ?? undefined} -->
        {@const comparisonValue = measure.name
          ? totalsComparisons?.[measure.name]
          : undefined}
        {@const isValidPercTotal = measure.name
          ? $page.data.measures?.find((m) => m.name === measure.name)
              ?.validPercentOfTotal
          : false}

        <div class="flex flex-row gap-x-7">
          {#await totals}
            <MeasureBigNumber
              {measure}
              value={undefined}
              isMeasureExpanded={!!expandedMeasureName}
              {showComparison}
              comparisonOption={$timeControlsStore?.selectedComparisonTimeRange
                ?.name}
              {comparisonValue}
              status={EntityStatus.Running}
              on:expand-measure={() => {
                metricsExplorerStore.setExpandedMeasureName(
                  metricViewName,
                  measure.name,
                );
              }}
            />
          {:then totalsData}
            <MeasureBigNumber
              {measure}
              value={totalsData?.data?.[0]?.[measure?.name ?? ""]}
              isMeasureExpanded={!!expandedMeasureName}
              {showComparison}
              comparisonOption={$timeControlsStore?.selectedComparisonTimeRange
                ?.name}
              {comparisonValue}
              status={EntityStatus.Idle}
              on:expand-measure={() => {
                metricsExplorerStore.setExpandedMeasureName(
                  metricViewName,
                  measure.name,
                );
              }}
            />
          {/await}

          {#if $timeSeriesDataStore?.isError}
            <div class="p-5"><CrossIcon /></div>
          {:else if formattedData && interval}
            <MeasureChart
              bind:mouseoverValue
              {measure}
              {isScrubbing}
              {scrubStart}
              {scrubEnd}
              {metricViewName}
              data={formattedData}
              {dimensionData}
              zone={$dashboardStore.selectedTimezone}
              xAccessor="ts_position"
              labelAccessor="ts"
              timeGrain={interval}
              yAccessor={measure.name}
              xMin={startValue}
              xMax={endValue}
              {showComparison}
              validPercTotal={isPercOfTotalAsContextColumn && isValidPercTotal
                ? undefined
                : null}
              mouseoverTimeFormat={(value) => {
                /** format the date according to the time grain */

                return interval
                  ? new Date(value).toLocaleDateString(
                      undefined,
                      TIME_GRAIN[interval].formatDate,
                    )
                  : value.toString();
              }}
            />
          {:else}
            <div class="flex items-center justify-center w-24">
              <Spinner status={EntityStatus.Running} />
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</TimeSeriesChartContainer>
