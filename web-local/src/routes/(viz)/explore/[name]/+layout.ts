import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ params, depends }) => {
  const exploreName = params.name;

  depends(exploreName, "explore");

  try {
    const { explore, metricsView, defaultExplorePreset } =
      await fetchExploreSpec(get(runtime), exploreName);

    return {
      explore,
      metricsView,
      defaultExplorePreset,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Explore not found");
  }
};
