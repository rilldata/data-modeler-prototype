<script lang="ts">
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import { getLocalComparison } from "@rilldata/web-common/features/canvas/components/kpi/util";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import {
    FormatPreset,
    NumberKind,
    numberKindForMeasure,
  } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import { type V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";
  import type { KPISpec } from ".";
  import KPISparkline from "./KPISparkline.svelte";
  import {
    useKPIComparisonTotal,
    useKPISparkline,
    useKPITotals,
    validateKPISchema,
  } from "./selector";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;
  export let topPadding = true;

  const ctx = getCanvasStateManagers();
  const {
    spec,
    timeControls: {
      showTimeComparison,
      selectedComparisonTimeRange,
      selectedTimeRange,
    },
  } = ctx.canvasEntity;

  let containerWidth: number;
  let containerHeight: number;

  $: kpiProperties = rendererProperties as KPISpec;

  $: ({
    metrics_view: metricsViewName,
    measure: measureName,
    sparkline,
    time_filters: timeFilters,
    comparison: comparisonOptions,
  } = kpiProperties);

  $: ({ showLocalTimeComparison, localComparisonTimeRange } =
    getLocalComparison(timeFilters));

  $: schema = validateKPISchema(ctx, kpiProperties);

  $: measure = spec.getMeasureForMetricView(measureName, metricsViewName);
  $: measureValue = useKPITotals(
    ctx,
    kpiProperties,
    timeAndFilterStore,
    $schema.isValid,
  );

  $: measureIsPercentage = $measure?.formatPreset === FormatPreset.PERCENTAGE;

  $: showSparkline = sparkline !== "none";
  $: isSparkRight = sparkline === "right";

  $: showComparison =
    ($showTimeComparison || showLocalTimeComparison) && comparisonOptions;
  $: comparisonValue = useKPIComparisonTotal(
    ctx,
    kpiProperties,
    timeAndFilterStore,
    $schema.isValid,
  );
  $: comparisonPercChange =
    $measureValue.data != null && $comparisonValue.data
      ? ($measureValue.data - $comparisonValue.data) / $comparisonValue.data
      : undefined;
  $: comparisonLabel =
    (localComparisonTimeRange?.name &&
      TIME_COMPARISON[localComparisonTimeRange.name]?.label) ||
    ($selectedComparisonTimeRange?.name &&
      TIME_COMPARISON[$selectedComparisonTimeRange.name]?.label);

  $: adjustForPad = topPadding ? 12 : 0;
  $: sparklineHeight = isSparkRight
    ? containerHeight
    : containerHeight -
      (showComparison && $comparisonValue?.data != null ? 100 : 60) -
      adjustForPad;
  $: sparklineWidth = isSparkRight ? containerWidth - 136 : containerWidth - 10;

  $: measureValueFormatter = $measure
    ? createMeasureValueFormatter<null>($measure, "big-number")
    : () => "no data";

  $: measureValueFormatted = $measureValue.data
    ? measureValueFormatter($measureValue.data)
    : "no data";

  $: numberKind = $measure ? numberKindForMeasure($measure) : NumberKind.ANY;

  $: sparklineData = useKPISparkline(
    ctx,
    kpiProperties,
    timeAndFilterStore,
    $schema.isValid && showSparkline,
  );
  $: sparkData = $sparklineData?.data || [];

  function getFormattedDiff(comparisonValue: number) {
    if (!$measureValue.data) return "";
    const delta = $measureValue.data - comparisonValue;
    return `${delta >= 0 ? "+" : ""}${measureValueFormatter(delta)}`;
  }
</script>

{#if $schema.isValid}
  {#if measure && !$measureValue.isFetching}
    <div
      bind:clientWidth={containerWidth}
      bind:clientHeight={containerHeight}
      class="flex h-full w-full bg-white items-center {isSparkRight
        ? 'flex-row'
        : 'flex-col'}"
      class:pt-2={topPadding && !isSparkRight && showSparkline}
      class:justify-center={!showSparkline || !sparkData.length}
    >
      <div
        class="flex flex-col {isSparkRight
          ? 'w-36 justify-center items-start pl-4 line-clamp-2'
          : 'w-full'} {!showSparkline || !isSparkRight ? 'items-center' : ''}"
      >
        <h2 class="measure-label">{$measure?.displayName || measureName}</h2>
        <div class="measure-value">{measureValueFormatted}</div>
        {#if showComparison && $comparisonValue.data}
          <div class="flex items-baseline gap-x-3 text-sm">
            {#if comparisonOptions?.includes("previous") && $comparisonValue.data != null}
              <div role="complementary" class="comparison-value">
                {measureValueFormatter($comparisonValue.data)}
              </div>
            {/if}
            {#if comparisonOptions?.includes("delta")}
              <div role="complementary" class="comparison-value">
                {#if $comparisonValue.data != null}
                  <span
                    class:text-red-500={$measureValue.data &&
                      $measureValue.data - $comparisonValue.data < 0}
                    >{getFormattedDiff($comparisonValue.data)}</span
                  >
                {:else}
                  <span
                    class="ui-copy-disabled-faint italic"
                    style:font-size=".9em">no change</span
                  >
                {/if}
              </div>
            {/if}
            {#if comparisonOptions?.includes("percent_change") && comparisonPercChange != null && !measureIsPercentage}
              <div
                role="complementary"
                class="w-fit font-semibold ui-copy-inactive"
                class:text-red-500={$measureValue.data &&
                  $measureValue.data < 0}
              >
                <PercentageChange
                  color="text-gray-500"
                  showPosSign
                  tabularNumber={false}
                  value={formatMeasurePercentageDifference(
                    comparisonPercChange,
                  )}
                />
              </div>
            {/if}
          </div>
          {#if comparisonLabel}
            <div class="comparison-range">
              vs {comparisonLabel?.toLowerCase()}
            </div>
          {/if}
        {/if}
      </div>
      {#if containerHeight && containerWidth && showSparkline && sparkData.length && $selectedTimeRange?.interval}
        <KPISparkline
          {sparkData}
          {measureName}
          {sparklineHeight}
          {sparklineWidth}
          {isSparkRight}
          timeGrain={$selectedTimeRange.interval}
          {measureValueFormatter}
          {numberKind}
        />
      {/if}
    </div>
  {:else}
    <div class="flex items-center justify-center w-full h-full">
      <Spinner status={EntityStatus.Running} />
    </div>
  {/if}
{:else}
  <ComponentError error={$schema.error} />
{/if}

<style lang="postcss">
  .measure-label {
    @apply font-medium text-sm whitespace-normal;
    @apply pr-2 text-gray-700;
  }
  .measure-value {
    @apply text-3xl font-medium text-gray-700 pb-1;
  }
  .comparison-range {
    @apply text-sm text-gray-500;
  }

  .comparison-value {
    @apply w-fit max-w-full overflow-hidden;
    @apply font-semibold text-ellipsis text-gray-500;
  }
</style>
