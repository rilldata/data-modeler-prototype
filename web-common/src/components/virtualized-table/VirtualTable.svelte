<script lang="ts" context="module">
  import { VirtualizedTableColumnSizes } from "@rilldata/web-common/components/virtualized-table/columnSizes";

  const columnSizes = new VirtualizedTableColumnSizes();

  export const ROW_HEIGHT = 24;
  export const MIN_COL_WIDTH = 108;
  export const MAX_COL_WIDTH = 400;
  export const HEADER_HEIGHT = 36;
  export const columnBuffer = 4;
  export const rowBuffer = 10;
  export const ROW_HEADER_WIDTH = 60;
</script>

<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { formatDataTypeAsDuckDbQueryString } from "@rilldata/web-common/lib/formatters";
  import type {
    V1MetricsViewColumn,
    V1MetricsViewRowsResponseDataItem,
  } from "@rilldata/web-common/runtime-client";
  import { tick } from "svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import { initColumnWidths } from "./init-widths";
  import { clamp } from "@rilldata/utils";
  import { copyToClipboard } from "@rilldata/actions";
  import VirtualTableCell from "./VirtualTableCell.svelte";
  import VirtualTableHeaderCellContent from "./VirtualTableHeaderCellContent.svelte";
  import VirtualTableRowHeader from "./VirtualTableRowHeader.svelte";
  import ColumnWidths from "./VirtualTableColumnWidths.svelte";
  import VirtualTableHeader from "./VirtualTableHeader.svelte";
  import VirtualTableRow from "./VirtualTableRow.svelte";
  import VirtualTooltip from "./VirtualTooltip.svelte";

  type HoveringData = {
    index: number;
    column: string;
    value: string | number | null;
    type: string;
    isHeader: boolean;
    isPin: boolean;
  };

  const dispatch = createEventDispatcher();

  export let rows: V1MetricsViewRowsResponseDataItem[];
  export let columns: (VirtualizedTableColumns | V1MetricsViewColumn)[];
  export let valueAccessor: (name: string) => string = (name: string) => name;
  export let columnAccessor: keyof VirtualizedTableColumns = "name";
  export let HeaderCell = VirtualTableHeaderCellContent;
  export let RowHeader = VirtualTableRowHeader;
  export let Cell = VirtualTableCell;
  export let PinnedCell = Cell;
  export let pinnedColumns = new Map<number, number>();
  export let cellBorders = false;
  export let stickyBorders = false;
  export let headerBorders = false;
  export let rightBorder = false;
  export let rowHeaders = true;
  export let resizableColumns = false;
  export let sortable = false;
  export let minColWidth = MIN_COL_WIDTH;
  export let maxColWidth = MAX_COL_WIDTH;
  export let rowHeight = ROW_HEIGHT;
  export let headerHeight = HEADER_HEIGHT;
  export let selectedIndexes: number[] = [];
  export let sortedColumn: string | null = null;
  export let name: string;

  let container: HTMLDivElement;
  let showTooltip = false;
  let resizing: null | {
    columnIndex: number;
    initialPixelWidth: number;
    initialCursorPosition: number;
  } = null;
  let hoverPosition: DOMRect;
  let hovering: HoveringData | null = null;
  let previousStartCol = 0;
  let previousPaddingLeft = 0;
  let newColStart = 0;
  let previousColumnWidth = 0;
  let previousScrollLeft = 0;
  let startColumn = 0;
  let paddingLeft = 0;
  let contentRect = new DOMRect(0, 0, 0, 0);
  let scrollTop = 0;
  let scrollLeft = 0;
  let nextPinnedColumnPosition = ROW_HEADER_WIDTH;

  $: columnWidths = columnSizes.get(name, columns, columnAccessor, () =>
    initColumnWidths({
      columns,
      rows,
      maxWidth: maxColWidth,
      minWidth: minColWidth,
      columnAccessor,
      resizableColumns,
    }),
  );

  $: firstColumnWidth = columnWidths[newColStart];

  $: clientWidth = contentRect.width;
  $: clientHeight = contentRect.height;

  $: rowCount = rows.length;
  $: columnCount = columns.length;

  $: rowChunk = Math.ceil(rowBuffer / 2);

  $: visibleRows = Math.ceil(clientHeight / rowHeight);
  $: renderedRows = Math.min(visibleRows + rowBuffer, rowCount);
  $: maxRowStart = rowCount - renderedRows;
  $: scrolledRows = Math.floor(scrollTop / rowHeight);
  $: rowTurnover = Math.floor((scrolledRows - rowChunk) / rowChunk);
  $: startRow = clamp(0, rowTurnover * rowChunk, maxRowStart);
  $: paddingTop = startRow * rowHeight;

  $: visibleColumns = Math.ceil(clientWidth / minColWidth);
  $: renderedColumns = Math.min(visibleColumns + columnBuffer, columnCount);
  $: maxColStart = columnCount - renderedColumns;

  $: totalRowSize = rowCount * rowHeight;
  $: totalColumnSize = columnWidths
    .slice(1)
    .reduce((acc, width) => acc + width, 0);

  function handleStartResize(
    event: MouseEvent & { currentTarget: HTMLButtonElement },
  ) {
    const parent = event.currentTarget.parentElement;
    if (!parent) return;
    const index = Number(event.currentTarget.dataset.index);

    resizing = {
      columnIndex: index,
      initialPixelWidth: parent.clientWidth,
      initialCursorPosition: event.clientX,
    };

    window.addEventListener("mousemove", handleResize);
    window.addEventListener("mouseup", handleEndResize);
  }

  function handleResize(event: MouseEvent) {
    if (!resizing) return;
    const delta = event.clientX - resizing.initialCursorPosition;

    requestAnimationFrame(() => {
      if (!resizing) return;
      columnWidths[resizing.columnIndex] = clamp(
        minColWidth,
        resizing.initialPixelWidth + delta,
        maxColWidth,
      );
      columnSizes.set(
        name,
        columns[resizing.columnIndex].name as string,
        columnWidths[resizing.columnIndex],
      );
    });
  }

  function handleEndResize() {
    resizing = null;

    window.removeEventListener("mousemove", handleResize);
    window.removeEventListener("mouseup", handleEndResize);
  }

  function handleHover(
    e: MouseEvent & {
      currentTarget: EventTarget & HTMLElement;
    },
  ) {
    hoverPosition = e.currentTarget.getBoundingClientRect();
    const isHeader = Boolean(e.currentTarget.tagName === "TH");
    const isPin = e.currentTarget.classList.contains("pin-button");
    const description = e.currentTarget.dataset.description;

    const column = String(e.currentTarget.dataset.column);
    const index = Number(e.currentTarget.dataset.index);
    const value =
      (description ?? isHeader)
        ? column
        : (rows[index]?.[column] as string | number | null);
    const type = columns.find((c) => c.name === column)?.type ?? "string";

    hovering = {
      index,
      column,
      type,
      value,
      isHeader,
      isPin,
    };

    showTooltip = true;
  }

  function handleLeave() {
    showTooltip = false;
    hovering = null;
  }

  function handleMouseDown(
    e: MouseEvent & {
      currentTarget: EventTarget & HTMLTableSectionElement;
    },
  ) {
    if (!hovering) return;

    if (e.shiftKey) {
      let exportedValue = formatDataTypeAsDuckDbQueryString(
        hovering.value,
        hovering.type,
      );

      copyToClipboard(exportedValue);

      return;
    }

    if (hovering.isHeader) {
      dispatch("column-click", hovering.column);
    } else {
      dispatch("select-item", {
        index: hovering.index,
        meta: e.ctrlKey || e.metaKey,
      });
    }
  }

  async function handleScroll(
    e: MouseEvent & { currentTarget: HTMLDivElement },
  ) {
    hovering = null;
    const target = e.currentTarget;
    scrollTop = target.scrollTop;
    scrollLeft = target.scrollLeft;

    let newColStart = previousStartCol;

    const xDelta = scrollLeft - previousScrollLeft;

    previousScrollLeft = scrollLeft;

    if (xDelta > 0) {
      while (
        scrollLeft - previousPaddingLeft >= firstColumnWidth &&
        newColStart < maxColStart
      ) {
        previousPaddingLeft += firstColumnWidth;
        newColStart++;
        previousColumnWidth = firstColumnWidth;
        firstColumnWidth = columnWidths[newColStart];
      }
    } else if (xDelta < 0) {
      while (scrollLeft - previousPaddingLeft <= 0 && newColStart > 0) {
        previousPaddingLeft -= previousColumnWidth;
        newColStart--;
        firstColumnWidth = previousColumnWidth;
        previousColumnWidth = columnWidths[newColStart - 1];
      }
    }

    if (newColStart !== startColumn) {
      previousStartCol = newColStart;
      paddingLeft = previousPaddingLeft;

      startColumn = newColStart;
    }

    // // This is to fix a weird quirk in Chrome
    await tick();
    container.scrollTo({ top: scrollTop, left: scrollLeft });
  }

  function togglePin(e: MouseEvent & { currentTarget: HTMLButtonElement }) {
    const index = Number(e.currentTarget.dataset.index);
    if (pinnedColumns.has(index)) {
      let found = false;
      let width = columnWidths[index];

      pinnedColumns.forEach((value, key) => {
        if (key === index) found = true;

        if (found) pinnedColumns.set(key, value - width);
      });

      nextPinnedColumnPosition -= width;

      pinnedColumns.delete(index);
    } else {
      pinnedColumns.set(index, nextPinnedColumnPosition);
      nextPinnedColumnPosition += columnWidths[index];
    }

    pinnedColumns = pinnedColumns;
  }
