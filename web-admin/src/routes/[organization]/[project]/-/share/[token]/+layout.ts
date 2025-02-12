import { fetchMagicAuthToken } from "@rilldata/web-admin/features/projects/selectors";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  fetchExploreSpec,
  fetchMetricsViewSchema,
} from "@rilldata/web-common/features/explores/selectors";
import { error } from "@sveltejs/kit";

export const load = async ({ params: { token }, parent }) => {
  const { runtime } = await parent();

  try {
    const tokenData = await fetchMagicAuthToken(token);

    if (!tokenData.token?.resourceName) {
      throw new Error("Token does not have an associated resource name");
    }

    const exploreName = tokenData.token?.resourceName;

    const {
      explore,
      metricsView,
      timeRanges,
      defaultExplorePreset,
      exploreStateFromYAMLConfig,
    } = await fetchExploreSpec(runtime?.instanceId, exploreName);
    const metricsViewSpec = metricsView.metricsView?.state?.validSpec ?? {};
    const exploreSpec = explore.explore?.state?.validSpec ?? {};

    let tokenExploreState: Partial<MetricsExplorerEntity> | undefined =
      undefined;
    if (tokenData.token?.state) {
      const schema = await fetchMetricsViewSchema(
        runtime?.instanceId,
        exploreSpec.metricsView ?? "",
      );
      tokenExploreState = getDashboardStateFromUrl(
        tokenData.token?.state,
        metricsViewSpec,
        exploreSpec,
        schema,
      );
    }

    return {
      explore,
      metricsView,
      timeRanges,
      defaultExplorePreset,
      exploreStateFromYAMLConfig,
      tokenExploreState,
      token: tokenData?.token,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Unable to find token");
  }
};
