<script lang="ts">
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { getComparisonOptionsForCanvas } from "@rilldata/web-common/features/canvas/filters/util";
  import { Comparison } from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components";
  import {
    TimeComparisonOption,
    TimeRangePreset,
    type DashboardTimeControls,
    type TimeRange,
  } from "@rilldata/web-common/lib/time/types";
  import { DateTime, Interval } from "luxon";

  export let allTimeRange: TimeRange;
  export let selectedTimeRange: DashboardTimeControls | undefined;
  export let selectedComparisonTimeRange: DashboardTimeControls | undefined;
  export let showFullRange = true;
  export let showTimeComparison = false;
  export let activeTimeZone: string;
  export let onDisplayTimeComparison: (show: boolean) => void;
  export let onSetSelectedComparisonRange: (range: TimeRange) => void;

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
      )
    : Interval.fromDateTimes(allTimeRange.start, allTimeRange.end);

  $: activeTimeGrain = selectedTimeRange?.interval;

  $: comparisonOptions = getComparisonOptionsForCanvas(selectedTimeRange);

  function onSelectComparisonRange(
    name: TimeComparisonOption,
    start: Date,
    end: Date,
  ) {
    onSetSelectedComparisonRange({
      name,
      start,
      end,
    });

    if (!showTimeComparison) {
      onDisplayTimeComparison(!showTimeComparison);
    }
  }

  $: disabled =
    selectedTimeRange?.name === TimeRangePreset.ALL_TIME || undefined;
</script>

<div
  class="wrapper"
  title={disabled && "Comparison not available when viewing all time range"}
>
  <button
    {disabled}
    class="flex gap-x-1.5 cursor-pointer"
    on:click={() => {
      onDisplayTimeComparison(!showTimeComparison);
    }}
  >
    <div class="pointer-events-none flex items-center gap-x-1.5">
      <Switch
        checked={showTimeComparison}
        id="comparing"
        small
        disabled={disabled ?? false}
      />

      <Label class="font-normal text-xs cursor-pointer" for="comparing">
        <span class:opacity-50={disabled}>Comparing</span>
      </Label>
    </div>
  </button>
  {#if activeTimeGrain && interval.isValid}
    <Comparison
      maxDate={DateTime.fromJSDate(allTimeRange.end)}
      minDate={DateTime.fromJSDate(allTimeRange.start)}
      timeComparisonOptionsState={comparisonOptions}
      selectedComparison={selectedComparisonTimeRange}
      showComparison={showTimeComparison}
      currentInterval={interval}
      grain={activeTimeGrain}
      zone={activeTimeZone}
      {onSelectComparisonRange}
      {showFullRange}
      disabled={disabled ?? false}
    />
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply flex w-fit;
    @apply h-7 rounded-full;
    @apply overflow-hidden select-none;
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
