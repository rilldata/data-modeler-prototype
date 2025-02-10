<script lang="ts">
  import { dev } from "$app/environment";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import ResourceWatcher from "@rilldata/web-common/features/entity-management/ResourceWatcher.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { initPylonWidget } from "@rilldata/web-common/features/help/initPylonWidget";
  import { RillTheme } from "@rilldata/web-common/layout";
  import BlockingOverlayContainer from "@rilldata/web-common/layout/BlockingOverlayContainer.svelte";
  import type { ApplicationBuildMetadata } from "@rilldata/web-common/layout/build-metadata";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import {
    initPosthog,
    posthogIdentify,
  } from "@rilldata/web-common/lib/analytics/posthog";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    errorEventHandler,
    initMetrics,
  } from "@rilldata/web-common/metrics/initMetrics";
  import { localServiceGetMetadata } from "@rilldata/web-common/runtime-client/local-service";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { initializeNodeStoreContexts } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import type { Query } from "@tanstack/query-core";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import type { AxiosError } from "axios";
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";

  /** This function will initialize the existing node stores and will connect them
   * to the Node server.
   */
  initializeNodeStoreContexts();

  const appBuildMetaStore: Writable<ApplicationBuildMetadata> =
    getContext("rill:app:metadata");

  queryClient.getQueryCache().config.onError = (
    error: AxiosError,
    query: Query,
  ) => errorEventHandler?.requestErrorEventHandler(error, query);
  initPylonWidget();

  let removeJavascriptListeners: () => void;
  onMount(async () => {
    const config = await localServiceGetMetadata();

    const shouldSendAnalytics =
      config.analyticsEnabled && !import.meta.env.VITE_PLAYWRIGHT_TEST && !dev;

    if (shouldSendAnalytics) {
      await initMetrics(config); // Proxies events through the Rill "intake" service
      initPosthog(config.version);
      posthogIdentify(config.userId, {
        installId: config.installId,
      });

      removeJavascriptListeners =
        errorEventHandler.addJavascriptErrorListeners();
    }

    featureFlags.set(false, "adminServer");
    featureFlags.set(config.readonly, "readOnly");

    // TODO: use TanStack Query, not this handmade store
    appBuildMetaStore.set({
      version: config.version,
      commitHash: config.buildCommit,
    });
  });

  /**
   * Async mount doesnt support an unsubscribe method.
   * So we need this to make sure javascript listeners for error handler is removed.
   */
  onMount(() => {
    return () => removeJavascriptListeners?.();
  });

  $: ({ host, instanceId } = $runtime);
</script>

<RillTheme>
  <QueryClientProvider client={queryClient}>
    <ResourceWatcher {host} {instanceId}>
      <div
        class="body h-screen w-screen overflow-hidden absolute flex flex-col"
      >
        <slot />
      </div>
    </ResourceWatcher>
  </QueryClientProvider>
</RillTheme>

{#if $overlay !== null}
  <BlockingOverlayContainer
    bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
  >
    <div slot="title" class="font-bold">
      {$overlay?.title}
    </div>
    <svelte:fragment slot="detail">
      {#if $overlay?.detail}
        <svelte:component
          this={$overlay.detail.component}
          {...$overlay.detail.props}
        />
      {/if}
    </svelte:fragment>
  </BlockingOverlayContainer>
{/if}

<NotificationCenter />

<style>
  /* Prevent trackpad navigation (like other code editors, like vscode.dev). */
  :global(body) {
    overscroll-behavior: none;
  }
</style>
