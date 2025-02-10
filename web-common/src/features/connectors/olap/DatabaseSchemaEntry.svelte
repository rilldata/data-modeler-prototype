<script lang="ts">
  import { Database, Folder } from "lucide-svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import type { V1AnalyzedConnector } from "../../../runtime-client";
  import type { ConnectorExplorerStore } from "../connector-explorer-store";
  import TableEntry from "./TableEntry.svelte";
  import { useTables } from "./selectors";

  export let instanceId: string;
  export let connector: V1AnalyzedConnector;
  export let database: string;
  export let databaseSchema: string;
  export let store: ConnectorExplorerStore;

  $: connectorName = connector?.name as string;

  $: expandedStore = store.getItem({
    connector: connectorName,
    database,
    databaseSchema,
  });
  $: expanded = $expandedStore;
  $: tablesQuery = useTables(
    instanceId,
    connectorName,
    database,
    databaseSchema,
  );
  $: ({ data } = $tablesQuery);

  $: typedData = data as
    | {
        name: string;
        database: string;
        databaseSchema: string;
        hasUnsupportedDataTypes: boolean;
      }[]
    | undefined;
</script>

<li aria-label={`${database}.${databaseSchema}`} class="database-schema-entry">
  <button
    class="database-schema-entry-header {database ? 'pl-[40px]' : 'pl-[22px]'}"
    class:open={expanded}
    on:click={() =>
      store.toggleItem({
        connector: connectorName,
        database,
        databaseSchema,
      })}
  >
    <CaretDownIcon
      className="transform transition-transform text-gray-400 {expanded
        ? 'rotate-0'
        : '-rotate-90'}"
      size="14px"
    />
    <!-- Some databases do not have a full "database -> databaseSchema -> table" hierarchy. 
      When there are only two organizational levels,the API returns "databaseSchema -> table". 
      However, in these cases, we should use a Database icon (not a Folder icon) to represent the organizational structure. -->
    {#if !database}
      <Database size="14px" class="shrink-0 text-gray-400" />
    {:else}
      <Folder size="14px" class="shrink-0 text-gray-400" />
    {/if}
    <span class="truncate">
      {databaseSchema}
    </span>
  </button>

  {#if expanded}
    {#if connector?.errorMessage}
      <div class="message">{connector.errorMessage}</div>
    {:else if !connector.driver || !connector.driver.name}
      <div class="message">Connector not found</div>
    {:else if !typedData || typedData.length === 0}
      <div class="message">No tables found</div>
    {:else if typedData.length > 0}
      <ol>
        {#each typedData as tableInfo (tableInfo)}
          <TableEntry
            {instanceId}
            driver={connector.driver.name}
            connector={connectorName}
            {database}
            {databaseSchema}
            table={tableInfo.name}
            hasUnsupportedDataTypes={tableInfo.hasUnsupportedDataTypes}
            {store}
          />
        {/each}
      </ol>
    {/if}
  {/if}
</li>

<style lang="postcss">
  .database-schema-entry {
    @apply w-full;
    @apply flex flex-col;
  }

  .database-schema-entry-header {
    @apply h-6 pr-2; /* left-padding is set dynamically above */
    @apply flex items-center gap-x-1;
  }

  button:hover {
    @apply bg-slate-100;
  }

  .message {
    @apply pl-2 pr-3.5 py-2;
    @apply text-gray-500;
  }
</style>
