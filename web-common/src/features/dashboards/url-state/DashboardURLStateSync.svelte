<script lang="ts">
  import { afterNavigate, beforeNavigate, goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { resolveTimeRanges } from "@rilldata/web-common/features/dashboards/time-controls/rill-time-ranges";
  import {
    getTimeControlState,
    type TimeControlState,
  } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    convertExploreStateToURLSearchParams,
    getUpdatedUrlForExploreState,
  } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
  import {
    clearExploreSessionStore,
    hasSessionStorageData,
    updateExploreSessionStore,
  } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-store";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
  import {
    createQueryServiceMetricsViewSchema,
    type V1ExplorePreset,
  } from "@rilldata/web-common/runtime-client";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";
  import { get } from "svelte/store";

  export let metricsViewName: string;
  export let exploreName: string;
  export let extraKeyPrefix: string | undefined = undefined;
  export let defaultExplorePreset: V1ExplorePreset;
  export let exploreStateFromYAMLConfig: Partial<MetricsExplorerEntity>;
  export let initExploreState: Partial<MetricsExplorerEntity> | undefined =
    undefined;
  export let partialExploreStateFromUrl: Partial<MetricsExplorerEntity>;
  export let exploreStateFromSessionStorage:
    | Partial<MetricsExplorerEntity>
    | undefined;

  const { dashboardStore, validSpecStore, timeRangeSummaryStore } =
    getStateManagers();
  $: exploreSpec = $validSpecStore.data?.explore;
  $: metricsSpec = $validSpecStore.data?.metricsView;

  $: ({ instanceId } = $runtime);

  const metricsViewSchema = createQueryServiceMetricsViewSchema(
    instanceId,
    metricsViewName,
  );
  $: ({ error: schemaError } = $metricsViewSchema);
  $: ({
    error,
    data: timeRangeSummaryResp,
    isLoading: timeRangeSummaryIsLoading,
  } = $timeRangeSummaryStore);
  $: timeRangeSummaryError = error as HTTPError;

  let timeControlsState: TimeControlState | undefined = undefined;
  $: if (metricsSpec && exploreSpec && $dashboardStore) {
    timeControlsState = getTimeControlState(
      metricsSpec,
      exploreSpec,
      timeRangeSummaryResp?.timeRangeSummary,
      $dashboardStore,
    );
  }

  let prevUrl = "";
  let initializing = false;

  onMount(() => {
    // in some cases afterNavigate is not always triggered
    // so this is the escape hatch to make sure dashboard store gets initialised
    setTimeout(() => {
      if (!$dashboardStore) {
        void handleExploreInit(true);
      }
    });
  });

  beforeNavigate(({ from, to }) => {
    if (!from || !to || from.url.pathname === to.url.pathname) {
      // routing to the same path but probably different url params
      return;
    }

    // session store is only used to save state for different views and not keep other params url
    // so, we clear the store when we navigate away
    clearExploreSessionStore(exploreName, extraKeyPrefix);
  });

  afterNavigate(({ from, to, type }) => {
    if (
      // null checks
      !metricsSpec ||
      !exploreSpec ||
      !to ||
      // seems like a sveltekit bug where an additional afterNavigate is triggered with invalid fields
      (from !== null && !from.url)
    ) {
      return;
    }

    void handleAfterNavigate(type, to.url);
  });

  async function handleExploreInit(isManualUrlChange: boolean) {
    if (!exploreSpec || !metricsSpec || initializing) return;
    initializing = true;

    let initState: Partial<MetricsExplorerEntity> | undefined;
    let shouldUpdateUrl = false;
    if (exploreStateFromSessionStorage && !isManualUrlChange) {
      // if there is state in session storage then merge state from config yaml with the state from session storage
      initState = {
        ...exploreStateFromYAMLConfig,
        ...exploreStateFromSessionStorage,
      };
      shouldUpdateUrl = true;
    } else if ($page.url.searchParams.size === 0) {
      // when there are no params set, state will be state from config yaml and any additional initial state like bookmark
      initState = {
        ...exploreStateFromYAMLConfig,
        // if the url changed manually then do not load data from initState, which is home bookmark or shared url's state
        ...(isManualUrlChange ? {} : (initExploreState ?? {})),
      };
      shouldUpdateUrl = !!initExploreState;
    } else {
      // else merge with explore from url
      initState = {
        ...exploreStateFromYAMLConfig,
        ...partialExploreStateFromUrl,
      };
    }

    if (metricsSpec.timeDimension) {
      // time range summary query has `enabled` based on `metricsSpec.timeDimension`
      // isLoading will never be true when the query is disabled, so we need this check before waiting for it.
      await waitUntil(() => !timeRangeSummaryIsLoading);

      [initState.selectedTimeRange, initState.selectedComparisonTimeRange] =
        await resolveTimeRanges(
          exploreSpec,
          [initState.selectedTimeRange, initState.selectedComparisonTimeRange],
          initState.selectedTimezone,
        );
    }
    metricsExplorerStore.init(exploreName, initState);
    timeControlsState ??= getTimeControlState(
      metricsSpec,
      exploreSpec,
      timeRangeSummaryResp?.timeRangeSummary,
      get(metricsExplorerStore).entities[exploreName],
    );
    const redirectUrl = new URL($page.url);
    redirectUrl.search = getUpdatedUrlForExploreState(
      exploreSpec,
      timeControlsState,
      defaultExplorePreset,
      initState,
      $page.url.searchParams,
    );
    // update session store to make sure updated to url or the initial state is propagated to the session store
    updateExploreSessionStore(
      exploreName,
      extraKeyPrefix,
      get(metricsExplorerStore).entities[exploreName],
      exploreSpec,
      timeControlsState,
    );
    prevUrl = redirectUrl.toString();

    if (!shouldUpdateUrl || redirectUrl.search === $page.url.search) {
      return;
    }

    // using `replaceState` directly messes up the navigation entries,
    // `from` and `to` have the old url before being replaced in `afterNavigate` calls leading to incorrect handling.
    void goto(redirectUrl, {
      replaceState: true,
      state: $page.state,
    });
  }

  async function handleAfterNavigate(type, to: URL) {
    if (!exploreSpec || !metricsSpec) return;

    const isInit =
      !$dashboardStore || !hasSessionStorageData(exploreName, extraKeyPrefix);
    if (isInit) {
      // When a user changes url manually and clears the params the `type` will be "enter"
      // This signal is used in handleExploreInit to make sure we do not use sessionStorage
      const isManualUrlChange = type === "enter";
      void handleExploreInit(isManualUrlChange);
      return;
    }

    // Pressing back button and going back to empty url state should not restore from session store
    const backButtonUsed = type === "popstate";
    const skipSessionStorage =
      backButtonUsed && $page.url.searchParams.size === 0;

    let partialExplore = partialExploreStateFromUrl;
    let shouldUpdateUrl = false;
    if (exploreStateFromSessionStorage && !skipSessionStorage) {
      partialExplore = exploreStateFromSessionStorage;
      shouldUpdateUrl = true;
    }

    if (metricsSpec.timeDimension) {
      [
        partialExplore.selectedTimeRange,
        partialExplore.selectedComparisonTimeRange,
      ] = await resolveTimeRanges(
        exploreSpec,
        [
          partialExplore.selectedTimeRange,
          partialExplore.selectedComparisonTimeRange,
        ],
        partialExplore.selectedTimezone ??
          get(metricsExplorerStore).entities[exploreName].selectedTimezone,
      );
    }

    const redirectUrl = new URL(to);
    metricsExplorerStore.mergePartialExplorerEntity(
      exploreName,
      partialExplore,
      metricsSpec,
    );
    if (shouldUpdateUrl) {
      // if we added extra url params from sessionStorage then update the url
      redirectUrl.search = getUpdatedUrlForExploreState(
        exploreSpec,
        timeControlsState,
        defaultExplorePreset,
        partialExplore,
        $page.url.searchParams,
      );
    }
    // update session store when back button was pressed.
    if (backButtonUsed) {
      updateExploreSessionStore(
        exploreName,
        extraKeyPrefix,
        $dashboardStore,
        exploreSpec,
        timeControlsState,
      );
    }

    if (
      !shouldUpdateUrl ||
      redirectUrl.search === to.toString() ||
      // redirect loop breaker
      (prevUrl && prevUrl === redirectUrl.toString())
    ) {
      prevUrl = redirectUrl.toString();
      return;
    }

    prevUrl = redirectUrl.toString();
    // using `replaceState` directly messes up the navigation entries,
    // `from` and `to` have the old url before being replaced in `afterNavigate` calls leading to incorrect handling.
    void goto(redirectUrl, {
      replaceState: true,
      state: $page.state,
    });
  }

  function gotoNewState() {
    if (!exploreSpec) return;

    const u = new URL(
      `${$page.url.protocol}//${$page.url.host}${$page.url.pathname}`,
    );
    u.search = convertExploreStateToURLSearchParams(
      $dashboardStore,
      exploreSpec,
      timeControlsState,
      defaultExplorePreset,
    );
    const newUrl = u.toString();
    if (!prevUrl || prevUrl === newUrl) return;

    prevUrl = newUrl;
    // dashboard changed so we should update the url
    void goto(newUrl);
    // also update the session store
    updateExploreSessionStore(
      exploreName,
      extraKeyPrefix,
      $dashboardStore,
      exploreSpec,
      timeControlsState,
    );
  }

  // reactive to only dashboardStore
  // but gotoNewState checks other fields
  $: if ($dashboardStore) {
    gotoNewState();
  }
</script>

{#if schemaError}
  <ErrorPage
    statusCode={schemaError?.response?.status}
    header="Error loading dashboard"
    body="Unable to fetch the schema for this dashboard."
    detail={schemaError?.response?.data?.message}
  />
{:else if timeRangeSummaryError}
  <ErrorPage
    statusCode={timeRangeSummaryError?.response?.status}
    header="Error loading dashboard"
    body="Unable to fetch the time range for this dashboard."
    detail={timeRangeSummaryError?.response?.data?.message}
  />
{:else if $dashboardStore}
  <slot />
{/if}
