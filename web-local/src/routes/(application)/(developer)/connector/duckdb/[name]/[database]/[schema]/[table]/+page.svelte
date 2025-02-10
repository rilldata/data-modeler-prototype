<script lang="ts">
  import { page } from "$app/stores";
  import TablePreviewWorkspace from "@rilldata/web-common/features/connectors/olap/TablePreviewWorkspace.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";

  const { readOnly } = featureFlags;

  $: name = $page.params.name;
  $: database = $page.params.database;
  $: databaseSchema = $page.params.schema;
  $: table = $page.params.table;

  onMount(() => {
    if ($readOnly) {
      throw error(404, "Page not found");
    }
  });
</script>

<svelte:head>
  <title>Rill Developer | {table}</title>
</svelte:head>

<TablePreviewWorkspace connector={name} {database} {databaseSchema} {table} />
