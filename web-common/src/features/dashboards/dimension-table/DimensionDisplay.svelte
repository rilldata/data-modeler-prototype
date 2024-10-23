<script lang="ts">
  /**
   * DimensionDisplay.svelte
   * -------------------------
   * Create a table with the selected dimension and measures
   * to be displayed in explore
   */
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { createQueryServiceMetricsViewAggregation } from "@rilldata/web-common/runtime-client";
  import { getDimensionFilterWithSearch } from "./dimension-table-utils";
  import DimensionHeader from "./DimensionHeader.svelte";
  import DimensionTable from "./DimensionTable.svelte";
  import { eventBus } from "@rilldata/events";
  import { metricsExplorerStore } from "../stores/dashboard-stores";

  const stateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      dashboardQueries: {
        dimensionTableSortedQueryBody,
        dimensionTableTotalQueryBody,
      },
      comparison: { isBeingCompared },
      dimensions: { dimensionTableDimName },
      dimensionFilters: { unselectedDimensionValues },
      dimensionTable: {
        virtualizedTableColumns,
        selectedDimensionValueNames,
        prepareDimTableRows,
      },
      activeMeasure: { activeMeasureName },
    },
    actions: {
      dimensionsFilter: {
        toggleDimensionValueSelection,
        selectItemsInFilter,
        deselectItemsInFilter,
      },
    },
    metricsViewName,
    exploreName,
    runtime,
  } = stateManagers;

  // cast is safe because dimensionTableDimName must be defined
  // for the dimension table to be open
  $: dimensionName = $dimensionTableDimName as string;

  let searchText = "";

  $: instanceId = $runtime.instanceId;

  const timeControlsStore = useTimeControlStore(stateManagers);

  $: filterSet = getDimensionFilterWithSearch(
    $dashboardStore?.whereFilter,
    searchText,
    dimensionName,
  );

  $: totalsQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    $metricsViewName,
    $dimensionTableTotalQueryBody,
    {
      query: {
        enabled: $timeControlsStore.ready,
      },
    },
  );

  $: unfilteredTotal = $totalsQuery?.data?.data?.[0]?.[$activeMeasureName] ?? 0;

  $: columns = $virtualizedTableColumns($totalsQuery);

  $: sortedQuery = createQueryServiceMetricsViewAggregation(
    $runtime.instanceId,
    $metricsViewName,
    $dimensionTableSortedQueryBody,
    {
      query: {
        enabled: $timeControlsStore.ready && !!filterSet,
      },
    },
  );

  $: tableRows = $prepareDimTableRows($sortedQuery, unfilteredTotal);

  $: areAllTableRowsSelected = tableRows.every((row) =>
    $selectedDimensionValueNames.includes(row[dimensionName] as string),
  );

  function onSelectItem(event) {
    const label = tableRows[event.detail.index][dimensionName] as string;
    toggleDimensionValueSelection(
      dimensionName,
      label,
      false,
      event.detail.meta,
    );
  }

  function toggleComparisonDimension(dimensionName, isBeingCompared) {
    metricsExplorerStore.setComparisonDimension(
      $exploreName,
      isBeingCompared ? undefined : dimensionName,
    );
  }

  function toggleAllSearchItems() {
    const labels = tableRows.map((row) => row[dimensionName] as string);

    if (areAllTableRowsSelected) {
      deselectItemsInFilter(dimensionName, labels);

      eventBus.emit("notification", {
        message: `Removed ${labels.length} items from filter`,
      });
      return;
    } else {
      const newValuesSelected = $unselectedDimensionValues(
        dimensionName,
        labels,
      );
      selectItemsInFilter(dimensionName, labels);
      eventBus.emit("notification", {
        message: `Added ${newValuesSelected.length} items to filter`,
      });
    }
  }

  function handleKeyDown(e) {
    // Select all items on Meta+A
    if ((e.ctrlKey || e.metaKey) && e.key === "a") {
      if (e.target.tagName === "INPUT") return;
      e.preventDefault();
      if (areAllTableRowsSelected) return;
      toggleAllSearchItems();
    }
  }
</script>

{#if $sortedQuery}
  <div class="h-full flex flex-col w-full" style:min-width="365px">
    <div class="flex-none" style:height="50px">
      <DimensionHeader
        {dimensionName}
        {areAllTableRowsSelected}
        isRowsEmpty={!tableRows.length}
        isFetching={$sortedQuery?.isFetching}
        on:search={(event) => {
          searchText = event.detail;
        }}
        on:toggle-all-search-items={() => toggleAllSearchItems()}
      />
    </div>

    {#if tableRows && columns.length && dimensionName}
      <div class="grow" style="overflow-y: hidden;">
        <DimensionTable
          on:select-item={(event) => onSelectItem(event)}
          on:toggle-dimension-comparison={() =>
            toggleComparisonDimension(dimensionName, isBeingCompared)}
          isFetching={$sortedQuery?.isFetching}
          {dimensionName}
          {columns}
          selectedValues={$selectedDimensionValueNames}
          rows={tableRows}
        />
      </div>
    {/if}
  </div>
{/if}

<svelte:window on:keydown={handleKeyDown} />
