<script lang="ts">
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import PivotDisplay from "@rilldata/web-common/features/dashboards/pivot/PivotDisplay.svelte";
  import { useModelHasTimeSeries } from "@rilldata/web-common/features/dashboards/selectors";
  import TabBar from "@rilldata/web-common/features/dashboards/tab-bar/TabBar.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { navigationOpen } from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { useExploreStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { runtime } from "../../../runtime-client/runtime-store";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import DimensionDisplay from "../dimension-table/DimensionDisplay.svelte";
  import Filters from "../filters/Filters.svelte";
  import { selectedMockUserStore } from "../granular-access-policies/stores";
  import LeaderboardDisplay from "../leaderboard/LeaderboardDisplay.svelte";
  import RowsViewerAccordion from "../rows-viewer/RowsViewerAccordion.svelte";
  import MetricsTimeSeriesCharts from "../time-series/MetricsTimeSeriesCharts.svelte";

  export let exploreName: string;
  export let metricsViewName: string;

  const { cloudDataViewer, readOnly } = featureFlags;

  let exploreContainerWidth: number;

  $: extraLeftPadding = !$navigationOpen;

  $: exploreStore = useExploreStore(exploreName);

  $: selectedDimensionName = $exploreStore?.selectedDimensionName;
  $: expandedMeasureName = $exploreStore?.tdd?.expandedMeasureName;
  $: showPivot = $exploreStore?.pivot?.active;
  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricsViewName,
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  $: isRillDeveloper = $readOnly === false;

  // Check if the mock user (if selected) has access to the explore
  $: explore = useExploreValidSpec($runtime.instanceId, exploreName);
  $: mockUserHasNoAccess =
    $selectedMockUserStore && $explore.error?.response?.status === 404;
</script>

<article
  class="flex flex-col size-full overflow-y-hidden dashboard-theme-boundary"
  bind:clientWidth={exploreContainerWidth}
>
  <div
    id="header"
    class="border-b w-fit min-w-full flex flex-col bg-slate-50 slide"
    class:left-shift={extraLeftPadding}
  >
    {#if mockUserHasNoAccess}
      <div class="mb-3" />
    {:else}
      {#key exploreName}
        <section class="flex relative justify-between gap-x-4 py-4 pb-6 px-4">
          <Filters />
          <div class="absolute bottom-0 flex flex-col right-0">
            <TabBar />
          </div>
        </section>
      {/key}
    {/if}
  </div>

  {#if mockUserHasNoAccess}
    <!-- Additional safeguard for mock users without dashboard access. -->
    <ErrorPage
      statusCode={$explore.error?.response?.status}
      header="This user can't access this dashboard"
      body="The security policy for this dashboard may make contents invisible to you. If you deploy this dashboard, {$selectedMockUserStore?.email} will see a 404."
    />
  {:else if showPivot}
    <PivotDisplay />
  {:else}
    <div
      class="flex gap-x-1 gap-y-2 size-full overflow-hidden pl-4 slide"
      class:flex-col={expandedMeasureName}
      class:flex-row={!expandedMeasureName}
      class:left-shift={extraLeftPadding}
    >
      <section class="pt-2 w-full" class:!w-fit={!expandedMeasureName}>
        {#key exploreName}
          {#if hasTimeSeries}
            <MetricsTimeSeriesCharts
              {exploreName}
              workspaceWidth={exploreContainerWidth}
            />
          {:else}
            <MeasuresContainer {exploreContainerWidth} {metricsViewName} />
          {/if}
        {/key}
      </section>

      {#if !expandedMeasureName}
        <section class="pt-2 pl-1 border-l overflow-auto w-full">
          {#if selectedDimensionName}
            <DimensionDisplay />
          {:else}
            <LeaderboardDisplay />
          {/if}
        </section>
      {/if}
    </div>
  {/if}
</article>

{#if (isRillDeveloper || $cloudDataViewer) && !expandedMeasureName && !mockUserHasNoAccess}
  <RowsViewerAccordion {metricsViewName} {exploreName} />
{/if}

<style lang="postcss">
  .left-shift {
    @apply pl-8;
  }
</style>
