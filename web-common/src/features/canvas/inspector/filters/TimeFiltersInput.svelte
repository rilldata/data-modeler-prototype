<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import CanvasComparisonPill from "@rilldata/web-common/features/canvas/filters/CanvasComparisonPill.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import {
    ALL_TIME_RANGE_ALIAS,
    deriveInterval,
  } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
  import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
  import {
    TimeComparisonOption,
    TimeRangePreset,
    type DashboardTimeControls,
    type TimeRange,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";

  export let selectedComponentName: string;
  export let id: string;
  export let timeFilter: string;
  export let showComparison: boolean;
  export let showGrain: boolean;
  export let onChange: (filter: string) => void = () => {};

  const {
    canvasEntity: {
      useComponent,
      spec: { canvasSpec },
    },
  } = getCanvasStateManagers();

  $: showLocalFilters = Boolean(timeFilter && timeFilter !== "");
  $: componentStore = useComponent(selectedComponentName);

  $: ({
    allTimeRange,
    timeRangeText,
    timeRangeStateStore,
    comparisonRangeStateStore,
    selectedTimezone,
    minTimeGrain,
    selectTimeRange,
    setTimeZone,
    displayTimeComparison,
    setSelectedComparisonRange,
  } = componentStore.localTimeControls);

  $: ({ selectedTimeRange, timeStart, timeEnd } = $timeRangeStateStore || {});

  $: selectedComparisonTimeRange =
    $comparisonRangeStateStore?.selectedComparisonTimeRange;

  $: baseTimeRange = selectedTimeRange?.start &&
    selectedTimeRange?.end && {
      name: selectedTimeRange?.name,
      start: selectedTimeRange.start,
      end: selectedTimeRange.end,
    };

  $: selectedRangeAlias = selectedTimeRange?.name;
  $: activeTimeGrain = selectedTimeRange?.interval;
  $: defaultTimeRange = $canvasSpec?.defaultPreset?.timeRange;
  $: timeRanges = $canvasSpec?.timeRanges ?? [];

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone($selectedTimezone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone($selectedTimezone),
      )
    : Interval.fromDateTimes($allTimeRange.start, $allTimeRange.end);

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    /** we should only reset the comparison range when the user has explicitly chosen a new
     * time range. Otherwise, the current comparison state should continue to be the
     * source of truth.
     */
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) {
    selectTimeRange(timeRange, timeGrain, comparisonTimeRange);
  }

  function selectRange(range: TimeRange) {
    const defaultTimeGrain = getDefaultTimeGrain(range.start, range.end).grain;

    const comparisonOption = DEFAULT_TIME_RANGES[range.name as TimeRangePreset]
      ?.defaultComparison as TimeComparisonOption;

    // Get valid option for the new time range
    const validComparison = allTimeRange && comparisonOption;

    makeTimeSeriesTimeRangeAndUpdateAppState(range, defaultTimeGrain, {
      name: validComparison,
    } as DashboardTimeControls);
  }

  function onSelectRange(name: string) {
    if (!$allTimeRange?.end) {
      return;
    }

    if (name === ALL_TIME_RANGE_ALIAS) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        $allTimeRange,
        "TIME_GRAIN_DAY",
        undefined,
      );
      return;
    }

    const includesTimeZoneOffset = name.includes("@");

    if (includesTimeZoneOffset) {
      const timeZone = name.match(/@ {(.*)}/)?.[1];

      if (timeZone) setTimeZone(timeZone);
    }

    const interval = deriveInterval(
      name,
      DateTime.fromJSDate($allTimeRange.end),
    );

    if (interval?.isValid) {
      const validInterval = interval as Interval<true>;
      const baseTimeRange: TimeRange = {
        // Temporary fix for custom syntax
        name: name as TimeRangePreset,
        start: validInterval.start.toJSDate(),
        end: validInterval.end.toJSDate(),
      };

      selectRange(baseTimeRange);
    }
  }

  function onTimeGrainSelect(timeGrain: V1TimeGrain) {
    if (baseTimeRange) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        baseTimeRange,
        timeGrain,
        selectedComparisonTimeRange,
      );
    }
  }

  $: if (timeFilter && timeFilter !== $timeRangeText) {
    onChange($timeRangeText);
  }
</script>

<div class="flex flex-col gap-y-1 pt-1">
  <div class="flex justify-between">
    <InputLabel
      capitalize={false}
      small
      label="Local time range"
      {id}
      faint={!showLocalFilters}
    />
    <Switch
      checked={showLocalFilters}
      on:click={() => {
        onChange(showLocalFilters ? "" : $timeRangeText);
      }}
      small
    />
  </div>
  <div class="text-gray-500">
    {#if showLocalFilters}
      Overriding inherited time filters from canvas.
    {:else}
      Overrides inherited time filters from canvas when ON.
    {/if}
  </div>

  {#if showLocalFilters}
    <div class="flex flex-row flex-wrap pt-2 gap-y-1.5 items-center">
      <SuperPill
        allTimeRange={$allTimeRange}
        {selectedRangeAlias}
        showPivot={!showGrain}
        minTimeGrain={$minTimeGrain}
        {defaultTimeRange}
        availableTimeZones={[]}
        {timeRanges}
        complete={false}
        {interval}
        {timeStart}
        {timeEnd}
        {activeTimeGrain}
        activeTimeZone={$selectedTimezone}
        canPanLeft={false}
        canPanRight={false}
        showFullRange={false}
        showDefaultItem={false}
        applyRange={selectRange}
        {onSelectRange}
        {onTimeGrainSelect}
        onSelectTimeZone={() => {}}
        onPan={() => {}}
      />

      {#if showComparison}
        <CanvasComparisonPill
          allTimeRange={$allTimeRange}
          {selectedTimeRange}
          showFullRange={false}
          {selectedComparisonTimeRange}
          showTimeComparison={$comparisonRangeStateStore?.showTimeComparison ??
            false}
          activeTimeZone={$selectedTimezone}
          onDisplayTimeComparison={displayTimeComparison}
          onSetSelectedComparisonRange={setSelectedComparisonRange}
        />
      {/if}
    </div>
  {/if}
</div>
