<script context="module" lang="ts">
  import { Interval } from "luxon";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Calendar from "@rilldata/web-common/components/date-picker/Calendar.svelte";
  import { DateTime } from "luxon";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import DateInput from "@rilldata/web-common/components/date-picker/DateInput.svelte";
</script>

<script lang="ts">
  export let interval: Interval<true> | undefined;
  export let zone: string;
  export let firstVisibleMonth: DateTime<true>;
  export let minDate: DateTime | undefined = undefined;
  export let maxDate: DateTime | undefined = undefined;
  export let applyRange: (range: Interval<true>) => void;
  export let closeMenu: () => void;

  let selectingStart = true;

  $: calendarInterval =
    interval?.set({
      start: interval.start.startOf("day"),
      end: interval.end.minus({ millisecond: 1 }).startOf("day"),
    }) ?? Interval.invalid("Invalid interval");

  function onValidDateInput(date: DateTime) {
    let newInterval: Interval;

    const selectedEndDateBeforeStart =
      calendarInterval?.start && date < calendarInterval.start;

    if (selectingStart || selectedEndDateBeforeStart) {
      newInterval = calendarInterval.set({ start: date });
      selectingStart = false;
    } else {
      newInterval = calendarInterval.set({ end: date });
      selectingStart = true;
    }

    if (newInterval.isValid) {
      calendarInterval = newInterval;
    } else {
      const singleDay = Interval.fromDateTimes(date, date.endOf("day"));
      if (singleDay.isValid) {
        calendarInterval = singleDay;
      }
    }

    if (calendarInterval.isValid) {
      firstVisibleMonth = calendarInterval.start;
    }
  }
</script>

<svelte:window
  on:keydown|capture={(e) => {
    if (e.key === "Tab") {
      e.stopImmediatePropagation();
    }
  }}
/>

<div class="flex flex-col bg-slate-50">
  <div class="p-4 pt-3 pb-2 border-b">
    <Calendar
      {maxDate}
      {minDate}
      selection={calendarInterval}
      {selectingStart}
      {firstVisibleMonth}
      onSelectDay={onValidDateInput}
    />
  </div>

  <div class="flex flex-col gap-y-2 px-4 py-3">
    <DateInput
      bind:selectingStart
      date={calendarInterval?.start ?? DateTime.now()}
      {zone}
      boundary="start"
      {minDate}
      {maxDate}
      currentYear={firstVisibleMonth.year}
      {onValidDateInput}
    />

    <DateInput
      bind:selectingStart
      date={calendarInterval?.end ?? DateTime.now()}
      {zone}
      boundary="end"
      {minDate}
      {maxDate}
      currentYear={firstVisibleMonth.year}
      {onValidDateInput}
    />
  </div>
  <div class="flex justify-end w-full py-1 px-2">
    <Button
      fit
      compact
      type="subtle"
      on:click={() => {
        const mapped = calendarInterval?.set({
          end: calendarInterval.end?.plus({ day: 1 }).startOf("day"),
        });

        if (mapped?.isValid) {
          applyRange(mapped);
        }

        closeMenu();
      }}
    >
      <span class="px-2 w-fit">Apply</span>
    </Button>
  </div>
</div>
