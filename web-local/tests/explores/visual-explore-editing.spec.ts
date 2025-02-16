import { expect } from "@playwright/test";
import { createExploreFromModel } from "web-local/tests/utils/exploreHelpers";
import { createAdBidsModel } from "../utils/dataSpecifcHelpers";
import { test } from "../utils/test";

test.describe("visual explore editing", () => {
  test("visual explore editor runthrough", async ({ page }) => {
    await createAdBidsModel(page);
    await createExploreFromModel(page);

    await page.waitForURL(
      "**/files/dashboards/AdBids_model_metrics_explore.yaml",
    );

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
      ' 4612345678910111213141516171819202122232425262728293031323334353637383940414243444546# Explore YAML# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboardstype: exploremetrics_view: AdBids_model_metricsdisplay_name: \"Adbids Model explore dashboard\"defaults:  time_range: P14Ddimensions:  - publisher  - domainmeasures:  - total_records  - bid_price_sumtime_zones:  - America/Los_Angeles  - America/Chicago  - America/New_York  - Europe/London  - Europe/Paristime_ranges:  - PT6H  - PT24H  - P7D  - P14D  - P4W  - P12M  - rill-TD  - rill-WTD  - rill-MTD  - rill-QTD  - rill-YTD  - rill-PDC  - rill-PWC  - rill-PMC  - rill-PQC  - rill-PYCtheme:  colors:    primary: hsl(180, 100%, 50%)    secondary: lightgreen',
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
      ' 30123456789101112131415161718192021222324252627282930# Explore YAML# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboardstype: exploredisplay_name: "Adbids Model dashboard"metrics_view: AdBids_model_metricsdimensions:  expr: "*"measures:  expr: "*"time_ranges:  - PT6H  - PT24H  - P7D  - P14D  - P4W  - P12M  - rill-TD  - rill-WTD  - rill-MTD  - rill-QTD  - rill-YTD  - rill-PDC  - rill-PWC  - rill-PMC  - rill-PQC  - rill-PYC',
    );
  });
});
