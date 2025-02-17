<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";

  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { ChartMetadata } from "@rilldata/web-common/features/canvas/components/charts/types";
  import { chartMetadata } from "@rilldata/web-common/features/canvas/components/charts/util";
  import { type CanvasComponentType } from "@rilldata/web-common/features/canvas/components/types";
  import { type CanvasComponentObj } from "@rilldata/web-common/features/canvas/components/util";

  export let componentType: CanvasComponentType;
  export let component: CanvasComponentObj;

  async function selectChartType(chartType: ChartMetadata) {
    component.updateChartType(chartType.type);
  }
</script>

<div class="section">
  <InputLabel small label="Chart type" id="chart-components" />
  <div class="chart-icons">
    {#each chartMetadata as chart}
      <Tooltip distance={8} location="right">
        <Button
          square
          small
          type="secondary"
          selected={componentType === chart.type}
          on:click={() => selectChartType(chart)}
        >
          <svelte:component this={chart.icon} size="20px" />
        </Button>
        <TooltipContent slot="tooltip-content">
          {chart.title}
        </TooltipContent>
      </Tooltip>
    {/each}
  </div>
</div>

<style lang="postcss">
  .section {
    @apply px-5 flex flex-col gap-y-2 pt-2;
    @apply border-t border-gray-200;
  }

  .chart-icons {
    @apply flex border-2 px-2 py-1 gap-x-4;
  }
</style>
