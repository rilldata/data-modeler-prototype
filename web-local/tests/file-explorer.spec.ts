import { expect } from "playwright/test";
import { test } from "./utils/test";

test.describe("File Explorer", () => {
  test.describe("File CRUD Operations", () => {
    test("should create, rename, edit, copy, and delete a file", async ({
      page,
    }) => {
      // Create a new file
      await page.getByLabel("Add Asset").click();
      await page.getByRole("menuitem", { name: "More" }).hover();
      await page.getByRole("menuitem", { name: "Blank file" }).click();
      await expect(
        page.getByRole("link", { name: "untitled_file" }),
      ).toBeVisible();
      await expect(
        page.getByRole("heading", { name: "untitled_file", exact: true }),
      ).toBeVisible();

      // Rename the file
      await page.getByRole("listitem", { name: "/untitled_file" }).hover();
      await page.getByLabel("/untitled_file actions menu").click();
      await page.getByRole("menuitem", { name: "Rename..." }).click();
      await page.getByLabel("File name").click();
      await page.getByLabel("File name").press("Meta+a");
      await page.getByLabel("File name").fill("README.md");
      await page.getByLabel("File name").press("Enter");
      await expect(
        page.getByRole("link", { name: "README.md" }).nth(1),
      ).toBeVisible();

      // Edit the file
      await page.getByLabel("Auto-save").click(); // Turn off auto-save
      await page.getByRole("textbox").nth(1).click();
      await page.keyboard.type("Here's a README.md file for the e2e test!");
      await page.getByRole("button", { name: "Save" }).click();
      // Wait half a second for the changes to be saved
      await page.waitForTimeout(500);
      // Navigate away from the file and back to it to verify the changes
      await page.getByRole("link", { name: "rill.yaml" }).click();
      await page.getByRole("link", { name: "README.md" }).first().click();
      await expect(
        page.getByText("Here's a README.md file for the e2e test!"),
      ).toBeVisible();

      // Duplicate the file
      await page.getByRole("listitem", { name: "/README.md" }).hover();
      await page.getByLabel("/README.md actions menu").click();
      await page.getByRole("menuitem", { name: "Duplicate" }).click();
      await expect(
        page
          .getByLabel("/README (copy).md Nav Entry")
          .getByRole("link", { name: "README (copy).md" }),
      ).toBeVisible();
      await expect(
        page.getByText("Here's a README.md file for the e2e test!"),
      ).toBeVisible();

      // Delete the file
      await page.getByRole("listitem", { name: "/README.md" }).hover();
      await page.getByLabel("/README.md actions menu").click();
      await page.getByRole("menuitem", { name: "Delete" }).click();
      await expect(
        page.getByRole("link", { name: "README.md" }),
      ).not.toBeVisible();
    });
  });

  test.describe("Folder CRUD Operations", () => {
    test("should create, rename, add sub-folder, and delete the folder", async ({
      page,
    }) => {
      // Create a new folder
      await page.getByLabel("Add Asset").click();
      await page.getByRole("menuitem", { name: "More" }).hover();
      await page.getByRole("menuitem", { name: "Folder" }).click();
      await expect(
        page.getByRole("directory", { name: "untitled_folder" }),
      ).toBeVisible();

      // Rename the folder
      await page.getByRole("directory", { name: "untitled_folder" }).hover();
      await page.getByLabel("untitled_folder actions menu").click();
      await page.getByRole("menuitem", { name: "Rename..." }).click();
      await page.getByLabel("Folder name").click();
      await page.getByLabel("Folder name").press("Meta+a");
      await page.getByLabel("Folder name").fill("my-directory");
      await page.getByLabel("Folder name").press("Enter");

      // Page reloads in test environment
      await page.waitForTimeout(2000);

      // Add something to the folder
      await page.getByRole("directory", { name: "my-directory" }).hover();
      await page.getByLabel("my-directory actions menu").click();
      await page.getByRole("menuitem", { name: "New folder" }).hover();
      const [createDirectoryResponse, getFilesResponse] = await Promise.all([
        page.waitForResponse("**/v1/instances/default/files/dir"),
        page.waitForResponse("**/v1/instances/default/files"),
        page.getByRole("menuitem", { name: "New folder" }).click(),
      ]);

      expect(createDirectoryResponse.status()).toBe(200);
      expect(getFilesResponse.status()).toBe(200);
      const resp = await getFilesResponse.json();
      expect(resp.files.length).toBe(4);
      await expect(
        page.getByRole("directory", {
          name: "my-directory/untitled_folder",
        }),
      ).toBeVisible();

      await page.waitForTimeout(2000);
      // Delete the folder
      await page
        .getByRole("button", { name: "my-directory my-directory" })
        .hover();
      await page.getByLabel("my-directory actions menu").click();
      await page.getByRole("menuitem", { name: "Delete" }).click();
      await page.getByRole("button", { name: "Delete" }).click();
    });
  });
});
