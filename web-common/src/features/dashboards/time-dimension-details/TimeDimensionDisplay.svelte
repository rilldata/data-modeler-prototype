<script lang="ts">
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    SortDirection,
    SortType,
  } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import { timeFormat } from "d3-time-format";
  import { onDestroy } from "svelte";
  import TDDHeader from "./TDDHeader.svelte";
  import TDDTable from "./TDDTable.svelte";
  import {
    chartInteractionColumn,
    tableInteractionStore,
    prepareDimensionData,
    prepareTimeData,
  } from "./time-dimension-data-store";
  import type { TableData } from "./types";
  import {
    DEFAULT_TIME_RANGES,
    TIME_COMPARISON,
    TIME_GRAIN,
  } from "@rilldata/web-common/lib/time/config";
  import type {
    MetricsViewSpecDimensionV2,
    MetricsViewSpecMeasureV2,
  } from "@rilldata/web-common/runtime-client";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
  import { selectedDimensionValues } from "../state-managers/selectors/dimension-filters";
  import { useDimensionTableData } from "./time-dimension-data-store";
  import type { TimeSeriesDatum } from "../time-series/timeseries-data-store";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";

  export let exploreName: string;
  export let formattedTimeSeriesData: TimeSeriesDatum[];
  export let primaryTotal: number;
  export let unfilteredTotal: number;
  export let comparisonTotal: number | undefined;
  export let measure: MetricsViewSpecMeasureV2;
  export let comparisonDimension: MetricsViewSpecDimensionV2 | undefined;
  export let showTimeComparison: boolean;
  export let error: HTTPError | null;

  const stateManagers = getStateManagers();
  const timeControlsStore = useTimeControlStore(stateManagers);
  const tableDimensionData = useDimensionTableData(stateManagers);

  const {
    dashboardStore,
    selectors: {
      dimensionFilters: { unselectedDimensionValues, isFilterExcludeMode },
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

  const timeControlStore = useTimeControlStore(stateManagers);

  let timeDimensionData: TableData | undefined;
  let highlitedRowIndex: number | undefined;

  $: ({ selectedTimeRange, selectedComparisonTimeRange } = $timeControlsStore);
  $: dimensionName = comparisonDimension?.name;
  $: dimensionLabel = comparisonDimension
    ? comparisonDimension.label
    : showTimeComparison
      ? "Time"
      : "No Comparison";

  $: pinIndex = $dashboardStore?.tdd.pinIndex;

  $: timeGrain = $timeControlStore.selectedTimeRange?.interval;

  $: measureLabel = measure?.label ?? "";

  $: isAllTime = selectedTimeRange?.name === TimeRangePreset.ALL_TIME;

  $: if (comparisonDimension && dimensionName && $tableDimensionData?.length) {
    const selectedValues = selectedDimensionValues({
      dashboard: $dashboardStore,
    })(dimensionName);
    timeDimensionData = prepareDimensionData(
      formattedTimeSeriesData,
      $tableDimensionData,
      primaryTotal,
      unfilteredTotal,
      measure,
      selectedValues,
      isAllTime,
      pinIndex,
    );
  } else if (!comparisonDimension && !dimensionName) {
    const currentRange = selectedTimeRange?.name;
    let currentLabel = "Custom Range";
    if (currentRange && currentRange in DEFAULT_TIME_RANGES)
      currentLabel = DEFAULT_TIME_RANGES[currentRange].label;
    const comparisonRange = selectedComparisonTimeRange?.name;
    let comparisonLabel = "Custom Range";
    if (comparisonRange && comparisonRange in TIME_COMPARISON)
      comparisonLabel = TIME_COMPARISON[comparisonRange].label;
    timeDimensionData = prepareTimeData(
      formattedTimeSeriesData,
      primaryTotal,
      comparisonTotal,
      currentLabel,
      comparisonLabel,
      measure,
      showTimeComparison,
      isAllTime,
    );
  }

  $: rowHeaderLabels =
    timeDimensionData?.rowHeaderData?.slice(1)?.map((row) => row[0]?.value) ??
    [];
  $: areAllTableRowsSelected = rowHeaderLabels?.every(
    (val) =>
      val !== undefined && timeDimensionData?.selectedValues?.includes(val),
  );

  $: columnHeaders = timeDimensionData?.columnHeaderData?.flat();

  $: if (timeDimensionData?.rowCount) {
    highlitedRowIndex = undefined;
    timeDimensionData.rowHeaderData.forEach((row, index) => {
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

    const dimensionValue = timeDimensionData?.rowHeaderData[y]?.[0]?.value;
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

  function toggleFilter(e: CustomEvent<string>) {
    if (dimensionName) toggleDimensionValueSelection(dimensionName, e.detail);
  }

  function toggleAllSearchItems() {
    const headerHasUndefined = rowHeaderLabels.some(
      (label) => label === undefined,
    );

    if (headerHasUndefined || !dimensionName) return;

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

    const length = timeDimensionData?.selectedValues?.length ?? 0;

    // Pin if some selected items are not pinned yet
    if (pinIndex > -1 && pinIndex < length - 1) {
      newPinIndex = length - 1;
    }
    // Pin if no items are pinned yet
    else if (pinIndex === -1) {
      newPinIndex = length - 1;
    }
    metricsExplorerStore.setPinIndex(exploreName, newPinIndex);
  }

  function handleKeyDown(
    e: KeyboardEvent & {
      currentTarget: EventTarget & Window;
    },
  ) {
    if (!comparisonDimension) return;
    // Select all items on Meta+A
    if ((e.ctrlKey || e.metaKey) && e.key === "a") {
      if (e.target instanceof HTMLElement && e.target.tagName === "INPUT")
        return;
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
  $: excludeMode = $isFilterExcludeMode(dimensionName ?? "");
</script>

<div class="h-full w-full flex flex-col">
  <TDDHeader
    {excludeMode}
    {areAllTableRowsSelected}
    expandedMeasureName={measure.name ?? ""}
    {dimensionName}
    isFetching={!timeDimensionData?.columnHeaderData}
    isRowsEmpty={!rowHeaderLabels.length}
    {exploreName}
    on:search={(e) => {
      metricsExplorerStore.setSearchText(exploreName, e.detail);
    }}
    on:toggle-all-search-items={() => toggleAllSearchItems()}
  />

  {#if error}
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
  {:else if timeDimensionData && measure}
    <TDDTable
      {measure}
      {excludeMode}
      {dimensionLabel}
      {measureLabel}
      sortDirection={$dashboardStore.sortDirection === SortDirection.ASCENDING}
      sortType={$dashboardStore.dashboardSortType}
      comparing={comparisonDimension
        ? "dimension"
        : showTimeComparison
          ? "time"
          : "none"}
      {timeFormatter}
      tableData={timeDimensionData}
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

  {#if !comparisonDimension && !showTimeComparison}
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
  {:else if comparisonDimension && timeDimensionData?.rowCount === 1}
    <div class="w-full h-full">
      <div class="flex flex-col items-center h-full text-sm">
        <div class="text-gray-600">No search results to show</div>
      </div>
    </div>
  {/if}
</div>

<svelte:window on:keydown={handleKeyDown} />
