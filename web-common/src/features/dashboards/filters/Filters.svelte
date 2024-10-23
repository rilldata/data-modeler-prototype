<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import { getMapFromArray } from "@rilldata/utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { useModelHasTimeSeries } from "../selectors";
  import { getStateManagers } from "../state-managers/state-managers";
  import ComparisonPill from "../time-controls/comparison-pill/ComparisonPill.svelte";
  import SuperPill from "../time-controls/super-pill/SuperPill.svelte";
  import { useTimeControlStore } from "../time-controls/time-control-store";
  import DimensionFilter from "./dimension-filters/DimensionFilter.svelte";
  import FilterButton from "./FilterButton.svelte";
  import TimeGrainSelector from "../time-controls/TimeGrainSelector.svelte";

  export let readOnly = false;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  const StateManagers = getStateManagers();
  const {
    metricsViewName,
    exploreName,
    actions: {
      dimensionsFilter: {
        toggleDimensionValueSelection,
        removeDimensionFilter,
        toggleDimensionFilterMode,
      },
      measuresFilter: { setMeasureFilter, removeMeasureFilter },
      filters: { clearAllFilters },
    },
    selectors: {
      dimensions: { allDimensions },
      dimensionFilters: {
        getDimensionFilterItems,
        getAllDimensionFilterItems,
        isFilterExcludeMode,
      },
      measures: { allMeasures },
      measureFilters: { getMeasureFilterItems, getAllMeasureFilterItems },
      pivot: { showPivot },
    },
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(StateManagers);

  $: ({
    selectedTimeRange,
    allTimeRange,
    showTimeComparison,
    selectedComparisonTimeRange,
    minTimeGrain,
  } = $timeControlsStore);

  $: ({ instanceId } = $runtime);

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
  $: metricTimeSeries = useModelHasTimeSeries(instanceId, $metricsViewName);
  $: hasTimeSeries = $metricTimeSeries.data;

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
</script>

<div class="flex flex-col gap-y-2 size-full">
  {#if hasTimeSeries}
    <div class="flex flex-row flex-wrap gap-x-2 gap-y-1.5 items-center">
      <Calendar size="16px" />
      {#if allTimeRange?.start && allTimeRange?.end}
        <SuperPill {allTimeRange} {selectedTimeRange} />
        <ComparisonPill
          {allTimeRange}
          {selectedTimeRange}
          showTimeComparison={!!showTimeComparison}
          {selectedComparisonTimeRange}
        />
        {#if !$showPivot && minTimeGrain}
          <TimeGrainSelector exploreName={$exploreName} />
        {/if}
      {/if}
    </div>
  {/if}

  <div class="relative flex flex-row gap-x-2 gap-y-2 items-start">
    {#if !readOnly}
      <Filter size="16px" className="ui-copy-icon flex-none mt-[5px]" />
    {/if}
    <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
      {#if !allDimensionFilters.length && !allMeasureFilters.length}
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
                {readOnly}
                {name}
                {label}
                {selectedValues}
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
              {name}
              {label}
              {dimensionName}
              {filter}
              on:remove={() => removeMeasureFilter(dimensionName, name)}
              on:apply={({ detail: { dimension, oldDimension, filter } }) =>
                handleMeasureFilterApply(dimension, name, oldDimension, filter)}
            />
          </div>
        {/each}
      {/if}

      {#if !readOnly}
        <FilterButton />
        <!-- if filters are present, place a chip at the end of the flex container 
      that enables clearing all filters -->
        {#if hasFilters}
          <Button type="text" on:click={clearAllFilters}>Clear filters</Button>
        {/if}
      {/if}
    </div>
  </div>
</div>
