<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasThemeProvider from "@rilldata/web-common/features/canvas/CanvasThemeProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/canvas/state-managers/StateManagersProvider.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: canvasName = data.dashboardName;
</script>

<StateManagersProvider {canvasName}>
  <CanvasThemeProvider>
    <CanvasDashboardEmbed
      showFilterBar={data.dashboard.spec?.filtersEnabled}
      items={data.dashboard.spec?.items || []}
      spec={data.dashboard.spec || {}}
    />
  </CanvasThemeProvider>
</StateManagersProvider>
