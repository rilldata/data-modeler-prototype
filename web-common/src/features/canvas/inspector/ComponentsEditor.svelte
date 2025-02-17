<script lang="ts">
  import {
    getComponentObj,
    getHeaderForComponent,
    isCanvasComponentType,
  } from "@rilldata/web-common/features/canvas/components/util";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import VegaConfigInput from "./chart/VegaConfigInput.svelte";
  import ComponentTabs from "./ComponentTabs.svelte";
  import FiltersMapper from "./filters/FiltersMapper.svelte";
  import ParamMapper from "./ParamMapper.svelte";

  export let selectedComponentIndex: number;
  export let fileArtifact: FileArtifact;

  const {
    canvasEntity: {
      spec: { getComponentFromIndex, getComponentNameFromIndex },
    },
  } = getCanvasStateManagers();
  let currentTab: string;

  $: componentSpec = getComponentFromIndex(selectedComponentIndex);
  $: componentName = getComponentNameFromIndex(selectedComponentIndex);

  $: ({ renderer, rendererProperties } = $componentSpec || {});

  $: componentType = isCanvasComponentType(renderer) ? renderer : null;
  $: path = ["items", selectedComponentIndex, "component", componentType || ""];

  $: component =
    componentType && rendererProperties
      ? getComponentObj(fileArtifact, path, componentType, rendererProperties)
      : null;
</script>

<SidebarWrapper
  type="secondary"
  disableHorizontalPadding
  title={getHeaderForComponent(componentType)}
>
  <svelte:fragment slot="header">
    {#if componentType}
      {#key componentType}
        <ComponentTabs {componentType} bind:currentTab />
      {/key}
    {/if}
  </svelte:fragment>

  {#if componentType && $componentName && component && rendererProperties}
    {#key $componentName}
      {#if currentTab === "options"}
        <ParamMapper
          {component}
          {componentType}
          paramValues={rendererProperties}
        />
      {:else if currentTab === "filters"}
        <FiltersMapper
          selectedComponentName={$componentName}
          {component}
          paramValues={rendererProperties}
        />
      {:else if currentTab === "config"}
        <VegaConfigInput {component} paramValues={rendererProperties} />
      {/if}
    {/key}
  {:else if !renderer}
    <div class="inspector-center">
      <Spinner status={EntityStatus.Running} size="16px" />
    </div>
  {:else}
    <div class="inspector-center">
      Unknown Component {renderer}
    </div>
  {/if}
</SidebarWrapper>

<style lang="postcss">
  .inspector-center {
    @apply flex items-center justify-center h-full w-full;
  }
</style>
