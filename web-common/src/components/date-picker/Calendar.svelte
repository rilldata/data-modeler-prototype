<script lang="ts">
  import Month from "./Month.svelte";
  import { DateTime } from "luxon";
  import { Interval } from "luxon";

  type MaybeDate = DateTime | undefined;

  export let selection: MaybeDate | Interval = undefined;
  export let minDate: MaybeDate = undefined;
  export let maxDate: MaybeDate = DateTime.now().startOf("day");
  export let visibleMonths = 1;
  export let selectingStart = true;
  export let firstVisibleMonth: MaybeDate = isValidDateTime(selection)
    ? selection
    : isValidInterval(selection)
      ? (selection.start ?? DateTime.now())
      : DateTime.now();
  export let singleDaySelection = isValidDateTime(selection);
  export let startOfWeek = 1;
  export let onSelectDay: (date: DateTime<true>) => void;

  let potentialEnd: DateTime<true> | undefined;
  let potentialStart: DateTime<true> | undefined;

  $: finalInterval = isValidDateTime(selection)
    ? (Interval.fromDateTimes(
        selection,
        selection.endOf("day"),
      ) as Interval<true>)
    : isValidInterval(selection)
      ? selection
      : undefined;

  $: firstMonth = isValidDateTime(firstVisibleMonth)
    ? firstVisibleMonth
    : DateTime.now();

  function onPan(direction: -1 | 1) {
    firstMonth = firstMonth.plus({ month: direction });
  }

  function isValidDateTime(
    value: MaybeDate | Interval<true>,
  ): value is DateTime<true> {
    return Boolean(value && value instanceof DateTime && value?.isValid);
  }

  function isValidInterval(
    value: MaybeDate | Interval,
  ): value is Interval<true> {
    return Boolean(value && value instanceof Interval && value?.isValid);
  }
</script>

<div class="flex gap-x-3 w-full min-w-56">
  {#each { length: visibleMonths } as month, i (month)}
    <Month
      {minDate}
      {maxDate}
      {singleDaySelection}
      {startOfWeek}
      interval={finalInterval}
      startDay={firstMonth.plus({ month: i }).set({ day: 1 }).startOf("day")}
      {selectingStart}
      {visibleMonths}
      visibleIndex={i}
      bind:potentialStart
      bind:potentialEnd
      {onSelectDay}
      {onPan}
    />
  {/each}
</div>
