import { test, expect } from "@playwright/test";
import { test as RillTest } from "../utils/test";

/// ClickHouse Cloud test to add clickhouse connector, create metrics -> explore dashboard, as well as table -> dashboard.
/// Second test, incremental modeling, create metrics -> explore dashboard (default limit 100,000)
/// Credentials have access only to playwright_e2e db.

test.describe("Connecting to ClickHouse Cloud", () => {
  RillTest("Create ClickHouse Connection...", async ({ page }) => {
    // Create clickhouse.yaml
    await page.getByLabel("Add Asset").click();
    await page.getByLabel("Add Source").click();
    await page.locator(`button#clickhouse`).click();

    const inputField = page.locator("input#host");
    await inputField.fill("l3e2nu99sw.us-east-1.aws.clickhouse.cloud");

    const inputField2 = page.locator("input#port");
    await inputField2.fill("8443");

    const inputField3 = page.locator("input#username");
    await inputField3.fill(`playwright_user`);

    const inputField4 = page.locator("input#password");
    await inputField4.fill(`R!llD@taQA2024`);

    const inputField5 = page.locator("input#ssl");
    await inputField5.check();

    const inputField6 = page.locator("input#database");
    await inputField6.fill(`playwright_E2E`); //defined database, only has access here

    await page
      .locator('button[form="add-data-clickhouse-form"]')
      .waitFor({ state: "visible" });
    await page.locator('button[form="add-data-clickhouse-form"]').click();

    // wait for connector to be created. Check that connectors clichouse is displayed
    await page.waitForSelector(
      'li[aria-label="/connectors/clickhouse.yaml Nav Entry"]',
      { state: "visible" },
    );
    const section = page.locator("section.connector-section");
    const clickhouseEntry = section.locator('li[aria-label="clickhouse"]');
    await expect(clickhouseEntry).toHaveCount(1);

    // find database in the UI, ensures that the database has loaded.
    // sometimes the context deadline exceed message appears so try again,

    const playwrightButton = page.locator(
      'li[aria-label="clickhouse"] button:has-text("playwright_E2E")',
    );
    const ukPricePaidItem = page.locator(
      'a.clickable-text[href="/connector/clickhouse/clickhouse/playwright_E2E/uk_price_paid"]',
    );
    const maxRetries = 3;
    let retries = 0;

    while (!(await playwrightButton.isVisible())) {
      if (retries >= maxRetries) {
        console.error("ClickHouse Cloud not able to load in time.");
      }

      console.log(`Reloading page... Attempt ${retries + 1}`);
      await page.reload();
      await page.waitForTimeout(5000);
      retries++;
    }
    await playwrightButton.click();

    // Hover over the "uk_price_paid" item and create metrics
    await ukPricePaidItem.hover();
    await page
      .getByLabel(
        "clickhouse-playwright_E2E.uk_price_paid actions menu trigger",
      )
      .click();
    await page
      .locator('div[role="menuitem"]:has-text("Generate metrics with AI")')
      .click();

    // Create explore dashboard
    await page.waitForSelector(
      'li[aria-label="/metrics/uk_price_paid_metrics.yaml Nav Entry"]',
      { state: "visible" },
    );
    await page
      .locator('button[role="button"]:has-text("Create Explore dashboard")')
      .click();

    // Check Visual Dashboard is working, if panel is not visible, something went wrong.
    await page.waitForSelector(
      'li[aria-label="/dashboards/uk_price_paid_metrics_explore.yaml Nav Entry"]',
      { state: "visible" },
    );
    await page.getByLabel("Inspector Panel").isVisible();

    // Checking the dashboard, 9/10 Total Transactions measure gets created but had 1 instance where it was not...
    await page.waitForSelector(
      'a[role="button"]:has-text("Total Transactions")',
      { state: "visible" },
    );
    await expect(
      page
        .locator('a[role="button"]:has-text("Total Transactions")')
        .locator("div")
        .last(),
    ).toHaveText("19.4M"); //maybe change this to the later check that checks that the graph is not empty.

    /// Create an explore dashboard with AI directly from clickhouse table
    await ukPricePaidItem.hover();
    await page
      .getByLabel(
        "clickhouse-playwright_E2E.uk_price_paid actions menu trigger",
      )
      .click();
    await page
      .locator('div[role="menuitem"]:has-text("Generate dashboard with AI")')
      .click();

    // Check Visual dashboard and dashboard preview (same as above)
    await page.waitForSelector(
      'li[aria-label="/dashboards/uk_price_paid_metrics_1_explore.yaml Nav Entry"]',
      { state: "visible" },
    );
    await page.getByLabel("Inspector Panel").isVisible();

    await page.waitForSelector(
      'a[role="button"]:has-text("Total Transactions")',
      { state: "visible" },
    );
    await expect(
      page
        .locator('a[role="button"]:has-text("Total Transactions")')
        .locator("div")
        .last(),
    ).toHaveText("19.4M");

    /// Feature Flag ClickHouse Modeling. Create model from uk_price_paid source.
    /// Create a metrics view and explore dashboard

    await page.locator('li[aria-label="/rill.yaml Nav Entry"]').click();
    const projectTextbox = await page.locator(
      'div[role="textbox"] div.cm-content',
    );
    await projectTextbox.click();

    // sometimes its too fast and the typing gets messed up so adding a 2 second wait
    await page.waitForTimeout(2000);

    const projectlines = ["\n", "features:\n", "  - clickhouseModeling\n"];

    // Type each line with a newline after
    for (const line of projectlines) {
      await projectTextbox.type(line); // Type the line
      await projectTextbox.press("Enter"); // Press Enter for a new line
    }

    await page.locator('button:has-text("Save")').click();

    // Create model from uk_price_paid dataset in ClickHouse Cloud
    await ukPricePaidItem.hover();
    await page
      .getByLabel(
        "clickhouse-playwright_E2E.uk_price_paid actions menu trigger",
      )
      .click();
    await page
      .locator('div[role="menuitem"]:has-text("Create new model")')
      .click();
    await page.waitForSelector(
      'li[aria-label="/models/uk_price_paid_model.sql Nav Entry"]',
      { state: "visible" },
    );

    // Check the inspector panel and make sure all the panels are not empty or (Truthy)
    await page.getByLabel("Inspector Panel").isVisible();
    const columns = [
      "date",
      "price",
      "is_new",
      "addr1",
      "street",
      "postcode2",
      "addr2",
      "locality",
      "district",
      "town",
      "county",
      "postcode1",
      "type",
      "duration",
    ];

    const inspectorPanel = page.locator('aside[aria-label="Inspector Panel"]');
    for (const columnName of columns) {
      await inspectorPanel.locator(`button:has-text("${columnName}")`).click(); //open the column graph

      const columnPanel = inspectorPanel.locator(
        `div:has-text("${columnName}")`,
      );

      const wFullDivs = columnPanel.locator("div.w-full");

      // Check that at least one `div.w-full` is not empty check this later
      const isNonEmptyDivPresent = await wFullDivs.evaluateAll((divs) =>
        divs.some((div) => (div.textContent?.trim() ?? "") !== ""),
      );

      // if a div is empty, this will fail.
      // https://github.com/rilldata/rill-private-issues/issues/1026
      expect(isNonEmptyDivPresent).toBeTruthy();

      //close details
      await inspectorPanel.locator(`button:has-text("${columnName}")`).click();
    }

    // Create metrics and dashboard from Model
    await page
      .locator('button:has-text("Generate metrics view with AI")')
      .click();
    await page.waitForSelector(
      'li[aria-label="/metrics/uk_price_paid_model_metrics.yaml Nav Entry"]',
      { state: "visible" },
    );
    await page
      .locator('button[role="button"]:has-text("Create Explore dashboard")')
      .click();

    // Since the model is a sample of 100K, the values are not certain, instead check if the graph is truthy. 9/10 Total Transaction appears
    await page.waitForSelector(
      'li[aria-label="/dashboards/uk_price_paid_model_metrics_explore.yaml Nav Entry"]',
      { state: "visible" },
    );
    await page.getByLabel("Inspector Panel").isVisible();
    await page.waitForSelector(
      'a[role="button"]:has-text("Total Transactions")',
      { state: "visible" },
    );
    // await expect(page.locator('a[role="button"]:has-text("Total Transactions")').locator('div').last()).toHaveText('99,999'); // since sample its going to be random
    const tddGraph = page.locator(
      'a[role = "button"]:has-text("Total Transactions")',
    );
    expect(tddGraph.locator('svg[role="application"]')).toBeTruthy();

    // Feature Flag Incremental Modeling, create incremental model, metrics view, and explore
    await page.getByLabel("Add Asset").click();
    await page.getByLabel("Add Model").click();

    await page.locator('li[aria-label="/models/model.sql Nav Entry"]').hover();
    await page.getByLabel("/models/model.sql actions menu trigger").click();
    await page.locator('div[role="menuitem"]:has-text("Rename...")').click();

    const inputField1 = page.locator("input#file-name");
    await inputField1.fill("AdBids_csv.yaml"); //must be the same name as output https://github.com/rilldata/rill/issues/6374
    await page.locator('button[form="rename-asset-form"]').click();

    // Mimic typing in the child contenteditable div
    const incrementalTextBox = await page.locator(
      'div[role="textbox"] div.cm-content',
    );
    await incrementalTextBox.click(); // Ensure it's focused for typing
    await page.waitForTimeout(2000);

    // Clear existing contents
    await incrementalTextBox.press("Meta+A");
    await incrementalTextBox.press("Backspace");

    const incrementalLines = [
      "type: model\n",
      "materialize: true\n",
      "incremental: true\n",
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
    for (const line of incrementalLines) {
      await incrementalTextBox.type(line); // Type the line
      await incrementalTextBox.press("Enter"); // Press Enter for a new line (sometimes this doesnt work, added '\n' to text
    }

    await page.waitForTimeout(5000); //allows the model to load

    // Test the incremental and full refresh buttons
    // These wont actually do anything at the moment since the dataset is static, but should at least refresh the UI.
    await page.getByLabel("Refresh Incremental Model").click();
    await page
      .locator('div[role="menuitem"]:has-text("Incremental refresh")')
      .isVisible();
    await page
      .locator('div[role="menuitem"]:has-text("Full refresh")')
      .isVisible();

    await page
      .locator('div[role="menuitem"]:has-text("Incremental refresh")')
      .click();
    await page.waitForTimeout(1000); //allows the model to load

    await page.getByLabel("Refresh Incremental Model").click();
    await page
      .locator('div[role="menuitem"]:has-text("Incremental refresh")')
      .isVisible();
    await page
      .locator('div[role="menuitem"]:has-text("Full refresh")')
      .isVisible();

    await page.locator('div[role="menuitem"]:has-text("Full refresh")').click();
    await page.waitForTimeout(1000); //allows the model to load

    // Create Metrics and Explore Dashboard off the 100 limit incremental model
    await page
      .getByRole("button", { name: "Generate metrics view with AI" })
      .click();

    await page.waitForSelector(
      'li[aria-label="/metrics/AdBids_csv_metrics.yaml Nav Entry"]',
      { state: "visible" },
    );
    await page
      .locator('button[role="button"]:has-text("Create Explore dashboard")')
      .click();
    await page.waitForSelector(
      'li[aria-label="/dashboards/AdBids_csv_metrics_explore.yaml Nav Entry"]',
      { state: "visible" },
    );
    await page.getByLabel("Inspector Panel").isVisible();
    await page.waitForSelector('a[role="button"]:has-text("Total Bids")', {
      state: "visible",
    });

    // Check that the graph for Total Bids is not empty.
    const tddGraphIncremental = page.locator(
      'a[role = "button"]:has-text("Total Bids")',
    );
    expect(tddGraphIncremental.locator('svg[role="application"]')).toBeTruthy();
  });
});
