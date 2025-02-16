import { expect } from "@playwright/test";
import { useDashboardFlowTestSetup } from "web-local/tests/explores/dashboard-flow-test-setup";
import { clickMenuButton } from "../utils/commonHelpers";
import { test } from "../utils/test";

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

    await expect(page.getByText("Sum of Bid Price 46,239")).toBeVisible();
    await expect(page.getByText("Total records 15,427")).not.toBeVisible();

    await measuresButton.click();
    await clickMenuItem("Total records");
    await clickMenuItem("Sum of Bid Price");
    await expect(measuresButton).toHaveText("1 of 2 Measures");
    await escape();

    await expect(page.getByText("Sum of Bid Price 46,239")).not.toBeVisible();
    await expect(page.getByText("Total records 15,427")).toBeVisible();

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
