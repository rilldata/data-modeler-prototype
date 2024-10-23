<script lang="ts">
  import { createEventDispatcher, getContext } from "svelte";
  import { dragTableCell } from "../drag-table-cell";
  import type { HeaderPosition, VirtualizedTableConfig } from "../types";
  import { modified } from "@rilldata/actions";

  const config: VirtualizedTableConfig = getContext("config");
  const isDimensionTable = config.table === "DimensionTable";

  const dispatch = createEventDispatcher();
  export let header;
  export let position: HeaderPosition = "top";
  export let enableResize = true;
  export let borderRight = false;
  export let onClick: undefined | (() => void) = undefined;
  export let onShiftClick: undefined | (() => void) = undefined;

  export let bgClass = "";

  let positionClasses: string;
  $: {
    if (position === "top") {
      positionClasses = "absolute left-0 top-0";
    } else if (position === "left") {
      positionClasses = "absolute left-0 top-0 text-center font-semibold";
    } else if (position === "top-left") {
      positionClasses = "sticky left-0 top-0 z-40 font-bold";
    }
  }

  $: borderClassesOuterDiv =
    (borderRight ? "border-r " : "") +
    (isDimensionTable
      ? position === "left"
        ? ""
        : "border-b"
      : "border-b border-b-1 border-r border-r-1 border border-gray-200 border-t-0 border-l-0 bg-gray-100");

  const borderClassesInnerDiv = isDimensionTable ? "" : "whitespace-nowrap";

  function focus() {
    dispatch("focus");
  }

  function blur() {
    dispatch("blur");
  }
</script>

<button
  on:mouseover={focus}
  on:mouseleave={blur}
  on:focus={focus}
  on:blur={blur}
  on:click={modified({ shift: onShiftClick, click: onClick })}
  style:transform="translate{position === 'left' ? 'Y' : 'X'}({header.start}px)"
  style:width="{header.size}px"
  style:height="{position === 'left'
    ? config.rowHeight
    : config.columnHeaderHeight}px"
  class="block {positionClasses}
   {bgClass} text-left {borderClassesOuterDiv}"
>
  <div
    class="
    ui-copy
    text-ellipsis overflow-hidden
    {isDimensionTable ? (position === 'left' ? '' : 'px-1') : 'px-4'}
    {borderClassesInnerDiv}
    {position === 'top' && `text-left`}
    {position === 'top-left' &&
      `${isDimensionTable ? 'font-normal' : 'text-center'}`}
    "
  >
    <slot />
    {#if enableResize}
      <div
        role="columnheader"
        tabindex="0"
        use:dragTableCell
        on:resize
        on:dblclick={() => {
          dispatch("reset-column-width");
        }}
        class="absolute top-0 right-0 cursor-col-resize grid place-items-end"
        style:padding-right="1.25px"
        style:width="12px"
        style:height="32px"
      >
        <!-- <div
          style:width="2px"
          style:height="34px"
          class="border border-l border-gray-200"
        /> -->
      </div>
    {/if}
  </div>
</button>
