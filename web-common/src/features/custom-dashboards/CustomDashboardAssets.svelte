<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { slide } from "svelte/transition";
  import { flip } from "svelte/animate";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import AddAssetButton from "../entity-management/AddAssetButton.svelte";
  import { getName } from "../entity-management/name-utils";
  import CustomDashboardMenuItems from "./CustomDashboardMenuItems.svelte";
  import { createCustomDashboard } from "./createCustomDashboard";
  import { useCustomDashboardFileNames } from "./selectors";

  let showCustomDashboards = true;

  $: customDashboardFileNames = useCustomDashboardFileNames(
    $runtime.instanceId,
  );

  async function handleAddCustomDashboard() {
    const newCustomDashboardName = getName(
      "dashboard",
      $customDashboardFileNames.data ?? [],
    );
    await createCustomDashboard($runtime.instanceId, newCustomDashboardName);
    await goto(`/custom-dashboard/${newCustomDashboardName}`);
  }
</script>

<div class="h-fit flex flex-col">
  <NavigationHeader bind:show={showCustomDashboards}>
    Custom Dashboards
  </NavigationHeader>

  {#if showCustomDashboards}
    <ol transition:slide={{ duration }}>
      {#if $customDashboardFileNames?.data}
        {#each $customDashboardFileNames.data as customDashboardName (customDashboardName)}
          <li animate:flip={{ duration }} aria-label={customDashboardName}>
            <NavigationEntry
              name={customDashboardName}
              context="custom-dashboard"
              open={$page.url.pathname ===
                `/custom-dashboard/${customDashboardName}`}
            >
              <CustomDashboardMenuItems
                slot="menu-items"
                {customDashboardName}
              />
            </NavigationEntry>
          </li>
        {/each}
      {/if}
      <AddAssetButton
        id="add-custom-dashboard"
        label="Add dashboard"
        bold={false}
        on:click={handleAddCustomDashboard}
      />
    </ol>
  {/if}
</div>
