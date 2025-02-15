<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { getComparisonRange } from "@rilldata/web-common/lib/time/comparisons";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import {
    type DashboardTimeControls,
    TimeComparisonOption,
  } from "@rilldata/web-common/lib/time/types";
  import { DateTime, Interval } from "luxon";
  import CalendarPlusDateInput from "./CalendarPlusDateInput.svelte";
  import RangeDisplay from "./RangeDisplay.svelte";

  type Option = {
    name: TimeComparisonOption;
    key: number;
    start: Date;
    end: Date;
  };

  export let currentInterval: Interval<true>;
  export let timeComparisonOptionsState: Option[];
  export let showComparison: boolean | undefined;
  export let selectedComparison: DashboardTimeControls | undefined;
  export let zone: string;
  export let disabled: boolean;
  export let grain: string;
  export let showFullRange: boolean;
  export let minDate: DateTime | undefined = undefined;
  export let maxDate: DateTime | undefined = undefined;
  export let onSelectComparisonRange: (
    name: string,
    start: Date,
    end: Date,
  ) => void;

  let open = false;
  let showSelector = false;

  $: interval = selectedComparison?.start
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedComparison.start).setZone(zone),
        DateTime.fromJSDate(selectedComparison.end).setZone(zone),
      )
    : undefined;

  $: firstVisibleMonth = interval?.start ?? DateTime.now();

  $: comparisonOption =
    (selectedComparison?.name as TimeComparisonOption | undefined) || null;
  $: firstOption = timeComparisonOptionsState[0];
  $: label =
    TIME_COMPARISON[comparisonOption ?? firstOption?.name]?.label ??
    "Custom range";

  $: selectedLabel = comparisonOption ?? firstOption?.name ?? "Custom range";

  function applyRange(range: Interval<true>) {
    onSelectComparisonRange(
      TimeComparisonOption.CUSTOM,
      range.start.toJSDate(),
      range.end.toJSDate(),
    );
  }

  function onCompareRangeSelect(comparisonOption: TimeComparisonOption) {
    if (
      currentInterval.isValid &&
      currentInterval.start &&
      currentInterval.end
    ) {
      const comparisonTimeRange = getComparisonRange(
        currentInterval.start.toJSDate(),
        currentInterval.end.toJSDate(),
        comparisonOption,
      );

      onSelectComparisonRange(
        comparisonOption,
        comparisonTimeRange.start,
        comparisonTimeRange.end,
      );
    }
  }
</script>

<DropdownMenu.Root
  bind:open
  closeOnItemClick={false}
  onOpenChange={(open) => {
    if (open && interval && interval?.isValid) {
      firstVisibleMonth = interval.start;
    }
    showSelector = !!(
      comparisonOption === TimeComparisonOption.CUSTOM && showComparison
    );
  }}
  typeahead={!showSelector}
>
  <DropdownMenu.Trigger asChild let:builder {disabled}>
    <button
      {disabled}
      aria-disabled={disabled}
      use:builder.action
      {...builder}
      aria-label="Select time comparison option"
    >
      <div class="gap-x-2 flex" class:opacity-50={!showComparison}>
        {#if !timeComparisonOptionsState.length && !showComparison}
          <p>no comparison period</p>
        {:else}
          <b class="line-clamp-1">{label}</b>
          {#if interval?.isValid && showFullRange}
            <RangeDisplay {interval} {grain} />
          {/if}
        {/if}
      </div>
      <span
        class="flex-none transition-transform"
        class:-rotate-180={open}
        class:opacity-50={!showComparison}
      >
        <CaretDownIcon />
      </span>
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start" class="p-0 overflow-hidden">
    <div class="flex">
      <div class="flex flex-col border-r w-48 p-1">
        {#each timeComparisonOptionsState as option (option.name)}
          {@const preset = TIME_COMPARISON[option.name]}
          {@const selected = selectedLabel === option.name}
          <DropdownMenu.Item
            class="flex gap-x-2"
            on:click={() => {
              onCompareRangeSelect(option.name);
              open = false;
            }}
          >
            <span class:font-bold={selected}>
              {preset?.label || option.name}
            </span>
          </DropdownMenu.Item>
          {#if option.name === TimeComparisonOption.CONTIGUOUS && timeComparisonOptionsState.length > 2}
            <DropdownMenu.Separator />
          {/if}
        {/each}
        {#if timeComparisonOptionsState.length}
          <DropdownMenu.Separator />
        {/if}

        <DropdownMenu.Item
          data-range="custom"
          on:click={() => {
            showSelector = !showSelector;
          }}
        >
          <span
            class:font-bold={comparisonOption === TimeComparisonOption.CUSTOM &&
              showComparison}
          >
            Custom
          </span>
        </DropdownMenu.Item>
      </div>
      {#if showSelector}
        <div class="bg-slate-50 flex flex-col w-64 px-2 py-1">
          {#if !interval || interval?.isValid}
            <CalendarPlusDateInput
              {maxDate}
              {minDate}
              {firstVisibleMonth}
              {interval}
              {zone}
              {applyRange}
              closeMenu={() => (open = false)}
            />
          {/if}
        </div>
      {/if}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  button {
    @apply gap-x-1;
  }

  .inactive {
    @apply opacity-50;
  }
</style>
