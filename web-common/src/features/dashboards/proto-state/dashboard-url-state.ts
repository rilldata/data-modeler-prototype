import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { createTimeRangeSummary } from "@rilldata/web-common/features/dashboards/selectors/index";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { getUrlForPath } from "@rilldata/utils";
import type { V1StructType } from "@rilldata/web-common/runtime-client";
import { type Readable, derived, get } from "svelte/store";

export type DashboardUrlState = {
  isReady: boolean;
  proto?: string;
  defaultProto?: string;
  urlName?: string;
  urlProto?: string;
  isPublicUrl?: boolean;
};
export type DashboardUrlStore = Readable<DashboardUrlState>;

/**
 * Creates a derived store that has the current proto and default proto of a dashboard along with the proto in the url.
 */
export function useDashboardUrlState(ctx: StateManagers): DashboardUrlStore {
  return derived(
    [useDashboardProto(ctx), useDashboardDefaultProto(ctx), page],
    ([proto, defaultProtoState, page], set) => {
      if (defaultProtoState.isFetching) {
        set({ isReady: false });
        return;
      }

      const defaultProto = defaultProtoState.proto;
      const urlProto = page.url.searchParams.get("state");
      const decodedUrlProto = urlProto
        ? decodeURIComponent(urlProto)
        : defaultProto;

      const urlName = getMetricsViewNameFromParams(page.params);
      set({
        isReady: true,
        proto,
        defaultProto,
        urlName,
        urlProto: decodedUrlProto,
        isPublicUrl: !urlName,
      });
    },
  );
}

/**
 * Code that looks at dashboard state and url state and decides which one to sync with.
 * This depends on the fact that the same dashboard state results in the same proto and vice-versa.
 *
 * Case 1
 * 1. The dashboard state changes due to user interactions.
 * 2. Proto in the state is updated to match the new state.
 * 3. `lastKnownProto` is now different from the proto in state.
 * 4. This triggers a goto to the url with the correct proto.
 * 5. After navigation `urlProto` changes. Since this will be equal to `lastKnownProto` there will be no more operations.
 *
 * Case 2
 * 1. The url is changed by using the back button (or any other way).
 * 2. `urlProto` changes to reflect the one in the new url.
 * 3. `lastKnownProto` is now different to the `urlProto`.
 * 4. This triggers a sync of state in the url to the dashboard store.
 * 5. After updating the store proto in the state will be the same as `lastKnownProto`. No navigations happen.
 */
export function useDashboardUrlSync(ctx: StateManagers, schema: V1StructType) {
  const dashboardUrlState = useDashboardUrlState(ctx);

  let lastKnownProto = get(dashboardUrlState)?.defaultProto;
  return dashboardUrlState.subscribe((state) => {
    const exploreName = get(ctx.exploreName);

    // Avoid a race condition when switching between metrics views
    // (It's not necessary for Public URLs because there's no UI flow for switching from one Public URL to another)
    if (
      !state.isPublicUrl &&
      state?.urlName?.toLowerCase() !== exploreName.toLowerCase()
    )
      return;

    if (!state.isReady || !state.proto) return;

    if (state.proto !== lastKnownProto) {
      // changed when filters etc are changed on the dashboard
      gotoNewDashboardUrl(get(page).url, state.proto, state.defaultProto!);

      lastKnownProto = state.proto;
    } else if (state.urlProto !== lastKnownProto) {
      // changed when user updated the url manually
      metricsExplorerStore.syncFromUrl(
        exploreName,
        state.urlProto!,
        get(ctx.validSpecStore).data?.metricsView ?? {},
        get(ctx.validSpecStore).data?.explore ?? {},
        schema,
      );
      lastKnownProto = state.urlProto;
    }
  });
}

function gotoNewDashboardUrl(url: URL, newState: string, defaultState: string) {
  // this store the actual state. for default state it will be empty
  let newStateInUrl = "";
  // changed when filters etc are changed on the dashboard

  const newUrl = getUrlForPath(get(page).url, url.pathname, [
    "features",
    "theme",
  ]);

  if (newState !== defaultState) {
    newStateInUrl = newState;
    newUrl.searchParams.set("state", newStateInUrl);
  }

  const currentStateInUrl = url.searchParams.get("state") ?? "";

  if (newStateInUrl === currentStateInUrl) return;
  goto(newUrl.toString());
}

/**
 * Pulls data from dashboard to create the url state.
 * Any subsections in the future would be added here to build the final url state.
 */
export function useDashboardProto(ctx: StateManagers) {
  return derived([ctx.dashboardStore], ([dashboard]) =>
    getProtoFromDashboardState(dashboard),
  );
}

// NOTE: the data here can be stale when metricsViewName changes in ctx, along with the metricsView.
//       but the time range summary is yet to be triggered to change causing it to have data from previous active dashboard
// Above issue is currently fixed in useDashboardUrlSync and DashboardURLStateProvider by create a new instance when the url is changed.
// TODO: we need to update the architecture to perhaps recreate all derived stores when the metricsViewName changes
export function useDashboardDefaultProto(ctx: StateManagers) {
  return derived(
    [ctx.validSpecStore, createTimeRangeSummary(ctx)],
    ([validSpec, timeRangeSummary]) => {
      const hasTimeSeries = Boolean(validSpec.data?.metricsView?.timeDimension);
      if (
        !validSpec.data?.metricsView ||
        !validSpec.data?.explore ||
        (hasTimeSeries && !timeRangeSummary.data)
      )
        return {
          isFetching: true,
          proto: "",
        };

      const metricsExplorer = getDefaultMetricsExplorerEntity(
        get(ctx.metricsViewName),
        validSpec.data.metricsView,
        validSpec.data.explore,
        timeRangeSummary.data,
      );
      return {
        isFetching: false,
        proto: getProtoFromDashboardState(metricsExplorer),
      };
    },
  );
}
/**
 * We have different ways of getting the metrics view name, depending on the context:
 * - The `dashboard` URL param is used in Cloud
 * - The `name` URL param is used in Ril Developer
 * - The `token` URL param is used in Public URLs. The metrics view name is embedded in the token.
 */
function getMetricsViewNameFromParams(
  params: Record<string, string>,
): string | undefined {
  const { dashboard, name } = params;

  if (dashboard) return dashboard;
  if (name) return name;
  // TODO: Add support for public urls' `token` param

  return undefined;
}
