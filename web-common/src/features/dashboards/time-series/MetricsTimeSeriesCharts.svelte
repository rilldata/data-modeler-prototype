<script lang="ts">
  import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import ReplacePivotDialog from "@rilldata/web-common/features/dashboards/pivot/ReplacePivotDialog.svelte";
  import {
    type PivotChipData,
    PivotChipType,
  } from "@rilldata/web-common/features/dashboards/pivot/types";

  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    metricsExplorerStore,
    useExploreStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import ChartTypeSelector from "@rilldata/web-common/features/dashboards/time-dimension-details/charts/ChartTypeSelector.svelte";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import BackToOverview from "@rilldata/web-common/features/dashboards/time-series/BackToOverview.svelte";
  import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
  import { getAdjustedChartTime } from "@rilldata/web-common/lib/time/ranges";
  import {
    type AvailableTimeGrain,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import {
    V1TimeGrain,
    type MetricsViewSpecMeasureV2,
  } from "@rilldata/web-common/runtime-client";
  import { TIME_GRAIN } from "../../../lib/time/config";
  import { DateTime, Duration, Interval } from "luxon";

  import ChartInteractions from "./ChartInteractions.svelte";
  import ChartWithTotal from "./ChartWithTotal.svelte";
  import { mergeMeasureFilters } from "../filters/measure-filters/measure-filter-utils";
  import { sanitiseExpression } from "../stores/filter-utils";
  import { adjustTimeInterval, updateChartInteractionStore } from "./utils";
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/shadcn/DashboardVisibilityDropdown.svelte";
  import { getEndOfPeriod } from "@rilldata/web-common/lib/time/transforms";

  export let exploreName: string;
  export let workspaceWidth: number;

  const StateManagers = getStateManagers();
  const {
    metricsViewName,
    selectors: {
      measures: {
        allMeasures,
        visibleMeasures,
        isMeasureValidPercentOfTotal,
        getMeasureByName,
        getFilteredMeasuresAndDimensions,
      },
      dimensionFilters: { includedDimensionValues },
      dimensions: { comparisonDimension },
    },
    actions: {
      measures: { toggleMeasureVisibility },
    },
    timeRangeSummaryStore,
    validSpecStore,
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(getStateManagers());

  $: timeControls = $timeControlsStore;

  let scrubStart: Date | undefined = undefined;
  let scrubEnd: Date | undefined = undefined;
  let mouseoverValue: DomainCoordinates | undefined = undefined;
  let parentElement: HTMLDivElement;

  $: exploreStore = useExploreStore(exploreName);

  $: ({
    selectedTimezone,
    // selectedTimeRange,
    leaderboardContextColumn,
    selectedScrubRange,
    // selectedComparisonTimeRange,
    tdd: { expandedMeasureName, chartType: tddChartType },
  } = $exploreStore);

  $: ({
    selectedTimeRange,
    allTimeRange,
    showTimeComparison,
    selectedComparisonTimeRange,
    // minTimeGrain,
  } = timeControls);

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start)
          .setZone(selectedTimezone)
          .startOf(TIME_GRAIN[timeGrain].label),
        DateTime.fromJSDate(selectedTimeRange.end)
          .setZone(selectedTimezone)
          .minus({ millisecond: 1 })
          .endOf(TIME_GRAIN[timeGrain].label),
      )
    : Interval.fromDateTimes(
        DateTime.fromJSDate(allTimeRange?.start).setZone(selectedTimezone),
        DateTime.fromJSDate(allTimeRange?.end).setZone(selectedTimezone),
      );

  $: comparisonInterval = selectedComparisonTimeRange?.start
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedComparisonTimeRange.start).setZone(
          selectedTimezone,
        ),
        DateTime.fromJSDate(selectedComparisonTimeRange.end).setZone(
          selectedTimezone,
        ),
      )
    : undefined;

  $: points = interval?.splitBy(
    Duration.fromObject({ [TIME_GRAIN[timeGrain].label]: 1 }),
  );

  $: console.log(points);
  // $: expandedMeasureName = tdd?.expandedMeasureName;
  $: isInTimeDimensionView = Boolean(expandedMeasureName);
  // $: comparisonDimension = $exploreStore?.selectedComparisonDimension;
  $: showComparison = Boolean(
    !$comparisonDimension && $timeControlsStore.showTimeComparison,
  );
  // $: tddChartType = tdd.chartType;
  $: timeGrain =
    timeControls.selectedTimeRange?.interval ??
    timeControls.minTimeGrain ??
    V1TimeGrain.TIME_GRAIN_DAY;
  $: isScrubbing = Boolean(selectedScrubRange?.isScrubbing);
  $: isAllTime =
    $timeControlsStore.selectedTimeRange?.name === TimeRangePreset.ALL_TIME;

  $: includedValuesForDimension = $comparisonDimension?.name
    ? $includedDimensionValues($comparisonDimension.name)
    : [];
  // $: leaderboardContextColumn = $exploreStore.leaderboardContextColumn;

  $: isAlternateChart = tddChartType !== TDDChart.DEFAULT;

  $: expandedMeasure = $getMeasureByName(expandedMeasureName);
  $: max = $timeRangeSummaryStore.data?.timeRangeSummary?.max;
  $: min = $timeRangeSummaryStore.data?.timeRangeSummary?.min;

  // List of measures which will be shown on the dashboard
  let renderedMeasures: MetricsViewSpecMeasureV2[];
  $: {
    renderedMeasures = expandedMeasure ? [expandedMeasure] : $visibleMeasures;
  }

  $: if (timeControls.ready) {
    // adjust scrub values for Javascript's timezone changes
    scrubStart = adjustOffsetForZone(
      selectedScrubRange?.start,
      selectedTimezone,
    );
    scrubEnd = adjustOffsetForZone(selectedScrubRange?.end, selectedTimezone);
  }

  $: ({ start: startValue, end: endValue } = getAdjustedChartTime(
    timeControls.selectedTimeRange?.start,
    timeControls.selectedTimeRange?.end,
    selectedTimezone,
    timeGrain,
    timeControls.selectedTimeRange?.name,
    $validSpecStore.data?.explore?.defaultPreset?.timeRange,
    tddChartType,
  ));

  $: visibleMeasureNames = $visibleMeasures
    .map(({ name }) => name)
    .filter(isDefined);
  $: allMeasureNames = $allMeasures.map(({ name }) => name).filter(isDefined);
  function isDefined(value: string | undefined): value is string {
    return value !== undefined;
  }

  $: activeTimeGrain = timeControls.selectedTimeRange?.interval;

  let showReplacePivotModal = false;
  function startPivotForTimeseries() {
    const pivot = $exploreStore?.pivot;

    if (
      pivot.rows.dimension.length ||
      pivot.columns.measure.length ||
      pivot.columns.dimension.length
    ) {
      showReplacePivotModal = true;
    } else {
      createPivot();
    }
  }

  function getTimeDimension() {
    return {
      id: timeControls.selectedTimeRange?.interval,
      title: TIME_GRAIN[activeTimeGrain as AvailableTimeGrain]?.label,
      type: PivotChipType.Time,
    } as PivotChipData;
  }

  function createPivot() {
    showReplacePivotModal = false;

    const measures = renderedMeasures
      .filter((m) => m.name !== undefined)
      .map((m) => {
        return {
          id: m.name as string,
          title: m.label || (m.name as string),
          type: PivotChipType.Measure,
        };
      });

    metricsExplorerStore.createPivot(
      exploreName,
      { dimension: [getTimeDimension()] },
      {
        dimension: [],
        measure: measures,
      },
    );
  }

  // $: xExtents = [
  //   timeControls?.adjustedStart
  //     ? new Date(timeControls?.adjustedStart)
  //     : new Date(min),
  //   timeControls.selectedTimeRange?.end
  //     ? getEnd(timeControls.selectedTimeRange?.end, timeGrain, selectedTimezone)
  //     : new Date(max),
  // ];

  function getEnd(date: Date, timeGrain: V1TimeGrain, zone: string) {
    const dateTime = DateTime.fromJSDate(date, { zone });
    console.log("format", dateTime.toRFC2822());

    return dateTime.endOf(TIME_GRAIN[timeGrain].label).toJSDate();
  }

  $: console.log(interval.toString());
