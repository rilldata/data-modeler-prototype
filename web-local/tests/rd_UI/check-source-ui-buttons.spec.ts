import { test, expect } from "@playwright/test";
import { test as RillTest } from "../utils/test";
import { cloud, waitForTable } from "../utils/sourceHelpers";
import {
  actionUsingMenu,
  renameFileUsingMenu,
  checkExistInConnector,
} from "../utils/commonHelpers";

// Testing functionality of all the buttons in the source page.
// Refresh, "Create Model", "Go to "(after creating model), panels

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

    // CHECK CONNECTORS for SOURCE (table name dynamic so wildcard)
    await checkExistInConnector(page, "duckdb", "main_db", "sales");

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
    await Promise.all([
      page.locator('button[aria-label="Refresh"]').click(),
      expect(page.getByText("Ingesting source sales").first()).toBeVisible(), // Test will fail if the text is not visible
    ]);
    // checking the panels ,
    await page.getByRole("button", { name: "Toggle table visibility" }).click(); // #6308
    const resultsPreviewTableClose = await page.locator(
      '[aria-label="Results Preview Table"]',
    ); // #6316
    await expect(resultsPreviewTableClose).toBeHidden();

    await page.getByRole("button", { name: "Toggle table visibility" }).click(); // #6308
    const resultsPreviewTableOpen = await page.locator(
      '[aria-label="Results Preview Table"]',
    ); // #6316
    await expect(resultsPreviewTableOpen).toBeVisible();

    await page
      .getByRole("button", { name: "Toggle inspector visibility" })
      .click(); // #6308
    const inspectorPanelClose = await page.locator(
      '[aria-label="Inspector Panel"]',
    ); // #6316
    await expect(inspectorPanelClose).toBeHidden();

    await page
      .getByRole("button", { name: "Toggle inspector visibility" })
      .click(); // #6308
    const inspectorPanelOpen = await page.locator(
      '[aria-label="Inspector Panel"]',
    ); // #6316
    await expect(inspectorPanelOpen).toBeVisible();

    // Create Model!
    console.log("Creating Create Model Button...");
    await Promise.all([
      //  waitForFileNavEntry(page, "/files/models/sales_model.sql", false), //set true?
      page.getByRole("button", { name: "Create model" }).click(),
    ]);

    // Return to Source, Select Go to. -> create model, check that its suffixed "_1"
    await page.locator('span:has-text("sales.yaml")').click();

    //   waitForFileNavEntry(page, "/models/sales_model_1.sql", false), //set true?
    await page.getByRole("button", { name: "Go to" }).click();
    await page
      .locator('div[role="menuitem"]:has-text("Create model")')
      .waitFor();
    await page.locator('div[role="menuitem"]:has-text("Create model")').click();

    // Return to source and check Go to for both models.
    await page.locator('span:has-text("sales.yaml")').click();
    await page.getByRole("button", { name: "Go to" }).click();
    await page
      .locator('div[role="menuitem"]:has-text("Create model")')
      .waitFor();
    await expect(
      page.locator('a[role="menuitem"][href="/files/models/sales_model.sql"]'),
    ).toBeVisible();
    await expect(
      page.locator(
        'a[role="menuitem"][href="/files/models/sales_model_1.sql"]',
      ),
    ).toBeVisible();

    // Delete and rename models and confirm back to Create model
    await page.locator('span:has-text("sales_model.sql")').hover();
    await actionUsingMenu(page, "/sales_model.sql", "Delete");

    await renameFileUsingMenu(
      page,
      "/models/sales_model_1.sql",
      "random_model.sql",
    );

    // Check the source and model are still linked
    await page.locator('span:has-text("sales.yaml")').click();
    await page.getByRole("button", { name: "Go to" }).click();
    await page
      .locator('div[role="menuitem"]:has-text("Create model")')
      .waitFor();
    await expect(
      page.locator('a[role="menuitem"][href="/files/models/random_model.sql"]'),
    ).toBeVisible();
    await page
      .locator('a[role="menuitem"][href="/files/models/random_model.sql"]')
      .click();

    // Delete new model
    await page.locator('span:has-text("random_model.sql")').hover();
    await actionUsingMenu(page, "/random_model.sql", "Delete");

    // Check the UI has returned to Create Model
    await page.locator('span:has-text("sales.yaml")').click();
    await expect(
      page.getByRole("button", { name: "Create model" }),
    ).toBeVisible();
  });
});
