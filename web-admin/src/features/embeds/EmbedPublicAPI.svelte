<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { get } from "svelte/store";

  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
  import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";

  import {
    getTimeControlState,
    type TimeControlState,
  } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    registerRPCMethod,
    emitNotification,
  } from "@rilldata/web-common/lib/rpc";

  export let instanceId: string;

  const {
    metricsViewName,
    validSpecStore,
    dashboardStore,
    timeRangeSummaryStore,
  } = getStateManagers();

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

  $: ({ data: timeRangeSummaryResp } = $timeRangeSummaryStore);

  let timeControlsState: TimeControlState | undefined = undefined;
  $: if (metricsViewSpec && exploreSpec && $dashboardStore) {
    timeControlsState = getTimeControlState(
      metricsViewSpec,
      exploreSpec,
      timeRangeSummaryResp?.timeRangeSummary,
      $dashboardStore,
    );
  }

  registerRPCMethod("getState", () => {
    return {
      state: decodeURIComponent(
        convertExploreStateToURLSearchParams(
          $dashboardStore,
          exploreSpec,
          timeControlsState,
          defaultExplorePreset,
        ),
      ),
    };
  });

  registerRPCMethod("setState", (state: string) => {
    if (typeof state !== "string") {
      return new Error("Expected state to be a string");
    }
    const currentUrl = new URL(get(page).url);
    currentUrl.search = state;
    goto(currentUrl, { replaceState: true });
    return true;
  });

  $: stateString = decodeURIComponent(
    convertExploreStateToURLSearchParams(
      $dashboardStore,
      exploreSpec,
      timeControlsState,
      defaultExplorePreset,
    ),
  );

  // Stream the state to the parent
  $: emitNotification("stateChange", {
    state: stateString,
  });
</script>

<slot />
