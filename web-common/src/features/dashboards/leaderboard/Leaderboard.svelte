<script lang="ts">
  /**
   * Leaderboard.svelte
   * -------------------------
   * This is the "implemented" feature of the leaderboard, meant to be used
   * in the application itself.
   */
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    V1MetricsViewComparisonResponse,
    createQueryServiceMetricsViewComparison,
    createQueryServiceMetricsViewTotals,
  } from "@rilldata/web-common/runtime-client";

  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import LeaderboardListItem from "./LeaderboardListItem.svelte";
  import {
    LeaderboardItemData,
    getLabeledComparisonFromComparisonRow,
    prepareLeaderboardItemData,
  } from "./leaderboard-utils";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { page } from "$app/stores";
  export let dimensionName: string;
  export let data: Promise<V1MetricsViewComparisonResponse>;
  export let activeMeasureName = "measure";

  /** The reference value is the one that the bar in the LeaderboardListItem
   * gets scaled with. For a summable metric, the total is a reference value,
   * or for a count(*) metric, the reference value is the total number of rows.
   */

  let slice = 7;

  $: measures = $page?.data?.measures;

  $: activeMeasure = measures.find((m) => m.name === activeMeasureName);

  $: formatter = createMeasureValueFormatter<null | undefined>(activeMeasure);

  // const {
  //   selectors: {
  //     activeMeasure: { activeMeasureName },
  //     dimensionFilters: { selectedDimensionValues },
  //     measureFilters: { getResolvedFilterForMeasureFilters },
  //     dashboardQueries: {
  //       leaderboardSortedQueryBody,
  //       leaderboardSortedQueryOptions,
  //       leaderboardDimensionTotalQueryBody,
  //       leaderboardDimensionTotalQueryOptions,
  //     },
  //   },
  //   actions: {
  //     dimensions: { setPrimaryDimension },
  //   },
  //   metricsViewName,
  //   runtime,
  // } = getStateManagers();

  // $: resolvedFilter = $getResolvedFilterForMeasureFilters;

  // $: sortedQuery = createQueryServiceMetricsViewComparison(
  //   $runtime.instanceId,
  //   $metricsViewName,
  //   $leaderboardSortedQueryBody(dimensionName, $resolvedFilter),
  //   $leaderboardSortedQueryOptions(dimensionName, $resolvedFilter),
  // );

  // $: totalsQuery = createQueryServiceMetricsViewTotals(
  //   $runtime.instanceId,
  //   $metricsViewName,
  //   $leaderboardDimensionTotalQueryBody(dimensionName, $resolvedFilter),
  //   $leaderboardDimensionTotalQueryOptions(dimensionName, $resolvedFilter),
  // );

  // let activeMeasureName = null;

  // $: console.log("WHAT", $page.data.totals.data);
  $: leaderboardTotal = $page.data.totals.data[0][activeMeasureName];

  // let aboveTheFold: LeaderboardItemData[] = [];
  // let selectedBelowTheFold: LeaderboardItemData[] = [];
  // let noAvailableValues = true;
  // let showExpandTable = false;
  // $: if (sortedQuery && !$sortedQuery?.isFetching) {
  //   const leaderboardData = prepareLeaderboardItemData(
  //     $sortedQuery?.data?.rows?.map((r) =>
  //       getLabeledComparisonFromComparisonRow(r, $activeMeasureName),
  //     ) ?? [],
  //     slice,
  //     $selectedDimensionValues(dimensionName),
  //     leaderboardTotal,
  //   );

  //   aboveTheFold = leaderboardData.aboveTheFold;
  //   selectedBelowTheFold = leaderboardData.selectedBelowTheFold;
  //   noAvailableValues = leaderboardData.noAvailableValues;
  //   showExpandTable = leaderboardData.showExpandTable;
  // }

  let hovered: boolean;

  function setPrimaryDimension(dimensionName: string) {
    // setPrimaryDimension(dimensionName);
  }
</script>

<!-- {#if $sortedQuery !== undefined} -->
<div
  role="grid"
  tabindex="0"
  style:width="315px"
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
>
  {#await data}
    <LeaderboardHeader isFetching={true} {dimensionName} {hovered} />
  {:then dataYeah}
    <LeaderboardHeader isFetching={false} {dimensionName} {hovered} />
    {@const leaderboardData = prepareLeaderboardItemData(
      dataYeah.rows?.map((r) =>
        getLabeledComparisonFromComparisonRow(r, "measure"),
      ) ?? [],
      slice,
      [],
      // $selectedDimensionValues(dimensionName),
      leaderboardTotal,
    )}

    {@const {
      aboveTheFold = [],
      selectedBelowTheFold = [],
      noAvailableValues,
      showExpandTable,
    } = leaderboardData}

    <div class="rounded-b border-gray-200 surface text-gray-800">
      {#each aboveTheFold as itemData (itemData.dimensionValue)}
        <LeaderboardListItem
          {dimensionName}
          {itemData}
          {formatter}
          on:click
          on:keydown
        />
      {/each}

      {#if selectedBelowTheFold?.length}
        <hr />
        {#each selectedBelowTheFold as itemData (itemData.dimensionValue)}
          <LeaderboardListItem {dimensionName} {itemData} on:click on:keydown />
        {/each}
        <hr />
      {/if}
      {#if noAvailableValues}
        <div style:padding-left="30px" class="p-1 ui-copy-disabled">
          No available values
        </div>
      {/if}
      {#if showExpandTable}
        <Tooltip location="right">
          <button
            on:click={() => setPrimaryDimension(dimensionName)}
            class="block flex-row w-full text-left transition-color ui-copy-muted"
            style:padding-left="30px"
          >
            (Expand Table)
          </button>
          <TooltipContent slot="tooltip-content"
            >Expand dimension to see more values</TooltipContent
          >
        </Tooltip>
      {/if}
    </div>
  {:catch _}
    <div class="ml-[22px] flex p-2 gap-x-1 items-center">
      <div class="text-gray-500">Unable to load leaderboard.</div>
      <!-- <button
      class="text-primary-500 hover:text-primary-600 font-medium"
      disabled={false}
      on:click={() => $sortedQuery.refetch()}>Try again</button
    > -->
    </div>
  {/await}
</div>
