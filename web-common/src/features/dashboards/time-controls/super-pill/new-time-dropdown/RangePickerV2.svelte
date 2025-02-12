<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { DateTime, Interval } from "luxon";
  import type { ISODurationString, NamedRange } from "../../new-time-controls";
  import {
    ALL_TIME_RANGE_ALIAS,
    getRangeLabel,
    RILL_TO_LABEL,
  } from "../../new-time-controls";
  import CalendarPlusDateInput from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/CalendarPlusDateInput.svelte";
  import RangeDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/RangeDisplay.svelte";
  import SyntaxElement from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/SyntaxElement.svelte";
  import { type V1ExploreTimeRange } from "@rilldata/web-common/runtime-client";
  import { bucketTimeRanges } from "../../time-range-store";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import TimeRangeMenuItem from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/TimeRangeMenuItem.svelte";
  import Switch from "@rilldata/web-common/components/button/Switch.svelte";
  import {
    LATEST_WINDOW_TIME_RANGES,
    PERIOD_TO_DATE_RANGES,
    PREVIOUS_COMPLETE_DATE_RANGES,
  } from "@rilldata/web-common/lib/time/config";
  import Timestamp from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/Timestamp.svelte";
  import TimeRangeSearch from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/TimeRangeSearch.svelte";

  export let timeRanges: V1ExploreTimeRange[];
  export let selected: string | undefined;
  export let interval: Interval<true>;
  export let zone: string;
  export let showDefaultItem: boolean;
  export let grain: string;
  export let context: string;
  export let minDate: DateTime;
  export let maxDate: DateTime;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let onSelectRange: (range: string, syntax?: boolean) => void;
  export let applyCustomRange: (range: Interval<true>) => void;

  let firstVisibleMonth: DateTime<true> = interval.start;
  let open = false;
  let allTimeAllowed = true;
  let searchComponent: TimeRangeSearch;

  $: rangeBuckets = bucketTimeRanges(timeRanges, "rill-TD");

  $: colloquialGrain = getColloquialGrain(selected);

  $: grainPhrase = getGrainPhrase(colloquialGrain);

  $: includesCurrentPeriod = interval.end.diff(maxDate).milliseconds > 0;

  $: selectedMeta = selected?.startsWith("P")
    ? LATEST_WINDOW_TIME_RANGES[selected]
    : selected?.startsWith("rill")
      ? (PERIOD_TO_DATE_RANGES[selected] ??
        PREVIOUS_COMPLETE_DATE_RANGES[selected])
      : undefined;

  function handleRangeSelect(range: string, syntax = false) {
    onSelectRange(range, syntax);

    closeMenu();
  }

  function closeMenu() {
    open = false;
  }

  // All of the below is hacky for the sake of development

  function getColloquialGrain(range: string | undefined) {
    // find the first character that matches H M D W Q Y in upper or lower case
    const grain = range?.match(/[HMDWQY]/i)?.[0];
    if (range === "CUSTOM") return undefined;
    return grain;
  }

  function getGrainPhrase(grain: string | undefined) {
    switch (grain?.toUpperCase()) {
      case "H":
        return "this hour";
      case "M":
        return "this month";
      case "D":
        return "today";
      case "W":
        return "this week";
      case "Q":
        return "this quarter";
      case "Y":
        return "this year";
      default:
        return "current period";
    }
  }
</script>

<DropdownMenu.Root
  bind:open
  onOpenChange={(open) => {
    if (open) {
      firstVisibleMonth = interval.start;
    }
  }}
  closeOnItemClick={false}
  typeahead={false}
