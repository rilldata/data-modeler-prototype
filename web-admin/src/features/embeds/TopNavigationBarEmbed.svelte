<script lang="ts">
  import { isErrorStoreEmpty } from "@rilldata/web-admin/components/errors/error-store";
  import BreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbItem.svelte";
  import TwoTieredBreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/TwoTieredBreadcrumbItem.svelte";
  import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import type {
    V1Resource,
    V1ResourceName,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import LastRefreshedDate from "../dashboards/listing/LastRefreshedDate.svelte";

  const dispatch = createEventDispatcher();

  export let instanceId: string;
  export let activeResource: V1ResourceName;

  const { twoTieredNavigation } = featureFlags;

  $: onProjectPage = !activeResource;
  $: onMetricsExplorerPage =
    !!activeResource &&
    activeResource.kind === ResourceKind.MetricsView.toString();

  // Dashboard breadcrumb
  $: dashboardsQuery = useValidDashboards(instanceId);
  $: ({ data: dashboards } = $dashboardsQuery);
  let currentResource: V1Resource;
  $: currentResource = dashboards?.find(
    (listing) => listing.meta.name.name === activeResource?.name,
  );
  $: currentResourceName = currentResource?.meta?.name?.name;

  $: breadcrumbOptions = dashboards?.reduce(
    (map, { meta, explore, canvas }) => {
      const name = meta.name.name;
      const isExplore = !!explore;
      return map.set(name.toLowerCase(), {
        label:
          (isExplore
            ? explore?.state?.validSpec?.displayName
            : canvas?.state?.validSpec?.displayName) || name,
      });
    },
    new Map(),
  );

  function onSelectResource(name: string) {
    // Because the breadcrumb only returns the identifying name, we need to look up the V1ResourceName (name + kind)
    const resource = dashboards?.find(
      (listing) => listing.meta.name.name.toLowerCase() === name,
    );
    if (!resource) {
      throw new Error(`Resource not found: ${name}`);
    }
    dispatch("select-resource", resource.meta.name);
  }
</script>

<div class="flex items-center w-full pr-4 py-1" class:border-b={!onProjectPage}>
  {#if $isErrorStoreEmpty}
    <nav>
      <ol class="flex items-center pl-4">
        {#if !onProjectPage}
          <div class="flex gap-x-2">
            <button
              class="text-gray-500 hover:text-gray-600"
              on:click={() => dispatch("go-home")}>Home</button
            >
            <span class="text-gray-600">/</span>
          </div>
        {/if}

        {#if currentResource}
          {#if $twoTieredNavigation}
            <TwoTieredBreadcrumbItem
              options={breadcrumbOptions}
              current={currentResourceName}
              onSelect={onSelectResource}
              isCurrentPage
            />
          {:else}
            <BreadcrumbItem
              options={breadcrumbOptions}
              current={currentResourceName}
              onSelect={onSelectResource}
              isCurrentPage
              isEmbedded
            />
          {/if}
        {/if}
      </ol>
    </nav>
  {:else}
    <div />
  {/if}

  {#if onMetricsExplorerPage}
    <div class="grow" />
    <div class="flex gap-x-4 items-center">
      <LastRefreshedDate dashboard={activeResource?.name} />
    </div>
  {/if}
</div>
