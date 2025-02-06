import { test, expect } from "@playwright/test";
import { test as RillTest } from "../utils/test";
import { addFolderWithCheck, addFileWithCheck } from "../utils/sourceHelpers";

/// Managed ClickHouse (return to this once managed CH is stable)

test.describe("Connecting to Managed ClickHouse.", () => {
  RillTest("Create ClickHouse Connection...", async ({ page }) => {
    /// Using Managed ClickHouse , create a metrics view then explore dashboard from a model. include checks on each step.
    addFolderWithCheck(page, "untitled_folder");
    await page.locator('button[id="/untitled_folder-nav-entry"]').hover();
    await page.getByLabel("untitled_folder actions menu trigger").click();
    await page.locator('div[role="menuitem"]:has-text("Rename...")').click();

    const inputField = page.locator("input#folder-name");
    await inputField.fill("connectors");

    await page.locator('button[form="rename-asset-form"]').click();

    // create blank file
    const navContainer = page.locator("#nav-\\/");
    await addFileWithCheck(page, "untitled_file");
    await navContainer
      .locator('li[aria-label="/untitled_file Nav Entry"]')
      .dragTo(page.locator('button[id="/connectors-nav-entry"]'));

    await page
      .locator('li[aria-label="/connectors/untitled_file Nav Entry"]')
      .hover();
    await page
      .getByLabel("/connectors/untitled_file actions menu trigger")
      .click();
    await page.locator('div[role="menuitem"]:has-text("Rename...")').click();

    const inputField1 = page.locator("input#file-name");
    await inputField1.fill("clickhouse.yaml");
    await page.locator('button[form="rename-asset-form"]').click();

    const childTextbox = await page.locator(
      'div[role="textbox"] div.cm-content',
    );
    await childTextbox.click(); // Ensure it's focused for typing

    const lines = [
      "type: connector\n",
      "driver: clickhouse\n",
      "managed: true\n",
    ];

    // Type each line with a newline after
    for (const line of lines) {
      await childTextbox.type(line); // Type the line
      await childTextbox.press("Enter"); // Press Enter for a new line
    }

    await page.locator('button:has-text("Save")').click();

    const section = page.locator("section.connector-section");
    const clickhouseEntry = section.locator('li[aria-label="clickhouse"]');
    await expect(clickhouseEntry).toHaveCount(1);

    await page.waitForSelector('span:has-text("No tables found")');

    // Now that Managed ClickHouse has been set up, enable featureflag for CH modeling and start workflow.

    await page.locator('li[aria-label="/rill.yaml Nav Entry"]').click();
    const projectTextbox = await page.locator(
      'div[role="textbox"] div.cm-content',
    );

    await projectTextbox.click();

    const projectlines = [
      "\n",
      "olap_connector: clickhouse\n",
      "",
      "features:\n",
      "  - clickhouseModeling\n",
    ];

    // Type each line with a newline after
    for (const line of projectlines) {
      await projectTextbox.type(line); // Type the line
      await projectTextbox.press("Enter"); // Press Enter for a new line
    }

    await page.locator('button:has-text("Save")').click();

    const ClickHouseReady = await page.locator(
      'li[aria-label="clickhouse"] button:has-text("No tables found")',
    );

    // loop and refresh page until ClickHouse is Ready
    const maxRetries = 10;
    let retries = 0;
    await page.reload();
    await page.waitForTimeout(2000);
    while (!(await ClickHouseReady.isVisible())) {
      if (retries >= maxRetries) {
        throw new Error(
          "playwrightButton did not become visible after maximum retries.",
        );
      }

      console.log(`Reloading page... Attempt ${retries + 1}`);
      await page.reload();
      await page.waitForTimeout(5000); // Wait for 2 seconds before checking again
      retries++;
    }

    // TODO: Create NORMAL model, metrics view and dashboard
    await page.getByLabel("Add Asset").click();
    await page.getByLabel("Add Model").click();

    await page.waitForSelector('li[aria-label="/models/model.sql Nav Entry"]');
    await page.locator('li[aria-label="/models/model.sql Nav Entry"]').hover();

    await page.getByLabel("/models/model.sql actions menu trigger").click();
    await page.locator('div[role="menuitem"]:has-text("Rename...")').click();

    const fileNameField = page.locator("input#file-name");
    await fileNameField.fill("AdBids_csv.yaml"); //must be the same name as output https://github.com/rilldata/rill/issues/6374
    await page.locator('button[form="rename-asset-form"]').click();

    // Mimic typing in the child contenteditable div
    const managedSourceTextBox = await page.locator(
      'div[role="textbox"] div.cm-content',
    );
    await managedSourceTextBox.click(); // Ensure it's focused for typing
    await page.waitForTimeout(2000);

    // Clear existing contents
    await managedSourceTextBox.press("Meta+A"); // need to check this
    await managedSourceTextBox.press("Backspace"); // Delete selected text

    const sourceLines = [
      "type: model\n",
      "materialize: true\n",
      "",
      "sql: >\n",
      "  SELECT timestamp, id, bid_price, domain, publisher\n",
      "  FROM gcs('https://storage.googleapis.com/playwright-gcs-qa/AdBids_csv.csv',\n",
      "          'CSV', 'timestamp DateTime, id UInt32, bid_price Double, domain String, publisher String'\n",
      "        )\n",
      "  {{ if dev }} LIMIT 100 {{ end }}\n",
      "output:\n",
      "  table: AdBids_csv\n",
      "  engine: MergeTree\n",
    ];

    // Type each line with a newline after
    for (const line of sourceLines) {
      await managedSourceTextBox.type(line); // Type the line
      await managedSourceTextBox.press("Enter"); // Press Enter for a new line
    }

    await page.waitForTimeout(5000); //allows the model to load

    // Check the UI panel for column graphs

    // Click Create metrics with AI

    // Click Create Explore Dashboard

    // Check that preview loads and has XYZ details

    // DONE

    // TODO: Incremental Modeling check UI buttons, metrics view, explore dashboard
    await page.getByLabel("Add Asset").click();
    await page.getByLabel("Add Model").click();

    await page.locator('li[aria-label="/models/model.sql Nav Entry"]').hover();
    await page.getByLabel("/models/model.sql actions menu trigger").click();
    await page.locator('div[role="menuitem"]:has-text("Rename...")').click();

    const fileNameField1 = page.locator("input#file-name");
    await fileNameField1.fill("AdBids_csv_incremental.yaml"); //must be the same name as output https://github.com/rilldata/rill/issues/6374
    await page.locator('button[form="rename-asset-form"]').click();

    // Mimic typing in the child contenteditable div
    const incrementalTextBox = await page.locator(
      'div[role="textbox"] div.cm-content',
    );
    await incrementalTextBox.click(); // Ensure it's focused for typing
    await page.waitForTimeout(2000);

    // Clear existing contents
    await incrementalTextBox.press("Meta+A"); // need to check this
    await incrementalTextBox.press("Backspace"); // Delete selected text

    const incrementalLines = [
      "type: model\n",
      "materialize: true\n",
      "incremental: true:\n",
      "",
      "sql: >\n",
      "  SELECT timestamp, id, bid_price, domain, publisher\n",
      "  FROM gcs('https://storage.googleapis.com/playwright-gcs-qa/AdBids_csv.csv',\n",
      "          'CSV', 'timestamp DateTime, id UInt32, bid_price Double, domain String, publisher String'\n",
      "        )\n",
      "  {{ if dev }} LIMIT 100 {{ end }}\n",
      "output:\n",
      "  table: AdBids_csv_incremental\n",
      "  engine: MergeTree\n",
    ];

    // Type each line with a newline after
    for (const line of incrementalLines) {
      await incrementalTextBox.type(line); // Type the line
      await incrementalTextBox.press("Enter"); // Press Enter for a new line
    }

    await page.waitForTimeout(5000); //allows the model to load

    // Check the Refresh buttons for incremental modeling

    // Check the UI panel for column graphs

    // Click Create metrics with AI

    // Click Create Explore Dashboard

    // Check that preview loads and has XYZ details

    // DONE
  });
});