</script>

<div class="size-full overflow-hidden">
  <div
    role="presentation"
    class="table-wrapper"
    class:border-r={rightBorder}
    class:cursor-col-resize={resizing}
    class:has-selection={selectedIndexes.length > 0}
    class:cell-borders={cellBorders}
    class:sticky-borders={stickyBorders}
    class:header-borders={headerBorders}
    style:--row-height="{rowHeight}px"
    style:--header-height="{headerHeight}px"
    bind:contentRect
    bind:this={container}
    on:mousedown={handleMouseDown}
    on:mouseleave={handleLeave}
    on:scroll={handleScroll}
  >
    <table
      class="relative"
      class:pointer-events-none={resizing}
      style:width="{totalColumnSize}px"
      style:height="{totalRowSize + headerHeight}px"
    >
      <ColumnWidths
        {rowHeaders}
        {paddingLeft}
        {startColumn}
        {columnWidths}
        {pinnedColumns}
        {renderedColumns}
      />

      <VirtualTableHeader
        {columns}
        {sortedColumn}
        {startColumn}
        {pinnedColumns}
        {resizableColumns}
        {renderedColumns}
        {rowHeaders}
        {RowHeader}
        {HeaderCell}
        on:click={togglePin}
        on:mousedown={handleStartResize}
        on:mouseenter={handleHover}
      />

      <tbody>
        <tr style:height="{paddingTop}px" />

        {#each { length: renderedRows } as _, index (index)}
          {@const rowIndex = index + startRow}
          {@const cells = rows[rowIndex]}
          {@const selected = selectedIndexes.includes(rowIndex)}
          <VirtualTableRow
            {columnAccessor}
            {sortedColumn}
            {startColumn}
            {columns}
            {Cell}
            {cells}
            {valueAccessor}
            {pinnedColumns}
            {rowHeaders}
            {selected}
            {rowIndex}
            {renderedColumns}
            {PinnedCell}
            {RowHeader}
            on:mouseenter={handleHover}
          />
        {/each}

        <tr class="h-full" />
      </tbody>
    </table>
  </div>
</div>

{#if showTooltip && hovering}
  <VirtualTooltip
    {sortable}
    {hovering}
    {hoverPosition}
    pinned={pinnedColumns.has(hovering.index)}
  />
{/if}

<style lang="postcss">
  table {
    @apply p-0 m-0 border-spacing-0 border-separate w-fit;
    @apply font-normal cursor-pointer select-none;
    @apply table-fixed;
  }

  .table-wrapper {
    @apply overflow-scroll w-fit max-w-full h-fit max-h-full relative bg-white;
  }

  .has-selection tbody {
    @apply text-gray-400;
    --bar-color: #f0f0f0;
  }
</style>
