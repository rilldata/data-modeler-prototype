import { expect } from "@playwright/test";
import { clickMenuButton } from "web-local/tests/utils/commonHelpers";
import { interactWithTimeRangeMenu } from "web-local/tests/utils/metricsViewHelpers";
import { ResourceWatcher } from "web-local/tests/utils/ResourceWatcher";
import { validateTableContents } from "web-local/tests/utils/tableHelpers";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../utils/test";

const pivotDashboard = `kind: metrics_view
display_name: Ad Bids
table: AdBids
timeseries: timestamp
dimensions:
  - display_name: Publisher
    column: publisher
    description: ""
  - display_name: Domain
    column: domain
    description: ""
measures:
  - name: total_records
    display_name: Total records
    expression: COUNT(*)
    description: ""
    format_preset: humanize
    valid_percent_of_total: true
  - name: bid_price
    display_name: Sum of Bid Price
    expression: SUM(bid_price)
    description: ""
    format_preset: humanize
    valid_percent_of_total: true
available_time_zones:
  - America/Los_Angeles
  - America/Chicago
  - America/New_York
  - Europe/London
  - Europe/Paris
  - Asia/Jerusalem
  - Europe/Moscow
  - Asia/Kolkata
  - Asia/Shanghai
  - Asia/Tokyo
  - Australia/Sydney
available_time_ranges:
  - PT6H
  - PT24H
  - P7D
  - P14D
  - P4W
  - P3M
  - P12M
  - rill-TD
  - rill-WTD
  - rill-MTD
  - rill-QTD
  - rill-YTD
  - rill-PDC
  - rill-PWC
  - rill-PMC
  - rill-PQC
  - rill-PYC
`;

const expectedOneMeasureOneDim = [
  [], // dummy row added for virtualization
  ["Total", "100.0k"],
  ["null", "32.9k"],
  ["Facebook", "19.3k"],
  ["Google", "18.8k"],
  ["Yahoo", "18.6k"],
  ["Microsoft", "10.4k"],
  [], // dummy row added for virtualization
];

const expectedTwoMeasureRowDimColDim = [
  [],
  [
    "Total",
    "100.0k",
    "300.6k",
    "15.6k",
    "48.8k",
    "15.5k",
    "48.5k",
    "15.1k",
    "46.8k",
    "14.9k",
    "45.7k",
    "13.1k",
    "37.1k",
    "12.9k",
    "37.0k",
    "12.9k",
    "36.7k",
  ],
  [
    "null",
    "32.9k",
    "98.8k",
    "5.1k",
    "15.9k",
    "5.1k",
    "16.0k",
    "5.0k",
    "15.5k",
    "4.9k",
    "15.1k",
    "4.3k",
    "12.1k",
    "4.3k",
    "12.1k",
    "4.2k",
    "12.1k",
  ],
  [
    "Facebook",
    "19.3k",
    "57.8k",
    "10.5k",
    "32.9k",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "8.8k",
    "25.0k",
    "-",
    "-",
    "-",
    "-",
  ],
  [
    "Google",
    "18.8k",
    "56.0k",
    "-",
    "-",
    "-",
    "-",
    "10.1k",
    "31.3k",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "8.6k",
    "24.7k",
  ],
  [
    "Yahoo",
    "18.6k",
    "55.5k",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "10.0k",
    "30.6k",
    "-",
    "-",
    "8.6k",
    "24.9k",
    "-",
    "-",
  ],
  [
    "Microsoft",
    "10.4k",
    "32.5k",
    "-",
    "-",
    "10.4k",
    "32.5k",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
  ],
  [],
];

