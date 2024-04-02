<script lang="ts">
  import { page } from "$app/stores";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { MetricsWorkspace } from "@rilldata/web-common/features/metrics-views";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";

  export let data;

  $: metricViewName = $page.params.name;

  const { readOnly } = featureFlags;

  onMount(() => {
    if ($readOnly) {
      throw error(404, "Page not found");
    }
  });

  $: yaml = data.file?.blob || "";

  $: initLocalUserPreferenceStore(metricViewName);
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

{#if yaml !== undefined}
  <MetricsWorkspace metricsDefName={metricViewName} />
{/if}
