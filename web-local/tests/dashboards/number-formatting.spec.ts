import { useDashboardFlowTestSetup } from "web-local/tests/dashboards/dashboard-flow-test-setup";
import { ResourceWatcher } from "web-local/tests/utils/ResourceWatcher";
import { interactWithTimeRangeMenu } from "../utils/dashboardHelpers";
import { expect } from "@playwright/test";
import { test } from "../utils/test";

test.describe("smoke tests for number formatting", () => {
  useDashboardFlowTestSetup();

  test("smoke tests for number formatting", async ({ page }) => {
    const watcher = new ResourceWatcher(page);

    // This is a metrics spec with all available formatting options
    const formatterFlowDashboard = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.
title: "AdBids_model_dashboard"
model: "AdBids_model"
default_time_range: ""
smallest_time_grain: ""
timeseries: "timestamp"
measures:
- label: humanized default
  expression: count(*)
  name: total_rows
  description: Total number of records present
- label: USD
  expression: sum(bid_price)
  name: total_bid_price
  format_preset: currency_usd
  valid_percent_of_total: true
- label: humanized chosen
  expression: sum(bid_price)
  name: total_humanize
  format_preset: humanize
  valid_percent_of_total: true
- label: No Format
  expression: sum(bid_price)
  name: total_none
  format_preset: none
  valid_percent_of_total: true
- label: percentage
  expression: sum(bid_price)
  name: total_percentage
  format_preset: percentage
  valid_percent_of_total: true
- label: interval_ms
  expression: sum(bid_price)
  name: total_interval_ms
  format_preset: interval_ms
  valid_percent_of_total: true
- label: d3 fixed
  expression: sum(bid_price)
  name: total_d3_fixed_points
  format_d3: ".3f"
  valid_percent_of_total: true
dimensions:
- name: publisher
  label: Publisher
  column: publisher
  description: ""
- name: domain
  label: Domain
  column: domain
  description: ""
`;

    // update the code editor with the new spec
    await watcher.updateAndWaitForDashboard(formatterFlowDashboard);

    const previewButton = page.getByRole("button", { name: "Preview" });

    await previewButton.click();

    /******************
     * check big nums
     ******************/
    for (const [name, bignum, tooltip_num] of [
      ["humanized default", "100.0k", "100000"],
      ["USD", "$300.6k", "300576.83999999857"],
      ["humanized chosen", "300.6k", "300576.83999999857"],
      ["No Format", "300576.83999999857", "300576.83999999857"],
      ["percentage", "30.1M%", "300576.83999999857"],
      ["interval_ms", "5 m", "5m 576ms"],
      ["d3 fixed", "300576.840", "300576.840"],
    ]) {
      // check bignum with correct format exists/is visible
      await expect(
        page.getByRole("button", { name: `${name} ${bignum}` }),
      ).toBeVisible();
      // hover over btn_name
      await page.getByRole("button", { name: `${name} ${bignum}` }).hover();
      // wait for a moment for the tooltip to appear
      await page
        .getByText(`${name} ${tooltip_num}`)
        .waitFor({ state: "visible" });
    }

    /******************
     * check leaderboard
     *
     * note that the leaderboard is shown with
     * "humanized default" format initially.
     *
     * This is a smoke test, so we won't check every format,
     * but a few combinations of format and context column.
     *
     ******************/

    const measuresButton = page.getByRole("button", {
      name: "Select a measure to filter by",
    });
    await measuresButton.click();
    await page.getByRole("menuitem", { name: "USD" }).click();
    await page
      .getByRole("menu", { name: "Showing USD" })
      .waitFor({ state: "hidden" });
    await expect(measuresButton).toHaveText("Showing USD");

    await expect(
      page.getByRole("row", { name: "null $98.8k 33%" }),
    ).toBeVisible();
    await measuresButton.click();
    await page.getByRole("menuitem", { name: "percentage" }).click();
    await expect(measuresButton).toHaveText("Showing percentage");
    await expect(
      page.getByRole("row", { name: "null 9.9M% 33%" }),
    ).toBeVisible();

    // try interval_ms...
    await measuresButton.click();
    await page.getByRole("menuitem", { name: "interval_ms" }).click();
    await expect(measuresButton).toHaveText("Showing interval_ms");
    // ...and add a time comparison to check deltas
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 4 Weeks" }).click();
    });
    await page.getByRole("button", { name: "Comparing" }).click();

    await expect(
      page.getByRole("row", { name: "null 27 s -4.3 s -14%" }),
    ).toBeVisible();

    // try No Format...
    await measuresButton.click();
    await page.getByRole("menuitem", { name: "No Format" }).click();
    await expect(measuresButton).toHaveText("Showing No Format");

    await expect(
      page.getByRole("row", {
        name: "null 26642.549999999974 -4348.7299999999705 -14%",
      }),
    ).toBeVisible();

    /******************
     * check dimension table
     *
     * just smoke testing, so we'll just check one value
     * per column.
     ******************/
    await page.getByText("Publisher").click();

    // humanized default
    await expect(
      page
        .locator("div")
        .filter({ hasText: /^8\.9k$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    ).toBeVisible();

    // USD
    await expect(
      page
        .locator("div")
        .filter({ hasText: /^\$26\.6k$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    ).toBeVisible();

    // humanized chosen
    await expect(
      page
        .locator("div")
        .filter({ hasText: /^26\.6k$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    ).toBeVisible();

    // No Format
    await expect(
      page
        .locator("div")
        .filter({ hasText: /^26642\.549999999974$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    ).toBeVisible();

    // No Format - context column, delta
    await expect(
      page
        .locator("div")
        .filter({ hasText: /^-4348\.7299999999705$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    ).toBeVisible();

    // No Format - context column, delta pct
    await expect(
      page
        .getByRole("table", { name: "Dimension table" })
        .locator("div")
        .filter({ hasText: /^-14%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    ).toBeVisible();

    // No Format - context column, pct of total
    await expect(
      page
        .locator("div")
        .filter({ hasText: /^33%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    ).toBeVisible();

    // percentage
    await expect(
      page
        .locator("div")
        .filter({ hasText: /^383\.4k%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    ).toBeVisible();

    // interval_ms
    await expect(
      page
        .locator("div")
        .filter({ hasText: /^3\.8 s$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    ).toBeVisible();

    // d3 fixed
    await expect(
      page
        .locator("div")
        .filter({ hasText: /^26642\.550$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    ).toBeVisible();
  });
});
