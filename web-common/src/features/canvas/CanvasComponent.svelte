<script lang="ts" context="module">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import ComponentRenderer from "@rilldata/web-common/features/canvas/components/ComponentRenderer.svelte";
  import {
    getComponentFilterProperties,
    isChartComponentType,
  } from "@rilldata/web-common/features/canvas/components/util";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { builderActions, getAttrs, type Builder } from "bits-ui";
</script>

<script lang="ts">
  export let i: number;
  export let builders: Builder[] = [];
  export let embed = false;
  export let selected = false;
  export let componentName: string;

  const {
    canvasEntity: {
      spec: { getComponentResourceFromName },
    },
  } = getCanvasStateManagers();

  let isHovered = false;

  $: component = getComponentResourceFromName(componentName);
  $: ({ renderer, rendererProperties } = $component ?? {});

  $: isChartType = isChartComponentType(renderer);

  $: title = rendererProperties?.title;
  $: description = rendererProperties?.description;
  $: componentFilters = getComponentFilterProperties(rendererProperties);

  function handleMouseEnter() {
    if (embed) return;
    isHovered = true;
  }

  function handleMouseLeave() {
    if (embed) return;
    isHovered = false;
  }
</script>

<div
  {...getAttrs(builders)}
  use:builderActions={{ builders }}
  role="presentation"
  data-index={i}
  class="canvas-component size-full"
  data-selected={selected}
  data-hovered={isHovered}
  class:!cursor-default={embed}
  style:z-index={renderer === "select" ? 100 : 0}
  on:contextmenu
  on:pointerenter
  on:pointerleave
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
>
  <div class="size-full relative">
    <div class="size-full overflow-hidden flex flex-col flex-none">
      <div class="size-full overflow-hidden flex flex-col flex-none relative">
        {#if !isChartType}
          <ComponentHeader {title} {description} filters={componentFilters} />
        {/if}
        {#if renderer && rendererProperties}
          <ComponentRenderer {renderer} {rendererProperties} {componentName} />
        {/if}
      </div>
    </div>
  </div>
</div>
