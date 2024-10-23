import type { Page } from "playwright";
import { clickMenuButton, openFileNavEntryContextMenu } from "./commonHelpers";
import { waitForFileNavEntry } from "./waitHelpers";

export async function createExploreFromSource(
  page: Page,
  sourcePath = "/sources/AdBids.yaml",
  metricsViewPath = "/metrics/AdBids_metrics.yaml",
) {
  await openFileNavEntryContextMenu(page, sourcePath);
  await clickMenuButton(page, "Generate metrics");
  await waitForFileNavEntry(page, metricsViewPath, true);
  await page.getByText("Create explore").click();
}

export async function createExploreFromModel(
  page: Page,
  navigateToFile = true,
  modelPath = "/models/AdBids_model.sql",
  metricsViewPath = "/metrics/AdBids_model_metrics.yaml",
) {
  await openFileNavEntryContextMenu(page, modelPath);
  await clickMenuButton(page, "Generate metrics");
  await waitForFileNavEntry(page, metricsViewPath, true);
  await page.getByText("Create explore").click();
  if (navigateToFile) {
    await page.getByRole("button", { name: "Edit" }).click();
    await page.getByRole("menuitem", { name: "Explore" }).click();
  }
}