const expectExpandedTable = [
  [],
  [
    "Total",
    "100.0k",
    "15.6k",
    "15.5k",
    "15.1k",
    "14.9k",
    "13.1k",
    "12.9k",
    "12.9k",
  ],
  ["Jan 2022", "34.9k", "2.5k", "2.5k", "2.5k", "2.5k", "8.4k", "8.3k", "8.2k"],
  ["null", "11.4k", "804", "802", "835", "827", "2.8k", "2.7k", "2.7k"],
  ["Facebook", "7.4k", "1.7k", "-", "-", "-", "5.7k", "-", "-"],
  ["Yahoo", "7.2k", "-", "-", "-", "1.7k", "-", "5.6k", "-"],
  ["Google", "7.2k", "-", "-", "1.7k", "-", "-", "-", "5.5k"],
  ["Microsoft", "1.7k", "-", "1.7k", "-", "-", "-", "-", "-"],
  [
    "Feb 2022",
    "31.7k",
    "2.4k",
    "2.3k",
    "10.2k",
    "10.0k",
    "2.3k",
    "2.3k",
    "2.2k",
  ],
  [
    "Mar 2022",
    "33.4k",
    "10.8k",
    "10.7k",
    "2.4k",
    "2.3k",
    "2.3k",
    "2.3k",
    "2.4k",
  ],
  [],
];

const expectedOneMeasureColDim = [
  [],
  ["15.6k", "15.5k", "15.1k", "14.9k", "13.1k", "12.9k", "12.9k"],
  [],
];

const expectedTimeComparison = [
  [],
  [
    "Total",
    "26.7k",
    "-4.7k",
    "-15%",
    "7.6k",
    "-167",
    "-2%",
    "7.8k",
    "-67",
    "-1%",
    "7.8k",
    "-153",
    "-2%",
    "3.4k",
    "-4.3k",
    "-56%",
  ],
  [
    "null",
    "8.9k",
    "-1.4k",
    "-14%",
    "2.5k",
    "-138",
    "-5%",
    "2.6k",
    "18",
    "1%",
    "2.6k",
    "13",
    "1%",
    "1.1k",
    "-1.3k",
    "-54%",
  ],
  [
    "Facebook",
    "7.0k",
    "2.8k",
    "67%",
    "2.0k",
    "1.3k",
    "175%",
    "2.1k",
    "1.3k",
    "168%",
    "2.0k",
    "1.2k",
    "159%",
    "889",
    "-1.0k",
    "-54%",
  ],
  [
    "Microsoft",
    "5.7k",
    "3.1k",
    "116%",
    "1.7k",
    "1.3k",
    "364%",
    "1.7k",
    "1.3k",
    "359%",
    "1.7k",
    "1.3k",
    "330%",
    "724",
    "-815",
    "-53%",
  ],
  [
    "Google",
    "2.6k",
    "-4.6k",
    "-65%",
    "716",
    "-1.3k",
    "-65%",
    "741",
    "-1.3k",
    "-64%",
    "733",
    "-1.4k",
    "-66%",
    "365",
    "-582",
    "-62%",
  ],
  [
    "Yahoo",
    "2.5k",
    "-4.5k",
    "-65%",
    "721",
    "-1.3k",
    "-64%",
    "699",
    "-1.4k",
    "-66%",
    "759",
    "-1.3k",
    "-63%",
    "324",
    "-585",
    "-64%",
  ],
  [],
];

const expectSortedDeltaCol = [
  [],
  [
    "Total",
    "1.1k",
    "-5",
    "~0%",
    "375",
    "11",
    "3%",
    "342",
    "-19",
    "-5%",
    "89",
    "12",
    "16%",
    "85",
    "3",
    "4%",
    "84",
    "9",
    "12%",
    "80",
    "4",
    "5%",
    "67",
    "-25",
    "-27%",
  ],
  [
    "null",
    "383",
    "2",
    "1%",
    "138",
    "12",
    "10%",
    "119",
    "1",
    "1%",
    "26",
    "-3",
    "-10%",
    "25",
    "0",
    "0%",
    "24",
    "3",
    "14%",
    "24",
    "2",
    "9%",
    "27",
    "-13",
    "-33%",
  ],
  [
    "Microsoft",
    "237",
    "-1",
    "~0%",
    "237",
    "-1",
    "~0%",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
  ],
  [
    "Facebook",
    "283",
    "-14",
    "-5%",
    "-",
    "-",
    "-",
    "223",
    "-20",
    "-8%",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "60",
    "6",
    "11%",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
  ],
  [
    "Yahoo",
    "103",
    "3",
    "3%",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "63",
    "15",
    "31%",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "40",
    "-12",
    "-23%",
  ],
  [
    "Google",
    "116",
    "5",
    "5%",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "-",
    "60",
    "3",
    "5%",
    "-",
    "-",
    "-",
    "56",
    "2",
    "4%",
    "-",
    "-",
    "-",
  ],
  [],
];

