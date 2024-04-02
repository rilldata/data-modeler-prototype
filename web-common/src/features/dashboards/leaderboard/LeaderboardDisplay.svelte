<script lang="ts">
  import VirtualizedGrid from "@rilldata/web-common/components/VirtualizedGrid.svelte";
  import { onDestroy, onMount } from "svelte";
  import Leaderboard from "./Leaderboard.svelte";
  import LeaderboardControls from "./LeaderboardControls.svelte";
  import type {
    MetricsViewSpecDimensionV2,
    V1MetricsViewComparisonResponse,
  } from "@rilldata/web-common/runtime-client";
  import { page } from "$app/stores";
  import { allSelectedDimensions } from "../workspace/dashboard-store";

  export let dimensions: MetricsViewSpecDimensionV2[];
  export let leaderBoards: Record<
    string,
    Promise<V1MetricsViewComparisonResponse>
  >;

  $: metricsViewName = $page.params.name;

  $: hiddenDimensions = allSelectedDimensions.get(metricsViewName);

  $: visibleDimensions = dimensions.filter(
    (d) => !$hiddenDimensions.has(d.name ?? ""),
  );

  // $: console.log($hiddenDimensions);

  /** Functionality for resizing the virtual leaderboard */
  let columns = 3;
  let availableWidth = 0;
  let leaderboardContainer: HTMLElement;
  let observer: ResizeObserver;

  function onResize() {
    if (!leaderboardContainer) return;
    availableWidth = leaderboardContainer.offsetWidth;
    columns = Math.max(1, Math.floor(availableWidth / (315 + 20)));
  }

  onMount(() => {
    onResize();
    const observer = new ResizeObserver(() => {
      onResize();
    });
    observer.observe(leaderboardContainer);
  });

  onDestroy(() => {
    observer?.disconnect();
  });

  function toggleDimension(e: CustomEvent<{ index: number; name: string }>) {
    const { name } = e.detail;
    hiddenDimensions.toggle(name);
  }
</script>

<svelte:window on:resize={onResize} />
<!-- container for the metrics leaderboard components and controls -->
<div
  bind:this={leaderboardContainer}
  class="flex flex-col overflow-y-scroll h-full w-full"
  style:min-width="365px"
>
  <div class="pl-1 pb-3">
    <LeaderboardControls
      {dimensions}
      {metricsViewName}
      on:item-clicked={toggleDimension}
      selectedDimensions={dimensions.map(
        (d) => !$hiddenDimensions.has(d.name ?? ""),
      )}
    />
  </div>
  <div class="grow overflow-hidden">
    {#if visibleDimensions.length > 0}
      <VirtualizedGrid
        {columns}
        height="100%"
        items={visibleDimensions}
        let:item
      >
        <!-- the single virtual element -->
        <Leaderboard dimensionName={item.name} data={leaderBoards[item.name]} />
      </VirtualizedGrid>
    {/if}
  </div>
</div>
