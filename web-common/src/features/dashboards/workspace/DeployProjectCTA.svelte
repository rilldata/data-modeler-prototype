<script lang="ts" context="module">
  export const allowPrimary = writable(false);
</script>

<script lang="ts">
  import { page } from "$app/stores";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import DeployIcon from "@rilldata/web-common/components/icons/DeployIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import PushToGitForDeployDialog from "@rilldata/web-common/features/project/PushToGitForDeployDialog.svelte";
  import { waitUntil } from "@rilldata/utils";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    createLocalServiceGetCurrentProject,
    createLocalServiceGetCurrentUser,
    createLocalServiceGetMetadata,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { get, writable } from "svelte/store";
  import { Button } from "../../../components/button";
  import Rocket from "svelte-radix/Rocket.svelte";
  import CloudIcon from "@rilldata/web-common/components/icons/CloudIcon.svelte";

  export let hasValidDashboard: boolean;

  let pushThroughGitOpen = false;
  let deployConfirmOpen = false;
  let deployCTAUrl: string;

  $: currentProject = createLocalServiceGetCurrentProject({
    query: {
      refetchOnWindowFocus: true,
    },
  });
  $: isDeployed = !!$currentProject.data?.project;

  $: allowPrimary.set(isDeployed || !hasValidDashboard);

  $: user = createLocalServiceGetCurrentUser();
  $: metadata = createLocalServiceGetMetadata();

  $: deployPageUrl = `${$page.url.protocol}//${$page.url.host}/deploy`;

  $: if (!$user.data?.user && $metadata.data) {
    deployCTAUrl = `${$metadata.data.loginUrl}?redirect=${deployPageUrl}`;
  } else {
    deployCTAUrl = deployPageUrl;
  }

  async function onShowRedeploy() {
    void behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeployIntent);

    await waitUntil(() => !get(currentProject).isFetching);
    if (get(currentProject).data?.project?.githubUrl) {
      pushThroughGitOpen = true;
      return;
    }

    window.open(deployCTAUrl, "_target");
  }

  function onShowDeploy() {
    deployConfirmOpen = true;
    void behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeployIntent);
  }
</script>

{#if isDeployed}
  <Tooltip distance={8}>
    <Button
      loading={$currentProject.isLoading}
      on:click={onShowRedeploy}
      type="secondary"
    >
      <CloudIcon size="16px" />
      Update
    </Button>
    <TooltipContent slot="tooltip-content">
      Push changes to Rill Cloud
    </TooltipContent>
  </Tooltip>
{:else}
  <Tooltip distance={8}>
    <Button
      loading={$currentProject.isLoading}
      on:click={onShowDeploy}
      type={hasValidDashboard ? "primary" : "secondary"}
    >
      <Rocket size="16px" />

      Deploy
    </Button>
    <TooltipContent slot="tooltip-content">
      Deploy this project to Rill Cloud
    </TooltipContent>
  </Tooltip>
{/if}

<AlertDialog bind:open={deployConfirmOpen}>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <div class="flex flex-row">
      <DeployIcon size="150px" />
      <div class="flex flex-col">
        <AlertDialogHeader>
          <AlertDialogTitle>Deploy this project</AlertDialogTitle>
          <AlertDialogDescription>
            You’re about to deploy to Rill Cloud, where you can set alerts,
            share dashboards, and more.
            <a
              href="https://www.rilldata.com/pricing"
              target="_blank"
              class="text-primary-600"
            >
              See pricing details
            </a>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter class="mt-5">
          <Button on:click={() => (deployConfirmOpen = false)} type="secondary">
            Back
          </Button>
          <Button
            on:click={() => (deployConfirmOpen = false)}
            type="primary"
            href={deployCTAUrl}
            target="_blank"
          >
            Continue
          </Button>
        </AlertDialogFooter>
      </div>
    </div>
  </AlertDialogContent>
</AlertDialog>

<PushToGitForDeployDialog
  bind:open={pushThroughGitOpen}
  githubUrl={$currentProject.data?.project?.githubUrl ?? ""}
  subpath={$currentProject.data?.project?.subpath ?? ""}
/>
