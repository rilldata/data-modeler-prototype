<script lang="ts">
  import * as Elements from "./components";
  import {
    ALL_TIME_RANGE_ALIAS,
    CUSTOM_TIME_RANGE_ALIAS,
    ISODurationString,
    NamedRange,
    RangeBuckets,
    deriveInterval,
  } from "../new-time-controls";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { getValidComparisonOption } from "@rilldata/web-common/features/dashboards/time-controls/time-range-store";
  import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
  import {
    DashboardTimeControls,
    TimeComparisonOption,
    TimeRange,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { DateTime, Interval } from "luxon";
  import { initLocalUserPreferenceStore } from "../../user-preferences";
  import { onMount } from "svelte";

  export let allTimeRange: TimeRange;
  export let selectedTimeRange: DashboardTimeControls | undefined;

  const ctx = getStateManagers();
  const metricsView = useMetricsView(ctx);
  const {
    metricsViewName,
    selectors: {
      timeRangeSelectors: { timeRangeSelectorState },
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
  } = ctx;

  $: localUserPreferences = initLocalUserPreferenceStore(metricViewName);

  $: metricViewName = $metricsViewName;

  $: dashboardStore = useDashboardStore(metricViewName);
  $: selectedRange =
    $dashboardStore?.selectedTimeRange?.name ?? ALL_TIME_RANGE_ALIAS;

  $: defaultTimeRange = $metricsView.data?.defaultTimeRange;

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
      )
    : Interval.fromDateTimes(allTimeRange.start, allTimeRange.end);

  $: activeTimeZone = $dashboardStore?.selectedTimezone;

  $: availableTimeZones = $metricsView?.data?.availableTimeZones ?? [];

  $: metricsViewSpec = $metricsView.data;

  $: ({
    latestWindowTimeRanges,
    periodToDateRanges,
    previousCompleteDateRanges,
    showDefaultItem,
  } = $timeRangeSelectorState);

  $: ranges = <RangeBuckets>{
    latest: latestWindowTimeRanges.map((range) => ({
      range: range.name,
      label: range.label,
    })),
    periodToDate: periodToDateRanges.map((range) => ({
      range: range.name,
      label: range.label,
    })),
    previous: previousCompleteDateRanges.map((range) => ({
      range: range.name,
      label: range.label,
    })),
  };

  $: activeTimeGrain = selectedTimeRange?.interval;

  function onSelectTimeZone(timeZone: string) {
    if (!interval.isValid) return;

    if (selectedRange === "CUSTOM") {
      selectRange({
        name: TimeRangePreset.CUSTOM,
        start: interval.start
          ?.setZone(timeZone, { keepLocalTime: true })
          .toJSDate(),
        end: interval.end
          ?.setZone(timeZone, { keepLocalTime: true })
          .toJSDate(),
      });
    }

    metricsExplorerStore.setTimeZone(metricViewName, timeZone);
    localUserPreferences.set({ timeZone });
  }

  function onSelectRange(name: NamedRange | ISODurationString) {
    if (!allTimeRange?.end) {
      return;
    }

    if (name === ALL_TIME_RANGE_ALIAS) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        allTimeRange,
        "TIME_GRAIN_DAY",
        undefined,
      );
      return;
    }

    const interval = deriveInterval(
      name,
      DateTime.fromJSDate(allTimeRange.end),
    );
    if (!interval?.isValid) return;

    const baseTimeRange: TimeRange = {
      name: name as TimeRangePreset,
      start: interval.start.toJSDate(),
      end: interval.end.toJSDate(),
    };

    selectRange(baseTimeRange);
  }

  // This is pulled directly from the old time controls and needs to be refactored
  onMount(() => {
    /**
     * Remove the timezone selector if no timezone key is present
     * or the available timezone list is empty. Set the default
     * timezone to UTC in such cases.
     *
     */
    if (
      !availableTimeZones.length &&
      $dashboardStore?.selectedTimezone !== "UTC"
    ) {
      metricsExplorerStore.setTimeZone(metricViewName, "UTC");
      localUserPreferences.set({ timeZone: "UTC" });
    }
  });

  function selectRange(range: TimeRange) {
    const defaultTimeGrain = getDefaultTimeGrain(range.start, range.end).grain;

    // Get valid option for the new time range
    const validComparison =
      $metricsView.data &&
      allTimeRange &&
      getValidComparisonOption(
        $metricsView.data,
        range,
        $dashboardStore.selectedComparisonTimeRange?.name as
          | TimeComparisonOption
          | undefined,
        allTimeRange,
      );

    makeTimeSeriesTimeRangeAndUpdateAppState(range, defaultTimeGrain, {
      name: validComparison,
    } as DashboardTimeControls);
  }

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    /** we should only reset the comparison range when the user has explicitly chosen a new
     * time range. Otherwise, the current comparison state should continue to be the
     * source of truth.
     */
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) {
    if (metricsViewSpec)
      metricsExplorerStore.selectTimeRange(
        metricViewName,
        timeRange,
        timeGrain,
        comparisonTimeRange,
        metricsViewSpec.measures,
      );
  }

  function onPan(direction: "left" | "right") {
    const panRange = $getNewPanRange(direction);
    if (!panRange) return;
    const { start, end } = panRange;

    const timeRange = {
      name: CUSTOM_TIME_RANGE_ALIAS,
      start: start,
      end: end,
    };

    const comparisonTimeRange = {
      name: TimeComparisonOption.CONTIGUOUS,
    } as DashboardTimeControls; // FIXME wrong typecasting across application

    if (!activeTimeGrain || !metricsViewSpec) return;
    metricsExplorerStore.selectTimeRange(
      metricViewName,
      timeRange as TimeRange,
      activeTimeGrain,
      comparisonTimeRange,
      metricsViewSpec.measures,
    );
  }
</script>

<div class="wrapper">
  <Elements.Nudge
    canPanLeft={$canPanLeft}
    canPanRight={$canPanRight}
    {onPan}
    direction="left"
  />
  <Elements.Nudge
    canPanLeft={$canPanLeft}
    canPanRight={$canPanRight}
    {onPan}
    direction="right"
  />
  <!-- TO DO -->
  <!-- <Elements.Zoom /> -->
  {#if interval.isValid && activeTimeGrain}
    <Elements.RangePicker
      {ranges}
      {showDefaultItem}
      {defaultTimeRange}
      selected={selectedRange}
      grain={activeTimeGrain}
      {onSelectRange}
      {interval}
      zone={activeTimeZone}
      applyCustomRange={(interval) => {
        selectRange({
          name: TimeRangePreset.CUSTOM,
          start: interval.start.toJSDate(),
          end: interval.end.toJSDate(),
        });
      }}
    />
  {/if}

  {#if availableTimeZones.length}
    <Elements.Zone
      watermark={allTimeRange?.end ?? new Date()}
      {activeTimeZone}
      {availableTimeZones}
      {onSelectTimeZone}
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
