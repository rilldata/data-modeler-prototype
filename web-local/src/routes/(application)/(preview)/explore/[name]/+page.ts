import { getExploreStates } from "@rilldata/web-common/features/explores/selectors";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset } = await parent();
  const { name: exploreName } = params;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  return {
    exploreName,
    ...getExploreStates(
      exploreName,
      undefined,
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    ),
  };
};
