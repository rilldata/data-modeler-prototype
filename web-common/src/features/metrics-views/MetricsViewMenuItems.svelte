<script lang="ts">
  import { goto } from "$app/navigation";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { extractFileName } from "@rilldata/utils";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WandIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import { createAndPreviewExplore } from "./create-and-preview-explore";

  export let filePath: string;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: metricsView = extractFileName(filePath);

  const dispatch = createEventDispatcher();
  const queryClient = useQueryClient();
  const { customDashboards, ai } = featureFlags;

  $: instanceId = $runtime.instanceId;
  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: resource = $resourceQuery.data;
  $: hasErrors = fileArtifact.getHasErrors(queryClient, instanceId);

  /**
   * Get the name of the dashboard's underlying model (if any).
   * Note that not all dashboards have an underlying model. Some dashboards are
   * underpinned by a source/table.
   */
  $: referenceModelName = $resourceQuery?.data?.meta?.refs?.filter(
    (ref) => ref.kind === ResourceKind.Model,
  )?.[0]?.name;

  const editModel = async () => {
    if (!referenceModelName) return;
    const artifact = fileArtifacts.findFileArtifact(
      ResourceKind.Model,
      referenceModelName,
    );
    if (!artifact) return;
    const previousScreenName = getScreenNameFromPage();
    await goto(`/files${artifact.path}`);
    await behaviourEvent.fireNavigationEvent(
      referenceModelName,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      previousScreenName,
      MetricsEventScreenName.Model,
    );
  };
</script>

{#if referenceModelName}
  <NavigationMenuItem on:click={editModel}>
    <Model slot="icon" />
    Edit model
  </NavigationMenuItem>
{/if}
{#if resource}
  <NavigationMenuItem
    on:click={() => createAndPreviewExplore(queryClient, instanceId, resource)}
  >
    <ExploreIcon slot="icon" />
    Generate dashboard
  </NavigationMenuItem>
{/if}
{#if $customDashboards}
  <NavigationMenuItem
    on:click={() => {
      dispatch("generate-chart", {
        metricsView,
      });
    }}
    disabled={$hasErrors}
  >
    <Explore slot="icon" />
    <div class="flex gap-x-2 items-center">
      Generate chart
      {#if $ai}
        with AI
        <WandIcon class="w-3 h-3" />
      {/if}
    </div>
    <svelte:fragment slot="description">
      {#if $hasErrors}
        Dashboard has errors
      {/if}
    </svelte:fragment>
  </NavigationMenuItem>
{/if}
