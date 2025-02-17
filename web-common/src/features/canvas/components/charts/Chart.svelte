<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import type { ChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import { getComponentFilterProperties } from "@rilldata/web-common/features/canvas/components/util";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type {
    MetricsViewSpecMeasureV2,
    V1ComponentSpecRendererProperties,
  } from "@rilldata/web-common/runtime-client";
  import type { View } from "svelte-vega";
  import type { Readable } from "svelte/store";
  import { getChartData, validateChartSchema } from "./selector";
  import type { ChartType } from "./types";
  import { generateSpec, getChartTitle, mergedVlConfig } from "./util";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let renderer: string;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;

  const ctx = getCanvasStateManagers();
  const {
    canvasEntity: {
      spec: { getMeasureForMetricView },
    },
  } = ctx;

  let viewVL: View;

  $: chartConfig = rendererProperties as ChartSpec;
  $: chartType = renderer as ChartType;

  $: schema = validateChartSchema(ctx, chartConfig);

  $: data = getChartData(ctx, chartConfig, timeAndFilterStore);
  $: hasNoData = !$data.isFetching && $data.data.length === 0;

  $: spec = generateSpec(chartType, chartConfig, $data);

  $: componentFilters = getComponentFilterProperties(rendererProperties);

  $: measure = getMeasureForMetricView(
    chartConfig.y?.field,
    chartConfig.metrics_view,
  );

  $: measureName = $measure?.name || "measure";

  $: measureFormatter = createMeasureValueFormatter<null | undefined>(
    $measure as MetricsViewSpecMeasureV2,
  );

  $: config = chartConfig.vl_config
    ? mergedVlConfig(chartConfig.vl_config)
    : undefined;

  $: title = chartConfig?.title || getChartTitle(chartConfig, $data);
  $: description = chartConfig?.description;
</script>

{#if $schema.isValid}
  {#if $data.isFetching}
    <div class="flex items-center h-full w-full">
      <Spinner status={EntityStatus.Running} size="16px" />
    </div>
  {:else if $data.error}
    <div class="text-red-500">{$data.error.message}</div>
  {:else}
    <ComponentHeader
      faint={!chartConfig?.title}
      {title}
      {description}
      filters={componentFilters}
    />
    {#if hasNoData}
      <div
        class="flex w-full h-full p-2 text-xl ui-copy-disabled items-center justify-center"
      >
        No Data to Display
      </div>
    {:else}
      <VegaLiteRenderer
        bind:viewVL
        canvasDashboard
        data={{ "metrics-view": $data.data }}
        {spec}
        expressionFunctions={{
          [measureName]: { fn: (val) => measureFormatter(val) },
        }}
        {config}
      />
    {/if}
  {/if}
{:else}
  <ComponentError error={$schema.error} />
{/if}
