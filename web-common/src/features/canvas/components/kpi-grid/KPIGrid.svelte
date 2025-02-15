<script lang="ts">
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";
  import type { KPIGridSpec } from ".";
  import ComponentError from "../ComponentError.svelte";
  import type { KPISpec } from "../kpi";
  import KPI from "../kpi/KPI.svelte";
  import { validateKPIGridSchema } from "./selector";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;

  let containerWidth: number;
  let containerHeight: number;
  let kpis: KPISpec[];

  $: kpiGridProperties = rendererProperties as KPIGridSpec;
  $: schema = validateKPIGridSchema(kpiGridProperties);

  // Convert measures to KPI specs
  $: kpis = (kpiGridProperties.measures || []).map((measure) => ({
    metrics_view: kpiGridProperties.metrics_view,
    measure,
    sparkline: kpiGridProperties.sparkline,
    comparison: kpiGridProperties.comparison,
    dimension_filters: kpiGridProperties.dimension_filters,
    time_filters: kpiGridProperties.time_filters,
  }));

  // Calculate individual KPI width based on container width and number of KPIs
  $: kpiWidth = containerWidth ? Math.floor(containerWidth / kpis.length) : 0;
</script>

{#if schema.isValid}
  <div
    bind:clientWidth={containerWidth}
    bind:clientHeight={containerHeight}
    class="flex flex-row w-full h-full bg-white py-4"
  >
    {#each kpis as kpi, i}
      <div
        style="width: {kpiWidth}px;"
        class="border-r border-gray-200"
        class:border-r-0={i === kpis.length - 1}
      >
        <KPI rendererProperties={kpi} topPadding={false} {timeAndFilterStore} />
      </div>
    {/each}
  </div>
{:else}
  <ComponentError error={schema.error} />
{/if}

<style lang="postcss">
  /* Add any custom styles here if needed */
</style>
