<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { isExpressionUnsupported } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { useModelHasTimeSeries } from "../selectors";
  import { getStateManagers } from "../state-managers/state-managers";
  import ComparisonPill from "../time-controls/comparison-pill/ComparisonPill.svelte";
  import SuperPill from "../time-controls/super-pill/SuperPill.svelte";
  import { useTimeControlStore } from "../time-controls/time-control-store";
  import FilterButton from "./FilterButton.svelte";
  import DimensionFilter from "./dimension-filters/DimensionFilter.svelte";
  import type {
    V1ExploreTimeRange,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import {
    ALL_TIME_RANGE_ALIAS,
    CUSTOM_TIME_RANGE_ALIAS,
    deriveInterval,
  } from "../time-controls/new-time-controls";
  import {
    TimeComparisonOption,
    TimeRangePreset,
    type DashboardTimeControls,
  } from "@rilldata/web-common/lib/time/types";
  import {
    metricsExplorerStore,
    useExploreState,
  } from "../stores/dashboard-stores";
  import type { TimeRange } from "@rilldata/web-common/lib/time/types";
  import { DateTime, Interval } from "luxon";
  import { initLocalUserPreferenceStore } from "../user-preferences";
  import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
  import { getValidComparisonOption } from "../time-controls/time-range-store";

  export let readOnly = false;
  export let timeRanges: V1ExploreTimeRange[];
  export let metricsViewName: string;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  const StateManagers = getStateManagers();
  const {
    exploreName,
    validSpecStore,
    actions: {
      dimensionsFilter: {
        toggleDimensionValueSelection,
        removeDimensionFilter,
        toggleDimensionFilterMode,
      },
      measuresFilter: { setMeasureFilter, removeMeasureFilter },
      filters: { clearAllFilters, setTemporaryFilterName },
    },
    selectors: {
      dimensions: { allDimensions },
      dimensionFilters: {
        dimensionHasFilter,
        getDimensionFilterItems,
        getAllDimensionFilterItems,
        isFilterExcludeMode,
      },

      measures: { allMeasures, filteredSimpleMeasures },
      measureFilters: {
        getMeasureFilterItems,
        getAllMeasureFilterItems,
        measureHasFilter,
      },
      pivot: { showPivot },
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
    dashboardStore,
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(StateManagers);

  let showDefaultItem = false;

  $: ({
    selectedTimeRange,
    allTimeRange,
    showTimeComparison,
    selectedComparisonTimeRange,
    minTimeGrain,
    timeStart,
    timeEnd,
    ready: timeControlsReady,
  } = $timeControlsStore);

  $: ({ instanceId } = $runtime);

  $: exploreSpec = $validSpecStore.data?.explore ?? {};
  $: metricsViewSpec = $validSpecStore.data?.metricsView ?? {};

  $: exploreState = useExploreState($exploreName);
  $: activeTimeZone = $exploreState?.selectedTimezone;

  $: selectedRangeAlias = selectedTimeRange?.name;
  $: activeTimeGrain = selectedTimeRange?.interval;
  $: defaultTimeRange = exploreSpec.defaultPreset?.timeRange;

  $: dimensions = $allDimensions;
  $: dimensionIdMap = getMapFromArray(
    dimensions,
    (dimension) => (dimension.name || dimension.column) as string,
  );

  $: measures = $allMeasures;
  $: measureIdMap = getMapFromArray(measures, (m) => m.name as string);

  $: currentDimensionFilters = $getDimensionFilterItems(dimensionIdMap);
  $: allDimensionFilters = $getAllDimensionFilterItems(
    currentDimensionFilters,
    dimensionIdMap,
  );

  $: currentMeasureFilters = $getMeasureFilterItems(measureIdMap);
  $: allMeasureFilters = $getAllMeasureFilterItems(
    currentMeasureFilters,
    measureIdMap,
  );

  // hasFilter only checks for complete filters and excludes temporary ones
  $: hasFilters =
    currentDimensionFilters.length > 0 || currentMeasureFilters.length > 0;
  $: metricTimeSeries = useModelHasTimeSeries(instanceId, metricsViewName);
  $: hasTimeSeries = $metricTimeSeries.data;

  $: isComplexFilter = isExpressionUnsupported($dashboardStore.whereFilter);

  $: availableTimeZones = exploreSpec.timeZones ?? [];

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
      )
    : allTimeRange
      ? Interval.fromDateTimes(allTimeRange.start, allTimeRange.end)
      : Interval.invalid("Invalid interval");

  $: localUserPreferences = initLocalUserPreferenceStore($exploreName);

  $: baseTimeRange = selectedTimeRange?.start &&
    selectedTimeRange?.end && {
      name: selectedTimeRange?.name,
      start: selectedTimeRange.start,
      end: selectedTimeRange.end,
    };

  function handleMeasureFilterApply(
    dimension: string,
    measureName: string,
    oldDimension: string,
    filter: MeasureFilterEntry,
  ) {
    if (oldDimension && oldDimension !== dimension) {
      removeMeasureFilter(oldDimension, measureName);
    }
    setMeasureFilter(dimension, filter);
  }

  function onPan(direction: "left" | "right") {
    const panRange = $getNewPanRange(direction);
    if (!panRange) return;
    const { start, end } = panRange;

    const timeRange = {
      name: CUSTOM_TIME_RANGE_ALIAS,
      start: start,
      end: end,
    };

    const comparisonTimeRange = {
      name: TimeComparisonOption.CONTIGUOUS,
    } as DashboardTimeControls; // FIXME wrong typecasting across application

    if (!activeTimeGrain) return;
    metricsExplorerStore.selectTimeRange(
      $exploreName,
      timeRange as TimeRange,
      activeTimeGrain,
      comparisonTimeRange,
      metricsViewSpec,
    );
  }

  function onSelectRange(name: string) {
    if (!allTimeRange?.end) {
      return;
    }

    if (name === ALL_TIME_RANGE_ALIAS) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        allTimeRange,
        "TIME_GRAIN_DAY",
        undefined,
      );
      return;
    }

    const includesTimeZoneOffset = name.includes("@");

    if (includesTimeZoneOffset) {
      const timeZone = name.match(/@ {(.*)}/)?.[1];

      if (timeZone) metricsExplorerStore.setTimeZone($exploreName, timeZone);
    }

    const interval = deriveInterval(
      name,
      DateTime.fromJSDate(allTimeRange.end),
    );

    if (interval?.isValid) {
      const validInterval = interval as Interval<true>;
      const baseTimeRange: TimeRange = {
        // Temporary fix for custom syntax
        name: name as TimeRangePreset,
        start: validInterval.start.toJSDate(),
        end: validInterval.end.toJSDate(),
      };

      selectRange(baseTimeRange);
    }
  }

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    /** we should only reset the comparison range when the user has explicitly chosen a new
     * time range. Otherwise, the current comparison state should continue to be the
     * source of truth.
     */
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) {
    metricsExplorerStore.selectTimeRange(
      $exploreName,
      timeRange,
      timeGrain,
      comparisonTimeRange,
      metricsViewSpec,
    );
  }

  function selectRange(range: TimeRange) {
    const defaultTimeGrain = getDefaultTimeGrain(range.start, range.end).grain;

    // Get valid option for the new time range
    const validComparison =
      allTimeRange &&
      getValidComparisonOption(
        exploreSpec.timeRanges,
        range,
        $exploreState.selectedComparisonTimeRange?.name as
          | TimeComparisonOption
          | undefined,
        allTimeRange,
      );

    makeTimeSeriesTimeRangeAndUpdateAppState(range, defaultTimeGrain, {
      name: validComparison,
    } as DashboardTimeControls);
  }

  function onSelectTimeZone(timeZone: string) {
    if (!interval.isValid) return;

    if (selectedRangeAlias === TimeRangePreset.CUSTOM) {
      selectRange({
        name: TimeRangePreset.CUSTOM,
        start: interval.start
          ?.setZone(timeZone, { keepLocalTime: true })
          .toJSDate(),
        end: interval.end
          ?.setZone(timeZone, { keepLocalTime: true })
          .toJSDate(),
      });
    }

    metricsExplorerStore.setTimeZone($exploreName, timeZone);
    localUserPreferences.set({ timeZone });
  }

  function onTimeGrainSelect(timeGrain: V1TimeGrain) {
    if (baseTimeRange) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        baseTimeRange,
        timeGrain,
        $dashboardStore?.selectedComparisonTimeRange,
      );
    }
  }
