<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { DateTime, Duration } from "luxon";

  export let date: DateTime<true> = DateTime.now();
  export let zone: string;

  $: zonedDate = date.setZone(zone);
  $: isoString = zonedDate.toISO();

  $: formattedString = zonedDate.toLocaleString(
    DateTime.DATETIME_MED_WITH_WEEKDAY,
  );

  $: humanReadableTimeOffset = Duration.fromObject(
    Object.fromEntries(
      Object.entries(DateTime.now().diff(date).rescale().toObject())
        .filter(([, value]) => value !== 0)
        .slice(0, 2),
    ),
  ).toHuman({
    listStyle: "narrow",
    maximumFractionDigits: 0,
  });
</script>

<Tooltip>
  <button
    class="text-gray-500 text-xs"
    on:click={() => {
      if (isoString) copyToClipboard(isoString);
    }}
  >
    {formattedString}
  </button>

  <TooltipContent slot="tooltip-content">
    <div class="flex flex-col gap-y-1 items-center">
      {#if humanReadableTimeOffset.length}
        <span>
          {humanReadableTimeOffset} ago
        </span>
      {/if}
      <span>
        {isoString}
      </span>
    </div>
  </TooltipContent>
</Tooltip>
