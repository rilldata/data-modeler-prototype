<script lang="ts">
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import TimeRangeReadOnly from "@rilldata/web-common/features/dashboards/filters/TimeRangeReadOnly.svelte";
  import DimensionFilterReadOnlyChip from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterReadOnlyChip.svelte";
  import MeasureFilterReadOnlyChip from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterReadOnlyChip.svelte";
  import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
  import { getDimensionFilters } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
  import { getMeasureFilters } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
  import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { getMapFromArray } from "@rilldata/utils";
  import type {
    V1Expression,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";

  export let metricsViewName: string;
  export let filters: V1Expression | undefined;
  export let timeRange: V1TimeRange | undefined;
  export let comparisonTimeRange: V1TimeRange | undefined;

  // time range could be an empty object sometimes
  $: hasTimeRange = timeRange && Object.keys(timeRange).length > 0;
  $: filtersLength =
    (filters?.cond?.exprs?.length ?? 0) + (hasTimeRange ? 1 : 0);

  let whereFilter: V1Expression;
  let havingFilter: DimensionThresholdFilter[];
  $: {
    const { dimensionFilters, dimensionThresholdFilters } =
      splitWhereFilter(filters);
    whereFilter = dimensionFilters;
    havingFilter = dimensionThresholdFilters;
  }

  $: metricsView = useMetricsView($runtime.instanceId, metricsViewName);
  $: dimensionIdMap = getMapFromArray(
    $metricsView.data?.metricsView?.state?.validSpec?.dimensions ?? [],
    (dimension) => dimension.name,
  );
  $: measureIdMap = getMapFromArray(
    $metricsView.data?.metricsView?.state?.validSpec?.measures ?? [],
    (measure) => measure.name,
  );

  $: currentDimensionFilters = getDimensionFilters(dimensionIdMap, whereFilter);
  $: currentMeasureFilters = getMeasureFilters(measureIdMap, havingFilter);
</script>

<div class="flex flex-col gap-y-3">
  <MetadataLabel>Filters ({filtersLength})</MetadataLabel>
  <div class="flex flex-wrap gap-2">
    {#if filtersLength}
      {#if hasTimeRange}
        <TimeRangeReadOnly {timeRange} {comparisonTimeRange} />
      {/if}
      {#each currentDimensionFilters as { name, label, selectedValues, isInclude } (name)}
        {@const dimension = dimensionIdMap.get(name)}
        <div animate:flip={{ duration: 200 }}>
          {#if dimension?.column}
            <DimensionFilterReadOnlyChip
              label={label ?? name}
              values={selectedValues}
              {isInclude}
            />
          {/if}
        </div>
      {/each}
      {#if currentMeasureFilters.length > 0}
        {#each currentMeasureFilters as { name, label, dimensionName, filter } (name)}
          <div animate:flip={{ duration: 200 }}>
            <MeasureFilterReadOnlyChip
              label={label ?? name}
              {dimensionName}
              {filter}
            />
          </div>
        {/each}
      {/if}
    {:else}
      <div
        in:fly|local={{ duration: 200, x: 8 }}
        class="ui-copy-disabled grid items-center"
        style:min-height="26px"
      >
        No filters selected
      </div>
    {/if}
  </div>
</div>
