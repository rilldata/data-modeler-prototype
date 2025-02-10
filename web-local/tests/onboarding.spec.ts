import { expect } from "@playwright/test";
import path from "path";
import { fileURLToPath } from "url";
import { ClickHouseTestContainer } from "./utils/clickhouse";
import { test } from "./utils/test";

test.describe("Onboarding", () => {
  test.use({ includeRillYaml: false });

  test("Example project", async ({ page }) => {
    // Should be redirected to the onboarding page
    await page.waitForURL("**/welcome");
    await expect(page.getByText("Welcome to Rill")).toBeVisible();

    // Click on "Example projects"
    await page.getByRole("button", { name: "Cost monitoring" }).click();

    // Expect to be navigated to the project's first dashboard
    await page.waitForURL("**/files/dashboards/metrics_margin_explore.yaml", {
      timeout: 20_000,
    });

    // Expect to see the `rill.yaml` file in the sidebar
    await expect(page.getByText("rill.yaml")).toBeVisible();

    // Expect to see the dashboard's default time range
    await expect(page.getByText("Last 3 weeks")).toBeVisible({
      timeout: 15_000,
    });
  });

  test.describe("Rill-managed OLAP", () => {
    test("Start with a blank project", async ({ page }) => {
      // Should be redirected to the onboarding page
      await page.waitForURL("**/welcome");
      await expect(page.getByText("Welcome to Rill")).toBeVisible();

      // Click on "Connect your data"
      await page.getByRole("button", { name: "Connect your data" }).click();

      // Click on "Start with a blank project"
      await page
        .getByRole("button", { name: "Or, start with a blank project" })
        .click();

      // Expect to be navigated to the home page
      await page.waitForURL("**/");

      // Expect to see the "Add data" button
      await expect(
        page.getByRole("button", { name: "Add data" }),
      ).toBeVisible();

      // Expect to see the `rill.yaml` file in the file explorer
      await expect(page.getByText("rill.yaml")).toBeVisible();
    });

    test("Local file", async ({ page }) => {
      test.setTimeout(20_000);

      // Should be redirected to the onboarding page
      await page.waitForURL("**/welcome");
      await expect(page.getByText("Welcome to Rill")).toBeVisible();

      // Click on "Connect your data"
      await page.getByRole("button", { name: "Connect your data" }).click();

      // Should get redirected to the select connectors page
      await page.waitForURL("**/welcome/select-connectors");

      // Click on button with aria-label "Local file"
      await page.getByLabel("local_file").click();

      // Click on "Continue"
      await page.getByRole("button", { name: "Continue" }).click();

      // Should get redirected to the add credentials page
      await page.waitForURL("**/welcome/add-credentials");

      // Upload a file (Note: this is copied from `sourceHelpers.ts` and should be cleaned-up.)
      const [fileChooser] = await Promise.all([
        page.waitForEvent("filechooser"),
        page.getByText("Upload a CSV, JSON or Parquet file").click(),
      ]);
      const __filename = fileURLToPath(import.meta.url);
      const __dirname = path.dirname(__filename);
      const adbidsCsvPath = path.join(__dirname, "./data/AdBids.csv");
      const fileUploadPromise = fileChooser.setFiles([adbidsCsvPath]);
      const fileRespWaitPromise = page.waitForResponse(/files\/entry/);
      await Promise.all([fileUploadPromise, fileRespWaitPromise]);

      // Expect to be navigated to the new source page
      await page.waitForURL("**/files/sources/AdBids.yaml");

      // Click "view this source"
      await page.getByRole("button", { name: "View this source" }).click();

      // Expect to see data
      await expect(page.getByLabel("Results Preview Table")).toBeVisible();
      await expect(
        page.getByLabel("Results Preview Table").getByText("timestamp"), // timestamp is the name of the second column
      ).toBeVisible();
      await expect(
        page.getByLabel("Results Preview Table").getByText("4000", {
          exact: true,
        }), // 4000 is the first row's ID
      ).toBeVisible();
      await expect(
        page.getByLabel("Results Preview Table").getByText("Facebook").first(), // Facebook is one of the publishers
      ).toBeVisible();
    });
  });

  test.describe("Self-managed OLAP", () => {
    test.describe("ClickHouse", () => {
      const clickhouse = new ClickHouseTestContainer();

      test.beforeAll(async () => {
        await clickhouse.start();
        await clickhouse.seed();
      });

      test.afterAll(async () => {
        await clickhouse.stop();
      });

      test("should connect to the ClickHouse instance and create a dashboard", async ({
        page,
      }) => {
        page.on("console", (msg) =>
          console.log(`📢 Console: ${msg.type()}: ${msg.text()}`),
        );

        // Should be redirected to the onboarding page
        await page.waitForURL("**/welcome");
        await expect(page.getByText("Welcome to Rill")).toBeVisible();

        // Pick a self-managed ClickHouse OLAP
        await page.getByRole("button", { name: "Connect your data" }).click();
        await page.getByRole("button", { name: "Self-managed" }).click();
        // "ClickHouse" is the default, so no need to pick it explicitly.
        await page.getByRole("button", { name: "Continue" }).click();

        // Add credentials that WILL NOT work
        await page.getByRole("textbox", { name: "Host" }).click();
        await page.getByRole("textbox", { name: "Host" }).fill("localhost");
        await page.getByRole("textbox", { name: "Host" }).press("Tab");
        await page
          .getByRole("textbox", { name: "Port (optional)" })
          .fill("9000");

        // Submit form and expect to see an error connecting to the dummy ClickHouse instance
        await page.getByRole("button", { name: "Connect" }).click();
        await expect(
          page
            .locator("#add-data-form div")
            .filter({ hasText: "dial tcp [::1]:9000: connect" })
            .nth(1),
        ).toBeVisible();

        // Add credentials that WILL work
        await page.getByRole("textbox", { name: "Host" }).click();
        await page
          .getByRole("textbox", { name: "Host" })
          .fill(clickhouse.getHost());
        await page
          .getByRole("textbox", { name: "Port (optional)" })
          .fill(clickhouse.getPort().toString());
        await page.getByRole("button", { name: "Connect" }).click();

        // Expect to advance to the next step
        await page.waitForURL("**/welcome/make-your-first-dashboard");
        await expect(
          page.getByText("Pick a table to power your first dashboard"),
        ).toBeVisible();

        // Select the AdBids table from the ClickHouse explorer
        await expect(page.getByText("clickhouse")).toBeVisible();
        await page.getByText("default").click();
        await page.getByText("ad_bids").click();

        // Create the dashboard
        await page.getByRole("button", { name: "Create dashboard" }).click();

        // Expect to be navigated to the new dashboard page
        await page.waitForURL(
          "**/files/dashboards/ad_bids_metrics_explore.yaml",
        );

        // Assert that we see the "Total records" Big Number
        await expect(
          page.getByRole("button", { name: "Total records" }),
        ).toHaveText("Total records 99,999");
      });
    });
  });
});
