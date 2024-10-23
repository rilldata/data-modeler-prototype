import { expect, type Locator } from "@playwright/test";
import { useDashboardFlowTestSetup } from "./dashboard-flow-test-setup";
import { test } from "../utils/test";

async function assertAAboveB(locA: Locator, locB: Locator) {
  const topA = await locA.boundingBox().then((box) => box?.y);
  const topB = await locB.boundingBox().then((box) => box?.y);

  expect(topA).toBeDefined();
  expect(topB).toBeDefined();

  // Safety: topB is defined
  expect(topA).toBeLessThan(topB as number);
}

test.describe("leaderboard and dimension table sorting", () => {
  useDashboardFlowTestSetup();

  test("leaderboard and dimension table sorting", async ({ page }) => {
    await page.getByRole("button", { name: "Preview" }).click();

    /**
     * LEADERBOARD
     */
    await assertAAboveB(
      page.getByRole("row", { name: "null 32.9k" }),
      page.getByRole("row", { name: "Microsoft 10.4k" }),
    );

    await page
      .getByLabel("publisher leaderboard")
      .getByLabel("Toggle sort leaderboards by value")
      .click();

    await assertAAboveB(
      page.getByRole("row", { name: "Microsoft 10.4k" }),
      page.getByRole("row", { name: "null 32.9k" }),
    );

    const timeRangeMenu = page.getByRole("button", {
      name: "Select time range",
    });

    async function openTimeRangeMenu() {
      await timeRangeMenu.click();
      await page
        .getByRole("menu", { name: "Select time range" })
        .waitFor({ state: "visible" });
    }

    await assertAAboveB(
      page.getByRole("row", { name: "Microsoft 10.4k" }),
      page.getByRole("row", { name: "null 32.9k" }),
    );

    // add time comparison and select Pct change
    await page.getByRole("button", { name: "Comparing" }).click();

    await openTimeRangeMenu();
    await page.getByRole("menuitem", { name: "Last 24 Hours" }).click();

    // need a slight delay for the time range to update
    // and the "Pct change" option to be available
    // in the context column dropdown
    await page.waitForTimeout(1000);

    await page
      .getByLabel("publisher leaderboard")
      .getByLabel("Toggle sort leaderboards by percent change")
      .click();

    // need a slight delay for the rankings to update
    await page.waitForTimeout(1000);

    // Broader selectors using RegEx to account for some Playwright runs triggering the display
    // of the starting value on hover
    await assertAAboveB(
      page.getByRole("row", { name: /^Google/ }),
      page.getByRole("row", { name: /^Facebook/ }),
    );

    await assertAAboveB(
      page.getByRole("row", { name: "news.yahoo.com 89 12 16%" }),
      page.getByRole("row", { name: "sports.yahoo.com 67 -25 -27%" }),
    );

    // Sort by absolute change
    await page
      .getByLabel("publisher leaderboard")
      .getByRole("button", {
        name: "Toggle sort leaderboards by absolute change",
      })
      .click();

    await assertAAboveB(
      page.getByRole("row", { name: "Google 116 5 5%" }),
      page.getByRole("row", { name: "Facebook 283 -14 -5%" }),
    );

    // toggle sort by absolute change
    await page
      .getByLabel("publisher leaderboard")
      .getByRole("button", {
        name: "Toggle sort leaderboards by absolute change",
      })
      .click();

    await assertAAboveB(
      page.getByRole("row", { name: "Facebook 283 -14 -5%" }),
      page.getByRole("row", { name: "Google 116 5 5%" }),
    );

    /**
     * DIMENSION TABLE
     */
    // click publisher leaderboard header to enter dimension table for that dim
    await page.locator("#svelte").getByText("Publisher").click();

    // click publisher column header to sort by publisher
    await page.getByRole("button", { name: "Publisher", exact: true }).click();
    await assertAAboveB(
      page
        .locator("div")
        .filter({ hasText: /^Yahoo$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .locator("div")
        .filter({ hasText: /^Microsoft$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );

    // sort by total records
    await page
      .getByRole("table", { name: "Dimension table" })
      .getByRole("button")
      .filter({ hasText: "Total records" })
      .click();
    await assertAAboveB(
      page
        .getByRole("table", { name: "Dimension table" })
        .locator("div")
        .filter({ hasText: /^383$/ }),
      page
        .locator("div")
        .filter({ hasText: /^283$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );

    // sort by absolute change once to sort by absolute change descending
    await page.locator(".w-full > button:nth-child(2)").click();
    await assertAAboveB(
      page
        .getByRole("table", { name: "Dimension table" })
        .locator("div")
        .filter({ hasText: /^2$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .getByRole("table", { name: "Dimension table" })
        .locator("div")
        .filter({ hasText: /^-14$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );

    // sort by absolute change TWICE to sort by absolute change ascending
    await page.locator(".w-full > button:nth-child(2)").click();
    await assertAAboveB(
      page
        .getByRole("table", { name: "Dimension table" })
        .locator("div")
        .filter({ hasText: /^-14$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .getByRole("table", { name: "Dimension table" })
        .locator("div")
        .filter({ hasText: /^2$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );

    // await page.waitForTimeout(60000);

    // sort by pct change ONCE to sort by pct change descending
    // await page.locator(".w-full > button:nth-child(3)").first().click();
    await page.getByRole("button", { name: "%" }).first().click();
    await assertAAboveB(
      page
        .locator("div")
        .filter({ hasText: /^5%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .locator("div")
        .filter({ hasText: /^3%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );

    // sort by pct change TWICE to sort by pct change ascending
    await page.getByRole("button", { name: "%" }).first().click();
    await assertAAboveB(
      page
        .locator("div")
        .filter({ hasText: /^3%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
      page
        .locator("div")
        .filter({ hasText: /^5%$/ })
        .getByRole("button", { name: "Filter dimension value" }),
    );
  });
});
