import { test, expect } from "@playwright/test";
import { test as RillTest } from "../utils/test";
import { cloud, waitForTable } from "../utils/sourceHelpers";
import { waitForFileNavEntry } from "../utils/waitHelpers";
import {
  actionUsingMenu,
  checkExistInConnector,
  renameFileUsingMenu,
} from "../utils/commonHelpers";

// GCS source ingestion test
// based on public bucket gs://playwright-gcs-qa/*
// Can add more files as required, currently parquet.gz files are erroring so removed.

test.describe("Check Source UI buttons.", () => {
  RillTest("Reading Source into Rill from GCS", async ({ page }) => {
    console.log("Testing cloud sales data ingestion...");
    await Promise.all([
      waitForTable(page, "/sources/sales.yaml", [
        "sale_date",
        "sale_id",
        "duration_ms",
        "customer_id",
        "sales_amount_usd",
        "products",
        "discounts",
        "region",
        "is_online",
      ]),
      cloud(page, "sales.csv", "gcs"),
    ]);
    console.log("Sales table validated...");

    // Create Model!
    console.log("Creating Create Model Button...");
    await Promise.all([
      waitForFileNavEntry(page, "/models/sales_model.sql", false), //set true?
      page.getByRole("button", { name: "Create model" }).click(),
    ]);

    // CHECK CONNECTORS for MODEL (table name dynamic so wildcard)
    await checkExistInConnector(page, "duckdb", "main_db", "sales_model");

    // CHECKING BUTTONS
    //Close File Explore Sidebar
    await page.locator('span[aria-label="Close sidebar"]').click();
    // Assert that the class changes
    const sidebarClose = page.locator(".sidebar.svelte-5nrsv4");
    await expect(sidebarClose).toHaveClass("sidebar svelte-5nrsv4 hide");

    await page.locator('span[aria-label="Show sidebar"]').click();
    // Assert that the class changes
    const sidebarOpen = page.locator(".sidebar.svelte-5nrsv4");
    await expect(sidebarOpen).toHaveClass("sidebar svelte-5nrsv4");

    // checking the refresh button
    await page.locator('button[aria-label="Refresh Model"]').click(); //#6316, need to find where this gets added
    await expect(
      page.getByText("Building model sales_model").first().isVisible(),
    ).toBeTruthy(); // Test will fail if the text is not visible

    // checking the panels ,
    await page.getByRole("button", { name: "Toggle table visibility" }).click(); // #6308
    const resultsPreviewTable = await page.locator(
      '[aria-label="Results Preview Table"]',
    ); // #6316
    await expect(resultsPreviewTable).toBeHidden();
    await expect(resultsPreviewTable.locator(`text="sale_id"`)).toHaveCount(0);

    await page
      .getByRole("button", { name: "Toggle inspector visibility" })
      .click(); // #6308
    const inspectorPanel = await page.locator('[aria-label="Inspector Panel"]'); // #6316
    await expect(inspectorPanel).toBeHidden();
    await expect(inspectorPanel.locator(`text="rows"`)).toHaveCount(0);

    // Wait for the download and confirm success (CSV, XLSX, Parquet)
    const [downloadCSV] = await Promise.all([
      page.waitForEvent("download"), // Wait for the download event
      page.getByLabel("Export Model Data").click(), // Dropdown
      page.getByRole("menuitem", { name: "Export as CSV" }).click(), // Export
    ]);

    const filePathCSV = await downloadCSV.path();
    if (filePathCSV) {
      console.log(`File successfully downloaded to: ${filePathCSV}`);
    } else {
      console.error("Download failed.");
    }

    const [downloadParquet] = await Promise.all([
      page.waitForEvent("download"), // Wait for the download event
      page.getByLabel("Export Model Data").click(), // Dropdown
      page
        .locator('div[role="menuitem"]:has-text("Export as Parquet")')
        .click(), // Export
    ]);

    const filePathParquet = await downloadParquet.path();
    if (filePathParquet) {
      console.log(`File successfully downloaded to: ${filePathParquet}`);
    } else {
      console.error("Download failed.");
    }

    const [downloadXSLX] = await Promise.all([
      page.waitForEvent("download"), // Wait for the download event
      page.getByLabel("Export Model Data").click(), // Dropdown
      page.locator('div[role="menuitem"]:has-text("Export as XLSX")').click(), // Export
    ]);

    const filePathXLSX = await downloadXSLX.path();
    if (filePathXLSX) {
      console.log(`File successfully downloaded to: ${filePathXLSX}`);
    } else {
      console.error("Download failed.");
    }

    // Select "Generate Metrics with AI",
    await Promise.all([
      waitForFileNavEntry(page, "/metrics/sales_model_metrics.yaml", false), //set true?
      page.getByRole("button", { name: "Generate metrics view" }).click(),
    ]);

    // Return to source and check Go to for both models.
    await page.locator('span:has-text("sales_model.sql")').click();

    await page.getByRole("button", { name: "Go to metrics view" }).click();

    await Promise.all([
      waitForFileNavEntry(page, "/metrics/sales_model_metrics_1.yaml", false), //set true?
      page.getByText("Create metrics view").click(),
    ]);

    await expect(
      page.getByRole("link", { name: "sales_model_metrics.yaml" }),
    ).toBeVisible();
    await expect(
      page.getByRole("link", { name: "sales_model_metrics_1.yaml" }),
    ).toBeVisible();

    // Delete a metrics vie and rename another
    await page.locator('span:has-text("sales_model_metrics.yaml")').hover();
    await actionUsingMenu(page, "/sales_model_metrics.yaml", "Delete");

    await renameFileUsingMenu(
      page,
      "/metrics/sales_model_metrics_1.yaml",
      "random_metrics.yaml",
    );

    // Check the model and metrics are still linked
    await page.locator('span:has-text("sales_model.sql")').click();
    await page.getByRole("button", { name: "Go to metrics view" }).click();
    await page
      .locator('div[role="menuitem"]:has-text("Create metrics view")')
      .waitFor();
    await expect(
      page.getByRole("menuitem", { name: "random_metrics", exact: true }),
    ).toBeVisible();
    await page
      .getByRole("menuitem", { name: "random_metrics", exact: true })
      .click();

    // Can add further testing like renaming files and creating metrics from button to see if number is correct.

    await page.locator('span:has-text("random_metrics.yaml")').hover();
    await actionUsingMenu(page, "/random_metrics.yaml", "Delete");

    // Check the UI has returned to Generate metrics view with AI
    await page.locator('span:has-text("sales_model.sql")').click();
    await expect(
      page.getByRole("button", { name: "Generate metrics view" }),
    ).toBeVisible();
  });
});
