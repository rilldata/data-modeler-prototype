<script lang="ts">
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import { eventBus } from "@rilldata/events";
  import {
    SortDirection,
    SortType,
  } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { debounce } from "@rilldata/utils";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { timeFormat } from "d3-time-format";
  import { onDestroy } from "svelte";
  import TDDHeader from "./TDDHeader.svelte";
  import TDDTable from "./TDDTable.svelte";
  import {
    chartInteractionColumn,
    tableInteractionStore,
    useTimeDimensionDataStore,
  } from "./time-dimension-data-store";
  import type { TDDComparison, TableData } from "./types";

  export let exploreName: string;
  export let expandedMeasureName: string;

  const stateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      dimensions: { allDimensions },
      dimensionFilters: { unselectedDimensionValues },
      measures: { allMeasures },
    },
    actions: {
      dimensionsFilter: {
        toggleDimensionValueSelection,
        selectItemsInFilter,
        deselectItemsInFilter,
      },
      sorting: { toggleSort },
    },
  } = getStateManagers();

  const timeDimensionDataStore = useTimeDimensionDataStore(stateManagers);
  const timeControlStore = useTimeControlStore(stateManagers);

  $: dimensionName = $dashboardStore?.selectedComparisonDimension ?? "";
  $: comparing = $timeDimensionDataStore?.comparing;

  $: pinIndex = $dashboardStore?.tdd.pinIndex;

  $: timeGrain = $timeControlStore.selectedTimeRange?.interval;

  $: measure = $allMeasures.find((m) => m.name === expandedMeasureName);

  $: measureLabel = measure?.label ?? "";

  let dimensionLabel = "";
  $: if (comparing === "dimension") {
    dimensionLabel =
      $allDimensions.find((d) => d.name === dimensionName)?.label ?? "";
  } else if (comparing === "time") {
    dimensionLabel = "Time";
  } else if (comparing === "none") {
    dimensionLabel = "No Comparison";
  }

  // Create a copy of the data to avoid flashing of table in transient states
  let timeDimensionDataCopy: TableData;
  let comparisonCopy: TDDComparison | undefined;
  $: if (
    $timeDimensionDataStore?.data &&
    $timeDimensionDataStore?.data?.columnHeaderData
  ) {
    comparisonCopy = comparing;
    timeDimensionDataCopy = $timeDimensionDataStore.data;
  }
  $: formattedData = timeDimensionDataCopy;
  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

  $: rowHeaderLabels =
    formattedData?.rowHeaderData?.slice(1)?.map((row) => row[0]?.value) ?? [];

  $: areAllTableRowsSelected = rowHeaderLabels?.every(
    (val) => val !== undefined && formattedData?.selectedValues?.includes(val),
  );

  $: columnHeaders = formattedData?.columnHeaderData?.flat();

  let highlitedRowIndex: number | undefined;
  $: if (formattedData?.rowCount) {
    highlitedRowIndex = undefined;
    formattedData.rowHeaderData.forEach((row, index) => {
      if (row[0]?.value === $chartInteractionColumn?.yHover) {
        highlitedRowIndex = index;
      }
    });
  }

  // Create a time formatter for the column headers
  $: timeFormatter = timeFormat(
    timeGrain ? TIME_GRAIN[timeGrain].d3format : "%H:%M",
  ) as (d: Date) => string;

  function highlightCell(e) {
    const { x, y } = e.detail;

    const dimensionValue = formattedData?.rowHeaderData[y]?.[0]?.value;
    let time: Date | undefined = undefined;

    const colHeader = columnHeaders?.[x]?.value;
    if (colHeader) {
      time = new Date(colHeader);
    }

    tableInteractionStore.set({
      dimensionValue,
      time: time,
    });
  }

  const debounceHighlightCell = debounce(highlightCell, 50);

  function toggleFilter(e) {
    toggleDimensionValueSelection(dimensionName, e.detail);
  }

  function toggleAllSearchItems() {
    const headerHasUndefined = rowHeaderLabels.some(
      (label) => label === undefined,
    );

    if (headerHasUndefined) return;

    if (areAllTableRowsSelected) {
      deselectItemsInFilter(
        dimensionName,
        rowHeaderLabels as (string | null)[],
      );

      eventBus.emit("notification", {
        message: `Removed ${rowHeaderLabels.length} items from filter`,
      });
      return;
    } else {
      const newValuesSelected = $unselectedDimensionValues(
        dimensionName,
        rowHeaderLabels,
      );
      selectItemsInFilter(dimensionName, rowHeaderLabels as (string | null)[]);
      eventBus.emit("notification", {
        message: `Added ${newValuesSelected.length} items to filter`,
      });
    }
  }

  function togglePin() {
    let newPinIndex = -1;

    // Pin if some selected items are not pinned yet
    if (pinIndex > -1 && pinIndex < formattedData?.selectedValues?.length - 1) {
      newPinIndex = formattedData?.selectedValues?.length - 1;
    }
    // Pin if no items are pinned yet
    else if (pinIndex === -1) {
      newPinIndex = formattedData?.selectedValues?.length - 1;
    }
    metricsExplorerStore.setPinIndex(exploreName, newPinIndex);
  }

  function handleKeyDown(e) {
    if (comparisonCopy !== "dimension") return;
    // Select all items on Meta+A
    if ((e.ctrlKey || e.metaKey) && e.key === "a") {
      if (e.target.tagName === "INPUT") return;
      e.preventDefault();
      if (areAllTableRowsSelected) return;
      toggleAllSearchItems();
    }
  }

  onDestroy(() => {
    tableInteractionStore.set({
      dimensionValue: undefined,
      time: undefined,
    });
  });
