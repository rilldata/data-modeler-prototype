<script lang="ts">
  import Pivot from "@rilldata/web-common/features/dashboards/pivot/RegularTable.svelte";
  import type {
    PivotPos,
    PivotRenderCallback,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { SortType } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import {
    ExcludeIcon,
    MeasureArrow,
    PieChart,
    PinHoverUnsetIcon,
    PinSetHoverIcon,
    PinSetIcon,
    PinUnsetIcon,
    SelectedCheckmark,
  } from "@rilldata/web-common/features/dashboards/time-dimension-details/TDDIcons";
  import { getClassForCell } from "@rilldata/web-common/features/dashboards/time-dimension-details/util";
  import { createEventDispatcher } from "svelte";
  import { lastKnownPosition } from "./time-dimension-data-store";
  import type { TDDComparison, TableData, TablePosition } from "./types";

  export let dimensionLabel: string;
  export let measureLabel: string;
  export let excludeMode: boolean;
  export let sortDirection: boolean;
  export let sortType: SortType;
  export let highlightedCol: number | undefined;
  export let scrubPos: { start?: number; end?: number };
  export let pinIndex: number;
  export let comparing: TDDComparison;
  export let tableData: TableData;
  export let markerColors: string[];

  /** Formatter for the time axis in the table*/
  export let timeFormatter: (date: Date) => string;

  const dispatch = createEventDispatcher();

  let pivot;

  let rowIdxHover: number | undefined;
  let colIdxHover: number | undefined;
  let hoveringPin = false;

  function getBodyData(pos: PivotPos) {
    return tableData?.body
      ?.slice(pos.x0, pos.x1)
      .map((row) => row.slice(pos.y0, pos.y1));
  }

  export function getRowHeaderData(pos: PivotPos) {
    return tableData?.rowHeaderData?.slice(pos.y0, pos.y1);
  }

  export function getColumnHeaderData(pos: PivotPos) {
    return tableData?.columnHeaderData?.slice(pos.x0, pos.x1);
  }

  const renderCell: PivotRenderCallback = (data) => {
    const classesToAdd = ["text-right"];
    const classesToRemove = [
      "border-b",
      "border-gray-200",
      "bg-white",
      "bg-gray-100",
      "bg-gray-200",
      "bg-primary-50",
      "bg-primary-100",
      "bg-primary-200",
      "bg-slate-50",
      "bg-slate-100",
      "bg-slate-200",
    ];

    if (pinIndex > -1 && comparing === "dimension" && data.y === pinIndex + 1) {
      classesToAdd.push("border-b", "border-gray-200");
    }

    if (comparing === "time" && data.y === 2) {
      classesToAdd.push("border-b", "border-gray-200");
    }

    const isScrubbed =
      scrubPos?.start !== undefined &&
      scrubPos?.end !== undefined &&
      data.x >= scrubPos.start &&
      data.x <= scrubPos.end - 1;

    const palette = isScrubbed
      ? "scrubbed"
      : data.y === 0
        ? "fixed"
        : "default";

    classesToAdd.push(
      getClassForCell(
        palette,
        rowIdxHover,
        colIdxHover ?? highlightedCol,
        data.y,
        data.x,
      ),
    );
    // Update DOM with consolidated class operations
    data.element.classList.toggle("font-semibold", Boolean(data.y == 0));
    data.element.classList.remove(...classesToRemove);
    data.element.classList.add(...classesToAdd);
  };

  const renderColumnHeader: PivotRenderCallback = (data) => {
    data.element.classList.add("font-medium", "text-right");
    return timeFormatter(data.value.value);
  };

  // Any time visible line list changes, redraw the table
  $: {
    scrubPos;
    highlightedCol;
    tableData?.selectedValues;
    pivot?.draw();
  }

  const getPinIcon = () => {
    if (comparing === "dimension") {
      if (tableData?.selectedValues.length === 0) return "";
      else if (pinIndex === tableData?.selectedValues.length - 1)
        return hoveringPin ? PinHoverUnsetIcon : PinSetIcon;
      else return hoveringPin ? PinSetHoverIcon : PinUnsetIcon;
    } else {
      return "";
    }
  };

  const getMarker = (value, y) => {
    if (y === 0) {
      return { icon: "", muted: false };
    }
    const visibleIdx = tableData?.selectedValues.indexOf(value.value);

    if (comparing === "time") {
      let icon = "";
      if (y == 1) icon = SelectedCheckmark("fill-primary-500");
      else if (y == 2) icon = SelectedCheckmark("fill-gray-300");
      return { icon, muted: false };
    }

    if (visibleIdx > -1) {
      if (excludeMode) return { icon: ExcludeIcon, muted: true };
      // Only show colored markers for first 11 selected values
      else
        return {
          icon: SelectedCheckmark(
            "fill-" + (visibleIdx < 11 ? markerColors[visibleIdx] : "gray-300"),
          ),
          muted: false,
        };
    }

    return {
      icon: ``,
      muted: !excludeMode && tableData?.selectedValues.length > 0,
    };
  };

  const renderRowHeader: PivotRenderCallback = ({ value, x, y, element }) => {
    const showBorder =
      (pinIndex > -1 && comparing === "dimension" && y === pinIndex + 1) ||
      (comparing === "time" && y === 2);
    if (showBorder) {
      element.classList.add("border-b", "border-gray-200");
    } else {
      element.classList.remove("border-b", "border-gray-200");
    }

    const cellBgColor = getClassForCell(
      "fixed",
      rowIdxHover,
      colIdxHover ?? highlightedCol,
      y,
      x - tableData?.fixedColCount,
    );
    if (x > 0) {
      element.classList.remove("bg-slate-50", "bg-slate-100", "bg-slate-200");
      element.classList.add(cellBgColor);
    }
    if (x === 0) {
      element.classList.add("pl-0");
      const marker = getMarker(value, y);

      // Gray out rows which are not included
      if (marker.muted) {
        element?.parentElement?.classList.add("ui-copy-disabled-faint");
      } else {
        element?.parentElement?.classList.remove("ui-copy-disabled-faint");
      }

      const fontWeight = y === 0 ? "font-semibold" : "font-normal";
      return `<div class="flex items-center w-full h-full overflow-hidden pr-2 gap-1">
        <div class="w-5 shrink-0 h-full flex items-center justify-center">${marker.icon}</div>
        <div class="truncate text-xs ${fontWeight}">${value.value}</div></div>`;
    } else if (x === 1)
      return `<div class="text-xs font-semibold text-right flex items-center justify-end gap-2" >
        ${value.value}
        ${value.spark}
        </div>`;
    else
      return `<div class="text-xs font-normal text-right" >${value.value}</div>`;
  };

  const renderRowCorner: PivotRenderCallback = (data) => {
    data.element.classList.add("bg-white", "z-10");
    if (data.x === 0) {
      const pinIcon = getPinIcon();
      return `
      <div class="flex items-center font-medium text-left">
        <span pin="true" class="pin w-5 pr-1 shrink-0 h-full flex items-center justify-center">${pinIcon} </span>
        <div sort="dimension" class="flex flex-grow items-center">
          <span  class="flex truncate">${dimensionLabel} </span>
        ${
          comparing === "dimension" && sortType === SortType.DIMENSION
            ? `<span>${MeasureArrow(sortDirection)}</span>`
            : ``
        }
        </div>
      </div>`;
    }

    if (data.x === 1)
      return `<div class="text-right font-medium flex items-center" sort="value">
        <span class="truncate">${measureLabel} </span>
        ${
          comparing === "dimension" &&
          tableData?.fixedColCount === 2 &&
          sortType === SortType.VALUE
            ? `<span>${MeasureArrow(sortDirection)}</span>`
            : ``
        }
      </div>`;
    if (data.x === 2)
      return `<div class="flex items-center justify-end" sort="value">${PieChart} % 
        ${
          comparing === "dimension" && sortType === SortType.VALUE
            ? MeasureArrow(sortDirection)
            : ``
        }
      </div>`;
  };

  const getColumnWidth = () => {
    return 75;
  };

  const getRowHeaderWidth = (x: number) => {
    return [250, 130, 50][x];
  };

  // Visible line list
  const toggleVisible = (n) => {
    n = parseInt(n);
    if (comparing != "dimension" || n == 0) return;
    const label = tableData?.rowHeaderData[n][0].value;
    dispatch("toggle-filter", label);
  };

  const togglePin = () => {
    dispatch("toggle-pin");
  };

  const handleEvent = (evt, table, attribute, callback) => {
    let currentNode = evt.target;

    let found = currentNode.hasAttribute(attribute);
    while (!found && currentNode !== table) {
      currentNode = currentNode.parentNode;
      found = currentNode.hasAttribute(attribute);
    }
    if (found) {
      const attributeValue = currentNode.getAttribute(attribute);
      callback(attributeValue);
    }
  };

  const handleMouseDown = (evt, table) => {
    handleEvent(evt, table, "__row", toggleVisible);
    handleEvent(evt, table, "sort", (type) => dispatch("toggle-sort", type));
    handleEvent(evt, table, "pin", togglePin);
  };

  const handleMouseHover = (evt, table) => {
    let newRowIdxHover;
    let newColIdxHover;
    let newHoveringPin = hoveringPin;
    if (evt.type === "mouseout") {
      newRowIdxHover = undefined;
      newColIdxHover = undefined;
      newHoveringPin = false;
    } else {
      handleEvent(evt, table, "__row", (n) => (newRowIdxHover = parseInt(n)));
      handleEvent(evt, table, "__col", (n) => (newColIdxHover = parseInt(n)));
      handleEvent(evt, table, "pin", () => (newHoveringPin = true));
    }

    if (hoveringPin !== newHoveringPin) {
      hoveringPin = newHoveringPin;
      pivot?.draw();
    }

    if (newRowIdxHover !== rowIdxHover && newColIdxHover !== colIdxHover) {
      rowIdxHover = newRowIdxHover;
      colIdxHover = newColIdxHover;
      dispatch("highlight", { x: colIdxHover, y: rowIdxHover });
      pivot?.draw();
    }
  };

  function resetHighlight() {
    rowIdxHover = undefined;
    colIdxHover = undefined;
    dispatch("highlight", { x: colIdxHover, y: rowIdxHover });
    pivot?.draw();
  }

  // Scroll to previous position in case of dashboard refresh during reconcile
  let currentPosition: TablePosition;
  let hasScrolled = false;
  let isInitialized;
  function handlePos(e) {
    const pos = e.detail;
    currentPosition = pos;
    isInitialized = pivot?.isInitialized();
    if (!$lastKnownPosition || hasScrolled) lastKnownPosition.set(pos);
  }

  $: if (isInitialized && !hasScrolled && $lastKnownPosition) {
    pivot?.scrollToCell($lastKnownPosition?.x0, $lastKnownPosition?.y0);
    hasScrolled =
      currentPosition?.x0 === $lastKnownPosition?.x0 &&
      currentPosition?.y0 === $lastKnownPosition?.y0;
  }

  // Hack: for some reason, not enough columns are being drawn on first render.
  // Force a second initial render to workaround it.
  $: {
    if (pivot) {
      setTimeout(pivot.draw, 0);
    }
  }

  // Add a CSS variable to control the cursor style as table element is not
  // directly available
  $: cssVarStyles = `--cursor: ${
    comparing === "dimension" ? "pointer" : "default"
  }`;
</script>

<div
  role="grid"
  tabindex="0"
  on:mouseleave={resetHighlight}
  style:height={comparing === "none" ? "80px" : "calc(100% - 50px)"}
  style={cssVarStyles}
  class="w-full relative h-full"
>
  <Pivot
    bind:this={pivot}
    {getRowHeaderData}
    {getColumnHeaderData}
    {getBodyData}
    rowCount={tableData?.rowCount}
    columnCount={tableData?.columnCount}
    rowHeaderDepth={4}
    columnHeaderDepth={1}
    {renderCell}
    {renderColumnHeader}
    {renderRowHeader}
    {renderRowCorner}
    {getColumnWidth}
    {getRowHeaderWidth}
    onMouseDown={handleMouseDown}
    onMouseHover={handleMouseHover}
    on:pos={handlePos}
  />
</div>

<style>
  /* Define cursor styles */
  :global(regular-table table, regular-table div[sort]) {
    cursor: var(--cursor, default);
  }
  :global(regular-table table tbody tr:first-child, regular-table thead) {
    cursor: default;
  }
  :global(.pin) {
    cursor: pointer;
    margin-top: 2px;
  }
</style>
