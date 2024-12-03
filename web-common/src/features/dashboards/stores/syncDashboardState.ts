import { page } from "$app/stores";
import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import {
  createMetricsViewSchema,
  createTimeRangeSummary,
} from "@rilldata/web-common/features/dashboards/selectors/index";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { derived, get } from "svelte/store";

/**
 * createDashboardStateSync creates a store that keeps the dashboard state in sync with metrics config.
 * It derives from metrics view spec, time range summary and metrics view schema.
 *
 * For the 1st time it is run it will call `metricsExplorerStore.init` to initialise the dashboard store.
 * Optionally loads an initial url state.
 *
 * For successive runs it will call `metricsExplorerStore.sync` to keep the store in sync with metrics config.
 * `sync` will make sure any removed measures and dimensions are not selected in anything in the dashboard.
 *
 * Note that this returns a readable so that the body of the `subscribe` is executed.
 *
 * @param ctx
 * @param initialUrlStateStore Initial url state to load when the dashboard store is initialised for the 1st time.
 * @returns A boolean readable that is true once the dashbaord store is created.
 */
export function createDashboardStateSync(
  ctx: StateManagers,
  initialUrlStateStore?: CompoundQueryResult<string>,
) {
  return derived(
    [
      ctx.validSpecStore,
      createTimeRangeSummary(ctx),
      createMetricsViewSchema(ctx),
      ...(initialUrlStateStore ? [initialUrlStateStore] : []),
    ],
    ([
      validSpecRes,
      timeRangeRes,
      metricsViewSchemaRes,
      initialUrlStateRes,
    ]) => {
      if (
        // still fetching
        validSpecRes.isFetching ||
        timeRangeRes.isFetching ||
        metricsViewSchemaRes.isFetching ||
        initialUrlStateRes?.isFetching
      ) {
        return { isFetching: true, error: false };
      }

      if (
        !validSpecRes.data?.metricsView ||
        (!!validSpecRes.data.metricsView?.timeDimension &&
          !timeRangeRes.data) ||
        !validSpecRes.data?.explore ||
        !metricsViewSchemaRes.data?.schema
      ) {
        return { isFetching: false, error: true };
      }

      const { metricsView, explore } = validSpecRes.data;

      const exploreName = get(ctx.exploreName);
      if (exploreName in get(metricsExplorerStore).entities) {
        // Successive syncs with metrics view spec
        // metricsExplorerStore.sync(exploreName, explore);
      } else {
        // Running for the 1st time. Initialise the dashboard store.
        metricsExplorerStore.init(
          exploreName,
          metricsView,
          explore,
          timeRangeRes.data,
        );
        const initialUrlState =
          get(page).url.searchParams.get("state") ?? initialUrlStateRes?.data;
        if (initialUrlState) {
          // If there is data to be loaded, load it during the init
          // metricsExplorerStore.syncFromUrl(
          //   exploreName,
          //   initialUrlState,
          //   metricsView,
          //   explore,
          //   metricsViewSchemaRes.data.schema,
          // );
          // Call sync to make sure changes in dashboard are honoured
          metricsExplorerStore.sync(exploreName, explore);
        }
      }
      return { isFetching: false, error: false };
    },
  );
}
