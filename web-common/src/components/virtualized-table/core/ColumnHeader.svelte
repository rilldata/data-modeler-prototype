<script lang="ts">
  import { DataTypeIcon } from "@rilldata/web-common/components/data-types";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import Pin from "@rilldata/web-common/components/icons/Pin.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { copyToClipboard, isClipboardApiSupported } from "@rilldata/actions";
  import { createEventDispatcher, getContext } from "svelte";
  import { fly } from "svelte/transition";
  import TooltipDescription from "../../tooltip/TooltipDescription.svelte";
  import type { ResizeEvent } from "../drag-table-cell";
  import type { HeaderPosition, VirtualizedTableConfig } from "../types";
  import StickyHeader from "./StickyHeader.svelte";

  export let pinned = false;
  export let noPin = false;
  export let showDataIcon = false;
  export let name;
  export let type: string;
  export let description = "";
  export let header;
  export let position: HeaderPosition = "top";
  export let enableResize = true;
  export let enableSorting = true;
  export let isSelected = false;
  export let highlight = false;
  export let sorted: SortDirection | undefined = undefined;

  const config: VirtualizedTableConfig = getContext("config");
  const dispatch = createEventDispatcher();

  let showMore = false;

  $: isDimensionTable = config.table === "DimensionTable";
  $: isDimensionColumn =
    isDimensionTable && (type === "VARCHAR" || type === "CODE_STRING");

  $: textAlignment = isDimensionColumn ? "text-left pl-1" : "text-right pr-1";

  $: columnFontWeight = isSelected
    ? "font-bold"
    : config.columnHeaderFontWeightClass;

  const handleResize = (event: ResizeEvent) => {
    dispatch("resize-column", {
      size: event.detail.size,
      name,
    });
  };
</script>

<StickyHeader
  {enableResize}
  bgClass={highlight
    ? config.headerBgColorHighlightClass
    : config.headerBgColorClass}
  on:reset-column-width={() => {
    dispatch("reset-column-width", { name });
  }}
  on:resize={handleResize}
  {position}
  {header}
  on:focus={() => {
    showMore = true;
  }}
  on:blur={() => {
    showMore = false;
  }}
  onClick={() => {
    dispatch("click-column");
  }}
  onShiftClick={() => {
    copyToClipboard(name, `copied column name "${name}" to clipboard`);
  }}
>
  <div
    class=" 
           flex
           justify-stretch
           select-none
           over
           {isDimensionTable ? '' : 'items-center gap-x-2'}
           "
  >
    <Tooltip location="top" alignment="middle" distance={16}>
      <div
        class="
        grid
        items-center cursor-pointer w-full"
        class:gap-x-2={!isSelected}
        style:grid-template-columns={isDimensionTable
          ? ""
          : `max-content auto ${!noPin && showMore ? "max-content" : ""}`}
      >
        {#if showDataIcon}
          <DataTypeIcon suppressTooltip color={"text-gray-500"} {type} />
        {/if}
        <span
          class="text-ellipsis
          {columnFontWeight}
          {isDimensionTable
            ? `${textAlignment} break-words line-clamp-2`
            : 'overflow-hidden whitespace-nowrap'}
          "
        >
          {#if typeof name !== "string"}
            <div class="flex justify-end">
              <svelte:component this={name} />
            </div>
          {:else}
            {name}
          {/if}
        </span>
      </div>
      <TooltipContent slot="tooltip-content" maxWidth="280px">
        {#if !isDimensionTable}
          <TooltipTitle>
            <svelte:fragment slot="name">
              {name}
            </svelte:fragment>
            <svelte:fragment slot="description">
              {showDataIcon ? type : ""}
            </svelte:fragment>
          </TooltipTitle>
        {/if}
        {#if isDimensionTable && description?.length}
          <TooltipDescription>
            {description}
          </TooltipDescription>
        {/if}
        {#if isDimensionTable || isClipboardApiSupported()}
          <TooltipShortcutContainer>
            {#if enableSorting}
              <div>Sort column</div>
              <Shortcut>Click</Shortcut>
            {/if}
            {#if isClipboardApiSupported()}
              <div>
                <StackingWord key="shift">Copy</StackingWord>
                column name to clipboard
              </div>
              <Shortcut>
                <span style="font-family: var(--system);">⇧</span> + Click
              </Shortcut>
            {/if}
          </TooltipShortcutContainer>
        {/if}
      </TooltipContent>
    </Tooltip>

    {#if sorted}
      <div class="mt-0.5 ui-copy-icon">
        {#if sorted === SortDirection.DESCENDING}
          <div in:fly|global={{ duration: 200, y: -8 }} style:opacity={1}>
            <ArrowDown size="12px" />
          </div>
        {:else if sorted === SortDirection.ASCENDING}
          <div in:fly|global={{ duration: 200, y: 8 }} style:opacity={1}>
            <ArrowDown flip size="12px" />
          </div>
        {/if}
      </div>
    {/if}

    {#if !noPin && showMore}
      <Tooltip location="top" alignment="middle" distance={16}>
        <button
          transition:fly={{ duration: 200, y: 4 }}
          class:text-gray-900={pinned}
          class:text-gray-400={!pinned}
          class="transition-colors duration-100 justify-self-end"
          on:click={() => {
            dispatch("pin");
          }}
        >
          <Pin size="16px" />
        </button>
        <TooltipContent slot="tooltip-content">
          {pinned
            ? "unpin this column from the right side of the table"
            : "pin this column to the right side of the table"}
        </TooltipContent>
      </Tooltip>
    {/if}
  </div>
</StickyHeader>

<style>
  .line-clamp-2 {
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
  }
</style>
