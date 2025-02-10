<script lang="ts">
  import { page } from "$app/stores";
  import ContextButton from "@rilldata/web-common/components/button/ContextButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import TableIcon from "../../../components/icons/TableIcon.svelte";
  import type { ConnectorExplorerStore } from "../connector-explorer-store";
  import TableMenuItems from "./TableMenuItems.svelte";
  import TableSchema from "./TableSchema.svelte";
  import UnsupportedTypesIndicator from "./UnsupportedTypesIndicator.svelte";
  import {
    makeFullyQualifiedTableName,
    makeTablePreviewHref,
  } from "./olap-config";

  export let instanceId: string;
  export let driver: string;
  export let connector: string;
  export let database: string; // The backend interprets an empty string as the default database
  export let databaseSchema: string; // The backend interprets an empty string as the default schema
  export let table: string;
  export let hasUnsupportedDataTypes: boolean;
  export let store: ConnectorExplorerStore;

  let contextMenuOpen = false;

  $: expandedStore = store.getItem({
    connector,
    database,
    databaseSchema,
    table,
  });
  $: showSchema = $expandedStore;

  const { allowContextMenu, allowNavigateToTable, allowShowSchema } = store;

  $: fullyQualifiedTableName = makeFullyQualifiedTableName(
    driver,
    database,
    databaseSchema,
    table,
  );
  $: tableId = `${connector}-${fullyQualifiedTableName}`;
  $: href = makeTablePreviewHref(
    driver,
    connector,
    database,
    databaseSchema,
    table,
  );

  $: open = $page.url.pathname === href;

  $: selected = store.isItemSelected({
    connector,
    database,
    databaseSchema,
    table,
  });

  $: element = allowNavigateToTable ? "a" : "button";
</script>

<li
  aria-label={tableId}
  class="table-entry group"
  class:open
  class:selected={$selected}
>
  <div
    class:pl-[58px]={database || !allowShowSchema}
    class="table-entry-header pl-10"
  >
    {#if allowShowSchema}
      <button
        on:click={() => {
          store.toggleItem({
            connector,
            database,
            databaseSchema,
            table,
          });
        }}
      >
        <CaretDownIcon
          className="flex-none transform transition-transform text-gray-400 {!showSchema &&
            '-rotate-90'}"
          size="14px"
        />
      </button>
    {/if}

    <svelte:element
      this={element}
      class="clickable-text"
      {...allowNavigateToTable ? { href } : {}}
      role="menuitem"
      tabindex="0"
      on:click={() => {
        store.toggleItem({
          connector,
          database,
          databaseSchema,
          table,
        });
      }}
    >
      <TableIcon size="14px" className="shrink-0 text-gray-400" />
      <span class="truncate">
        {table}
      </span>
    </svelte:element>

    {#if hasUnsupportedDataTypes}
      <UnsupportedTypesIndicator
        {instanceId}
        {connector}
        {database}
        {databaseSchema}
        {table}
      />
    {/if}

    {#if allowContextMenu}
      <DropdownMenu.Root bind:open={contextMenuOpen}>
        <DropdownMenu.Trigger asChild let:builder>
          <ContextButton
            id="more-actions-{tableId}"
            tooltipText="More actions"
            label="{tableId} actions menu trigger"
            builders={[builder]}
            suppressTooltip={contextMenuOpen}
          >
            <MoreHorizontal />
          </ContextButton>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content
          class="min-w-60"
          align="start"
          side="right"
          sideOffset={16}
        >
          <TableMenuItems {connector} {database} {databaseSchema} {table} />
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    {/if}
  </div>

  {#if allowShowSchema && showSchema}
    <TableSchema {connector} {database} {databaseSchema} {table} />
  {/if}
</li>

<style lang="postcss">
  .table-entry {
    @apply w-full justify-between;
    @apply flex flex-col;
  }

  .table-entry-header {
    @apply h-6 pr-2; /* left-padding is set dynamically above */
    @apply flex justify-between items-center gap-x-1;
  }

  .table-entry-header:hover {
    @apply bg-slate-100;
  }

  .open {
    @apply bg-slate-100;
  }

  .clickable-text {
    @apply flex grow items-center gap-x-1;
    @apply text-gray-900 truncate;
  }

  .selected {
    @apply bg-blue-50;
  }
</style>
