<script lang="ts">
  import PivotError from "@rilldata/web-common/features/dashboards/pivot/PivotError.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { derived } from "svelte/store";
  import { getPivotConfig } from "./pivot-data-config";
  import { usePivotForExplore } from "./pivot-data-store";
  import PivotEmpty from "./PivotEmpty.svelte";
  import PivotHeader from "./PivotHeader.svelte";
  import PivotSidebar from "./PivotSidebar.svelte";
  import PivotTable from "./PivotTable.svelte";
  import PivotToolbar from "./PivotToolbar.svelte";

  const stateManagers = getStateManagers();
  const {
    exploreName,
    dashboardStore,
    selectors: {
      pivot: { columns },
    },
  } = stateManagers;
  const { cloudDataViewer, readOnly } = featureFlags;

  $: isRillDeveloper = $readOnly === false;
  $: canShowDataViewer = Boolean($cloudDataViewer || isRillDeveloper);

  const pivotExploreState = derived(dashboardStore, (dashboard) => {
    return dashboard?.pivot;
  });

  let showPanels = true;

  $: pivotDataStore = usePivotForExplore(stateManagers);
  $: pivotConfig = getPivotConfig(stateManagers);

  $: ({ isFetching, assembled } = $pivotDataStore);

  $: hasColumnAndNoMeasure =
    $columns.dimension.length > 0 && $columns.measure.length === 0;

  function removeActiveCell() {
    if (!$dashboardStore.pivot.activeCell) return;
    metricsExplorerStore.removePivotActiveCell($exploreName);
  }
</script>

<div class="layout">
  {#if showPanels}
    <PivotSidebar />
  {/if}
  <div class="flex flex-col size-full overflow-hidden">
    {#if showPanels}
      <PivotHeader />
    {/if}
    <div
      class="content"
      role="presentation"
      on:mousedown|self={removeActiveCell}
    >
      <PivotToolbar {isFetching} bind:showPanels />

      {#if $pivotDataStore?.error?.length}
        <PivotError errors={$pivotDataStore.error} />
      {:else if !$pivotDataStore?.data || $pivotDataStore?.data?.length === 0}
        <PivotEmpty {assembled} {isFetching} {hasColumnAndNoMeasure} />
      {:else}
        <PivotTable
          {pivotDataStore}
          config={pivotConfig}
          pivotState={pivotExploreState}
          setPivotExpanded={(expanded) =>
            metricsExplorerStore.setPivotExpanded($exploreName, expanded)}
          setPivotSort={(sorting) =>
            metricsExplorerStore.setPivotSort($exploreName, sorting)}
          setPivotRowPage={(page) =>
            metricsExplorerStore.setPivotRowPage($exploreName, page)}
          {canShowDataViewer}
          setPivotActiveCell={(rowId, columnId) =>
            metricsExplorerStore.setPivotActiveCell(
              $exploreName,
              rowId,
              columnId,
            )}
        />
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .layout {
    @apply flex box-border h-full overflow-hidden;
  }

  .content {
    @apply flex w-full flex-col bg-slate-100 overflow-hidden size-full;
    @apply p-2 gap-y-2;
  }
</style>
