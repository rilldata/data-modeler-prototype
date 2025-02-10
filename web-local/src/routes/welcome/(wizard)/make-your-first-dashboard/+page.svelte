<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import { connectorExplorerStore } from "@rilldata/web-common/features/connectors/connector-explorer-store";
  import ConnectorExplorer from "@rilldata/web-common/features/connectors/ConnectorExplorer.svelte";
  import { useCreateMetricsViewFromTableUIAction } from "@rilldata/web-common/features/metrics-views/ai-generation/generateMetricsView";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  export let data: PageData;
  const { onboardingState } = data;

  const store = connectorExplorerStore.duplicateStore();

  const explorerState = store.store;
  $: selectedItem = $explorerState.selectedItem;

  async function createDashboard() {
    await createMetricsViewFromTable();
    onboardingState.complete();
  }

  async function skip() {
    await goto("/");
    onboardingState.complete();
  }

  let createMetricsViewFromTable: () => Promise<void>;
  $: {
    if (selectedItem && selectedItem?.table) {
      createMetricsViewFromTable = useCreateMetricsViewFromTableUIAction(
        $runtime.instanceId,
        selectedItem.connector,
        selectedItem?.database ?? "",
        selectedItem?.databaseSchema ?? "",
        selectedItem.table,
        true,
        BehaviourEventMedium.Button,
        MetricsEventSpace.Workspace,
      );
    }
  }
</script>

<div class="w-[544px] p-6 flex flex-col gap-y-4">
  <h2 class="text-lead">Pick a table to power your first dashboard</h2>
  <ConnectorExplorer {store} border />
  {#if selectedItem?.table}
    <Button type="primary" large on:click={createDashboard}>
      Create dashboard
    </Button>
  {:else}
    <Button type="secondary" large on:click={skip}>
      Or, start with a blank project
    </Button>
  {/if}
</div>
