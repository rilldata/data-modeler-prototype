<script lang="ts">
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";
  import { Tag } from "../../components/tag";
  import {
    createRuntimeServiceGetInstance,
    type V1AnalyzedConnector,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import type { ConnectorExplorerStore } from "./connector-explorer-store";
  import { symbolIconMapping } from "./connector-icon-mapping";
  import DatabaseExplorer from "./olap/DatabaseExplorer.svelte";

  export let connector: V1AnalyzedConnector;
  export let store: ConnectorExplorerStore;

  $: connectorName = connector?.name as string;
  $: expandedStore = store.getItem(connectorName);
  $: expanded = $expandedStore;
  $: ({ instanceId } = $runtime);
  $: instance = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  $: olapConnector = $instance.data?.instance?.olapConnector;
  $: isOlapConnector = olapConnector === connector.name;
  $: implementsOlap = connector.driver?.implementsOlap;
</script>

<!-- For now, only show OLAP connectors -->
{#if implementsOlap}
  {#if connector.name}
    <li aria-label={connector.name} class="connector-entry">
      <button
        class="connector-entry-header"
        on:click={() => {
          store.toggleItem(connectorName);
        }}
      >
        <CaretDownIcon
          className="transform transition-transform text-gray-400 {expanded
            ? 'rotate-0'
            : '-rotate-90'}"
          size="14px"
        />
        <span class="flex-none">
          {#if connector.driver?.name}
            <svelte:component
              this={symbolIconMapping[connector.driver.name]}
              size="16px"
            />
          {/if}
        </span>

        <h4>{connector.name}</h4>

        {#if isOlapConnector}
          <Tag height={16} class="ml-auto">OLAP</Tag>
        {/if}
      </button>

      {#if expanded}
        <DatabaseExplorer {instanceId} {connector} {store} />
      {/if}
    </li>
  {/if}
{/if}

<style lang="postcss">
  .connector-entry {
    @apply flex flex-col flex-none;
  }

  .connector-entry-header {
    @apply flex gap-x-1 items-center flex-none;
    @apply w-full px-2 h-6 outline-none;
    @apply z-10 bg-white;
  }

  button:hover {
    @apply bg-slate-100;
  }

  h4 {
    @apply text-xs font-medium;
  }
</style>
