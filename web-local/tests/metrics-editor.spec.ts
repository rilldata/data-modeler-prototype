import { expect } from "@playwright/test";
import { useDashboardFlowTestSetup } from "./explores/dashboard-flow-test-setup";
import {
  AD_BIDS_EXPLORE_PATH,
  AD_BIDS_METRICS_PATH,
} from "./utils/dataSpecifcHelpers";
import { gotoNavEntry } from "./utils/waitHelpers";
import { updateCodeEditor, wrapRetryAssertion } from "./utils/commonHelpers";
import { test } from "./utils/test";

test.describe("Metrics editor", () => {
  useDashboardFlowTestSetup();

  test("Metrics editor", async ({ page }) => {
    await gotoNavEntry(page, AD_BIDS_METRICS_PATH);

    await page.getByLabel("code").click();

    await updateCodeEditor(page, "");

    // inspector should point to the documentation
    await expect(page.getByText("For help building dashboards")).toBeVisible();

    // skeleton should result in an empty skeleton YAML file
    await page.getByText("start with a skeleton").click();

    // check to see that the placeholder is gone by looking for the button that was once there
    await wrapRetryAssertion(async () => {
      await expect(page.getByText("start with a skeleton")).toBeHidden();
    });

    await gotoNavEntry(page, AD_BIDS_EXPLORE_PATH);
    // the Preview button should be disabled
    await expect(page.getByRole("button", { name: "Preview" })).toBeDisabled();
    await gotoNavEntry(page, AD_BIDS_METRICS_PATH);

    // the editor should show a validation error
    await expect(
      page.getByText(
        'must set a value for either the "model" field or the "table" field',
      ),
    ).toBeVisible();

    // now let's create a working metrics view
    await updateCodeEditor(page, "");
    await wrapRetryAssertion(async () => {
      await expect(
        page.getByText("metrics configuration from an existing model"),
      ).toBeVisible();
    });

    // select the first menu item.
    await page
      .getByText("metrics configuration from an existing model")
      .click();
    await page.getByRole("menuitem").getByText("AdBids_model").click();
    // Wait for menu to close
    await expect(
      page.getByRole("menuitem", { name: "Last 6 Hours" }),
    ).not.toBeVisible();

    // let's check the inspector.
    await expect(page.getByText("Table columns")).toBeVisible();

    // go to the dashboard and make sure the metrics and dimensions are there.
    await gotoNavEntry(page, AD_BIDS_EXPLORE_PATH);
    await page.getByRole("button", { name: "Preview" }).click();

    // check to see metrics make sense.
    await expect(page.getByText("Total Records 100.0k")).toBeVisible();

    // double-check that leaderboards make sense.
    await expect(
      page.getByRole("row", { name: "google.com 15.1k" }),
    ).toBeVisible();

    // go back to the metrics page.
    await page.getByRole("button", { name: "Edit" }).click();
    await page.getByRole("menuitem", { name: "Metrics View" }).click();

    // Create a metrics view from a table (rather than a model)
    await updateCodeEditor(
      page,
      `version: 1
type: metrics_view
title: "AdBids table"
table: "AdBids"
timeseries: "timestamp"
measures:
  - name: "Total Records"
    expression: count(*)
dimensions:
  - name: publisher
    label: Publisher
    column: publisher
  - name: domain
    label: Domain
    column: domain
  `,
    );

    // Check that the metrics inspector shows the table columns
    await expect(page.getByText("Table columns")).toBeVisible();
  });
});
