<script>
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";
  import EmptyMeasureIcon from "./EmptyMeasureIcon.svelte";
  import EmptyTableIcon from "./EmptyTableIcon.svelte";

  export let isFetching = false;
  export let assembled = false;
  export let hasColumnAndNoMeasure = false;
</script>

<div class="flex flex-col items-center w-full h-full justify-center gap-y-6">
  {#if isFetching}
    <Spinner size="64px" status={EntityStatus.Running} />
    <div class="font-semibold text-gray-800 mt-1 text-lg">
      Hang tight! We're building your pivot table...
    </div>
    <div class="text-gray-600">
      Need help? Reach out to us on <a
        target="_blank"
        rel="noopener"
        href="http://bit.ly/3jg4IsF">Discord</a
      >
    </div>
  {:else if hasColumnAndNoMeasure}
    <EmptyMeasureIcon />
    <div class="flex flex-col items-center gap-y-2">
      <div class="font-semibold text-gray-800 mt-1 text-lg">Keep it up!</div>
      <div class="text-gray-600 text-base">
        Add a measure to complete your table.
      </div>
    </div>
    <div class="text-gray-600">
      Learn more about pivot tables in our <a
        target="_blank"
        rel="noopener"
        href="https://docs.rilldata.com/explore/filters/pivot">docs</a
      >.
    </div>
  {:else if assembled}
    <EmptyTableIcon />
    <div class="text-gray-600 text-base">
      No data to show for the selected filters.
    </div>
  {:else}
    <EmptyTableIcon />
    <div class="flex flex-col items-center gap-y-2">
      <div class="font-semibold text-gray-800 mt-1 text-lg">
        Your pivot table looks lonely
      </div>
      <div class="text-gray-600 text-base">
        Give it some data to keep it company.
      </div>
    </div>
    <div class="text-gray-600">
      Learn more about pivot tables in our <a
        target="_blank"
        href="https://docs.rilldata.com/explore/filters/pivot">docs</a
      >.
    </div>
  {/if}
</div>