</script>

{#if startValue && endValue}
  <div class="h-full w-[460px] lg:w-[620px]">
    <div class:mb-6={isAlternateChart} class="flex items-center gap-x-1 px-2.5">
      {#if isInTimeDimensionView}
        <BackToOverview {exploreName} />
        <ChartTypeSelector
          hasComparison={Boolean(
            showComparison || includedValuesForDimension.length,
          )}
          {exploreName}
          chartType={tddChartType}
        />
      {:else}
        <DashboardVisibilityDropdown
          category="Measures"
          tooltipText="Choose measures to display"
          onSelect={(name) => toggleMeasureVisibility(allMeasureNames, name)}
          selectableItems={$allMeasures.map(({ name, label }) => ({
            name: name ?? "",
            label: label ?? name ?? "",
          }))}
          selectedItems={visibleMeasureNames}
          onToggleSelectAll={() => {
            toggleMeasureVisibility(allMeasureNames);
          }}
        />

        <button
          class="h-6 px-1.5 py-px rounded-sm hover:bg-gray-200 text-gray-700 ml-auto"
          on:click={() => {
            startPivotForTimeseries();
          }}
        >
          Start Pivot
        </button>
      {/if}
    </div>

    <div class="z-10 gap-x-9 flex flex-row pt-4" style:padding-left="118px">
      <div class="relative w-full">
        <ChartInteractions {exploreName} {showComparison} {timeGrain} />
        {#if tddChartType === TDDChart.DEFAULT}
          <div class="translate-x-5">
            {#if selectedTimeRange && startValue && endValue}
              <SimpleDataGraphic
                height={26}
                overflowHidden={false}
                top={29}
                bottom={0}
                right={isInTimeDimensionView ? 10 : 25}
                xMin={startValue}
                xMax={endValue}
              >
                <Axis superlabel side="top" placement="start" />
              </SimpleDataGraphic>
            {/if}
          </div>
        {/if}
      </div>
    </div>

    {#if renderedMeasures}
      <div
        bind:this={parentElement}
        class:pb-4={!isInTimeDimensionView}
        class="flex flex-col gap-y-2 w-full overflow-y-scroll h-full"
      >
        {#each renderedMeasures as measure (measure.name)}
          <ChartWithTotal
            {isScrubbing}
            {scrubStart}
            {scrubEnd}
            {max}
            {min}
            {points}
            {startValue}
            {endValue}
            {interval}
            {isAllTime}
            {parentElement}
            bind:mouseoverValue
            {expandedMeasureName}
            isComparison={showComparison}
            {measure}
            metricsViewName={$metricsViewName}
            {exploreName}
            showComparison={showTimeComparison}
            {timeControls}
            {comparisonInterval}
            {timeGrain}
            {tddChartType}
            {leaderboardContextColumn}
            onExpandMeasure={() => {
              metricsExplorerStore.setExpandedMeasureName(
                exploreName,
                measure.name,
              );
            }}
            whereFilter={sanitiseExpression(
              mergeMeasureFilters($exploreStore),
              undefined,
            )}
            selectedTimeZone={selectedTimezone}
            filteredMeasures={$getFilteredMeasuresAndDimensions(
              $validSpecStore.data?.metricsView ?? {},
              $allMeasures.map((m) => m.name ?? "") ?? [],
            ).measures}
            isValidPercTotal={measure.name
              ? $isMeasureValidPercentOfTotal(measure.name)
              : false}
            comparisonDimension={$comparisonDimension}
            stateManagers={StateManagers}
            onChartHover={(dimension, ts, formattedData) => {
              updateChartInteractionStore(
                ts,
                dimension,
                isAllTime,
                formattedData,
              );
            }}
            onChartBrush={(interval) => {
              const { start, end } = adjustTimeInterval(
                interval,
                selectedTimezone,
              );

              metricsExplorerStore.setSelectedScrubRange(exploreName, {
                start,
                end,
                isScrubbing: true,
              });
            }}
            onChartBrushEnd={(interval) => {
              const { start, end } = adjustTimeInterval(
                interval,
                selectedTimezone,
              );

              metricsExplorerStore.setSelectedScrubRange(exploreName, {
                start,
                end,
                isScrubbing: false,
              });
            }}
            onChartBrushClear={({ start, end }) => {
              metricsExplorerStore.setSelectedScrubRange(exploreName, {
                start,
                end,
                isScrubbing: false,
              });
            }}
          />
        {/each}
      </div>
    {/if}
  </div>
{/if}

<ReplacePivotDialog
  open={showReplacePivotModal}
  onCancel={() => {
    showReplacePivotModal = false;
  }}
  onReplace={createPivot}
/>
