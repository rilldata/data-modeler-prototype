import { test } from "@playwright/test";
import { test as RillTest } from "../utils/test";
import { cloud, waitForTable } from "../utils/sourceHelpers";
import {
  checkInspectorSource,
  checkInspectorModel,
} from "../utils/inspectorHelpers";
import { createModel } from "../utils/modelHelpers";

// Testing the contents of the Inspector Panel
// Does the correct rows and columns appear, and does each column have a visible graph?
test.describe("Checking the Inspector Panel for Source and Model. Check if values are correct as well as if the UI populates graph.", () => {
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
    console.log("Sales table validated.");

    await checkInspectorSource(page, "100,000", "9", [
      "sale_date",
      "sale_id",
      "duration_ms",
      "customer_id",
      "sales_amount_usd",
      "products",
      "discounts",
      "region",
      "is_online",
    ]);
    console.log("Testing cloud customer data ingestion...");
    await Promise.all([
      waitForTable(page, "/sources/customer_data.yaml", [
        "customer_id",
        "name",
        "email",
        "signup_date",
        "preferences",
        "total_spent_usd",
        "loyalty_tier",
        "is_active",
      ]),
      cloud(page, "customer_data.csv", "gcs"),
    ]);
    console.log("Customer data table validated.");
    await checkInspectorSource(page, "10,000", "8", [
      "signup_date",
      "customer_id",
      "name",
      "email",
      "preferences",
      "total_spent_usd",
      "loyalty_tier",
      "is_active",
    ]),
      console.log("Creating model to join sources.");
    await createModel(page, "joined_model.sql");
    // wait for textbox to appear for model
    await page.waitForSelector('div[role="textbox"]');

    await page.evaluate(() => {
      // Ensure the parent textbox is focused for typing
      const parentTextbox = document.querySelector(
        'div[role="textbox"]',
      ) as HTMLElement;
      if (parentTextbox) {
        parentTextbox.focus();
      } else {
        console.error("Parent textbox not found!");
      }
    });

    // Mimic typing in the child contenteditable div
    const childTextbox = await page.locator(
      'div[role="textbox"] div.cm-content',
    );
    await childTextbox.click(); // Ensure it's focused for typing

    // Clear existing contents
    await childTextbox.press("Meta+A"); // need to check this
    await childTextbox.press("Backspace"); // Delete selected text

    const lines = [
      "-- Model SQL\n",
      "-- Reference documentation: https://docs.rilldata.com/reference/project-files/models\n",
      "SELECT a.*,\n",
      "    b.* exclude customer_id\n",
      "FROM sales AS a\n",
      "LEFT JOIN customer_data AS b\n",
      "ON a.customer_id = b.customer_id\n",
    ];

    // Type all lines at once
    for (const line of lines) {
      await childTextbox.type(line);
      await childTextbox.press("Enter"); // Press Enter for a new line
    }

    console.log("Content typed successfully.");
    await checkInspectorModel(page, "100,000", "16", [
      "sale_date",
      "sale_id",
      "duration_ms",
      "customer_id",
      "sales_amount_usd",
      "products",
      "discounts",
      "region",
      "is_online",
      "signup_date",
      "customer_id",
      "name",
      "email",
      "preferences",
      "total_spent_usd",
      "loyalty_tier",
      "is_active",
    ]);
  });
});