>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      {...builder}
      use:builder.action
      class="flex gap-x-1"
      aria-label="Select time range"
    >
      {#if selected}
        <b class="mr-1 line-clamp-1 flex-none">{getRangeLabel(selected)}</b>
      {/if}

      {#if interval.isValid}
        <RangeDisplay {interval} {grain} />
      {/if}

      <span class="flex-none transition-transform" class:-rotate-180={open}>
        <CaretDownIcon />
      </span>
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start" class="p-0 overflow-hidden w-[480px]">
    <TimeRangeSearch
      bind:this={searchComponent}
      {context}
      onSelectRange={(range, syntax) => {
        open = false;
        onSelectRange(range, syntax);
      }}
    />

    <div class="flex">
      <div class="flex flex-col w-[216px] flex-none">
        <div
          class="flex-flex-col max-h-[600px] overflow-y-auto overflow-x-hidden p-1"
        >
          {#if showDefaultItem && defaultTimeRange}
            <DropdownMenu.Item
              on:click={() => {
                handleRangeSelect(defaultTimeRange);
              }}
            >
              <div class:font-bold={selected === defaultTimeRange}>
                Last {humaniseISODuration(defaultTimeRange)}
              </div>
            </DropdownMenu.Item>

            <DropdownMenu.Separator />
          {/if}

          {#each rangeBuckets.customTimeRanges as range, i (i)}
            <TimeRangeMenuItem
              {range}
              selected={selected === range.range}
              onClick={handleRangeSelect}
            />
          {/each}

          {#if rangeBuckets.customTimeRanges.length}
            <DropdownMenu.Separator />
          {/if}

          {#each rangeBuckets.ranges as ranges, i (i)}
            {#each ranges as range, i (i)}
              <TimeRangeMenuItem
                {range}
                selected={selected === range.range}
                onClick={handleRangeSelect}
              />
            {/each}
            {#if ranges.length}
              <DropdownMenu.Separator />
            {/if}
          {/each}

          {#if allTimeAllowed}
            <DropdownMenu.Item
              on:click={() => {
                handleRangeSelect(ALL_TIME_RANGE_ALIAS);
              }}
            >
              <span class:font-bold={selected === ALL_TIME_RANGE_ALIAS}>
                {RILL_TO_LABEL[ALL_TIME_RANGE_ALIAS]}
              </span>
            </DropdownMenu.Item>
          {/if}
        </div>

        <DropdownMenu.Separator />

        <!-- <DropdownMenu.Item
          on:click={() => (showSelector = !showSelector)}
          data-range="custom"
        >
          <span class:font-bold={selected === "CUSTOM"}> Calendar </span>
        </DropdownMenu.Item> -->

        <!-- <DropdownMenu.Separator /> -->

        <div class="flex justify-between items-center py-2 px-3">
          <span>Include {grainPhrase}</span>
          <Switch
            id="Show comparison"
            checked={includesCurrentPeriod}
            on:click={() => {
              if (includesCurrentPeriod) {
                const updatedString = `${selectedMeta.rillSyntax}, latest/${colloquialGrain}`;
                console.log({ updatedString });
                onSelectRange(updatedString, true);
              } else {
                onSelectRange(selectedMeta.rillSyntax, true);
              }
            }}
          />
        </div>
      </div>

      <div class="bg-slate-50 border-l flex flex-col w-full gap-y-2">
        <!-- <div class="size-full flex-col gap-y-5 flex"> -->
        <CalendarPlusDateInput
          {firstVisibleMonth}
          {interval}
          {zone}
          {maxDate}
          {minDate}
          applyRange={applyCustomRange}
          closeMenu={() => (open = false)}
        />
        <!-- <div class="flex flex-col gap-y-3">
              <span class="text-gray-500 text-xs">Example searches</span>
              <div class="flex gap-x-2 flex-wrap gap-y-1">
                {#each exampleSearches as search, i (i)}
                  <SyntaxElement range={search} onClick={updateSearch} />
                {/each}
              </div>
            </div> -->

        <!-- <div class="flex flex-col gap-y-3">
              <span class="text-gray-500 text-xs">Custom ranges</span>
              <div class="flex gap-x-2 flex-wrap gap-y-1">
                {#each rangeBuckets.customTimeRanges as range, i (i)}
                  <SyntaxElement range={range.range} onClick={updateSearch} />
                {:else}
                  <span class="text-gray-500"> No ranges available </span>
                {/each}
              </div>
            </div> -->

        <div class="flex flex-col gap-y-3 border-t p-3 mt-auto">
          <span class="text-gray-500 text-xs">Timeframe</span>

          <div class="flex flex-col gap-y-1">
            {#if minDate}
              <div class="flex justify-between items-center">
                <SyntaxElement
                  range="earliest"
                  onClick={searchComponent?.updateSearch}
                />
                <Timestamp date={minDate} {zone} />
              </div>
            {/if}

            {#if maxDate}
              <div class="flex justify-between items-center">
                <SyntaxElement
                  range="latest"
                  onClick={searchComponent?.updateSearch}
                />
                <Timestamp date={maxDate} {zone} />
              </div>
            {/if}

            <div class="flex justify-between items-center">
              <SyntaxElement
                range="now"
                onClick={searchComponent?.updateSearch}
              />
              <Timestamp {zone} />
            </div>
          </div>
        </div>

        <div class="h-7">
          <a href="https://www.rilldata.com" class="mt-auto">
            Syntax documentation
          </a>
        </div>

        <!-- </div> -->
      </div>
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  button {
  }
</style>
