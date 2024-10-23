<!-- @component
The main feature-set component for dashboard filters
 -->
<script lang="ts">
  import TimeRangeReadOnly from "@rilldata/web-common/features/dashboards/filters/TimeRangeReadOnly.svelte";
  import { allDimensions } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimensions";
  import { allMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { getMapFromArray } from "@rilldata/utils";
  import type {
    V1Expression,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { flip } from "svelte/animate";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getDimensionFilters } from "../state-managers/selectors/dimension-filters";
  import { getMeasureFilters } from "../state-managers/selectors/measure-filters";
  import DimensionFilterReadOnlyChip from "./dimension-filters/DimensionFilterReadOnlyChip.svelte";
  import MeasureFilterReadOnlyChip from "./measure-filters/MeasureFilterReadOnlyChip.svelte";

  export let exploreName: string;
  export let filters: V1Expression | undefined;
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let timeRange: V1TimeRange | undefined;
  export let comparisonTimeRange: V1TimeRange | undefined = undefined;

  $: validExploreSpecs = useExploreValidSpec($runtime.instanceId, exploreName);

  // Get dimension filters
  $: dimensions = allDimensions({
    validExplore: $validExploreSpecs.data?.explore,
    validMetricsView: $validExploreSpecs.data?.metricsView,
  });
  $: dimensionIdMap = getMapFromArray(
    dimensions,
    (dimension) => dimension.name as string,
  );
  $: dimensionFilters = getDimensionFilters(dimensionIdMap, filters);

  // Get measure filters
  $: measures = allMeasures({
    validExplore: $validExploreSpecs.data?.explore,
    validMetricsView: $validExploreSpecs.data?.metricsView,
  });
  $: measureIdMap = getMapFromArray(
    measures,
    (measure) => measure.name as string,
  );
  $: measureFilters = getMeasureFilters(
    measureIdMap,
    dimensionThresholdFilters,
  );
</script>

<div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2 items-center">
  {#if timeRange}
    <TimeRangeReadOnly {timeRange} {comparisonTimeRange} />
  {/if}
  {#if dimensionFilters.length > 0}
    {#each dimensionFilters as { name, label, selectedValues, isInclude } (name)}
      {@const dimension = dimensions.find((d) => d.name === name)}
      <div animate:flip={{ duration: 200 }}>
        {#if dimension?.column || dimension?.expression}
          <DimensionFilterReadOnlyChip
            label={label ?? name}
            values={selectedValues}
            {isInclude}
          />
        {/if}
      </div>
    {/each}
  {/if}
  {#if measureFilters.length > 0}
    {#each measureFilters as { name, label, dimensionName, filter } (name)}
      <div animate:flip={{ duration: 200 }}>
        <MeasureFilterReadOnlyChip
          label={label ?? name}
          {dimensionName}
          {filter}
        />
      </div>
    {/each}
  {/if}
</div>
