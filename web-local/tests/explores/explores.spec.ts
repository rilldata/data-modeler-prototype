import { expect } from "@playwright/test";
import {
  createExploreFromModel,
  createExploreFromSource,
} from "web-local/tests/utils/exploreHelpers";
import {
  assertLeaderboards,
  interactWithTimeRangeMenu,
} from "web-local/tests/utils/metricsViewHelpers";
import { ResourceWatcher } from "web-local/tests/utils/ResourceWatcher";
import { updateCodeEditor, wrapRetryAssertion } from "../utils/commonHelpers";
import {
  AD_BIDS_EXPLORE_PATH,
  AD_BIDS_METRICS_PATH,
  assertAdBidsDashboard,
  createAdBidsModel,
} from "../utils/dataSpecifcHelpers";
import { createSource } from "../utils/sourceHelpers";
import { test } from "../utils/test";
import { gotoNavEntry } from "../utils/waitHelpers";

test.describe("explores", () => {
  test("Autogenerate explore from source", async ({ page }) => {
    await createSource(page, "AdBids.csv", "/sources/AdBids.yaml");
    await createExploreFromSource(page);
    // Temporary timeout while the issue is looked into
    await page.waitForTimeout(1000);
    await assertAdBidsDashboard(page);
  });

  test("Autogenerate explore from model", async ({ page }) => {
    await createAdBidsModel(page);
    await createExploreFromModel(page, true);
    await assertAdBidsDashboard(page);

    // Change the time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "All Time" }).click();
    });

    // click on publisher=Facebook leaderboard value
    await page.getByRole("row", { name: "Facebook 19.3k" }).click();
    await wrapRetryAssertion(() =>
      assertLeaderboards(page, [
        {
          label: "Publisher",
          values: ["null", "Facebook", "Microsoft", "Google", "Yahoo"],
        },
        {
          label: "Domain",
          values: ["facebook.com", "instagram.com"],
        },
      ]),
    );
  });

  test("Dashboard runthrough", async ({ page }) => {
    // Enable to get logs in CI
    // page.on("console", async (msg) => {
    //   console.log(msg.text());
    // });
    // page.on("pageerror", (exception) => {
    //   console.log(
    //     `Uncaught exception: "${exception.message}"\n${exception.stack}`
    //   );
    // });
    const watcher = new ResourceWatcher(page);

    await createAdBidsModel(page);
    await createExploreFromModel(page, true);

    // Change the time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "All Time" }).click();
    });

    // Check the total records are 100k
    await expect(page.getByText("Total records 100k")).toBeVisible();

    // Check the row viewer accordion is visible
    await expect(page.getByText("Model Data 100k of 100k rows")).toBeVisible();

    // Change the metric trend granularity

    const timeGrainSelector = page.getByRole("button", {
      name: "Select a time grain",
    });
    await timeGrainSelector.click();
    await page.getByRole("menuitem", { name: "day" }).click();

    // Change the time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    });

    await page.getByRole("button", { name: "Comparing" }).click();

    // Check that the total records are 272 and have comparisons
    await expect(page.getByText("272 -23 -8%")).toBeVisible();

    // Check the row viewer accordion is updated
    await expect(page.getByText("Model Data 272 of 100k rows")).toBeVisible();

    // Check row viewer is collapsed by looking for the cell value "7029", which should be in the table
    await expect(page.getByRole("cell", { name: "7029" })).not.toBeVisible();

    // Expand row viewer and check data is there
    await page.getByRole("button", { name: "Toggle rows viewer" }).click();
    await expect(page.getByRole("cell", { name: "7029" })).toBeVisible();

    await page.getByRole("button", { name: "Toggle rows viewer" }).click();
    // Check row viewer is collapsed
    await expect(page.getByRole("cell", { name: "7029" })).not.toBeVisible();

    // Download the data as CSV
    // Start waiting for download before clicking. Note no await.
    const downloadCSVPromise = page.waitForEvent("download");
    await page.getByLabel("Export model data").click();
    await page.getByRole("menuitem", { name: "Export as CSV" }).click();
    const downloadCSV = await downloadCSVPromise;
    await downloadCSV.saveAs("temp/" + downloadCSV.suggestedFilename());
    const csvRegex = /^AdBids_model_filtered_.*\.csv$/;
    expect(csvRegex.test(downloadCSV.suggestedFilename())).toBe(true);

    // Download the data as XLSX
    // Start waiting for download before clicking. Note no await.
    const downloadXLSXPromise = page.waitForEvent("download");
    await page.getByLabel("Export model data").click();
    await page.getByRole("menuitem", { name: "Export as XLSX" }).click();
    const downloadXLSX = await downloadXLSXPromise;
    await downloadXLSX.saveAs("temp/" + downloadXLSX.suggestedFilename());
    const xlsxRegex = /^AdBids_model_filtered_.*\.xlsx$/;
    expect(xlsxRegex.test(downloadXLSX.suggestedFilename())).toBe(true);

    // Download the data as Parquet
    // Start waiting for download before clicking. Note no await.
    const downloadParquetPromise = page.waitForEvent("download");
    await page.getByLabel("Export model data").click();
    await page.getByRole("menuitem", { name: "Export as Parquet" }).click();
    const downloadParquet = await downloadParquetPromise;
    await downloadParquet.saveAs("temp/" + downloadParquet.suggestedFilename());

    const parquetRegex = /^AdBids_model_filtered_.*\.parquet$/;
    expect(parquetRegex.test(downloadParquet.suggestedFilename())).toBe(true);

    // Turn off comparison
    await page.getByRole("button", { name: "Comparing" }).click();

    // Check number
    await expect(page.getByText("272", { exact: true })).toBeVisible();

    // Add comparison back
    await page.getByRole("button", { name: "Comparing" }).click();

    /*
      There is a bug where if you programmatically click the Time Range Selector button right after clicking the "Previous Period" menu item,
      the comparison menu closes, the time range menu opens, and then the comparison menu opens again. You can reproduce with a script like this in console
      after opening up comparison menu when "no comparison" is selected:
      (() => {
        document.evaluate("//button[contains(., 'previous period')]", document, null, XPathResult.ANY_TYPE, null ).iterateNext().click();
        document.querySelector('[aria-label="Select time range"]').click();
      })()

      For now, we will wait for the menu to disappear before clicking the next menu
     */
    await expect(page.getByLabel("Comparison selector")).not.toBeVisible();

    await page.getByLabel("Select time range").click();
    await page.getByRole("menuitem", { name: "Custom" }).click();

    await page.getByLabel("start date").fill("2022-02-01");
    await page.getByLabel("start date").blur();
    await page.getByRole("button", { name: "Apply" }).click();

    // Check number
    await expect(page.getByText("Total records 65,091")).toBeVisible();

    // Flip back to All Time
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "All Time" }).click();
    });

    // Check number
    await expect(page.getByText("Total records 100k")).toBeVisible();

    // Filter to Facebook via leaderboard
    await page.getByRole("row", { name: "Facebook 19.3k" }).click();

    await page.waitForSelector("text=Publisher Facebook");

    // Change filter to excluded
    await page.getByText("Publisher Facebook").click();
    await page.getByRole("button", { name: "Exclude" }).click();
    await page.getByText("Exclude Publisher Facebook").click();

    // Check number
    await expect(page.getByText("Total records 80,659")).toBeVisible();

    // Clear the filter from filter bar
    await page.getByLabel("View filter").getByLabel("Remove").click();

    // Apply a different filter
    await page.getByRole("row", { name: "google.com 15.1k" }).click();

    // Check number
    await expect(page.getByText("Total records 15,119")).toBeVisible();

    // Clear all filters button
    await page.getByRole("button", { name: "Clear filters" }).click();

    // Check number
    await expect(page.getByText("Total records 100k")).toBeVisible();

    // TODO
    //    Change time range to last 6 hours
    //    Check that the data is updated for last 6 hours
    //    Change time range back to all time

    // Edit Explore
    await page.getByRole("button", { name: "Edit" }).click();
    await page.getByRole("menuitem", { name: "Explore" }).click();

    // Get the dashboard name field and change it

    const changeDisplayNameDoc = `
type: explore

display_name: "Adbids dashboard renamed"
metrics_view: AdBids_model_metrics

dimensions: '*'
measures: '*'
`;
    await page.getByLabel("code").click();
    await watcher.updateAndWaitForExplore(changeDisplayNameDoc);

    // Remove timestamp column
    // await page.getByLabel("Remove timestamp column").click();

    await page.getByRole("button", { name: "Preview" }).click();

    // Assert that name changed
    await expect(
      page.getByRole("link", { name: "Adbids dashboard renamed" }),
    ).toBeVisible();

    // Assert that no time dimension specified
    // await expect(page.getByText("No time dimension specified")).toBeVisible();

    // Edit Explore
    await page.getByRole("button", { name: "Edit" }).click();
    await page.getByRole("menuitem", { name: "Metrics View" }).click();

    // Add timestamp column back

    const addBackTimestampColumnDoc = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

    version: 1
    type: metrics_view
    display_name: "AdBids_model_dashboard"
    model: "AdBids_model"
    default_time_range: ""
    smallest_time_grain: "week"
    timeseries: "timestamp"
    measures:
      - label: Total records
        expression: count(*)
        name: total_records
        description: Total number of records present
        format_preset: humanize
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

    await page.getByLabel("code").click();
    await watcher.updateAndWaitForDashboard(addBackTimestampColumnDoc);
    await page.getByRole("button", { name: "Go to dashboard" }).click();
    await page
      .getByRole("menuitem", { name: "Adbids dashboard renamed" })
      .click();

    // Preview
    await page.getByRole("button", { name: "Preview" }).click();

    // Assert that time dimension is now week
    await expect(timeGrainSelector).toHaveText("by Week");

    // Edit Explore
    await page.getByRole("button", { name: "Edit" }).click();
    await page.getByRole("menuitem", { name: "Explore" }).click();

    await gotoNavEntry(page, AD_BIDS_METRICS_PATH);

    // Write an incomplete measure
    const docWithIncompleteMeasure = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

    version: 1
    type: metrics_view
    display_name: "AdBids_model_dashboard"
    model: "AdBids_model"
    default_time_range: ""
    smallest_time_grain: "week"
    timeseries: "timestamp"
    measures:
      - label: Avg Bid Price
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

    await updateCodeEditor(page, docWithIncompleteMeasure);
    await gotoNavEntry(page, AD_BIDS_EXPLORE_PATH);
    await expect(page.getByRole("button", { name: "Preview" })).toBeDisabled();
    await gotoNavEntry(page, AD_BIDS_METRICS_PATH);

    // Complete the measure
    const docWithCompleteMeasure = `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

version: 1
type: metrics_view
display_name: "AdBids_model_dashboard_rename"
model: "AdBids_model"
default_time_range: ""
smallest_time_grain: "week"
timeseries: "timestamp"
measures:
  - label: Total rows
    expression: count(*)
    name: total_rows
    format_preset: humanize
    description: Total number of records present
  - label: Avg Bid Price
    expression: avg(bid_price)
    name: avg_bid_price
    format_preset: currency_usd
dimensions:
  - name: publisher
    label: Publisher
    column: publisher
    description: ""
  - name: domain
    label: Domain Name
    column: domain
    description: ""
        `;

    await updateCodeEditor(page, docWithCompleteMeasure);
    await gotoNavEntry(page, AD_BIDS_EXPLORE_PATH);
    await expect(page.getByRole("button", { name: "Preview" })).toBeEnabled();

    // Preview
    await page.getByRole("button", { name: "Preview" }).click();

    // Check Avg Bid Price
    await expect(page.getByText("Avg Bid Price $3.01")).toBeVisible();

    // Change the leaderboard metric
    await page
      .getByRole("button", { name: "Select a measure to filter by" })
      .click();
    await page.getByRole("option", { name: "Avg Bid Price" }).click();

    // Check domain and sample value in leaderboard
    await expect(page.getByText("Domain Name")).toBeVisible();
    await expect(page.getByText("facebook.com $3.13")).toBeVisible();

    // Open the Publisher details table
    await page
      .getByLabel("Open dimension details")
      .filter({ hasText: "Publisher" })
      .click();

    // Check that table is shown
    await expect(
      page.getByRole("table", { name: "Dimension table" }),
    ).toBeVisible();

    // Check for a table value
    // Can do better table checking in the future when table is refactored to use proper row setup
    // For now, just check the dimensions
    await expect(
      page.locator("button").filter({ hasText: /^Microsoft$/ }),
    ).toBeVisible();

    // TODO when table is better formatted
    //    Change sort direction
    //    Check new sort direction worked in first table row
    //    Change sort column and check

    // Click a table value to filter
    await page
      .locator("button")
      .filter({ hasText: /^Microsoft$/ })
      .click();

    // Check that filter was applied
    await expect(
      page.getByLabel("View filter").getByText("Publisher Microsoft"),
    ).toBeVisible();

    // go back to the leaderboards.
    await page.getByText("All dimensions").click();
    // clear all filters
    await page.getByText("Clear filters").click();

    // run through TDD table view
    await page.getByText("Total rows 100k").click();

    await expect(
      page.getByText("No comparison dimension selected"),
    ).toBeVisible();

    await page
      .getByRole("button", { name: "Select a comparison dimension" })
      .first()
      .click();
    await page.getByRole("menuitem", { name: "Domain Name" }).click();

    await page.waitForTimeout(500);

    await page.getByRole("cell", { name: "google.com", exact: true }).click();
    await page
      .getByRole("cell", { name: "instagram.com", exact: true })
      .click();
    await page.getByRole("cell", { name: "msn.com", exact: true }).click();

    await expect(page.getByText("Total rows 43,749")).toBeVisible();

    await page.getByRole("cell", { name: "Total rows" }).locator("div").click();

    await page.getByRole("button", { name: "Total rows", exact: true }).click();
    await page.getByRole("menuitem", { name: "Avg Bid Price" }).click();

    await expect(page.getByText(" Avg Bid Price $3.02")).toBeVisible();

    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 4 Weeks" }).click();
    });

    await page
      .getByRole("button", { name: "Select a comparison dimension" })
      .first()
      .click();
    await page
      .getByRole("menuitem", { name: "No comparison dimension" })
      .click();
    await page.getByRole("button", { name: "Comparing" }).click();

    await expect(page.getByText("~0%")).toBeVisible();

    // await page.getByRole("button", { name: "Edit Explore" }).click();

    // go back to the dashboard

    // TODO
    //    Check that details table can exclude
    //    Add search criteria
    //    Check that table got search
    //    Clear search
    //    Change the sort column to total rows
    //    Go back to leaderboard
    //    Check that selected metric is total rows
    //    Change the leaderboard metric to avg bid price
    //    await page.getByRole("button", { name: "Total records" }).click();
  });
});
