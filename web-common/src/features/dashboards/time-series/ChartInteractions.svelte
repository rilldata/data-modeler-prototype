<script lang="ts">
  import Zoom from "@rilldata/web-common/components/icons/Zoom.svelte";
  import MetaKey from "@rilldata/web-common/components/tooltip/MetaKey.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import {
    type DashboardTimeControls,
    TimeComparisonOption,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";
  import RangeDisplay from "../time-controls/super-pill/components/RangeDisplay.svelte";
  import { scrubRangeStore } from "@rilldata/web-common/components/time-series-chart/Chart.svelte";

  export let exploreName: string;
  export let showComparison = false;
  export let timeGrain: V1TimeGrain | undefined;

  let priorRange: DashboardTimeControls | null = null;
  let button: HTMLButtonElement;

  const { validRange } = scrubRangeStore;

  const StateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
    validSpecStore,
  } = StateManagers;

  $: activeTimeZone = $dashboardStore?.selectedTimezone;

  $: subInterval = $validRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate($validRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate($validRange.end).setZone(activeTimeZone),
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
        metricsExplorerStore.setSelectedScrubRange(exploreName, undefined);
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

    metricsExplorerStore.selectTimeRange(
      exploreName,
      timeRange,
      timeGrain,
      comparisonTimeRange,
      $validSpecStore.data?.metricsView ?? {},
    );
  }

  function zoomScrub() {
    if (
      $validRange?.start instanceof Date &&
      $validRange?.end instanceof Date
    ) {
      if ($dashboardStore.selectedTimeRange) {
        priorRange = $dashboardStore.selectedTimeRange;
      }

      const { start, end } = getOrderedStartEnd(
        $validRange.start,
        $validRange.end,
      );
      metricsExplorerStore.setSelectedTimeRange(exploreName, {
        name: TimeRangePreset.CUSTOM,
        start,
        end,
      });

      scrubRangeStore.reset();

      window.addEventListener("click", cancelUndo, true);
    }
  }

  function clearPriorRange() {
    priorRange = null;
  }

  function undoZoom() {
    if (priorRange) {
      metricsExplorerStore.setSelectedTimeRange(exploreName, priorRange);
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
