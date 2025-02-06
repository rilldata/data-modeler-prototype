import { test } from "@playwright/test";
import { test as RillTest } from "../utils/test";
import { addFolderWithCheck } from "../utils/sourceHelpers";

/// Blank Folder test
/// In this test we create a `untilted_file`, create a second one to ensure `_1` is appended

test.describe("Creating a Folder... and making a source.", () => {
  RillTest("Creating Folder", async ({ page }) => {
    // create folder file
    await Promise.all([addFolderWithCheck(page, "untitled_folder")]);
    await Promise.all([addFolderWithCheck(page, "untitled_folder_1")]);
    await Promise.all([addFolderWithCheck(page, "untitled_folder_2")]);
    // create folder in subfolder
    await page.locator('span:has-text("untitled_folder_2")').last().hover();
    await page.getByLabel("untitled_folder_2 actions menu trigger").click();
    await page.getByRole("menuitem", { name: "New Folder" }).first().click();

    // check that the folder exists,
    await page.waitForSelector("#nav-\\/untitled_folder_2\\/untitled_folder", {
      timeout: 5000,
    });
    await page
      .locator('[aria-label="/untitled_folder_2/untitled_folder"]')
      .isVisible();

    // create another for proper "_1" append
    await page.locator('span:has-text("untitled_folder_2")').last().hover();
    await page.getByLabel("untitled_folder_2 actions menu trigger").click();
    await page.getByRole("menuitem", { name: "New Folder" }).first().click();

    await page.waitForSelector(
      "#nav-\\/untitled_folder_2\\/untitled_folder_1",
      { timeout: 5000 },
    );
    await page
      .locator('[aria-label="/untitled_folder_2/untitled_folder_1"]')
      .isVisible();
  });
});
