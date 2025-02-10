<script lang="ts">
  import { Database } from "lucide-svelte";
  import { slide } from "svelte/transition";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import { LIST_SLIDE_DURATION as duration } from "../../../layout/config";
  import type { V1AnalyzedConnector } from "../../../runtime-client";
  import type { ConnectorExplorerStore } from "../connector-explorer-store";
  import DatabaseSchemaEntry from "./DatabaseSchemaEntry.svelte";
  import { useDatabaseSchemas } from "./selectors";

  export let instanceId: string;
  export let connector: V1AnalyzedConnector;
  export let database: string;
  export let store: ConnectorExplorerStore;

  $: connectorName = connector?.name as string;
  $: expandedStore = store.getItem({
    connector: connectorName,
    database,
  });
  $: expanded = $expandedStore;
  $: databaseSchemasQuery = useDatabaseSchemas(
    instanceId,
    connector?.name as string,
    database,
  );

  $: ({ data, error, isLoading } = $databaseSchemasQuery);
</script>

<li aria-label={database} class="database-entry">
  {#if database}
    <button
      class="database-entry-header"
      class:open={expanded}
      on:click={() => store.toggleItem({ connector: connectorName, database })}
    >
      <CaretDownIcon
        className="transform transition-transform text-gray-400 {expanded
          ? 'rotate-0'
          : '-rotate-90'}"
        size="14px"
      />
      <Database size="14px" class="shrink-0 text-gray-400" />
      <span class="truncate">
        {database}
      </span>
    </button>
  {/if}

  <ol transition:slide={{ duration }}>
    {#if expanded}
      {#if error}
        <span class="message">Error: {error.response.data?.message}</span>
      {:else if isLoading}
        <span class="message">Loading schemas...</span>
      {:else if data}
        {#if data.length === 0}
          <span class="message">No schemas found</span>
        {:else}
          {#each data as schema (schema)}
            <DatabaseSchemaEntry
              {instanceId}
              {connector}
              {database}
              {store}
              databaseSchema={schema ?? ""}
            />
          {/each}
        {/if}
      {/if}
    {/if}
  </ol>
</li>

<style lang="postcss">
  .database-entry {
    @apply w-full justify-between;
    @apply flex flex-col;
  }

  .database-entry-header {
    @apply h-6 pl-[22px] pr-2;
    @apply flex items-center gap-x-1;
  }

  button:hover {
    @apply bg-slate-100;
  }

  .message {
    @apply pl-2 pr-3.5 pt-2 pb-2 text-gray-500;
  }
</style>
