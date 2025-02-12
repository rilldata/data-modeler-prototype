<script lang="ts">
  import { DateTime, Interval } from "luxon";
  import { ChevronLeft } from "lucide-svelte";
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";
  import Day from "./Day.svelte";

  export let startDay: DateTime<true>;
  export let startOfWeek: number;
  export let interval: Interval<true> | undefined;
  export let selectingStart: boolean;
  export let visibleMonths = 2;
  export let visibleIndex: number;
  export let potentialEnd: DateTime | undefined;
  export let potentialStart: DateTime | undefined;
  export let singleDaySelection = false;
  export let minDate: DateTime<true> | DateTime<false> | undefined;
  export let maxDate: DateTime<true> | DateTime<false> | undefined;
  export let onPan: (direction: 1 | -1) => void;
  export let onSelectDay: (date: DateTime<true>) => void;

  $: firstDayOfMonth = startDay.startOf("month");
  $: firstWeekday = firstDayOfMonth.weekday;

  $: diff = (firstWeekday - startOfWeek) % 7;

  $: firstVisibleDay = firstDayOfMonth.minus({ days: diff });

  $: visibleWeeks = Math.ceil((firstWeekday + startDay.daysInMonth) / 7);

  $: inclusiveEnd = interval?.end?.minus({ millisecond: 0 });

  $: forwardPanEnabled = !maxDate || startDay.plus({ month: 1 }) < maxDate;

  $: days = Array.from({ length: visibleWeeks * 7 }, (_, i) => {
    return firstVisibleDay.plus({ days: i });
  });

  $: weekdays = Array.from({ length: 7 }, (_, i) =>
    new Date(0, 0, i + startOfWeek).toLocaleString("default", {
      weekday: "short",
    }),
  );

  function resetPotentialDates() {
    potentialEnd = undefined;
    potentialStart = undefined;
  }
</script>

<div class="flex flex-col gap-2 w-full">
  <div class="flex justify-between px-2">
    <button
      type="button"
      class="hover:opacity-50"
      class:hide={visibleIndex !== 0}
      on:click={() => onPan(-1)}
    >
      <ChevronLeft size="14px" />
    </button>
    <div
      class="w-full text-center px-2 py-1 text-sm flex gap-x-1 justify-center select-none"
    >
      <b>{startDay.monthLong}</b>
      <p>{startDay.year}</p>
    </div>
    {#if forwardPanEnabled}
      <button
        type="button"
        class="hover:opacity-50"
        class:hide={visibleIndex !== visibleMonths - 1}
        on:click={() => onPan(1)}
      >
        <ChevronRight size="14px" />
      </button>
    {/if}
  </div>

  <div
    role="presentation"
    class="grid grid-cols-7 gap-y-1 w-full"
    class:single-day-selection={singleDaySelection}
    class:selecting-start={selectingStart}
    on:mouseleave={resetPotentialDates}
  >
    {#each weekdays as weekday (weekday)}
      <div class="weekday">{weekday}</div>
    {/each}
    {#each days as date (date.toISO())}
      <Day
        {date}
        {selectingStart}
        {inclusiveEnd}
        bind:potentialEnd
        bind:potentialStart
        {singleDaySelection}
        {onSelectDay}
        {resetPotentialDates}
        start={interval?.start}
        outOfMonth={date.month !== startDay.month}
        disabled={Boolean(
          (minDate && date < minDate) || (maxDate && date > maxDate),
        )}
      />
    {/each}
  </div>
</div>

<style lang="postcss">
  .weekday {
    @apply text-center w-full aspect-[2/1] text-slate-500 select-none;
  }

  .hide {
    @apply opacity-0 pointer-events-none;
  }
</style>
