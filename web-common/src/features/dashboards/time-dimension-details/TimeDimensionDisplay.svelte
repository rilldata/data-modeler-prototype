<script lang="ts">
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";

  import {
    SortDirection,
    SortType,
  } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { timeFormat } from "d3-time-format";
  import TDDHeader from "./TDDHeader.svelte";
  import TDDTable from "./TDDTable.svelte";
  import {
    chartInteractionColumn,
    tableInteractionStore,
    useTimeDimensionDataStore,
  } from "./time-dimension-data-store";
  import type { TDDComparison, TableData } from "./types";
  import { colorGetter } from "../filters/colorGetter";

  export let metricViewName: string;

  const stateManagers = getStateManagers();
  const {
    metricsViewName,
    dashboardStore,
    selectors: {
      dimensionFilters: { unselectedDimensionValues, selectedDimensionValues },
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

  $: metricsView = useMetricsView(stateManagers);
  $: dimensionName = $dashboardStore?.selectedComparisonDimension ?? "";
  $: values = $selectedDimensionValues(dimensionName);

  $: timeGrain = $timeControlStore.selectedTimeRange?.interval;

  // Get labels for table headers
  $: measureLabel =
    $metricsView?.data?.measures?.find(
      (m) => m.name === $dashboardStore?.expandedMeasureName,
    )?.label ?? "";

  let dimensionLabel = "";
  $: if ($timeDimensionDataStore?.comparing === "dimension") {
    dimensionLabel =
      $metricsView?.data?.dimensions?.find((d) => d.name === dimensionName)
        ?.label ?? "";
  } else if ($timeDimensionDataStore?.comparing === "time") {
    dimensionLabel = "Time";
  } else if ($timeDimensionDataStore?.comparing === "none") {
    dimensionLabel = "No Comparison";
  }

  // Create a copy of the data to avoid flashing of table in transient states
  let timeDimensionDataCopy: TableData;
  let comparisonCopy: TDDComparison | undefined;
  $: if (
    $timeDimensionDataStore?.data &&
    $timeDimensionDataStore?.data?.columnHeaderData
  ) {
    comparisonCopy = $timeDimensionDataStore?.comparing;
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

  $: markerColors =
    values.map((value) =>
      colorGetter.get($metricsViewName, dimensionName, value),
    ) ?? [];

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

      notifications.send({
        message: `Removed ${rowHeaderLabels.length} items from filter`,
      });
      return;
    } else {
      const newValuesSelected = $unselectedDimensionValues(
        dimensionName,
        rowHeaderLabels,
      );
      selectItemsInFilter(dimensionName, rowHeaderLabels as (string | null)[]);
      notifications.send({
        message: `Added ${newValuesSelected.length} items to filter`,
      });
    }
  }

  function togglePin() {
    const pinIndex = $dashboardStore?.pinIndex;
    let newPinIndex = -1;

    // Pin if some selected items are not pinned yet
    if (pinIndex > -1 && pinIndex < formattedData?.selectedValues?.length - 1) {
      newPinIndex = formattedData?.selectedValues?.length - 1;
    }
    // Pin if no items are pinned yet
    else if (pinIndex === -1) {
      newPinIndex = formattedData?.selectedValues?.length - 1;
    }
    metricsExplorerStore.setPinIndex(metricViewName, newPinIndex);
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
</script>

<div class="h-full w-full flex flex-col">
  <TDDHeader
    {areAllTableRowsSelected}
    comparing={$timeDimensionDataStore?.comparing}
    {dimensionName}
    isFetching={!$timeDimensionDataStore?.data?.columnHeaderData}
    isRowsEmpty={!rowHeaderLabels.length}
    {metricViewName}
    on:search={(e) => {
      metricsExplorerStore.setSearchText(metricViewName, e.detail);
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
  {:else if formattedData && comparisonCopy}
    <TDDTable
      {markerColors}
      {excludeMode}
      {dimensionLabel}
      {measureLabel}
      sortDirection={$dashboardStore.sortDirection === SortDirection.ASCENDING}
      sortType={$dashboardStore.dashboardSortType}
      comparing={comparisonCopy}
      {timeFormatter}
      tableData={formattedData}
      highlightedCol={$chartInteractionColumn?.hover}
      pinIndex={$dashboardStore?.pinIndex}
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
      on:highlight={highlightCell}
    />
  {/if}

  {#if $timeDimensionDataStore?.comparing === "none"}
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
  {/if}
</div>

<svelte:window on:keydown={handleKeyDown} />
