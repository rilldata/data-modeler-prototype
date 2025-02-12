<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import { getRangeLabel } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import SyntaxElement from "./SyntaxElement.svelte";

  import type { UITimeRange } from "../../time-range-store";

  export let range: UITimeRange;
  export let selected: boolean;
  export let onClick: ((range: string, syntax: boolean) => void) | undefined =
    undefined;

  $: label = getRangeLabel(range.range ?? "");

  $: finalRange = range?.meta?.rillSyntax ?? range.range;
</script>

<DropdownMenu.Item
  on:click={() => {
    if (onClick && finalRange)
      onClick(finalRange, !range.meta && !range.range?.startsWith("P"));
  }}
  class="group h-7"
>
  <div class="size-full flex justify-between items-center" title={range.range}>
    <span class:font-bold={selected} class="truncate">
      {#if label.endsWith(", complete")}
        {label.replace(", complete", "")}
        <span class="text-gray-400 text-[11px]">(complete)</span>
      {:else}
        {label}
      {/if}
    </span>

    <!-- {#if range.meta} -->
    <span class="group-hover:block hidden">
      <SyntaxElement range={finalRange} />
    </span>
    <!-- {/if} -->
  </div>
</DropdownMenu.Item>
