<script lang="ts">
  import { Interval, DateTime } from "luxon";

  export let interval: Interval<true>;
  export let grain: string;
  export let abbreviation: string | undefined = undefined;

  $: intervalStartsAndEndsAtMidnight =
    interval.start.hour === 0 &&
    interval.start.minute === 0 &&
    interval.end.hour === 0 &&
    interval.end.minute === 0;

  $: intervalStartsAndEndsOnHour =
    interval.start.minute === 0 && interval.end.minute === 0;

  $: timeFormat = intervalStartsAndEndsOnHour ? "h a" : "h:mm a";

  $: showTime =
    !intervalStartsAndEndsAtMidnight ||
    grain === "TIME_GRAIN_HOUR" ||
    grain === "TIME_GRAIN_MINUTE";

  $: inclusiveInterval = interval.set({
    end: interval.end.minus({ millisecond: 1 }),
  });

  $: displayedInterval = showTime ? interval : inclusiveInterval;

  $: date = displayedInterval.toLocaleString(DateTime.DATE_MED);

  $: time = displayedInterval.toFormat(timeFormat, { separator: "-" });
</script>

<div class="flex gap-x-1 whitespace-nowrap" title="{date} {time}">
  <span class="line-clamp-1 text-left">
    {date}
    {#if showTime}
      ({time})
    {/if}
    {#if abbreviation}
      {abbreviation}
    {/if}
  </span>
</div>
