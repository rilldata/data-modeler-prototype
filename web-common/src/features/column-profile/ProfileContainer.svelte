<script lang="ts">
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import {
    COLUMN_PROFILE_CONFIG,
    LIST_SLIDE_DURATION,
  } from "@rilldata/web-common/layout/config";
  import { modified } from "@rilldata/actions";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";

  export let active = false;
  export let emphasize = false;
  export let type;
  export let example;
  export let isFetching;
  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;
  export let mode = "summaries";
  export let onShiftClick: () => void;

  const dispatch = createEventDispatcher();

  let columns: string;
  $: summarySize =
    COLUMN_PROFILE_CONFIG.summaryVizWidth[compact ? "small" : "medium"];
  $: if (hideNullPercentage) {
    columns = `${summarySize}px`;
  } else {
    columns = `${summarySize}px ${COLUMN_PROFILE_CONFIG.nullPercentageWidth}px`;
  }
</script>

<div>
  <button
    class="
        px-4
        select-none
        flex
        space-between
        gap-2
        transition-colors
        hover:bg-gray-100
        focus:bg-gray-100
        focus:ring-gray-500
        focus:outline-gray-300 flex-1
        justify-between w-full"
    class:ui-copy-disabled-faint={isFetching}
    class:bg-gray-50={active}
    on:click={modified({
      shift: onShiftClick,
      click: () => {
        dispatch("select");
      },
    })}
  >
    <div class="flex gap-2 items-baseline flex-1" style:min-width="0px">
      <div class="self-center flex items-center ui-copy-icon-muted">
        <slot name="icon" />
      </div>
      <div
        class:font-bold={emphasize}
        class="text-left w-full text-ellipsis overflow-hidden whitespace-nowrap"
      >
        <slot name="left" />
      </div>
    </div>
    <div
      class:hidden={hideRight || mode !== "summaries"}
      class="grid gap-x-2"
      style:grid-template-columns={columns}
    >
      {#if mode === "summaries"}
        <div>
          <slot name="summary" />
        </div>
        {#if !hideNullPercentage}
          <div>
            <slot name="nullity" />
          </div>
        {/if}
      {/if}
      <div>
        <slot name="context-button" />
      </div>
    </div>
    <div
      class:hidden={mode !== "example" || hideRight}
      class="pl-8 text-ellipsis overflow-hidden whitespace-nowrap text-right"
      style:max-width="240px"
    >
      <FormattedDataType
        {type}
        isNull={example === null || example === ""}
        value={example}
      />
    </div>
  </button>
  {#if active && $$slots["details"]}
    <div class="w-full" transition:slide={{ duration: LIST_SLIDE_DURATION }}>
      <slot name="details" />
    </div>
  {/if}
</div>
