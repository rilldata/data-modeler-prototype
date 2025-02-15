<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import { KPIGrid } from "@rilldata/web-common/features/canvas/components/kpi-grid";
  import {
    isCanvasComponentType,
    isChartComponentType,
  } from "@rilldata/web-common/features/canvas/components/util";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";
  import { Chart } from "./charts";
  import { Image } from "./image";
  import { KPI } from "./kpi";
  import { Markdown } from "./markdown";
  import { Table } from "./table";

  export let renderer: string;
  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let componentName: string;

  const ctx = getCanvasStateManagers();
  const { componentTimeAndFilterStore } = ctx.canvasEntity;

  const filterableComponents = new Map([
    ["kpi", KPI],
    ["kpi_grid", KPIGrid],
    ["table", Table],
  ]);

  const nonFilterableComponents = new Map([
    ["markdown", Markdown],
    ["image", Image],
  ]);

  $: isFilterable = filterableComponents.has(renderer);

  let timeAndFilterStore: Readable<TimeAndFilterStore> | undefined;
  $: if (
    (isChartComponentType(renderer) || isFilterable) &&
    rendererProperties?.metrics_view
  ) {
    timeAndFilterStore = componentTimeAndFilterStore(componentName);
  }
</script>

{#if rendererProperties && isCanvasComponentType(renderer)}
  {#if isChartComponentType(renderer) && timeAndFilterStore}
    <Chart {rendererProperties} {renderer} {timeAndFilterStore} />
  {:else if isFilterable && timeAndFilterStore}
    <svelte:component
      this={filterableComponents.get(renderer)}
      {rendererProperties}
      {timeAndFilterStore}
    />
  {:else}
    <svelte:component
      this={nonFilterableComponents.get(renderer)}
      {rendererProperties}
    />
  {/if}
{:else}
  <ComponentError error="Invalid component type" />
{/if}
