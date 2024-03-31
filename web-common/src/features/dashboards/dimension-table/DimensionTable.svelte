<!-- @component 
Creates a virtualized dimension table. This consists of four sub-components:
ColumnHeaders – sticky column headers. Utilizes the columnVirtualizer (for now).
TableCells – the cell contents.
-->
<script lang="ts">
  import ColumnHeaders from "@rilldata/web-common/components/virtualized-table/sections/ColumnHeaders.svelte";
  import TableCells from "@rilldata/web-common/components/virtualized-table/sections/TableCells.svelte";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { createEventDispatcher, setContext } from "svelte";
  import type { DimensionTableRow } from "./dimension-table-types";
  import {
    estimateColumnCharacterWidths,
    estimateColumnSizes,
  } from "./dimension-table-utils";
  import DimensionFilterGutter from "./DimensionFilterGutter.svelte";
  import { DIMENSION_TABLE_CONFIG as config } from "./DimensionTableConfig";
  import DimensionValueHeader from "./DimensionValueHeader.svelte";

  import { getStateManagers } from "../state-managers/state-managers";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";

  const dispatch = createEventDispatcher();

  export let rows: DimensionTableRow[];
  export let columns: VirtualizedTableColumns[];
  export let selectedValues: string[];

  export let dimensionName: string;
  export let isFetching: boolean;

  const {
    actions: { dimensionTable },
    selectors: {
      sorting: { sortMeasure },
      dimensions: { dimensionTableColumnName },
      dimensionFilters: { isFilterExcludeMode },
      comparison: { isBeingCompared: isBeingComparedReadable },
    },
  } = getStateManagers();

  $: excludeMode = $isFilterExcludeMode(dimensionName);
  $: isBeingCompared = $isBeingComparedReadable(dimensionName);

  /** the overscan values tell us how much to render off-screen. These may be set by the consumer
   * in certain circumstances. The tradeoff: the higher the overscan amount, the more DOM elements we have
   * to render on initial load.
   */
  export let rowOverscanAmount = 40;
  export let columnOverscanAmount = 5;

  let container: HTMLDivElement;

  let containerWidth: number;

  /** this is a perceived character width value, in pixels, when our monospace
   * font is 12px high. */
  const CHARACTER_LIMIT_FOR_WRAPPING = 9;
  const FILTER_COLUMN_WIDTH = config.indexWidth;

  $: selectedIndex = selectedValues.map((label) => {
    return rows.findIndex((row) => row[dimensionColumnName] === label);
  });

  let rowScrollOffset = 0;
  $: rowScrollOffset = $rowVirtualizer?.scrollOffset || 0;
  let colScrollOffset = 0;
  $: colScrollOffset = $columnVirtualizer?.scrollOffset || 0;

  /** if we're inferring the column widths from static-ish data, let's
   * find the largest strings in the column and use that to bootstrap the
   * column widths.
   */
  const { columnWidths, largestColumnLength } = estimateColumnCharacterWidths(
    columns,
    rows,
  );

  /* check if column header requires extra space for larger column names  */
  if (largestColumnLength > CHARACTER_LIMIT_FOR_WRAPPING) {
    config.columnHeaderHeight = 46;
  }

  /* set context for child components */
  setContext("config", config);

  let estimateColumnSize: number[] = [];

  /* Separate out dimension column */
  $: dimensionColumnName = $dimensionTableColumnName(dimensionName);
  $: dimensionColumn = columns?.find(
    (c) => c.name == dimensionColumnName,
  ) as VirtualizedTableColumns;
  $: measureColumns =
    columns?.filter((c) => c.name !== dimensionColumnName) ?? [];

  let horizontalScrolling = false;

  let manualDimensionColumnWidth: number | null = null;

  $: rowVirtualizer = createVirtualizer({
    getScrollElement: () => container,
    count: rows.length,
    estimateSize: () => config.rowHeight,
    overscan: rowOverscanAmount,
    paddingStart: config.columnHeaderHeight,
    initialOffset: rowScrollOffset,
  });

  $: if (rows && columns) {
    estimateColumnSize = estimateColumnSizes(
      columns,
      columnWidths,
      containerWidth,
      config,
    );

    if (manualDimensionColumnWidth !== null) {
      estimateColumnSize[0] = manualDimensionColumnWidth;
    }
  }

  $: columnVirtualizer = createVirtualizer({
    getScrollElement: () => container,
    horizontal: true,
    count: measureColumns.length,
    getItemKey: (index) => measureColumns[index].name,
    estimateSize: (index) => {
      return estimateColumnSize[index + 1];
    },
    overscan: columnOverscanAmount,
    paddingStart: estimateColumnSize[0] + FILTER_COLUMN_WIDTH,
    initialOffset: colScrollOffset,
  });

  $: virtualRows = $rowVirtualizer?.getVirtualItems() ?? [];
  $: virtualHeight = $rowVirtualizer?.getTotalSize() ?? 0;

  $: virtualColumns = $columnVirtualizer?.getVirtualItems() ?? [];
  $: virtualWidth = $columnVirtualizer?.getTotalSize() ?? 0;

  let activeIndex;
  function setActiveIndex(event) {
    activeIndex = event.detail;
  }
  function clearActiveIndex() {
    activeIndex = false;
  }

  /** handle scrolling tooltip suppression */
  let scrolling = false;
  let timeoutID;
  $: {
    if (scrolling) {
      if (timeoutID) clearTimeout(timeoutID);
      timeoutID = setTimeout(() => {
        scrolling = false;
      }, 200);
    }
  }

  function onSelectItem(event) {
    // store previous scroll position before re-render
    rowScrollOffset = $rowVirtualizer.scrollOffset;
    colScrollOffset = $columnVirtualizer.scrollOffset;
    dispatch("select-item", event.detail);
  }

  async function handleColumnHeaderClick(event) {
    colScrollOffset = $columnVirtualizer.scrollOffset;
    const columnName = event.detail;
    dimensionTable.handleMeasureColumnHeaderClick(columnName);
  }

  async function handleResizeDimensionColumn(event) {
    rowScrollOffset = $rowVirtualizer.scrollOffset;
    colScrollOffset = $columnVirtualizer.scrollOffset;

    const { size } = event.detail;
    manualDimensionColumnWidth = Math.max(config.minColumnWidth, size);
  }
