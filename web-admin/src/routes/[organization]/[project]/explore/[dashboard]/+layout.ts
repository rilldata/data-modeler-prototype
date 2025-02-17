import type { V1Bookmark } from "@rilldata/web-admin/client";
import {
  fetchBookmarks,
  isHomeBookmark,
} from "@rilldata/web-admin/features/bookmarks/selectors";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  fetchExploreSpec,
  fetchMetricsViewSchema,
} from "@rilldata/web-common/features/explores/selectors";
import {
  type V1ExplorePreset,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";

export const load = async ({ params, depends, parent }) => {
  const { user, project, runtime } = await parent();

  const { dashboard: exploreName } = params;

  depends(`explore:${exploreName}`);

  let explore: V1Resource | undefined;
  let metricsView: V1Resource | undefined;
  let defaultExplorePreset: V1ExplorePreset | undefined;
  let exploreStateFromYAMLConfig: Partial<MetricsExplorerEntity> = {};
  let bookmarks: V1Bookmark[] | undefined;

  try {
    [
      {
        explore,
        metricsView,
        defaultExplorePreset,
        exploreStateFromYAMLConfig,
      },
      bookmarks,
    ] = await Promise.all([
      fetchExploreSpec(runtime?.instanceId, exploreName),
      // public projects might not have a logged-in user. bookmarks are not available in this case
      user ? fetchBookmarks(project.id, exploreName) : Promise.resolve([]),
    ]);
  } catch {
    // error handled in +page.svelte for now
    // TODO: move it here
    return {
      explore: <V1Resource>{},
      metricsView: <V1Resource>{},
      defaultExplorePreset: <V1ExplorePreset>{},
      exploreStateFromYAMLConfig,
    };
  }

  const metricsViewSpec = metricsView.metricsView?.state?.validSpec ?? {};
  const exploreSpec = explore.explore?.state?.validSpec ?? {};

  let homeBookmarkExploreState: Partial<MetricsExplorerEntity> | undefined =
    undefined;
  try {
    const homeBookmark = bookmarks.find(isHomeBookmark);
    const schema = await fetchMetricsViewSchema(
      runtime?.instanceId,
      exploreSpec.metricsView ?? "",
    );

    if (homeBookmark) {
      homeBookmarkExploreState = getDashboardStateFromUrl(
        homeBookmark.data ?? "",
        metricsViewSpec,
        exploreSpec,
        schema,
      );
    }
  } catch {
    // TODO
  }

  return {
    explore,
    metricsView,
    defaultExplorePreset,
    exploreStateFromYAMLConfig,
    homeBookmarkExploreState,
  };
};
