<script lang="ts">
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import { createRuntimeServiceAnalyzeConnectors } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import ConnectorEntry from "./ConnectorEntry.svelte";
  import type { ConnectorExplorerStore } from "./connector-explorer-store";

  export let store: ConnectorExplorerStore;
  export let border: boolean = false;

  $: ({ instanceId } = $runtime);

  $: connectors = createRuntimeServiceAnalyzeConnectors(instanceId, {
    query: {
      // sort alphabetically
      select: (data) => {
        if (!data?.connectors) return;

        const connectors = data.connectors.sort((a, b) =>
          (a?.name as string).localeCompare(b?.name as string),
        );
        return { connectors };
      },
    },
  });
  $: ({ data, error } = $connectors);
</script>

<div class="wrapper" class:outer-border={border}>
  {#if error}
    <span class="message">
      {error.message}
    </span>
  {:else if data?.connectors}
    {#if data.connectors.length === 0}
      <span class="message">
        No connectors found. Add data to get started!
      </span>
    {:else}
      <ol transition:slide={{ duration }}>
        {#each data.connectors as connector (connector.name)}
          <ConnectorEntry {connector} {store} />
        {/each}
      </ol>
    {/if}
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply overflow-auto px-0 pb-4;
  }

  .wrapper.outer-border {
    @apply border border-gray-300 rounded-md;
    @apply p-2;
  }

  .message {
    @apply pl-2 pr-3.5 py-2;
    @apply flex flex-none;
    @apply text-gray-500;
  }
</style>