</script>

<div
  bind:clientWidth={containerWidth}
  style="height: 100%;"
  role="table"
  aria-label="Dimension table"
>
  <div
    bind:this={container}
    on:scroll={() => {
      horizontalScrolling = container?.scrollLeft > 0;
    }}
    style:width="100%"
    style:height="100%"
    class="overflow-auto grid max-w-fit"
    style:grid-template-columns="max-content auto"
    on:scroll={() => {
      /** capture to suppress cell tooltips. Otherwise,
       * there's quite a bit of rendering jank.
       */
      scrolling = true;
    }}
  >
    {#if rowVirtualizer}
      <div
        role="grid"
        tabindex="0"
        class="relative surface"
        on:mouseleave={clearActiveIndex}
        on:blur={clearActiveIndex}
        style:will-change="transform, contents"
        style:width="{virtualWidth}px"
        style:height="{virtualHeight}px"
      >
        <!-- measure column headers -->
        <ColumnHeaders
          virtualColumnItems={virtualColumns}
          noPin={true}
          selectedColumn={$sortMeasure}
          columns={measureColumns}
          on:click-column={handleColumnHeaderClick}
        />
        <!-- dimension value and gutter column -->
        <div class="flex">
          <!-- Gutter for Include Exlude Filter -->
          <DimensionFilterGutter
            virtualRowItems={virtualRows}
            totalHeight={virtualHeight}
            {rows}
            column={dimensionColumn}
            {dimensionName}
            {selectedIndex}
            {isBeingCompared}
            {excludeMode}
            on:toggle-dimension-comparison
            on:select-item={(event) => onSelectItem(event)}
          />
          <DimensionValueHeader
            on:resize-column={handleResizeDimensionColumn}
            virtualRowItems={virtualRows}
            totalHeight={virtualHeight}
            width={estimateColumnSize[0]}
            column={dimensionColumn}
            {rows}
            {activeIndex}
            {selectedIndex}
            {excludeMode}
            {scrolling}
            {horizontalScrolling}
            on:dimension-sort
            on:select-item={(event) => onSelectItem(event)}
            on:inspect={setActiveIndex}
          />
        </div>
        {#if rows.length}
          <!-- VirtualTableBody -->
          <TableCells
            virtualColumnItems={virtualColumns}
            virtualRowItems={virtualRows}
            columns={measureColumns}
            {rows}
            {activeIndex}
            {selectedIndex}
            {scrolling}
            {excludeMode}
            on:select-item={(event) => onSelectItem(event)}
            on:inspect={setActiveIndex}
            cellLabel="Filter dimension value"
          />
        {:else if isFetching}
          <div class="flex text-gray-500 justify-center mt-[30vh]">
            Loading...
          </div>
        {:else}
          <div class="flex text-gray-500 justify-center mt-[30vh]">
            No results to show
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
