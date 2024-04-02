<script lang="ts">
  import {
    useMetricsView,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { getValidComparisonOption } from "@rilldata/web-common/features/dashboards/time-controls/time-range-store";
  import {
    getAllowedTimeGrains,
    getDefaultTimeGrain,
  } from "@rilldata/web-common/lib/time/grains";
  import type {
    DashboardTimeControls,
    TimeComparisonOption,
    TimeGrain,
    TimeRange,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { initLocalUserPreferenceStore } from "../user-preferences";
  import ComparisonSelector from "./ComparisonSelector.svelte";
  import NoTimeDimensionCTA from "./NoTimeDimensionCTA.svelte";
  import TimeComparisonSelector from "./TimeComparisonSelector.svelte";
  import TimeGrainSelector from "./TimeGrainSelector.svelte";
  import TimeRangeSelector from "./TimeRangeSelector.svelte";
  import TimeZoneSelector from "./TimeZoneSelector.svelte";

  export let metricViewName: string;

  const localUserPreferences = initLocalUserPreferenceStore(metricViewName);

  $: dashboardStore = useDashboardStore(metricViewName);

  let baseTimeRange: TimeRange | undefined;
  let minTimeGrain: V1TimeGrain | undefined;
  let availableTimeZones: string[] = [];

  $: metricsView = useMetricsView($runtime.instanceId, metricViewName);

  $: hasTimeSeriesQuery = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName,
  );
  $: hasTimeSeries = $hasTimeSeriesQuery?.data;

  const timeControlsStore = useTimeControlStore(getStateManagers());
  $: allTimeRange = $timeControlsStore.allTimeRange;
  $: minTimeGrain = $timeControlsStore.minTimeGrain;

  $: if (
    $timeControlsStore.ready &&
    !!$metricsView?.data?.table &&
    !!$metricsView?.data?.timeDimension
  ) {
    availableTimeZones = $metricsView?.data?.availableTimeZones ?? [];

    /**
     * Remove the timezone selector if no timezone key is present
     * or the available timezone list is empty. Set the default
     * timezone to UTC in such cases.
     *
     */
    if (
      !availableTimeZones?.length &&
      $dashboardStore?.selectedTimezone !== "UTC"
    ) {
      metricsExplorerStore.setTimeZone(metricViewName, "UTC");
      localUserPreferences.set({ timeZone: "UTC" });
    }

    baseTimeRange = $timeControlsStore.selectedTimeRange?.start &&
      $timeControlsStore.selectedTimeRange?.end && {
        name: $timeControlsStore.selectedTimeRange?.name,
        start: $timeControlsStore.selectedTimeRange.start,
        end: $timeControlsStore.selectedTimeRange.end,
      };
  }

  // we get the timeGrainOptions so that we can assess whether or not the
  // activeTimeGrain is valid whenever the baseTimeRange changes
  let timeGrainOptions: TimeGrain[];
  $: timeGrainOptions =
    $timeControlsStore.timeStart && $timeControlsStore.timeEnd
      ? getAllowedTimeGrains(
          new Date($timeControlsStore.timeStart),
          new Date($timeControlsStore.timeEnd),
        )
      : [];

  function onSelectTimeRange(name: TimeRangePreset, start: Date, end: Date) {
    baseTimeRange = {
      name,
      start: new Date(start),
      end: new Date(end),
    };

    const defaultTimeGrain = getDefaultTimeGrain(
      baseTimeRange.start,
      baseTimeRange.end,
    ).grain;

    // Get valid option for the new time range
    const validComparison =
      $metricsView.data &&
      $timeControlsStore.allTimeRange &&
      getValidComparisonOption(
        $metricsView.data,
        baseTimeRange,
        $dashboardStore.selectedComparisonTimeRange?.name as
          | TimeComparisonOption
          | undefined,
        $timeControlsStore.allTimeRange,
      );

    makeTimeSeriesTimeRangeAndUpdateAppState(
      baseTimeRange,
      defaultTimeGrain,
      $dashboardStore?.showTimeComparison
        ? ({
            name: validComparison,
          } as DashboardTimeControls)
        : undefined,
    );
  }

  function onSelectTimeZone(timeZone: string) {
    metricsExplorerStore.setTimeZone(metricViewName, timeZone);
    localUserPreferences.set({ timeZone });
  }

  function onSelectComparisonRange(
    name: TimeComparisonOption,
    start: Date,
    end: Date,
  ) {
    metricsExplorerStore.setSelectedComparisonRange(metricViewName, {
      name,
      start,
      end,
    });
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
    metricsExplorerStore.selectTimeRange(
      metricViewName,
      timeRange,
      timeGrain,
      comparisonTimeRange,
    );
  }
</script>

<div class="flex flex-row items-center gap-x-1">
  {#if !hasTimeSeries}
    <NoTimeDimensionCTA />
  {:else if allTimeRange?.start && minTimeGrain && $timeControlsStore.selectedTimeRange}
    <TimeRangeSelector
      {metricViewName}
      {minTimeGrain}
      boundaryStart={allTimeRange.start}
      boundaryEnd={allTimeRange.end}
      selectedRange={$timeControlsStore?.selectedTimeRange}
      on:select-time-range={(e) =>
        onSelectTimeRange(e.detail.name, e.detail.start, e.detail.end)}
      on:remove-scrub={() => {
        metricsExplorerStore.setSelectedScrubRange(metricViewName, undefined);
      }}
    />
    {#if availableTimeZones?.length}
      <TimeZoneSelector
        on:select-time-zone={(e) => onSelectTimeZone(e.detail.timeZone)}
        {metricViewName}
        {availableTimeZones}
        now={allTimeRange?.end}
      />
    {/if}
    <ComparisonSelector {metricViewName} />
    {#if $dashboardStore?.showTimeComparison}
      <TimeComparisonSelector
        on:select-comparison={(e) => {
          onSelectComparisonRange(e.detail.name, e.detail.start, e.detail.end);
        }}
        {minTimeGrain}
        currentStart={$timeControlsStore.selectedTimeRange.start}
        currentEnd={$timeControlsStore.selectedTimeRange.end}
        boundaryStart={allTimeRange.start}
        boundaryEnd={allTimeRange.end}
        showComparison={$timeControlsStore?.showComparison}
        selectedComparison={$timeControlsStore?.selectedComparisonTimeRange}
        zone={$dashboardStore.selectedTimezone}
      />
    {/if}
    <TimeGrainSelector {timeGrainOptions} {minTimeGrain} />
  {/if}
</div>
