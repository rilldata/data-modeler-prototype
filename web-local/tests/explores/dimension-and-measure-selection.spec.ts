import { expect } from "@playwright/test";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { clickMenuButton } from "../utils/commonHelpers";
import { test } from "../utils/test";

test.describe("dimension and measure selectors", () => {
  test.use({
    projectInit: { name: "AdBids" },
  });

  test("dimension and measure selectors flow", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "Preview" }).click();

    const measuresButton = page.getByRole("button", {
      name: "Choose measures to display",
    });
    const dimensionsButton = page.getByRole("button", {
      name: "Choose dimensions to display",
    });

    async function escape() {
      await page.keyboard.press("Escape");
      await page.getByRole("menu").waitFor({ state: "hidden" });
    }

    async function clickMenuItem(itemName: string) {
      await clickMenuButton(page, itemName);
    }

    await measuresButton.click();
    await clickMenuItem("Total records");
    await escape();
    await expect(measuresButton).toHaveText("1 of 2 Measures");

    await expect(page.getByText("Sum of Bid Price 301k")).toBeVisible();
    await expect(page.getByText("Total records 100k")).not.toBeVisible();

    await measuresButton.click();
    await clickMenuItem("Total records");
    await clickMenuItem("Sum of Bid Price");
    await expect(measuresButton).toHaveText("1 of 2 Measures");
    await escape();

    await expect(page.getByText("Sum of Bid Price 301k")).not.toBeVisible();
    await expect(page.getByText("Total records 100k")).toBeVisible();

    await dimensionsButton.click();
    await clickMenuItem("Publisher");
    await expect(dimensionsButton).toHaveText("1 of 2 Dimensions");
    await escape();

    await expect(page.getByText("Publisher")).not.toBeVisible();
    await expect(page.getByText("Domain")).toBeVisible();

    await dimensionsButton.click();
    await clickMenuItem("Publisher");
    await clickMenuItem("Domain");
    await expect(dimensionsButton).toHaveText("1 of 2 Dimensions");
    await escape();

    await expect(page.getByText("Publisher")).toBeVisible();
    await expect(page.getByText("Domain")).not.toBeVisible();
  });
});
