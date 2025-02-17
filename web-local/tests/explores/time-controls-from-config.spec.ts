import { expect } from "@playwright/test";
import { interactWithTimeRangeMenu } from "web-local/tests/utils/metricsViewHelpers";
import { ResourceWatcher } from "web-local/tests/utils/ResourceWatcher";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../utils/test";

test.describe("time controls settings from explore preset", () => {
  test.use({
    projectInit: { name: "AdBids" },
  });

  test("preset time_range", async ({ page }) => {
    const watcher = new ResourceWatcher(page);

    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByLabel("code").click();

    // Set a time range that is one of the supported presets
    await watcher.updateAndWaitForExplore(
      getDashboardYaml(`time_range: "P4W"
  comparison_mode: time
`),
    );

    await page.waitForTimeout(1000);
    // Preview
    await page.getByRole("button", { name: "Preview" }).click();

    // Time range has changed
    await expect(page.getByText("Last 4 Weeks")).toBeVisible();
    // Data has changed as well
    await expect(
      page.getByText("Total records 26,687 -4,732 -15%"),
    ).toBeVisible();
    await expect(page.getByText("Facebook 7.0k 2.8k 67%")).toBeVisible();
    await page.getByRole("button", { name: "Edit" }).click();
    await page.getByRole("menuitem", { name: "Explore" }).click();

    // Set a time range that is one of the period to date preset
    await watcher.updateAndWaitForExplore(
      getDashboardYaml(`time_range: "rill-WTD"
  comparison_mode: time
`),
    );

    // Preview
    await page.getByRole("button", { name: "Preview" }).click();

    // Time range has changed
    await expect(page.getByText("Week to Date")).toBeVisible();
    // Data has changed as well
    await expect(page.getByText("Total records 3,435 +156 5%")).toBeVisible();
    await expect(page.getByText("Facebook 889 36 4%")).toBeVisible();

    // Select a different time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 7 Days" }).click();
    });
    // Wait for menu to close
    await expect(
      page.getByRole("menuitem", { name: "Last 7 Days" }),
    ).not.toBeVisible();
    // Data has changed
    await expect(page.getByText("Total records 7,863 -15 ~0%")).toBeVisible();
    await expect(page.getByText("Facebook 2.0k -51 -2%")).toBeVisible();
    await page.getByRole("button", { name: "Edit" }).click();
    await page.getByRole("menuitem", { name: "Explore" }).click();

    // Set a time range that is not one of the supported presets
    await watcher.updateAndWaitForExplore(
      getDashboardYaml(`time_range: "P2W"
  comparison_mode: time
`),
    );
    // Preview
    await page.getByRole("button", { name: "Preview" }).click();

    // Time range has changed
    await expect(page.getByText("Last 2 Weeks")).toBeVisible();
    // Data has changed as well
    await expect(
      page.getByText("Total records 11,193 -4,301 -28%"),
    ).toBeVisible();
    await expect(page.getByText("Facebook 2.9k -1.2k -29%")).toBeVisible();
  });

  test("preset comparison_modes", async ({ page }) => {
    const watcher = new ResourceWatcher(page);

    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByLabel("code").click();

    // Set comparison to time
    await watcher.updateAndWaitForExplore(
      getDashboardYaml(`time_range: "P4W"
  comparison_mode: time
`),
    );
    // Preview
    await page.getByRole("button", { name: "Preview" }).click();
    // Comparison is selected
    await expect(page.getByRole("switch", { name: "Comparing" })).toBeChecked();
    // Go back to metrics editor
    await page.getByRole("button", { name: "Edit" }).click();
    await page.getByRole("menuitem", { name: "Explore" }).click();
    // Set comparison to dimension
    await watcher.updateAndWaitForExplore(
      getDashboardYaml(`time_range: "P4W"
  comparison_mode: dimension
  comparison_dimension: publisher
`),
    );
    // Preview
    await page.getByRole("button", { name: "Preview" }).click();
    // Comparison is selected
    await expect(
      page
        .getByLabel("publisher leaderboard")
        .getByLabel("Comparison column")
        .getByLabel("Toggle breakdown for publisher dimension"),
    ).toBeVisible();
    // Go back to metrics editor
    await page.getByRole("button", { name: "Edit" }).click();
    await page.getByRole("menuitem", { name: "Explore" }).click();
    // Set comparison to none
    await watcher.updateAndWaitForExplore(
      getDashboardYaml(`time_range: "P4W"
  comparison_mode: dimension
`),
    );
    // Preview
    await page.getByRole("button", { name: "Preview" }).click();
  });

  test("preset time_ranges", async ({ page }) => {
    const watcher = new ResourceWatcher(page);

    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByLabel("code").click();

    await watcher.updateAndWaitForExplore(
      getDashboardYaml(
        `time_range: "P4W"
  comparison_mode: time
`,
        `time_ranges:
- PT6H
- range: P5D
  comparison_offsets:
    - rill-PP
    - rill-PW
- P4W
- rill-WTD
- rill-MTD`,
      ),
    );
    // Preview
    await page.getByRole("button", { name: "Preview" }).click();

    // Open the time range menu
    await page.getByLabel("Select time range").click();
    // Assert the options available
    await Promise.all(
      [
        "All Time",
        "Last 6 Hours",
        "Last 5 Days",
        "Last 4 Weeks",
        "Week To Date",
        "Month To Date",
      ].map((label) =>
        expect(page.getByRole("menuitem", { name: label })).toBeVisible(),
      ),
    );
    // Select Last 6 hours
    await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    // Wait for time range menu to close
    await expect(
      page.getByRole("menu", { name: "Select time range" }),
    ).not.toBeVisible();
    // Assert data has changed
    await expect(page.getByText("Total records 272 -23 -8%")).toBeVisible();
    await expect(page.getByText("Facebook 68 -3 -4%")).toBeVisible();

    // Open the time comparison
    await page
      .getByLabel("Select time comparison option")
      .click({ force: true });
    // Assert the options available
    await Promise.all(
      [
        "Previous Period",
        "Previous day",
        "Previous week",
        "Previous month",
      ].map((label) =>
        expect(page.getByRole("menuitem", { name: label })).toBeVisible(),
      ),
    );
    // Select Previous week
    await page.getByRole("menuitem", { name: "Previous week" }).click();
    // Wait for time range menu to close
    await expect(
      page.getByRole("menu", { name: "Time comparison selector" }),
    ).not.toBeVisible();
    // Assert data has changed
    await expect(page.getByText("Total records 272 -18 -6%")).toBeVisible();
    await expect(page.getByText("Facebook 68 -24 -26%")).toBeVisible();

    // Select Last 5 days
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 5 Days" }).click();
    });
    // Assert data has changed
    await expect(page.getByText("Total records 5,585 +16 ~0%")).toBeVisible();
    await page.pause();
    await expect(page.getByText("Facebook 1.5k -25 -2%")).toBeVisible();

    // Open the time comparison
    await page.getByLabel("Select time comparison option").click();
    // Assert the options available
    await Promise.all(
      ["Previous Period", "Previous week"].map((label) =>
        expect(page.getByRole("menuitem", { name: label })).toBeVisible(),
      ),
    );
    // Select Last 6 hours
    await page.getByRole("menuitem", { name: "Previous Period" }).click();
    // Wait for time range menu to close
    await expect(
      page.getByRole("menu", { name: "Time comparison selector" }),
    ).not.toBeVisible();
    // Assert data has changed
    await expect(page.getByText("Total records 5,585 -23 ~0%")).toBeVisible();
    await expect(page.getByText("Facebook 1.5k -6 ~0%")).toBeVisible();
  });
});

function getDashboardYaml(defaults: string, extras = "") {
  return `
type: explore
title: "AdBids_metrics_explore"
metrics_view: "AdBids_metrics"
${extras}

measures: '*'
dimensions: '*'

defaults:
  ${defaults}
  `;
}
