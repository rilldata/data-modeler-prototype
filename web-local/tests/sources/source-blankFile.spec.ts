import { test } from "@playwright/test";
import { test as RillTest } from "../utils/test";
import { addFileWithCheck, waitForTable } from "../utils/sourceHelpers";
import {
  renameFileUsingMenu,
  actionUsingMenu,
  checkExistInConnector,
} from "../utils/commonHelpers";

/// Blank File test
/// In this test we create a `untilted_file`, create a second one to ensure `_1` is appended
/// Following that we rename and modify the contents of a file to a source
/// Re-create a file to ensure it makes a `untitiled_file`, then duplicate it.

test.describe("Creating a blank file... and making a source.", () => {
  RillTest("Creating Blank file", async ({ page }) => {
    // create blank file
    await addFileWithCheck(page, "untitled_file");
    // Wait for the file `untitled_file` to be present on the page
    await page.waitForSelector('li[aria-label="/untitled_file Nav Entry"]', {
      state: "visible",
    });

    //create another blank file and expected untitled_file_1
    await addFileWithCheck(page, "untitled_file_1");
    await page.waitForSelector('li[aria-label="/untitled_file_1 Nav Entry"]', {
      state: "visible",
    });

    await renameFileUsingMenu(page, "/untitled_file", "source.yaml");

    await page.waitForSelector('li[aria-label="/source.yaml Nav Entry"]', {
      state: "visible",
    });
    console.log("File renamed successfully to source.yaml!");

    const textBox = page
      .getByLabel("Code editor") // Locate the labeled parent
      .getByRole("textbox"); // Find the inner textbox

    // Wait for the textbox to be visible
    await textBox.waitFor({ state: "visible" });

    // Rewrite the contents of the textbox
    await textBox.fill(`# Testing manual file creation

        type: source

        connector: "duckdb"
        sql: "select * from read_csv('gs://playwright-gcs-qa/AdBids_csv.csv', auto_detect=true, ignore_errors=1, header=true)"`);

    console.log("Successfully Modified Contents. Checking for data.");

    await waitForTable(page, "/source.yaml", [
      "timestamp",
      "id",
      "bid_price",
      "domain",
      "publisher",
    ]);

    // CREATING A NEW BLANK FILE, EXPECT IT TO BE `untitled_file` as we modified the original

    console.log("Creating a new file, expecting `untitled_file`");
    // create new blank file
    await addFileWithCheck(page, "untitled_file");
    await page.waitForSelector('li[aria-label="/untitled_file Nav Entry"]', {
      state: "visible",
    });

    // TEST FOR DUPLICATES and refresh

    // Locate and click the ellipsis menu button for `untitled_file`
    await actionUsingMenu(page, "/source.yaml", "Duplicate");
    await page.getByText("View this source").click();

    // checks that the file exists in the duckdb connector
    await page.waitForSelector(
      'li[aria-label="/source (copy).yaml Nav Entry"]',
      { state: "visible" },
    );
    await checkExistInConnector(page, "duckdb", "main_db", "source (copy)");
  });
});
