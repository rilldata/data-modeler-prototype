<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import AddAssetButton from "../entity-management/AddAssetButton.svelte";
  import { getName } from "../entity-management/name-utils";
  import ChartMenuItems from "./ChartMenuItems.svelte";
  import { createChart } from "./createChart";
  import { useChartFileNames } from "./selectors";
  import { flip } from "svelte/animate";

  let showCharts = true;

  $: chartFileNames = useChartFileNames($runtime.instanceId);

  async function handleAddChart() {
    const newChartName = getName("chart", $chartFileNames.data ?? []);
    await createChart($runtime.instanceId, newChartName);
    await goto(`/chart/${newChartName}`);
  }
</script>

<div class="h-fit flex flex-col">
  <NavigationHeader bind:show={showCharts}>Charts</NavigationHeader>

  {#if showCharts}
    <ol transition:slide={{ duration }}>
      {#if $chartFileNames?.data}
        {#each $chartFileNames.data as chartName (chartName)}
          <li animate:flip={{ duration }} aria-label={chartName}>
            <NavigationEntry
              name={chartName}
              context="chart"
              open={$page.url.pathname === `/chart/${chartName}`}
            >
              <ChartMenuItems slot="menu-items" {chartName} />
            </NavigationEntry>
          </li>
        {/each}
      {/if}
      <AddAssetButton
        id="add-chart"
        label="Add chart"
        bold={false}
        on:click={handleAddChart}
      />
    </ol>
  {/if}
</div>
