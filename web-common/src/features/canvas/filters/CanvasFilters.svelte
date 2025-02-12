<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import FilterButton from "@rilldata/web-common/features/dashboards/filters/FilterButton.svelte";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { getPanRangeForTimeRange } from "@rilldata/web-common/features/dashboards/state-managers/selectors/charts";
  import { isExpressionUnsupported } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import {
    ALL_TIME_RANGE_ALIAS,
    CUSTOM_TIME_RANGE_ALIAS,
    deriveInterval,
  } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
  import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
  import {
    TimeComparisonOption,
    TimeRangePreset,
    type DashboardTimeControls,
    type TimeRange,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";
  import { onDestroy } from "svelte";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import CanvasComparisonPill from "./CanvasComparisonPill.svelte";

  export let readOnly = false;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";
  const {
    canvasName,
    canvasEntity: {
      filters: {
        whereFilter,
        toggleDimensionValueSelection,
        removeDimensionFilter,
        toggleDimensionFilterMode,
        setMeasureFilter,
        removeMeasureFilter,
        setTemporaryFilterName,
        clearAllFilters,
        dimensionHasFilter,
        getDimensionFilterItems,
        getAllDimensionFilterItems,
        isFilterExcludeMode,
        getMeasureFilterItems,
        getAllMeasureFilterItems,
        measureHasFilter,
      },
      spec: { canvasSpec, allDimensions, allSimpleMeasures, metricViewNames },
      timeControls: {
        allTimeRange,
        timeRangeStateStore,
        comparisonRangeStateStore,
        selectedTimezone,
        minTimeGrain,
        selectTimeRange,
        setTimeZone,
        destroy,
      },
    },
  } = getCanvasStateManagers();

  let showDefaultItem = false;

  $: ({ selectedTimeRange, timeStart, timeEnd } = $timeRangeStateStore || {});

  $: selectedComparisonTimeRange =
    $comparisonRangeStateStore?.selectedComparisonTimeRange;

  $: baseTimeRange = selectedTimeRange?.start &&
    selectedTimeRange?.end && {
      name: selectedTimeRange?.name,
      start: selectedTimeRange.start,
      end: selectedTimeRange.end,
    };

  $: selectedRangeAlias = selectedTimeRange?.name;
  $: activeTimeGrain = selectedTimeRange?.interval;
  $: defaultTimeRange = $canvasSpec?.defaultPreset?.timeRange;
  $: availableTimeZones = $canvasSpec?.timeZones ?? [];
  $: timeRanges = $canvasSpec?.timeRanges ?? [];

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone($selectedTimezone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone($selectedTimezone),
      )
    : Interval.fromDateTimes($allTimeRange.start, $allTimeRange.end);

  $: localUserPreferences = initLocalUserPreferenceStore($canvasName);

  $: dimensionIdMap = getMapFromArray(
    $allDimensions,
    (dimension) => (dimension.name || dimension.column) as string,
  );

  $: measureIdMap = getMapFromArray(
    $allSimpleMeasures,
    (m) => m.name as string,
  );

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

  $: isComplexFilter = isExpressionUnsupported($whereFilter);

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
    const getPanRange = getPanRangeForTimeRange(
      selectedTimeRange,
      $selectedTimezone,
    );
    const panRange = getPanRange(direction);
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
    selectTimeRange(
      timeRange as TimeRange,
      activeTimeGrain,
      comparisonTimeRange,
    );
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
    selectTimeRange(timeRange, timeGrain, comparisonTimeRange);
  }

  function selectRange(range: TimeRange) {
    const defaultTimeGrain = getDefaultTimeGrain(range.start, range.end).grain;

    const comparisonOption = DEFAULT_TIME_RANGES[range.name as TimeRangePreset]
      ?.defaultComparison as TimeComparisonOption;

    // Get valid option for the new time range
    const validComparison = allTimeRange && comparisonOption;

    makeTimeSeriesTimeRangeAndUpdateAppState(range, defaultTimeGrain, {
      name: validComparison,
    } as DashboardTimeControls);
  }

  async function onSelectRange(name: string, syntax?: boolean) {
    if (!$allTimeRange?.end) {
      return;
    }

    if (name === ALL_TIME_RANGE_ALIAS) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        $allTimeRange,
        "TIME_GRAIN_DAY",
        undefined,
      );
      return;
    }

    const includesTimeZoneOffset = name.includes("@");

    if (includesTimeZoneOffset) {
      const timeZone = name.match(/@ {(.*)}/)?.[1];

      if (timeZone) setTimeZone(timeZone);
    }

    const interval = await deriveInterval(
      syntax && !includesTimeZoneOffset
        ? name + ` @ {${$selectedTimezone}}`
        : name,
      DateTime.fromJSDate($allTimeRange.end),
      $metricViewNames?.[0], // TODO: handle multiple metrics
    );

    if (interval.isValid) {
      const validInterval = interval as Interval<true>;
      const baseTimeRange: TimeRange = {
        // Temporary fix for custom syntax
        name: name,
        start: validInterval.start.toJSDate(),
        end: validInterval.end.toJSDate(),
      };

      selectRange(baseTimeRange);
    }
  }

  function onTimeGrainSelect(timeGrain: V1TimeGrain) {
    if (baseTimeRange) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        baseTimeRange,
        timeGrain,
        selectedComparisonTimeRange,
      );
    }
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

    setTimeZone(timeZone);
    localUserPreferences.set({ timeZone });
  }

  onDestroy(destroy);
