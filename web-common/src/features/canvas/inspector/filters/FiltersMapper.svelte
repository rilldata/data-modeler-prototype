<script lang="ts">
  import { type CanvasComponentObj } from "@rilldata/web-common/features/canvas/components/util";
  import DimensionFiltersInput from "@rilldata/web-common/features/canvas/inspector/filters/DimensionFiltersInput.svelte";
  import TimeFiltersInput from "@rilldata/web-common/features/canvas/inspector/filters/TimeFiltersInput.svelte";
  import { type V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";

  export let selectedComponentName: string;
  export let component: CanvasComponentObj;
  export let paramValues: V1ComponentSpecRendererProperties;

  $: localParamValues = localParamValues || {};
  let oldParamValuesRef: V1ComponentSpecRendererProperties = {};

  // TODO: Make this robust possibly a store.
  $: if (JSON.stringify(paramValues) !== JSON.stringify(oldParamValuesRef)) {
    localParamValues = structuredClone(paramValues) || {};
    oldParamValuesRef = paramValues;
  }

  $: inputParams = component.inputParams().filter;

  $: metricsView =
    "metrics_view" in paramValues ? paramValues.metrics_view : null;

  onMount(() => {
    localParamValues = structuredClone(paramValues) || {};
  });
</script>

<div>
  {#each Object.entries(inputParams) as [key, config]}
    <div class="component-param">
      {#if config.type === "time_filters"}
        <TimeFiltersInput
          {selectedComponentName}
          id={key}
          timeFilter={localParamValues[key]}
          showComparison={config?.meta?.hasComparison}
          showGrain={config?.meta?.hasGrain}
          onChange={async (filter) => {
            localParamValues[key] = filter;
            component.updateProperty(key, localParamValues[key]);
          }}
        />
      {:else if config.type == "dimension_filters" && metricsView}
        <DimensionFiltersInput
          {metricsView}
          {selectedComponentName}
          id={key}
          filter={localParamValues[key]}
          onChange={async (filter) => {
            localParamValues[key] = filter;
            component.updateProperty(key, localParamValues[key]);
          }}
        />
      {/if}
    </div>
  {/each}
</div>

<style lang="postcss">
  .component-param {
    @apply py-3 px-5;
    @apply border-t border-gray-200;
  }
</style>
