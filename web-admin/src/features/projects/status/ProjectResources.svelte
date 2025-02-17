<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    createRuntimeServiceCreateTrigger,
    createRuntimeServiceListResources,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
    type V1ListResourcesResponse,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient, type Query } from "@tanstack/svelte-query";
  import Button from "web-common/src/components/button/Button.svelte";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";
  import RefreshAllSourcesAndModelsConfirmDialog from "./RefreshAllSourcesAndModelsConfirmDialog.svelte";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  let isConfirmDialogOpen = false;
  let isReconciling = false;

  const INITIAL_REFETCH_INTERVAL = 200; // Start at 200ms for immediate feedback
  const MAX_REFETCH_INTERVAL = 2_000; // Cap at 2s
  const BACKOFF_FACTOR = 1.5;
  let currentRefetchInterval = INITIAL_REFETCH_INTERVAL;

  $: ({ instanceId } = $runtime);

  function isResourceErrored(resource: V1Resource) {
    return !!resource.meta.reconcileError;
  }

  function isResourceReconciling(resource: V1Resource) {
    return (
      resource.meta.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
      resource.meta.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING
    );
  }

  function calculateRefetchInterval(
    currentInterval: number,
    data: V1ListResourcesResponse,
    query: Query<V1ListResourcesResponse>,
  ): number | false {
    if (query.state.error) return false;
    if (!data) return INITIAL_REFETCH_INTERVAL;

    const hasErrors = data.resources.some(isResourceErrored);
    const hasReconcilingResources = data.resources.some(isResourceReconciling);

    if (hasErrors || !hasReconcilingResources) {
      currentRefetchInterval = INITIAL_REFETCH_INTERVAL;
      return false;
    }

    currentRefetchInterval = Math.min(
      currentInterval * BACKOFF_FACTOR,
      MAX_REFETCH_INTERVAL,
    );
    return currentRefetchInterval;
  }

  $: resources = createRuntimeServiceListResources(
    instanceId,
    {
      // Ensure admins can see all resources, regardless of the security policy
      skipSecurityChecks: true,
    },
    {
      query: {
        select: (data: V1ListResourcesResponse) => ({
          ...data,
          // Filter out project parser and refresh triggers
          resources: data?.resources?.filter(
            (resource: V1Resource) =>
              resource.meta.name.kind !== ResourceKind.ProjectParser &&
              resource.meta.name.kind !== ResourceKind.RefreshTrigger,
          ),
        }),
        refetchInterval: (data, query) =>
          calculateRefetchInterval(currentRefetchInterval, data, query),
      },
    },
  );

  $: hasReconcilingResources = $resources.data?.resources?.some(
    isResourceReconciling,
  );

  $: isReconciling = Boolean(hasReconcilingResources);

  $: isRefreshButtonDisabled = hasReconcilingResources;

  function refreshAllSourcesAndModels() {
    isReconciling = false;

    void $createTrigger
      .mutateAsync({
        instanceId,
        data: { allSourcesModels: true },
      })
      .then(() => {
        currentRefetchInterval = INITIAL_REFETCH_INTERVAL;
        void queryClient.invalidateQueries(
          getRuntimeServiceListResourcesQueryKey(instanceId, undefined),
        );
      });
  }
</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Resources</h2>
    <Button
      type="secondary"
      on:click={() => {
        isConfirmDialogOpen = true;
      }}
      disabled={isRefreshButtonDisabled}
    >
      Refresh all sources and models
    </Button>
  </div>

  {#if $resources.isLoading}
    <DelayedSpinner isLoading={$resources.isLoading} size="16px" />
  {:else if $resources.isError}
    <div class="text-red-500">
      Error loading resources: {$resources.error?.message}
    </div>
  {:else if $resources.data}
    <ProjectResourcesTable data={$resources?.data?.resources} {isReconciling} />
  {/if}
</section>

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>