</script>

<div class="flex flex-col gap-y-2 w-full h-20 justify-center">
  <div class="flex flex-row flex-wrap gap-x-2 gap-y-1.5 items-center ml-2">
    <Calendar size="16px" />
    <SuperPill
      context={$canvasName}
      allTimeRange={$allTimeRange}
      {selectedRangeAlias}
      showPivot={false}
      minTimeGrain={$minTimeGrain}
      {defaultTimeRange}
      {availableTimeZones}
      {timeRanges}
      complete={false}
      toggleComplete={() => {}}
      {interval}
      {timeStart}
      {timeEnd}
      {activeTimeGrain}
      activeTimeZone={$selectedTimezone}
      canPanLeft
      canPanRight
      showPan
      {showDefaultItem}
      applyRange={selectRange}
      {onSelectRange}
      {onTimeGrainSelect}
      {onSelectTimeZone}
      {onPan}
    />
    <CanvasComparisonPill
      allTimeRange={$allTimeRange}
      {selectedTimeRange}
      {selectedComparisonTimeRange}
    />
  </div>
  <div class="relative flex flex-row gap-x-2 gap-y-2 items-start ml-2">
    {#if !readOnly}
      <Filter size="16px" className="ui-copy-icon flex-none mt-[5px]" />
    {/if}
    <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
      {#if isComplexFilter}
        <AdvancedFilter advancedFilter={$whereFilter} />
      {:else if !allDimensionFilters.length && !allMeasureFilters.length}
        <div
          in:fly={{ duration: 200, x: 8 }}
          class="ui-copy-disabled grid ml-1 items-center"
          style:min-height={ROW_HEIGHT}
        >
          No filters selected
        </div>
      {:else}
        {#each allDimensionFilters as { name, label, selectedValues, metricsViewNames } (name)}
          {@const dimension = $allDimensions.find(
            (d) => d.name === name || d.column === name,
          )}
          {@const dimensionName = dimension?.name || dimension?.column}
          <div animate:flip={{ duration: 200 }}>
            {#if dimensionName && metricsViewNames?.length}
              <DimensionFilter
                {metricsViewNames}
                {readOnly}
                {name}
                {label}
                {selectedValues}
                {timeStart}
                {timeEnd}
                timeControlsReady={!!$timeRangeStateStore}
                excludeMode={$isFilterExcludeMode(name)}
                onRemove={() => removeDimensionFilter(name)}
                onToggleFilterMode={() => toggleDimensionFilterMode(name)}
                onSelect={(value) =>
                  toggleDimensionValueSelection(name, value, true)}
              />
            {/if}
          </div>
        {/each}
        {#each allMeasureFilters as { name, label, dimensionName, filter, dimensions: dimensionsForMeasure } (name)}
          <div animate:flip={{ duration: 200 }}>
            <MeasureFilter
              allDimensions={dimensionsForMeasure || $allDimensions}
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
          allDimensions={$allDimensions}
          filteredSimpleMeasures={$allSimpleMeasures}
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
