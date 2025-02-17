<script lang="ts">
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { ComponentFilterProperties } from "@rilldata/web-common/features/canvas/components/types";

  export let filters: ComponentFilterProperties;
  export let atleastOneFilter: boolean;
</script>

{#if atleastOneFilter}
  <Tooltip distance={8} location="top">
    <Filter size="16px" className="text-gray-400" />
    <TooltipContent slot="tooltip-content">
      <div class="p-2">
        <div class="font-semibold mb-2">Local Filters Applied</div>
        {#if filters.time_filters}
          <div>
            <span class="font-medium">Time Filters:</span>
            {filters.time_filters}
          </div>
        {/if}
        {#if filters.dimension_filters}
          <div>
            <span class="font-medium">Dimension Filters:</span>
            {filters.dimension_filters}
          </div>
        {/if}
      </div>
    </TooltipContent>
  </Tooltip>
{/if}
