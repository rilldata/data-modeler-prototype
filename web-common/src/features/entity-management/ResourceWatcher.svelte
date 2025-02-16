<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { WatchFilesClient } from "@rilldata/web-common/features/entity-management/WatchFilesClient";
  import { WatchResourcesClient } from "@rilldata/web-common/features/entity-management/WatchResourcesClient";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { onMount } from "svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import Banner from "@rilldata/web-common/components/banner/Banner.svelte";

  const fileWatcher = new WatchFilesClient().client;
  const resourceWatcher = new WatchResourcesClient().client;
  const {
    retryAttempts: fileAttempts,
    closed: fileWatcherClosed,
    error: fileError,
  } = fileWatcher;
  const {
    retryAttempts: resourceAttempts,
    closed: resourceWatcherClosed,
    error: resourceError,
  } = fileWatcher;

  export let host: string;
  export let instanceId: string;

  $: fileWatcher.watch(`${host}/v1/instances/${instanceId}/files/watch`);

  $: resourceWatcher.watch(
    `${host}/v1/instances/${instanceId}/resources/-/watch`,
  );

  $: failed = $fileAttempts >= 2 || $resourceAttempts >= 2;

  onMount(() => {
    void fileArtifacts.init(queryClient, instanceId);

    return () => {
      fileWatcher.close();
      resourceWatcher.close();
    };
  });

  function handleVisibilityChange() {
    if (document.visibilityState === "visible") {
      fileWatcher.heartbeat();
      resourceWatcher.heartbeat();
    } else {
      fileWatcher.throttle(true);
      resourceWatcher.throttle(true);
    }
  }
</script>

<svelte:window
  on:visibilitychange={handleVisibilityChange}
  on:blur={() => {
    fileWatcher.throttle();
    resourceWatcher.throttle();
  }}
  on:click={() => {
    fileWatcher.heartbeat();
    resourceWatcher.heartbeat();
  }}
  on:keydown={() => {
    fileWatcher.heartbeat();
    resourceWatcher.heartbeat();
  }}
  on:focus={() => {
    fileWatcher.heartbeat();
    resourceWatcher.heartbeat();
  }}
/>

{#if failed}
  <ErrorPage
    fatal
    statusCode={500}
    header="Error connecting to runtime"
    body="Try restarting the server"
  />
{:else}
  {#if $fileWatcherClosed || $resourceWatcherClosed}
    <Banner
      banner={{
        message:
          "Connection closed due to inactivity. Interact with the page to reconnect.",
        type: "warning",
        iconType: "alert",
      }}
    />
  {:else if $fileError || $resourceError}
    <Banner
      banner={{
        message: `Runtime connection unstable. Made ${$fileError ? $fileAttempts : $resourceAttempts} attempt(s) to reconnect.`,
        type: "error",
        iconType: "alert",
      }}
    />
  {/if}
  <slot />
{/if}
