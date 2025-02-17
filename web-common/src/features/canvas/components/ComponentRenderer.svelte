<script lang="ts">
  import { isChartComponentType } from "@rilldata/web-common/features/canvas/components/util";
  import { Chart } from "./charts";
  import { Image } from "./image";
  import { KPI } from "./kpi";
  import { Markdown } from "./markdown";
  import { Table } from "./table";
  import { KPIGrid } from "@rilldata/web-common/features/canvas/components/kpi-grid";
  import {
    createQueryServiceResolveComponent,
    type V1ComponentSpecRendererProperties,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let renderer: string;
  export let componentName: string;

  const components = new Map([
    ["kpi", KPI],
    ["kpi_grid", KPIGrid],
    ["table", Table],
    ["markdown", Markdown],
    ["image", Image],
  ]);

  $: ({ instanceId } = $runtime);

  $: componentQuery = createQueryServiceResolveComponent(
    instanceId,
    componentName,
    { args: {} },
  );
  $: componentData = $componentQuery?.data;
  $: rendererProperties =
    componentData?.rendererProperties as V1ComponentSpecRendererProperties;

  $: Component = components.get(renderer);
</script>

{#if rendererProperties}
  {#if isChartComponentType(renderer)}
    <Chart {rendererProperties} {renderer} />
  {:else}
    <svelte:component this={Component} {rendererProperties} />
  {/if}
{/if}
