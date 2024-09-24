<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import type {
    RangeBuckets,
    NamedRange,
    ISODurationString,
  } from "../../new-time-controls";
  import { getRangeLabel } from "../../new-time-controls";
  import TimeRangeMenu from "./TimeRangeMenu.svelte";
  import { DateTime, Interval } from "luxon";
  import RangeDisplay from "./RangeDisplay.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CalendarPlusDateInput from "./CalendarPlusDateInput.svelte";

  export let ranges: RangeBuckets;
  export let selected: NamedRange | ISODurationString;
  export let interval: Interval<true>;
  export let zone: string;
  export let showDefaultItem: boolean;
  export let grain: string;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let hideRanges = false;
  export let onSelectRange: (range: NamedRange | ISODurationString) => void;
  export let applyCustomRange: (range: Interval<true>) => void;

  let firstVisibleMonth: DateTime<true> = interval.start;
  let open = false;
  let showSelector = false;
</script>

<DropdownMenu.Root
  bind:open
  onOpenChange={(open) => {
    if (open) {
      firstVisibleMonth = interval.start;
    }
    showSelector = selected === "CUSTOM";
  }}
  closeOnItemClick={false}
  typeahead={!showSelector}
>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      {...builder}
      use:builder.action
      class="flex gap-x-1"
      aria-label="Select time range"
    >
      <b class=" line-clamp-1 flex-none">{getRangeLabel(selected)}</b>
      {#if interval.isValid && (selected === "CUSTOM" || !hideRanges)}
        <span class="ml-1">
          <RangeDisplay {interval} {grain} />
        </span>
      {/if}
      <span class="flex-none transition-transform" class:-rotate-180={open}>
        <CaretDownIcon />
      </span>
    </button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" class="p-0 overflow-y-auto  !bg-white">
    <div class="flex">
      <div class="flex flex-col w-48 p-1">
        <TimeRangeMenu
          {ranges}
          {selected}
          {showDefaultItem}
          {defaultTimeRange}
          onSelectRange={(selected) => {
            onSelectRange(selected);

            open = false;
          }}
          onSelectCustomOption={() => (showSelector = !showSelector)}
        />
      </div>
      {#if showSelector}
        <div class="bg-slate-50 border-l flex flex-col w-64 p-2 py-1">
          <CalendarPlusDateInput
            {firstVisibleMonth}
            {interval}
            {zone}
            applyRange={applyCustomRange}
            closeMenu={() => (open = false)}
          />
        </div>
      {/if}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
