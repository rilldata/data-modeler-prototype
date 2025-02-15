<script lang="ts">
  import { DateTime, Interval } from "luxon";

  export let interval: Interval<true>;
  export let grain: string;
  export let abbreviation: string | undefined = undefined;

  $: showTime = grain === "TIME_GRAIN_HOUR" || grain === "TIME_GRAIN_MINUTE";
  $: timeFormat = grain === "TIME_GRAIN_MINUTE" ? "h:mm a" : "h a";

  $: inclusiveInterval = interval.set({
    end: interval.end.minus({ millisecond: 1 }),
  });

  $: displayedInterval = showTime ? interval : inclusiveInterval;

  $: date = displayedInterval.toLocaleString(DateTime.DATE_MED);

  $: time = displayedInterval.toFormat(timeFormat, { separator: "-" });
</script>

<div class="flex gap-x-1 whitespace-nowrap truncate" title="{date} {time}">
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
