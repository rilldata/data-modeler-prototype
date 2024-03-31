<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import StickyHeader from "@rilldata/web-common/components/virtualized-table/core/StickyHeader.svelte";
  import { getContext } from "svelte";
  import type { VirtualizedTableConfig } from "../../../components/virtualized-table/types";
  import DimensionCompareMenu from "@rilldata/web-common/features/dashboards/leaderboard/DimensionCompareMenu.svelte";
  import { getStateManagers } from "../state-managers/state-managers";
  import { colorGetter } from "../filters/colorGetter";
  import type { VirtualizedTableColumns } from "../../../components/virtualized-table/types";
  import type { DimensionTableRow } from "./dimension-table-types";
  import type { VirtualItem } from "@tanstack/svelte-virtual";

  export let totalHeight: number;
  export let virtualRowItems: VirtualItem[];
  export let selectedIndex: number[] = [];
  export let excludeMode = false;
  export let isBeingCompared = false;
  export let column: VirtualizedTableColumns;
  export let rows: DimensionTableRow[];
  export let dimensionName: string;

  const {
    metricsViewName,
    selectors: {
      dimensions: { dimensionTableDimName },
    },
  } = getStateManagers();

  const config: VirtualizedTableConfig = getContext("config");
</script>

<div
  class="sticky left-0 top-0 z-20 bg-white"
  style:height="{totalHeight}px"
  style:width="{config.indexWidth}px"
>
  <div
    style:height="{config.columnHeaderHeight}px"
    class="sticky left-0 top-0 surface z-40 flex items-center"
  >
    <DimensionCompareMenu dimensionName={$dimensionTableDimName} />
  </div>
  {#each virtualRowItems as row (`row-${row.key}`)}
    {@const isSelected = selectedIndex.includes(row.index)}
    {@const indexOf = selectedIndex.indexOf(row.index)}
    {@const dimensionValue = String(rows[row.index][column.name])}
    <StickyHeader
      enableResize={false}
      position="left"
      header={{ size: config.indexWidth, start: row.start }}
    >
      <div class="py-0.5 grid place-items-center">
        {#if isSelected && !excludeMode && isBeingCompared}
          {@const color = colorGetter.get(
            $metricsViewName,
            dimensionName,
            dimensionValue,
          )}
          <CheckCircle
            className="fill-{indexOf >= 7 ? 'gray-300' : color}"
            size="18px"
          />
        {:else if isSelected && !excludeMode}
          <Check size="20px" />
        {:else if isSelected && excludeMode}
          <Cancel size="20px" />
        {:else}
          <Spacer />
        {/if}
      </div>
    </StickyHeader>
  {/each}
</div>
