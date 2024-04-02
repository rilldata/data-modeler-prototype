<script lang="ts">
  import PivotDisplay from "@rilldata/web-common/features/dashboards/pivot/PivotDisplay.svelte";
  import {
    useDashboard,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import TabBar from "@rilldata/web-common/features/dashboards/tab-bar/TabBar.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { useDashboardStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import DimensionDisplay from "../dimension-table/DimensionDisplay.svelte";
  import Filters from "../filters/Filters.svelte";
  import MockUserHasNoAccess from "../granular-access-policies/MockUserHasNoAccess.svelte";
  import { selectedMockUserStore } from "../granular-access-policies/stores";
  import LeaderboardDisplay from "../leaderboard/LeaderboardDisplay.svelte";
  import RowsViewerAccordion from "../rows-viewer/RowsViewerAccordion.svelte";
  import TimeControls from "../time-controls/TimeControls.svelte";
  import TimeDimensionDisplay from "../time-dimension-details/TimeDimensionDisplay.svelte";
  import MetricsTimeSeriesCharts from "../time-series/MetricsTimeSeriesCharts.svelte";
  import DashboardCTAs from "./DashboardCTAs.svelte";
  import DashboardTitle from "./DashboardTitle.svelte";
  import { navigationOpen } from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import type {
    MetricsViewSpecDimensionV2,
    V1MetricsViewAggregationResponse,
    V1MetricsViewComparisonResponse,
    V1MetricsViewToplistResponse,
  } from "@rilldata/web-common/runtime-client";

  export let metricViewName: string;
  export let dimensions: MetricsViewSpecDimensionV2[];
  export let totals: void | V1MetricsViewAggregationResponse;
  export let leaderBoards: Record<
    string,
    Promise<void | V1MetricsViewComparisonResponse>
  >;

  const { cloudDataViewer, readOnly } = featureFlags;

  let exploreContainerWidth: number;

  $: extraLeftPadding = !$navigationOpen;

  // $: metricsExplorer = useDashboardStore(metricViewName);

  // $: selectedDimensionName = $metricsExplorer?.selectedDimensionName;
  // $: expandedMeasureName = $metricsExplorer?.expandedMeasureName;
  // // $: showPivot = $metricsExplorer?.pivot?.active;
  // $: metricTimeSeries = useModelHasTimeSeries(
  //   $runtime.instanceId,
  //   metricViewName,
  // );

  let expandedMeasureName = null;
  let selectedDimensionName = null;
  $: hasTimeSeries = true;

  let showPivot = false;

  $: isRillDeveloper = $readOnly === false;

  // // Check if the mock user (if selected) has access to the dashboard
  $: dashboard = useDashboard($runtime.instanceId, metricViewName);
  $: mockUserHasNoAccess =
    $selectedMockUserStore && $dashboard.error?.response?.status === 404;
</script>

<section
  class="flex flex-col h-screen w-full overflow-y-hidden dashboard-theme-boundary"
  bind:clientWidth={exploreContainerWidth}
>
  <div
    id="header"
    class="border-b w-fit min-w-full flex flex-col bg-slate-50 pl-4 slide"
    class:left-shift={extraLeftPadding}
  >
    {#if isRillDeveloper}
      <!-- FIXME: adding an -mb-3 fixes the spacing issue incurred by changes to the header 
        to accommodate the cloud dashboard. We should go back and reconcile these headers so we 
        don't need to do this. -->
      <div
        class="flex items-center justify-between -mb-3 w-full pl-1 pr-4"
        style:height="var(--header-height)"
      >
        <DashboardTitle {metricViewName} />
        <DashboardCTAs {metricViewName} />
      </div>
    {/if}

    {#if mockUserHasNoAccess}
      <div class="mb-3" />
    {:else}
      <div class="-ml-3 px-1 pt-2 space-y-2">
        <TimeControls {metricViewName} />

        {#key metricViewName}
          <section class="flex justify-between gap-x-4">
            <!-- <Filters /> -->
            <div class="flex flex-col justify-end">
              <TabBar />
            </div>
          </section>
        {/key}
      </div>
    {/if}
  </div>

  {#if mockUserHasNoAccess}
    <MockUserHasNoAccess />
  {:else}
    <div class="size-full overflow-hidden">
      {#if showPivot}
        <PivotDisplay />
      {:else}
        <div
          class="flex gap-x-1 gap-y-4 pt-3 size-full overflow-hidden pl-4 slide"
          class:flex-col={expandedMeasureName}
          class:flex-row={!expandedMeasureName}
          class:left-shift={extraLeftPadding}
        >
          {#key metricViewName}
            {#if hasTimeSeries}
              <MetricsTimeSeriesCharts
                {totals}
                {metricViewName}
                workspaceWidth={exploreContainerWidth}
              />
            {:else}
              <MeasuresContainer {exploreContainerWidth} {metricViewName} />
            {/if}
          {/key}

          {#if expandedMeasureName}
            <TimeDimensionDisplay {metricViewName} />
          {:else if selectedDimensionName}
            <DimensionDisplay />
          {:else}
            <LeaderboardDisplay {dimensions} {leaderBoards} />
          {/if}
        </div>
      {/if}
    </div>

    {#if (isRillDeveloper || $cloudDataViewer) && !expandedMeasureName && !showPivot}
      <RowsViewerAccordion {metricViewName} />
    {/if}
  {/if}
</section>

<style lang="postcss">
  .left-shift {
    @apply pl-8;
  }
</style>
