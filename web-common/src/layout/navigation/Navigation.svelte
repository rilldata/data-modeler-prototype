<script lang="ts" context="module">
  export const navigationOpen = (() => {
    const store = writable(true);
    return {
      ...store,
      toggle: () => store.update((open) => !open),
    };
  })();
</script>

<script lang="ts">
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { ModelAssets } from "@rilldata/web-common/features/models";
  import ProjectTitle from "@rilldata/web-common/features/project/ProjectTitle.svelte";
  import SourceAssets from "@rilldata/web-common/features/sources/navigation/SourceAssets.svelte";
  import { writable } from "svelte/store";
  import ChartAssets from "../../features/charts/ChartAssets.svelte";
  import CustomDashboardAssets from "../../features/custom-dashboards/CustomDashboardAssets.svelte";
  import DashboardAssets from "../../features/dashboards/DashboardAssets.svelte";
  import OtherFiles from "../../features/project/OtherFiles.svelte";
  import TableAssets from "../../features/tables/TableAssets.svelte";
  import { useIsModelingSupportedForCurrentOlapDriver } from "../../features/tables/selectors";
  import type { V1Instance, V1Resource } from "../../runtime-client";

  import { DEFAULT_NAV_WIDTH } from "../config";
  import Footer from "./Footer.svelte";
  import Resizer from "../Resizer.svelte";
  import SurfaceControlButton from "./SurfaceControlButton.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  export let instance: V1Instance;
  export let resources: V1Resource[];

  $: mapped = resources.reduce((acc, resource) => {
    if (!resource.meta?.name?.kind) return acc;

    const resources = acc.get(resource.meta.name.kind as ResourceKind);

    if (!resources) {
      acc.set(resource.meta.name.kind as ResourceKind, [resource]);
    } else {
      resources.push(resource);
    }

    return acc;
  }, new Map<ResourceKind, V1Resource[]>());

  const { customDashboards, readOnly } = featureFlags;

  let width = DEFAULT_NAV_WIDTH;
  let previousWidth: number;
  let container: HTMLElement;
  let resizing = false;

  $: isModelerEnabled = $readOnly === false;

  $: olapConnector = instance.olapConnector;
  $: isModelingSupportedForCurrentOlapDriver =
    useIsModelingSupportedForCurrentOlapDriver(instance.instanceId ?? "");

  function handleResize(
    e: UIEvent & {
      currentTarget: EventTarget & Window;
    },
  ) {
    const currentWidth = e.currentTarget.innerWidth;

    if (currentWidth < previousWidth && currentWidth < 768) {
      $navigationOpen = false;
    }

    previousWidth = currentWidth;
  }
</script>

<svelte:window on:resize={handleResize} />

<nav
  class="sidebar"
  class:hide={!$navigationOpen}
  class:resizing
  style:width="{width}px"
  bind:this={container}
>
  <Resizer
    min={DEFAULT_NAV_WIDTH}
    basis={DEFAULT_NAV_WIDTH}
    max={440}
    bind:dimension={width}
    bind:resizing
    side="right"
  />
  <div class="inner" style:width="{width}px">
    <ProjectTitle />

    <div class="scroll-container">
      <div class="nav-wrapper">
        {#if isModelerEnabled}
          <TableAssets />

          {#if olapConnector === "duckdb"}
            <SourceAssets />
          {/if}
          {#if $isModelingSupportedForCurrentOlapDriver.data}
            <ModelAssets assets={mapped.get(ResourceKind.Model) ?? []} />
          {/if}
        {/if}
        <DashboardAssets />
        {#if $customDashboards}
          <ChartAssets />
          <CustomDashboardAssets />
        {/if}
        {#if isModelerEnabled}
          <OtherFiles />
        {/if}
      </div>
    </div>
    <Footer />
  </div>
</nav>

<SurfaceControlButton
  {resizing}
  navWidth={width}
  navOpen={$navigationOpen}
  on:click={navigationOpen.toggle}
/>

<style lang="postcss">
  .sidebar {
    @apply flex flex-col flex-none relative overflow-hidden;
    @apply h-screen border-r z-0;
    transition-property: width;
    will-change: width;
  }

  .inner {
    @apply h-full overflow-hidden flex flex-col;
    will-change: width;
  }

  .nav-wrapper {
    @apply flex flex-col h-fit w-full gap-y-3;
  }

  .scroll-container {
    @apply overflow-y-auto overflow-x-hidden;
    @apply transition-colors h-full bg-white pb-8;
  }

  .sidebar:not(.resizing) {
    transition-duration: 400ms;
    transition-timing-function: ease-in-out;
  }

  .hide {
    width: 0px !important;
  }
</style>
