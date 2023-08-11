import { expect, test } from "@playwright/test";
import {
  deleteEntity,
  renameEntityUsingMenu,
  updateCodeEditor,
} from "./utils/commonHelpers";
import {
  waitForAdBids,
  waitForAdImpressions,
} from "./utils/dataSpecifcHelpers";
import { TestEntityType } from "./utils/helpers";
import {
  TestDataPath,
  createOrReplaceSource,
  uploadFile,
} from "./utils/sourceHelpers";
import { startRuntimeForEachTest } from "./utils/startRuntimeForEachTest";
import { entityNotPresent, waitForEntity } from "./utils/waitHelpers";

test.describe("sources", () => {
  startRuntimeForEachTest();

  test("Import sources", async ({ page }) => {
    await page.goto("/");

    await Promise.all([
      waitForAdBids(page, "AdBids"),
      uploadFile(page, "AdBids.csv"),
    ]);

    await Promise.all([
      waitForAdImpressions(page, "AdImpressions"),
      uploadFile(page, "AdImpressions.tsv"),
    ]);

    // upload existing table and keep both
    await Promise.all([
      waitForEntity(page, TestEntityType.Source, "AdBids", false),
      waitForAdBids(page, "AdBids_1"),
      uploadFile(page, "AdBids.csv", true, true),
    ]);

    // upload existing table and replace
    await uploadFile(page, "AdBids.csv", true, false);
    await entityNotPresent(page, "AdBids_2");
  });

  test("Rename and delete sources", async ({ page }) => {
    await page.goto("/");

    await createOrReplaceSource(page, "AdBids.csv", "AdBids");

    // rename
    await renameEntityUsingMenu(page, "AdBids", "AdBids_new");
    await waitForEntity(page, TestEntityType.Source, "AdBids_new", true);
    await entityNotPresent(page, "AdBids");

    // delete
    await deleteEntity(page, "AdBids_new");
    await entityNotPresent(page, "AdBids_new");
    await entityNotPresent(page, "AdBids");
  });

  test("Edit source", async ({ page }) => {
    await page.goto("/");

    // Upload data & create two sources
    await createOrReplaceSource(page, "AdImpressions.tsv", "AdImpressions");
    await createOrReplaceSource(page, "AdBids.csv", "AdBids");

    // Edit source path to a non-existent file
    const nonExistentSource = `type: local_file
path: ${TestDataPath}/non_existent_file.csv`;
    await updateCodeEditor(page, nonExistentSource);
    await page.getByRole("button", { name: "Save and refresh" }).click();

    // Observe error message "file does not exist"
    await expect(page.getByText("file does not exist")).toBeVisible();

    // Edit source path to an existent file
    const adImpressionsSource = `type: local_file
path: ${TestDataPath}/AdImpressions.tsv`;
    await updateCodeEditor(page, adImpressionsSource);
    await page.getByRole("button", { name: "Save and refresh" }).click();

    // Check that the source data is updated
    // (The column "user_id" exists in AdImpressions, but not in AdBids)
    await expect(
      page.getByRole("button").filter({ hasText: "user_id" }).first()
    ).toBeVisible();
  });
});
