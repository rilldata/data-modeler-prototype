<script lang="ts">
  import { page } from "$app/stores";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
  import DashboardURLStateSync from "@rilldata/web-common/features/dashboards/url-state/DashboardURLStateSync.svelte";
  import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
  import { getExploreStates } from "@rilldata/web-common/features/explores/selectors";
  import type { V1ExplorePreset } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  /**
   * Temporary wrapper component that mimics the parsing and loading of url into metrics.
   * This is ideally done in the loader function but for embed it needs to be done here.
   * TODO: Fix embed to update the URL and get rid of this.
   */

  const { exploreName, metricsViewName, validSpecStore } = getStateManagers();

  const orgName = $page.params.organization;
  const projectName = $page.params.project;
  const storeKeyPrefix =
    orgName && projectName ? `__${orgName}__${projectName}` : "";

  $: ({ instanceId } = $runtime);
  $: exploreSpec = $validSpecStore.data?.explore ?? {};
  $: metricsViewSpec = $validSpecStore.data?.metricsView ?? {};
  $: metricsViewTimeRange = useMetricsViewTimeRange(
    instanceId,
    $metricsViewName,
  );
  $: defaultExplorePreset = getDefaultExplorePreset(
    exploreSpec,
    $metricsViewTimeRange.data,
  );
  $: ({ partialExploreState: exploreStateFromYAMLConfig } =
    convertPresetToExploreState(
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
      [], // TODO
    ));

  let partialExploreStateFromUrl: Partial<MetricsExplorerEntity> = {};
  let exploreStateFromSessionStorage:
    | Partial<MetricsExplorerEntity>
    | undefined = undefined;
  function parseUrl(url: URL, defaultExplorePreset: V1ExplorePreset) {
    ({ partialExploreStateFromUrl, exploreStateFromSessionStorage } =
      getExploreStates(
        $exploreName,
        storeKeyPrefix,
        url.searchParams,
        metricsViewSpec,
        exploreSpec,
        defaultExplorePreset,
      ));
  }

  // only reactive to url and defaultExplorePreset
  $: parseUrl($page.url, defaultExplorePreset);

  $: validSpec = $validSpecStore.data;
</script>

{#if !$validSpecStore.isLoading && (!validSpec?.metricsView?.timeDimension || !$metricsViewTimeRange.isLoading)}
  <DashboardURLStateSync
    metricsViewName={$metricsViewName}
    exploreName={$exploreName}
    {defaultExplorePreset}
    {exploreStateFromYAMLConfig}
    {partialExploreStateFromUrl}
    {exploreStateFromSessionStorage}
  >
    <slot />
  </DashboardURLStateSync>
{/if}
