import { expect } from "@playwright/test";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../utils/test";

test.describe("visual explore editing", () => {
  test.use({
    projectInit: { name: "AdBids" },
  });

  test("visual explore editor runthrough", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByLabel("code").click();

    await page.getByRole("button", { name: "Subset" }).first().click();
    await page.getByRole("button", { name: "Subset" }).nth(1).click();

    await page.getByRole("button", { name: "Custom" }).first().click();
    await page.getByRole("button", { name: "Custom" }).nth(1).click();
    await page.getByRole("button", { name: "Custom" }).nth(2).click();

    let text = await page
      .getByRole("textbox", { name: "Code editor" })
      .textContent();

    expect(text).toEqual(
      ' 48123456789101112131415161718192021222324252627282930313233343536373839404142434445464748# Explore YAML# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboardstype: exploretitle: "Adbids dashboard"metrics_view: AdBids_metricsdimensions:  - publisher  - domainmeasures:  - total_records  - bid_price_sumtime_ranges:  - PT6H  - PT24H  - P7D  - P14D  - P4W  - P12M  - rill-TD  - rill-WTD  - rill-MTD  - rill-QTD  - rill-YTD  - rill-PDC  - rill-PWC  - rill-PMC  - rill-PQC  - rill-PYCtime_zones:  - America/Los_Angeles  - America/Chicago  - America/New_York  - Europe/London  - Europe/Paris  - Asia/Jerusalem  - Europe/Moscow  - Asia/Kolkata  - Asia/Shanghai  - Asia/Tokyo  - Australia/Sydneytheme:  colors:    primary: hsl(180, 100%, 50%)    secondary: lightgreen',
    );

    await page.getByRole("button", { name: "Expression" }).first().click();
    await page.getByRole("button", { name: "Expression" }).nth(1).click();

    await page.getByRole("button", { name: "Default" }).first().click();
    await page.getByRole("button", { name: "None" }).first().click();
    await page.getByRole("button", { name: "Presets" }).click();

    text = await page
      .getByRole("textbox", { name: "Code editor" })
      .textContent();

    expect(text).toEqual(
      ' 30123456789101112131415161718192021222324252627282930# Explore YAML# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboardstype: exploretitle: "Adbids dashboard"metrics_view: AdBids_metricsdimensions:  expr: "*"measures:  expr: "*"time_ranges:  - PT6H  - PT24H  - P7D  - P14D  - P4W  - P12M  - rill-TD  - rill-WTD  - rill-MTD  - rill-QTD  - rill-YTD  - rill-PDC  - rill-PWC  - rill-PMC  - rill-PQC  - rill-PYC',
    );
  });
});
