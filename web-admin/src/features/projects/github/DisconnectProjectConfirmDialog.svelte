<script lang="ts">
  import {
    createAdminServiceUploadProjectAssets,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { extractGithubDisconnectError } from "@rilldata/web-admin/features/projects/github/github-errors";
  import { invalidateProjectQueries } from "@rilldata/web-admin/features/projects/invalidations";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import { eventBus } from "@rilldata/events";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { AxiosError } from "axios";

  export let open = false;
  export let organization: string;
  export let project: string;

  const deleteProjectConnection = createAdminServiceUploadProjectAssets();
  $: ({ error, isLoading } = $deleteProjectConnection);
  $: parsedError = extractGithubDisconnectError(
    error as unknown as AxiosError<RpcStatus>,
  );

  async function onDisconnect() {
    await $deleteProjectConnection.mutateAsync({
      organization,
      project,
      data: {},
    });
    open = false;

    void invalidateProjectQueries($runtime.instanceId, organization, project);

    eventBus.emit("notification", {
      message: `Disconnected github repo`,
      type: "success",
    });
    void behaviourEvent?.fireGithubIntentEvent(
      BehaviourEventAction.GithubDisconnect,
      {
        is_fresh_connection: true,
      },
    );
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Disconnect from GitHub?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          This project will be disconnected from GitHub and no longer under
          version control. Are you sure want to disconnect?
          <a
            href="https://docs.rilldata.com/deploy/deploy-dashboard/github-101"
            target="_blank"
            class="text-primary-600"
          >
            Learn more ->
          </a>
        </div>
      </AlertDialogDescription>
      {#if parsedError?.message && !isLoading}
        <div class="text-red-500 text-sm py-px">
          {parsedError.message}
        </div>
      {/if}
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="secondary"
        on:click={() => (open = false)}
        disabled={isLoading}>Cancel</Button
      >
      <Button
        type="primary"
        on:click={onDisconnect}
        loading={isLoading}
        disabled={isLoading}
      >
        Yes, disconnect
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
