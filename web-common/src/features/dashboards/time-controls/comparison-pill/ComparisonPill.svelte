<script lang="ts">
  import * as Elements from "../super-pill/components";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    DashboardTimeControls,
    TimeComparisonOption,
    TimeRange,
  } from "@rilldata/web-common/lib/time/types";
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { DateTime, Interval } from "luxon";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";

  export let allTimeRange: TimeRange;
  export let selectedTimeRange: DashboardTimeControls | undefined;
  export let showTimeComparison: boolean;
  export let selectedComparisonTimeRange: DashboardTimeControls | undefined;

  const ctx = getStateManagers();
  const metricsViewQuery = useMetricsView(ctx);
  const {
    metricsViewName,
    selectors: {
      timeRangeSelectors: { timeComparisonOptionsState },
    },
  } = ctx;

  $: metricViewName = $metricsViewName;

  $: dashboardStore = useDashboardStore(metricViewName);

  $: activeTimeZone = $dashboardStore?.selectedTimezone;

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
      )
    : Interval.fromDateTimes(allTimeRange.start, allTimeRange.end);

  $: metricsViewSpec = $metricsViewQuery.data;

  $: activeTimeGrain = selectedTimeRange?.interval;

  function onSelectComparisonRange(
    name: TimeComparisonOption,
    start: Date,
    end: Date,
  ) {
    if (!showTimeComparison) {
      metricsExplorerStore.displayTimeComparison(
        metricViewName,
        !showTimeComparison,
      );
    }
    if (metricsViewSpec)
      metricsExplorerStore.setSelectedComparisonRange(
        metricViewName,
        {
          name,
          start,
          end,
        },
        metricsViewSpec,
      );
  }
</script>

<div class="wrapper">
  <button
    class="flex gap-x-1.5 cursor-pointer"
    on:click={() => {
      metricsExplorerStore.displayTimeComparison(
        metricViewName,
        !showTimeComparison,
      );
    }}
  >
    <div class="pointer-events-none flex items-center gap-x-1.5">
      <Switch checked={showTimeComparison} id="comparing" small />

      <Label class="font-normal text-xs cursor-pointer" for="comparing">
        <span>Comparing</span>
      </Label>
    </div>
  </button>
  {#if activeTimeGrain && interval.isValid}
    <Elements.Comparison
      timeComparisonOptionsState={$timeComparisonOptionsState}
      selectedComparison={selectedComparisonTimeRange}
      showComparison={showTimeComparison}
      currentInterval={interval}
      grain={activeTimeGrain}
      zone={activeTimeZone}
      {onSelectComparisonRange}
    />
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply flex w-fit;
    @apply h-7 rounded-full;
    @apply overflow-hidden;
  }

  :global(.wrapper > button) {
    @apply border;
  }

  :global(.wrapper > button:not(:first-child)) {
    @apply -ml-[1px];
  }

  :global(.wrapper > button) {
    @apply border;
    @apply px-2 flex items-center justify-center bg-white;
  }

  :global(.wrapper > button:first-child) {
    @apply pl-2.5 rounded-l-full;
  }
  :global(.wrapper > button:last-child) {
    @apply pr-2.5 rounded-r-full;
  }

  :global(.wrapper > button:hover:not(:disabled)) {
    @apply bg-gray-50 cursor-pointer;
  }

  :global(.wrapper > [data-state="open"]) {
    @apply bg-gray-50 border-gray-400 z-50;
  }
</style>
