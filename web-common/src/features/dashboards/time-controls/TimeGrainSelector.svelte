<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
  import type {
    AvailableTimeGrain,
    DashboardTimeControls,
    TimeGrain,
  } from "@rilldata/web-common/lib/time/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { V1TimeGrain } from "../../../runtime-client";
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "../stores/dashboard-stores";
  import { getAllowedTimeGrains } from "@rilldata/web-common/lib/time/grains";
  import Chip from "@rilldata/web-common/components/chip/core/Chip.svelte";
  import { timeChipColors } from "@rilldata/web-common/components/chip/chip-types";
  import type { TimeRange } from "@rilldata/web-common/lib/time/types";
  import { useMetricsView } from "../selectors";

  export let metricViewName: string;
  export let pill = false;

  const timeControlsStore = useTimeControlStore(getStateManagers());

  let timeGrainOptions: TimeGrain[];
  let open = false;

  $: ({ instanceId } = $runtime);

  $: dashboardStore = useDashboardStore(metricViewName);
  $: metricsViewQuery = useMetricsView(instanceId, metricViewName);
  $: metricsView = $metricsViewQuery.data;

  $: ({ minTimeGrain, timeStart, timeEnd, selectedTimeRange } =
    $timeControlsStore);

  $: timeGrainOptions =
    timeStart && timeEnd
      ? getAllowedTimeGrains(new Date(timeStart), new Date(timeEnd))
      : [];

  $: baseTimeRange = selectedTimeRange?.start &&
    selectedTimeRange?.end && {
      name: selectedTimeRange?.name,
      start: selectedTimeRange.start,
      end: selectedTimeRange.end,
    };

  $: activeTimeGrain = $timeControlsStore.selectedTimeRange?.interval;
  $: activeTimeGrainLabel =
    activeTimeGrain && TIME_GRAIN[activeTimeGrain as AvailableTimeGrain]?.label;

  $: capitalizedLabel = activeTimeGrainLabel
    ?.split(" ")
    .map((word) => {
      return word.charAt(0).toUpperCase() + word.slice(1);
    })
    .join(" ");

  $: timeGrains = minTimeGrain
    ? timeGrainOptions
        .filter((timeGrain) => !isGrainBigger(minTimeGrain, timeGrain.grain))
        .map((timeGrain) => {
          return {
            main: timeGrain.label,
            key: timeGrain.grain,
          };
        })
    : [];

  function onTimeGrainSelect(timeGrain: V1TimeGrain) {
    if (baseTimeRange) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        baseTimeRange,
        timeGrain,
        $dashboardStore?.selectedComparisonTimeRange,
      );
    }
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
    if (metricsView)
      metricsExplorerStore.selectTimeRange(
        metricViewName,
        timeRange,
        timeGrain,
        comparisonTimeRange,
        metricsView.measures,
      );
  }
</script>

{#if activeTimeGrain && timeGrainOptions.length && minTimeGrain}
  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger asChild let:builder>
      {#if pill}
        <Chip
          builders={[builder]}
          {...timeChipColors}
          extraRounded
          outline
          extraPadding={false}
        >
          <div slot="body" class="flex gap-x-2 pl-1.5 items-center">
            <b>Time</b>

            <div class="flex gap-x-1 items-center">
              <span class="font-medium">{capitalizedLabel}</span>

              <div class="transition-transform" class:-rotate-180={open}>
                <CaretDownIcon size="14px" />
              </div>
            </div>
          </div>
        </Chip>
      {:else}
        <button
          use:builder.action
          {...builder}
          aria-label="Select a time grain"
          class="pill"
        >
          <span class="ml-1">by</span>

          <span class="font-bold">{capitalizedLabel}</span>

          <div class="flex-none transition-transform" class:-rotate-180={open}>
            <CaretDownIcon />
          </div>
        </button>
      {/if}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="min-w-40" align="start">
      {#each timeGrains as option (option.key)}
        <DropdownMenu.CheckboxItem
          role="menuitem"
          checked={option.key === activeTimeGrain}
          class="text-xs cursor-pointer"
          on:click={() => onTimeGrainSelect(option.key)}
        >
          {option.main}
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

<style lang="postcss">
  .pill {
    @apply bg-background rounded-full;
    @apply border overflow-hidden flex gap-x-1.5 w-fit h-7;
    @apply px-2 items-center justify-center;
  }
</style>
