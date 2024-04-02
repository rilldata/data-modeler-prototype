<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { startWatchFilesClient } from "@rilldata/web-common/features/entity-management/watch-files-client";
  import { startWatchResourcesClient } from "@rilldata/web-common/features/entity-management/watch-resources-client";
  import { retainFeaturesFlags } from "@rilldata/web-common/features/feature-flags";
  import RillDeveloperLayout from "@rilldata/web-common/layout/RillDeveloperLayout.svelte";
  import { errorEventHandler } from "@rilldata/web-common/metrics/initMetrics";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { RuntimeUrl } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import type { Query } from "@tanstack/query-core";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import type { AxiosError } from "axios";
  import { onMount } from "svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  import { page } from "$app/stores";
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";

  export let data;

  queryClient.getQueryCache().config.onError = (
    error: AxiosError,
    query: Query,
  ) => errorEventHandler?.requestErrorEventHandler(error, query);

  beforeNavigate(retainFeaturesFlags);

  $: showNavigation = $page.route.id !== "/(application)/welcome";

  onMount(() => {
    const stopWatchFilesClient = startWatchFilesClient(queryClient);
    const stopWatchResourcesClient = startWatchResourcesClient(queryClient);
    const stopJavascriptErrorListeners =
      errorEventHandler?.addJavascriptErrorListeners();
    void fileArtifacts.init(queryClient, "default");

    return () => {
      stopWatchFilesClient();
      stopWatchResourcesClient();
      stopJavascriptErrorListeners?.();
    };
  });
</script>

<QueryClientProvider client={queryClient}>
  <RuntimeProvider host={RuntimeUrl} instanceId="default">
    <RillDeveloperLayout>
      {#if showNavigation}
        <Navigation instance={data.instance} resources={data.resources} />
      {/if}
      <slot />
    </RillDeveloperLayout>
  </RuntimeProvider>
</QueryClientProvider>