</script>

<div class="h-full w-full flex flex-col">
  <TDDHeader
    {areAllTableRowsSelected}
    {comparing}
    {expandedMeasureName}
    {dimensionName}
    isFetching={!$timeDimensionDataStore?.data?.columnHeaderData}
    isRowsEmpty={!rowHeaderLabels.length}
    {exploreName}
    on:search={(e) => {
      metricsExplorerStore.setSearchText(exploreName, e.detail);
    }}
    on:toggle-all-search-items={() => toggleAllSearchItems()}
  />

  {#if $timeDimensionDataStore?.isError}
    <div
      style:height="calc(100% - 200px)"
      class="w-full flex items-center justify-center text-sm"
    >
      <div class="text-center">
        <div class="text-red-600 mt-1 text-lg">
          We encountered an error while loading the data. Please try refreshing
          the page.
        </div>
        <div class="text-gray-600">
          If the issue persists, please contact us on <a
            target="_blank"
            rel="noopener noreferrer"
            href="http://bit.ly/3jg4IsF">Discord</a
          >.
        </div>
      </div>
    </div>
  {:else if formattedData && comparisonCopy && measure}
    <TDDTable
      {measure}
      {excludeMode}
      {dimensionLabel}
      {measureLabel}
      sortDirection={$dashboardStore.sortDirection === SortDirection.ASCENDING}
      sortType={$dashboardStore.dashboardSortType}
      comparing={comparisonCopy}
      {timeFormatter}
      tableData={formattedData}
      highlightedRow={highlitedRowIndex}
      highlightedCol={$chartInteractionColumn?.xHover}
      {pinIndex}
      scrubPos={{
        start: $chartInteractionColumn?.scrubStart,
        end: $chartInteractionColumn?.scrubEnd,
      }}
      on:toggle-pin={togglePin}
      on:toggle-filter={toggleFilter}
      on:toggle-sort={(e) => {
        toggleSort(
          e.detail === "dimension" ? SortType.DIMENSION : SortType.VALUE,
        );
      }}
      on:highlight={debounceHighlightCell}
    />
  {/if}

  {#if comparing === "none"}
    <!-- Get height by subtracting table and header heights -->
    <div class="w-full" style:height="calc(100% - 200px)">
      <div class="flex flex-col items-center justify-center h-full text-sm">
        <Compare size="32px" />
        <div class="font-semibold text-gray-600 mt-1">
          No comparison dimension selected
        </div>
        <div class="text-gray-600">
          To see more values, select a comparison dimension above.
        </div>
      </div>
    </div>
  {:else if comparing === "dimension" && formattedData?.rowCount === 1}
    <div class="w-full h-full">
      <div class="flex flex-col items-center h-full text-sm">
        <div class="text-gray-600">No search results to show</div>
      </div>
    </div>
  {/if}
</div>

<svelte:window on:keydown={handleKeyDown} />
