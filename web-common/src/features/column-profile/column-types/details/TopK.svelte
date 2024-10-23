<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { copyToClipboard, isClipboardApiSupported } from "@rilldata/actions";
  import { modified } from "@rilldata/actions";
  import { isNested } from "@rilldata/web-common/lib/duckdb-data-types";
  import {
    formatBigNumberPercentage,
    formatDataType,
    formatInteger,
  } from "@rilldata/web-common/lib/formatters";
  import type { TopKEntry } from "@rilldata/web-common/runtime-client";
  import { format } from "d3-format";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import TopKListItem from "./TopKListItem.svelte";
  import type { Location } from "@rilldata/web-common/lib/place-element";

  export let colorClass = "bg-primary-200";
  export let topK: TopKEntry[] | undefined;
  export let totalRows: number;
  export let k = 15;
  export let type: string;

  const dispatch = createEventDispatcher();

  $: smallestPercentage =
    topK && topK.length
      ? Math.min(
          ...topK.slice(0, 5).map((entry) => (entry?.count ?? 0) / totalRows),
        )
      : undefined;
  $: formatPercentage =
    smallestPercentage && smallestPercentage < 0.01
      ? format("0.2%")
      : smallestPercentage
        ? format("0.1%")
        : () => "";

  // We need this to get transition working properly.
  // Since the topk query is in a reactive statement with `enable`, `topK` can be undefined.
  // This leads to unexpected issues when paired with transition
  $: topKCopy = topK ?? topKCopy;

  function ensureSpaces(str: string, n = 6) {
    return `${Array.from({ length: n - str.length })
      .fill("&nbsp;")
      .join("")}${str}`;
  }

  function getCopyValue(type, value) {
    return isNested(type) ? formatDataType(value, type) : value;
  }

  let tooltipProps: { location: Location; distance: number } = {
    location: "right",
    distance: 16,
  };

  function handleFocus(value: TopKEntry) {
    return () => dispatch("focus-top-k", value);
  }

  function handleBlur(value: TopKEntry) {
    return () => dispatch("blur-top-k", value);
  }

  /** handle LISTs and STRUCTs */
</script>

{#if topKCopy && totalRows}
  <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
    {#each topKCopy.slice(0, k) as item (item.value)}
      {@const negligiblePercentage = item.count / totalRows < 0.0002}
      {@const percentage = negligiblePercentage
        ? "<.01%"
        : formatPercentage(item.count / totalRows)}
      <TopKListItem
        value={item.count / totalRows}
        color={colorClass}
        on:focus={handleFocus(item)}
        on:blur={handleBlur(item)}
      >
        <svelte:fragment slot="title">
          <Tooltip {...tooltipProps} suppress={!isClipboardApiSupported()}>
            <button
              style:font-size="12px"
              class="text-ellipsis overflow-hidden whitespace-nowrap"
              on:click={modified({
                shift: () =>
                  copyToClipboard(
                    getCopyValue(type, item.value),
                    `copied column value "${
                      item.value === null
                        ? "NULL"
                        : getCopyValue(type, item.value)
                    }" to clipboard`,
                  ),
              })}
            >
              {formatDataType(item.value, type)}
            </button>
            <TooltipContent slot="tooltip-content" maxWidth="300px">
              <TooltipTitle>
                <svelte:fragment slot="name">
                  {`${formatDataType(item.value, type)}`}
                </svelte:fragment>
                <svelte:fragment slot="description"
                  >{formatBigNumberPercentage(item.count / totalRows)} of rows</svelte:fragment
                >
              </TooltipTitle>
              <TooltipShortcutContainer>
                <div>
                  <StackingWord key="shift">Copy</StackingWord> column value to clipboard
                </div>
                <Shortcut>
                  <span style="font-family: var(--system);">⇧</span> + Click
                </Shortcut>
              </TooltipShortcutContainer>
            </TooltipContent>
          </Tooltip>
        </svelte:fragment>
        <svelte:fragment slot="right">
          <Tooltip {...tooltipProps} suppress={!isClipboardApiSupported()}>
            <button
              on:click={modified({
                shift: () =>
                  copyToClipboard(
                    item.count,
                    `copied ${item.count} to clipboard`,
                  ),
              })}
            >
              {formatInteger(item.count)}
              <span class="ui-copy-inactive pl-2">
                {@html ensureSpaces(percentage)}</span
              >
            </button>
            <TooltipContent slot="tooltip-content">
              <TooltipTitle>
                <svelte:fragment slot="name"
                  >{`${formatDataType(item.value, type)}`}</svelte:fragment
                >
                <svelte:fragment slot="description"
                  >{formatBigNumberPercentage(item.count / totalRows)} of rows</svelte:fragment
                >
              </TooltipTitle>
              <TooltipShortcutContainer>
                <div>
                  <StackingWord key="shift">Copy</StackingWord> count to clipboard
                </div>
                <Shortcut>
                  <span style="font-family: var(--system);">⇧</span> + Click
                </Shortcut>
              </TooltipShortcutContainer>
            </TooltipContent>
          </Tooltip>
        </svelte:fragment>
      </TopKListItem>
    {/each}
  </div>
{/if}
