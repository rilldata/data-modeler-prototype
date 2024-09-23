<script lang="ts">
  import Zoom from "@rilldata/web-common/components/icons/Zoom.svelte";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import {
    DashboardTimeControls,
    TimeComparisonOption,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import RangeDisplay from "../time-controls/super-pill/components/RangeDisplay.svelte";
  import { Interval, DateTime } from "luxon";
  import MetaKey from "@rilldata/web-common/components/tooltip/MetaKey.svelte";

  export let metricViewName: string;
  export let showComparison = false;
  export let timeGrain: V1TimeGrain | undefined;

  let priorRange: DashboardTimeControls | null = null;
  let button: HTMLButtonElement;

  const StateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
  } = StateManagers;

  $: activeTimeZone = $dashboardStore?.selectedTimezone;

  $: metricsView = useMetricsView($runtime.instanceId, metricViewName);

  $: validSpec = $metricsView?.data;

  $: ({ selectedScrubRange } = $dashboardStore);

  $: selectedSubRange =
    selectedScrubRange?.start && selectedScrubRange?.end
      ? getOrderedStartEnd(selectedScrubRange.start, selectedScrubRange.end)
      : null;

  $: subInterval = selectedSubRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedSubRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedSubRange.end).setZone(activeTimeZone),
      )
    : null;

  function onKeyDown(e: KeyboardEvent) {
    const targetTagName = (e.target as HTMLElement).tagName;
    if (["INPUT", "TEXTAREA", "SELECT"].includes(targetTagName)) {
      return;
    }

    const isMac = window.navigator.userAgent.includes("Macintosh");

    if (e.key === "ArrowLeft" && !e.metaKey && !e.altKey) {
      if ($canPanLeft) {
        const panRange = $getNewPanRange("left");
        if (panRange) updatePanRange(panRange.start, panRange.end);
      }
    } else if (e.key === "ArrowRight" && !e.metaKey && !e.altKey) {
      if ($canPanRight) {
        const panRange = $getNewPanRange("right");
        if (panRange) updatePanRange(panRange.start, panRange.end);
      }
    } else if ($dashboardStore?.selectedScrubRange?.end) {
      if (e.key === "z" && !e.metaKey && !e.ctrlKey) {
        zoomScrub();
      } else if (
        !$dashboardStore.selectedScrubRange?.isScrubbing &&
        e.key === "Escape"
      ) {
        metricsExplorerStore.setSelectedScrubRange(metricViewName, undefined);
      }
    } else if (
      priorRange &&
      e.key === "z" &&
      ((isMac && e.metaKey) || (!isMac && e.ctrlKey))
    ) {
      e.preventDefault();
      undoZoom();
    }
  }

  function updatePanRange(start: Date, end: Date) {
    if (!timeGrain) return;
    const timeRange = {
      name: TimeRangePreset.CUSTOM,
      start: start,
      end: end,
    };

    const comparisonTimeRange = showComparison
      ? ({
          name: TimeComparisonOption.CONTIGUOUS,
        } as DashboardTimeControls) // FIXME wrong typecasting across application
      : undefined;

    if (validSpec)
      metricsExplorerStore.selectTimeRange(
        metricViewName,
        timeRange,
        timeGrain,
        comparisonTimeRange,
        validSpec.measures,
      );
  }

  function zoomScrub() {
    if (
      selectedScrubRange?.start instanceof Date &&
      selectedScrubRange?.end instanceof Date
    ) {
      if ($dashboardStore.selectedTimeRange) {
        priorRange = $dashboardStore.selectedTimeRange;
      }

      const { start, end } = getOrderedStartEnd(
        selectedScrubRange.start,
        selectedScrubRange.end,
      );
      metricsExplorerStore.setSelectedTimeRange(metricViewName, {
        name: TimeRangePreset.CUSTOM,
        start,
        end,
      });

      window.addEventListener("click", cancelUndo, true);
    }
  }

  function clearPriorRange() {
    priorRange = null;
  }

  function undoZoom() {
    if (priorRange) {
      metricsExplorerStore.setSelectedTimeRange(metricViewName, priorRange);
      clearPriorRange();
    }
  }

  function cancelUndo(e: MouseEvent) {
    window.removeEventListener("click", cancelUndo, true);

    if (
      !priorRange ||
      (e.target instanceof HTMLElement && e.target === button)
    ) {
      return;
    }

    clearPriorRange();
  }

  function handleClick() {
    if (priorRange) {
      undoZoom();
    } else {
      zoomScrub();
    }
  }
</script>

{#if priorRange || (subInterval?.isValid && !subInterval.start.equals(subInterval.end))}
  <button
    bind:this={button}
    on:click|stopPropagation={handleClick}
    aria-label={priorRange ? "Undo zoom" : "Zoom"}
  >
    <div class="content-wrapper">
      <span class="flex-none">
        <Zoom size="16px" />
      </span>

      {#if subInterval?.isValid && timeGrain}
        <RangeDisplay interval={subInterval} grain={timeGrain} />
      {/if}

      <span class="font-medium line-clamp-1 flex-none whitespace-nowrap">
        {#if priorRange}
          Undo Zoom (<MetaKey plusses={false} action="Z" />)
        {:else}
          Zoom (Z)
        {/if}
      </span>
    </div>
  </button>
{/if}

<!-- Only to be used on singleton components to avoid multiple state dispatches -->
<svelte:window on:keydown={onKeyDown} />

<style lang="postcss">
  button {
    @apply border rounded-[2px] bg-background pointer-events-auto;
    @apply absolute left-1/2 -top-8 -translate-x-1/2 z-50;
  }

  .content-wrapper {
    @apply py-1 px-2 flex gap-x-1 w-fit flex-none;
    @apply pointer-events-none;
  }
</style>
