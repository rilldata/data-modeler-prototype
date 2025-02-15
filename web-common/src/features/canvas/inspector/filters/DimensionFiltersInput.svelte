<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import FilterButton from "@rilldata/web-common/features/dashboards/filters/FilterButton.svelte";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { isExpressionUnsupported } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import { flip } from "svelte/animate";

  export let metricsView: string;
  export let selectedComponentName: string;
  export let id: string;
  export let filter: string;
  export let onChange: (filter: string) => void = () => {};

  const {
    canvasEntity: {
      useComponent,
      spec: { getDimensionsForMetricView, getSimpleMeasuresForMetricView },
    },
  } = getCanvasStateManagers();

  let filterToggle = false;

  $: showFilter = !!filter || filterToggle;
  $: componentStore = useComponent(selectedComponentName);

  $: allDimensions = getDimensionsForMetricView(metricsView);
  $: allSimpleMeasures = getSimpleMeasuresForMetricView(metricsView);

  $: ({
    whereFilter,
    toggleDimensionValueSelection,
    removeDimensionFilter,
    toggleDimensionFilterMode,
    setMeasureFilter,
    removeMeasureFilter,
    setTemporaryFilterName,
    clearAllFilters,
    filterText,
    dimensionHasFilter,
    getDimensionFilterItems,
    getAllDimensionFilterItems,
    isFilterExcludeMode,
    getMeasureFilterItems,
    getAllMeasureFilterItems,
    measureHasFilter,
  } = componentStore.localFilters);

  $: dimensionIdMap = getMapFromArray(
    $allDimensions,
    (dimension) => (dimension.name || dimension.column) as string,
  );

  $: measureIdMap = getMapFromArray(
    $allSimpleMeasures,
    (m) => m.name as string,
  );

  $: if ((filter ?? "") !== ($filterText ?? "")) {
    onChange($filterText);
  }

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
</script>

<div class="flex flex-col gap-y-2 pt-1">
  <div class="flex justify-between">
    <InputLabel
      capitalize={false}
      small
      label="Local filters"
      {id}
      faint={!showFilter}
    />
    <Switch
      checked={showFilter}
      on:click={() => {
        if (filter) {
          filterToggle = false;
          onChange("");
        } else {
          filterToggle = true;
        }
      }}
      small
    />
  </div>
  <div class="text-gray-500">
    {#if showFilter}
      Overriding inherited filters from canvas.
    {:else}
      Overrides inherited filters from canvas when ON.
    {/if}
  </div>
  {#if showFilter}
    <div class="flex justify-between gap-x-2">
      <InputLabel small label="Filters" {id} />

      <FilterButton
        allDimensions={$allDimensions}
        filteredSimpleMeasures={$allSimpleMeasures}
        dimensionHasFilter={$dimensionHasFilter}
        measureHasFilter={$measureHasFilter}
        {setTemporaryFilterName}
        addBorder={false}
      />
    </div>

    <div class="relative flex flex-col gap-x-2 gap-y-2 items-start">
      <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
        {#if isComplexFilter}
          <AdvancedFilter advancedFilter={$whereFilter} />
        {:else if allDimensionFilters.length || allMeasureFilters.length}
          {#each allDimensionFilters as { name, label, selectedValues } (name)}
            {@const dimension = $allDimensions.find(
              (d) => d.name === name || d.column === name,
            )}
            {@const dimensionName = dimension?.name || dimension?.column}
            <div animate:flip={{ duration: 200 }}>
              {#if dimensionName}
                <DimensionFilter
                  metricsViewNames={[metricsView]}
                  readOnly={false}
                  smallChip
                  {name}
                  {label}
                  {selectedValues}
                  timeStart={new Date(0).toISOString()}
                  timeEnd={new Date().toISOString()}
                  timeControlsReady
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
                allDimensions={$allDimensions}
                {name}
                {label}
                {dimensionName}
                {filter}
                onRemove={() => removeMeasureFilter(dimensionName, name)}
                onApply={({ dimension, oldDimension, filter }) =>
                  handleMeasureFilterApply(
                    dimension,
                    name,
                    oldDimension,
                    filter,
                  )}
              />
            </div>
          {/each}
        {/if}
      </div>
      <div class="ml-auto">
        {#if hasFilters}
          <Button type="text" on:click={clearAllFilters}>Clear filters</Button>
        {/if}
      </div>
    </div>
  {/if}
</div>
