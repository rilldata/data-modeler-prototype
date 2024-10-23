import { expect } from "@playwright/test";
import { useDashboardFlowTestSetup } from "./dashboard-flow-test-setup";
import { test } from "../utils/test";
import { clickMenuButton } from "../utils/commonHelpers";

test.describe("dimension and measure selectors", () => {
  // dashboard test setup
  useDashboardFlowTestSetup();

  test("dimension and measure selectors flow", async ({ page }) => {
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

    await expect(page.getByText("Sum of Bid Price 300.6k")).toBeVisible();
    await expect(page.getByText("Total records 100.0k")).not.toBeVisible();

    await measuresButton.click();
    await clickMenuItem("Total records");
    await clickMenuItem("Sum of Bid Price");
    await expect(measuresButton).toHaveText("1 of 2 Measures");
    await escape();

    await expect(page.getByText("Sum of Bid Price 300.6k")).not.toBeVisible();
    await expect(page.getByText("Total records 100.0k")).toBeVisible();

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
