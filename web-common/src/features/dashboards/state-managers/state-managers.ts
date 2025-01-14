import {
  contextColWidthDefaults,
  type ContextColWidths,
  type MetricsExplorerEntity,
} from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { createPersistentDashboardStore } from "@rilldata/web-common/features/dashboards/stores/persistent-dashboard-state";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
import {
  type ExploreValidSpecResponse,
  useExploreValidSpec,
} from "@rilldata/web-common/features/explores/selectors";
import { dedupe } from "@rilldata/web-common/lib/arrayUtils";
import {
  createQueryServiceMetricsViewTimeRange,
  createQueryServiceMetricsViewTimeRanges,
  type RpcStatus,
  type V1ExplorePreset,
  type V1MetricsViewTimeRangeResponse,
  type V1MetricsViewTimeRangesResponse,
} from "@rilldata/web-common/runtime-client";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient, QueryObserverResult } from "@tanstack/svelte-query";
import { getContext } from "svelte";
import {
  derived,
  get,
  type Readable,
  type Writable,
  writable,
} from "svelte/store";
import {
  metricsExplorerStore,
  type MetricsExplorerStoreType,
  updateMetricsExplorerByName,
  useExploreState,
} from "web-common/src/features/dashboards/stores/dashboard-stores";
import { createStateManagerActions, type StateManagerActions } from "./actions";
import type { DashboardCallbackExecutor } from "./actions/types";
import {
  createStateManagerReadables,
  type StateManagerReadables,
} from "./selectors";

export type StateManagers = {
  runtime: Writable<Runtime>;
  metricsViewName: Writable<string>;
  exploreName: Writable<string>;
  metricsStore: Readable<MetricsExplorerStoreType>;
  dashboardStore: Readable<MetricsExplorerEntity>;
  timeRangeSummaryStore: Readable<
    QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>
  >;
  timeRanges: Readable<
    QueryObserverResult<V1MetricsViewTimeRangesResponse, RpcStatus>
  >;
  validSpecStore: Readable<
    QueryObserverResult<ExploreValidSpecResponse, RpcStatus>
  >;
  queryClient: QueryClient;
  updateDashboard: DashboardCallbackExecutor;
  /**
   * A collection of Readables that can be used to select data from the dashboard.
   */
  selectors: StateManagerReadables;
  /**
   * A collection of functions that update the dashboard data model.
   */
  actions: StateManagerActions;
  /**
   * Store to track the width of the context columns in leaderboards.
   * FIXME: this was implemented as a low-risk fix for in advance of
   * the new branding release 2024-01-31, but should be revisted since
   * it's a one-off solution that introduces another new pattern.
   */
  contextColumnWidths: Writable<ContextColWidths>;
  defaultExploreState: Readable<V1ExplorePreset>;
};

export const DEFAULT_STORE_KEY = Symbol("state-managers");

export function getStateManagers(): StateManagers {
  return getContext(DEFAULT_STORE_KEY);
}

export function createStateManagers({
  queryClient,
  metricsViewName,
  exploreName,
  extraKeyPrefix,
}: {
  queryClient: QueryClient;
  metricsViewName: string;
  exploreName: string;
  extraKeyPrefix?: string;
}): StateManagers {
  const metricsViewNameStore = writable(metricsViewName);
  const exploreNameStore = writable(exploreName);

  const dashboardStore: Readable<MetricsExplorerEntity> = derived(
    [exploreNameStore],
    ([name], set) => {
      const exploreState = useExploreState(name);
      return exploreState.subscribe(set);
    },
  );

  const validSpecStore: Readable<
    QueryObserverResult<ExploreValidSpecResponse, RpcStatus>
  > = derived([runtime, exploreNameStore], ([r, exploreName], set) =>
    useExploreValidSpec(r.instanceId, exploreName, { queryClient }).subscribe(
      set,
    ),
  );

  const timeRangeSummaryStore: Readable<
    QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>
  > = derived(
    [runtime, metricsViewNameStore, validSpecStore],
    ([runtime, mvName, validSpec], set) =>
      createQueryServiceMetricsViewTimeRange(
        runtime.instanceId,
        mvName,
        {},
        {
          query: {
            queryClient,
            enabled: !!validSpec?.data?.metricsView?.timeDimension,
            staleTime: Infinity,
            cacheTime: Infinity,
          },
        },
      ).subscribe(set),
  );

  const timeRanges: Readable<
    QueryObserverResult<V1MetricsViewTimeRangesResponse, RpcStatus>
  > = derived(
    [runtime, metricsViewNameStore, validSpecStore],
    ([runtime, mvName, validSpec], set) => {
      if (!validSpec.data?.explore) {
        return;
      }

      const explore = validSpec.data.explore;
      const defaultPreset = explore.defaultPreset ?? {};
      const rillTimes = dedupe([
        "inf",
        ...(defaultPreset.timeRange ? [defaultPreset.timeRange] : []),
        ...(explore.timeRanges?.length
          ? explore.timeRanges.map((t) => t.range!)
          : []),
      ]);

      createQueryServiceMetricsViewTimeRanges(runtime.instanceId, mvName, {
        expressions: rillTimes,
      }).subscribe(set);
    },
  );

  const updateDashboard = (
    callback: (metricsExplorer: MetricsExplorerEntity) => void,
  ) => {
    const name = get(dashboardStore).name;
    // TODO: Remove dependency on MetricsExplorerStore singleton and its exports
    updateMetricsExplorerByName(name, callback);
  };

  const contextColumnWidths = writable<ContextColWidths>(
    contextColWidthDefaults,
  );

  const defaultExploreState = derived(
    [validSpecStore, timeRangeSummaryStore],
    ([validSpec, timeRangeSummary]) => {
      if (!validSpec.data?.explore) {
        return {};
      }
      return getDefaultExplorePreset(
        validSpec.data?.explore ?? {},
        timeRangeSummary.data,
      );
    },
  );

  const persistentDashboardStore = createPersistentDashboardStore(
    (extraKeyPrefix || "") + exploreName,
  );
  initLocalUserPreferenceStore(exploreName);

  return {
    runtime: runtime,
    metricsViewName: metricsViewNameStore,
    exploreName: exploreNameStore,
    metricsStore: metricsExplorerStore,
    timeRangeSummaryStore,
    timeRanges,
    validSpecStore,
    queryClient,
    dashboardStore,

    updateDashboard,
    /**
     * A collection of Readables that can be used to select data from the dashboard.
     */
    selectors: createStateManagerReadables({
      dashboardStore,
      validSpecStore,
      timeRangeSummaryStore,
      timeRanges,
      queryClient,
    }),
    /**
     * A collection of functions that update the dashboard data model.
     */
    actions: createStateManagerActions({
      updateDashboard,
      persistentDashboardStore,
    }),
    contextColumnWidths,
    defaultExploreState,
  };
}