test.describe("pivot run through", () => {
  test.use({
    projectInit: { name: "AdBids" },
  });

  test("pivot run through", async ({ page }) => {
    const watcher = new ResourceWatcher(page);

    await page.getByLabel("/metrics").click();
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/metrics/AdBids_metrics.yaml");
    await page.getByLabel("code").click();

    // update the code editor with the new spec
    await watcher.updateAndWaitForDashboard(pivotDashboard);
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    const previewButton = page.getByRole("button", { name: "Preview" });
    await previewButton.click();

    const pivotButton = page.getByRole("link", {
      name: "Pivot",
      exact: true,
    });

    await pivotButton.click();

    const rowZone = page.locator(".dnd-zone.horizontal").nth(0);
    const columnZone = page.locator(".dnd-zone.horizontal").nth(1);

    // measures buttons
    const totalRecords = page.getByRole("button", { name: "Total records" });

    // dimensions buttons
    const publisher = page.getByRole("button", { name: "Publisher" });
    const domain = page.getByRole("button", { name: "Domain" });

    // single measure
    await totalRecords.dragTo(columnZone);
    await expect(
      page.locator("td").filter({ hasText: "100.0k" }),
    ).toBeVisible();

    // one measure and one dimension
    await publisher.dragTo(rowZone);
    await expect(page.locator(".status.running")).toHaveCount(0);
    await validateTableContents(page, "table", expectedOneMeasureOneDim);

    // add second measure using menu and add column dimension
    await domain.dragTo(columnZone);
    const addColumnField = page
      .getByRole("button", { name: "add-field" })
      .nth(1);
    await addColumnField.click();
    await clickMenuButton(page, "Sum of Bid Price");
    await expect(page.locator(".status.running")).toHaveCount(0);
    await validateTableContents(page, "table", expectedTwoMeasureRowDimColDim);

    // Remove the row dimension and second measure
    await page.getByRole("button", { name: "Remove" }).nth(3).click();
    await page.getByRole("button", { name: "Remove" }).nth(0).click();
    await expect(page.locator(".status.running")).toHaveCount(0);
    await validateTableContents(page, "table", expectedOneMeasureColDim);

    const timeMonth = page.getByRole("button", { name: "month" });
    await timeMonth.dragTo(rowZone);

    const addRowField = page.getByRole("button", { name: "add-field" }).nth(0);
    await addRowField.click();
    await clickMenuButton(page, "Publisher");

    const expandButton = page
      .locator("td")
      .filter({ hasText: "Jan" })
      .getByRole("button");
    await expandButton.click();
    await expect(page.locator(".status.running")).toHaveCount(0);
    await validateTableContents(page, "table", expectExpandedTable);

    // Remove the time dimension and column dimension and measure
    await page.getByRole("button", { name: "Remove" }).nth(3).click();
    await page.getByRole("button", { name: "Remove" }).nth(2).click();
    await page.getByRole("button", { name: "Remove" }).nth(0).click();

    // Change the time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 4 Weeks" }).click();
    });

    // add measure and time week to column
    const timeWeek = page.getByRole("button", { name: "week" });
    await totalRecords.dragTo(columnZone);
    await timeWeek.dragTo(columnZone);

    // enable time comparison
    await page.getByRole("button", { name: "Comparing" }).click();
    await expect(page.locator(".status.running")).toHaveCount(0);
    await validateTableContents(page, "table", expectedTimeComparison);

    await addColumnField.click();
    await clickMenuButton(page, "Domain");

    // sort the delta column for the first dimension value column
    const firstColumnDeltaButton = page.locator(
      "tr:nth-child(3) > th:nth-child(6) > .header-cell",
    );
    await firstColumnDeltaButton.click();
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 24 Hours" }).click();
    });
    await expect(page.locator(".status.running")).toHaveCount(0);
    await validateTableContents(page, "table", expectSortedDeltaCol);
  });
});
