import { createExploreFromModel } from "../utils/exploreHelpers";
import { createAdBidsModel } from "../utils/dataSpecifcHelpers";
import { test } from "../utils/test";

export function useDashboardFlowTestSetup() {
  test.beforeEach(async ({ page }) => {
    await createAdBidsModel(page);
    await createExploreFromModel(page);
  });
}