</script>

<div class="flex flex-col gap-y-2 size-full">
  {#if hasTimeSeries}
    <div class="flex flex-row flex-wrap gap-x-2 gap-y-1.5 items-center">
      <Calendar size="16px" />
      {#if allTimeRange?.start && allTimeRange?.end}
        <SuperPill
          {allTimeRange}
          {selectedRangeAlias}
          showPivot={$showPivot}
          {minTimeGrain}
          {defaultTimeRange}
          {availableTimeZones}
          {timeRanges}
          complete={false}
          {interval}
          {timeStart}
          {timeEnd}
          {activeTimeGrain}
          {activeTimeZone}
          canPanLeft={$canPanLeft}
          canPanRight={$canPanRight}
          showPan
          {showDefaultItem}
          applyRange={selectRange}
          {onSelectRange}
          {onTimeGrainSelect}
          {onSelectTimeZone}
          {onPan}
        />
        <ComparisonPill
          {allTimeRange}
          {selectedTimeRange}
          showTimeComparison={!!showTimeComparison}
          {selectedComparisonTimeRange}
        />
      {/if}
    </div>
  {/if}

  <div class="relative flex flex-row gap-x-2 gap-y-2 items-start">
    {#if !readOnly}
      <Filter size="16px" className="ui-copy-icon flex-none mt-[5px]" />
    {/if}
    <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
      {#if isComplexFilter}
        <AdvancedFilter advancedFilter={$dashboardStore.whereFilter} />
      {:else if !allDimensionFilters.length && !allMeasureFilters.length}
        <div
          in:fly={{ duration: 200, x: 8 }}
          class="ui-copy-disabled grid ml-1 items-center"
          style:min-height={ROW_HEIGHT}
        >
          No filters selected
        </div>
      {:else}
        {#each allDimensionFilters as { name, label, selectedValues } (name)}
          {@const dimension = dimensions.find(
            (d) => d.name === name || d.column === name,
          )}
          {@const dimensionName = dimension?.name || dimension?.column}
          <div animate:flip={{ duration: 200 }}>
            {#if dimensionName}
              <DimensionFilter
                metricsViewNames={[metricsViewName]}
                {readOnly}
                {name}
                {label}
                {selectedValues}
                {timeStart}
                {timeEnd}
                {timeControlsReady}
                excludeMode={$isFilterExcludeMode(name)}
                onRemove={() => removeDimensionFilter(name)}
                onToggleFilterMode={() => toggleDimensionFilterMode(name)}
                onSelect={(value) =>
                  toggleDimensionValueSelection(name, value, true)}
              />
            {/if}
          </div>
        {/each}
        {#each allMeasureFilters as { name, label, dimensionName, filter } (name)}
          <div animate:flip={{ duration: 200 }}>
            <MeasureFilter
              allDimensions={dimensions}
              {name}
              {label}
              {dimensionName}
              {filter}
              onRemove={() => removeMeasureFilter(dimensionName, name)}
              onApply={({ dimension, oldDimension, filter }) =>
                handleMeasureFilterApply(dimension, name, oldDimension, filter)}
            />
          </div>
        {/each}
      {/if}

      {#if !readOnly}
        <FilterButton
          allDimensions={dimensions}
          filteredSimpleMeasures={$filteredSimpleMeasures()}
          dimensionHasFilter={$dimensionHasFilter}
          measureHasFilter={$measureHasFilter}
          {setTemporaryFilterName}
        />
        <!-- if filters are present, place a chip at the end of the flex container 
      that enables clearing all filters -->
        {#if hasFilters}
          <Button type="text" on:click={clearAllFilters}>Clear filters</Button>
        {/if}
      {/if}
    </div>
  </div>
</div>
