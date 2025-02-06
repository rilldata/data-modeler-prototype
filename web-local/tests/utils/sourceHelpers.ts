import { expect } from "@playwright/test";
import {
  extractFileName,
  splitFolderAndFileName,
} from "@rilldata/web-common/features/entity-management/file-path-utils";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
import path from "node:path";
import type { Page } from "playwright";
import { fileURLToPath } from "url";
import { clickModalButton, waitForProfiling } from "./commonHelpers";
import { waitForFileNavEntry } from "./waitHelpers";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
export const TestDataPath = path.join(__dirname, "../data");

/**
 * Used to upload local file as a source
 * @param page
 * @param file File name relative to test data folder
 * @param isDuplicate
 * @param keepBoth
 */
export async function uploadFile(
  page: Page,
  file: string,
  isDuplicate = false,
  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#local_file").click();
  // wait for file chooser while clicking on upload button
  const [fileChooser] = await Promise.all([
    page.waitForEvent("filechooser"),
    page.getByText("Upload a CSV, JSON or Parquet file").click(),
  ]);
  // input the `file` after joining with `testDataPath`
  const fileUploadPromise = fileChooser.setFiles([
    path.join(TestDataPath, file),
  ]);
  const fileRespWaitPromise = page.waitForResponse(/files\/entry/);

  // TODO: infer duplicate
  if (isDuplicate) {
    await fileUploadPromise;
    let duplicatePromise;
    if (keepBoth) {
      // click on `Keep Both` if `isDuplicate`=true and `keepBoth`=true
      duplicatePromise = clickModalButton(page, "Keep Both");
    } else {
      // else click on `Replace Existing Source`
      duplicatePromise = clickModalButton(page, "Replace Existing Source");
    }
    await Promise.all([fileRespWaitPromise, duplicatePromise]);
  } else {
    await Promise.all([fileRespWaitPromise, fileUploadPromise]);
    // if not duplicate wait and make sure `Duplicate source name` modal is not open
    await asyncWait(100);
    await expect(page.getByText("Duplicate source name")).toBeHidden();
  }
}

export async function createSource(page: Page, file: string, filePath: string) {
  await uploadFile(page, file);
  await Promise.all([
    page.getByText("View this source").click(),
    waitForFileNavEntry(page, filePath, true),
  ]);
}

export async function waitForSource(
  page: Page,
  filePath: string,
  columns: Array<string>,
) {
  const [, fileName] = splitFolderAndFileName(filePath);
  const name = extractFileName(fileName);

  await Promise.all([
    page.getByText("View this source").click(),
    waitForFileNavEntry(page, filePath, true),
    waitForProfiling(page, name, columns),
  ]);
}

//additions by Roy

export async function addFileWithCheck(
  page: Page,
  file: string, //used to check URL for creation
) {
  const filePath = `/files/${file}`;
  // add asset button
  await page.getByLabel("Add Asset").click();
  // Hover over "More" to reveal the dropdown menu
  await page.locator("text=More").hover();
  // Click the "Blank file" button
  await page.locator("text=Blank file").click();
  //check the URL for created file.
  await page.waitForURL(`**${filePath}`);
}

export async function addFolderWithCheck(
  page: Page,
  file: string, //used to check URL for creation
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // Hover over "More" to reveal the dropdown menu
  await page.locator("text=More").hover();
  // Click the "Blank file" button
  await page.getByRole("menuitem", { name: "Folder" }).click();
  // check folder exists
  await page.waitForSelector(`button[id="/${file}-nav-entry"]`, {
    state: "visible",
  });
}

export async function waitForTable(
  page: Page,
  filePath: string,
  columns: Array<string>,
) {
  const [, fileName] = splitFolderAndFileName(filePath);
  const name = extractFileName(fileName);

  await Promise.all([
    page.getByText("View this source").click(),
    waitForFileNavEntry(page, filePath, false),
    waitForProfiling(page, name, columns), //this one is imported bc failed sources will still navigate
  ]);
}

//S3, GCS, ABS
export async function cloud(
  page: Page,
  file: string,
  comp: string,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator(`button#${comp}`).click();

  // input the needed details for mysql
  // Locate the  SQL text box and modify
  // console.log("opened the UI");

  const inputField = page.locator("input#path");
  // Modify the text
  // Get the current working directory
  await inputField.fill(
    comp === "gcs"
      ? `gs://playwright-${comp}-qa/${file}`
      : `${comp}://playwright-${comp}-qa/${file}`,
  );

  // Locate the DSN text box and modify
  const inputField2 = page.locator("input#name");
  const fileName = file.split(".")[0];

  // Modify the text
  await inputField2.fill(fileName);

  // add source menu item
  await page
    .locator(`button[form="add-data-${comp}-form"]`)
    .waitFor({ state: "visible" });
  await page.locator(`button[form="add-data-${comp}-form"]`).click();
  //  await page.waitForTimeout(5000); // Waits for 5000 milliseconds (5 seconds) -- refreshing page bc idk whats happnening forever load
  //  await page.reload();

  // TODO: infer duplicate
  //currently we dont check anything when duplicate source is added, or sometimes it overwrites.
}

// Additions MySQL
export async function mySQLDataset(
  page: Page,
  file: string,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#mysql").click();

  // input the needed details for mysql
  // Locate the  SQL text box and modify
  const inputField = page.locator("input#sql");
  // Modify the text
  await inputField.fill(`SELECT * FROM ${file};`);

  // Locate the DSN text box and modify
  const inputField2 = page.locator("input#dsn");

  // Modify the text
  await inputField2.fill("root:rootpass@tcp(127.0.0.1:3306)/default");

  // Locate the source_name text box and modify
  const inputField3 = page.locator("input#name");

  // Modify the text
  await inputField3.fill(`${file}`);

  // add source menu item
  await page
    .locator('button[form="add-data-mysql-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-mysql-form"]').click();
  await page.waitForTimeout(2000); // Waits for 5000 milliseconds (5 seconds) -- refreshing page bc idk whats happnening forever load
  // await page.reload();

  // TODO: infer duplicate
  //currently we dont check anything when duplicate source is added, or sometimes it overwrites.
}

// Additions SQLITE
export async function sqlLiteDataset(
  page: Page,
  file: string,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#sqlite").click();

  // input the needed details for mysql
  // Locate the  SQL text box and modify
  // console.log("opened the UI");

  const currentDir = process.cwd(); // Returns the directory where the script is executed
  const dbPath = path.join(currentDir, "mydb.sqlite");

  const inputField = page.locator("input#db");
  // Modify the text
  // Get the current working directory

  await inputField.fill(dbPath);

  // Locate the DSN text box and modify
  const inputField2 = page.locator("input#table");

  // Modify the text
  await inputField2.fill(`${file}`);

  // Locate the source_name text box and modify
  const inputField3 = page.locator("input#name");

  // Modify the text
  await inputField3.fill(`${file}`);

  // add source menu item
  await page
    .locator('button[form="add-data-sqlite-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-sqlite-form"]').click();
  await page.waitForTimeout(5000); // Waits for 5000 milliseconds (5 seconds) -- refreshing page bc idk whats happnening forever load
  //  await page.reload();

  // TODO: infer duplicate
  //currently we dont check anything when duplicate source is added, or sometimes it overwrites.
}

// Additions postgre
export async function pgDataset(
  page: Page,
  file: string,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#postgres").click();

  // input the needed details for mysql
  // Locate the  SQL text box and modify
  // console.log("opened the UI");

  const inputField = page.locator("input#sql");
  // Modify the text
  // Get the current working directory
  await inputField.fill(`SELECT * FROM ${file};`);

  // Locate the DSN text box and modify
  const inputField2 = page.locator("input#database_url");

  // Modify the text
  await inputField2.fill(
    "postgresql://postgres:postgrespass@localhost:5432/default",
  );

  // Locate the source_name text box and modify
  const inputField3 = page.locator("input#name");

  // Modify the text
  await inputField3.fill(`${file}`);

  // add source menu item
  await page
    .locator('button[form="add-data-postgres-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-postgres-form"]').click();
  await page.waitForTimeout(5000); // Waits for 5000 milliseconds (5 seconds) -- refreshing page bc idk whats happnening forever load
  //  await page.reload();

  // TODO: infer duplicate
  //currently we dont check anything when duplicate source is added, or sometimes it overwrites.
}

export async function DuckDB(
  page: Page,
  file: string,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#duckdb").click();

  // input the needed details for mysql
  // Locate the  SQL text box and modify
  const currentDir = process.cwd(); // Returns the directory where the script is executed
  const dbPath = path.resolve(currentDir, "tests/data/playwright.db"); //need to fix this when ready to deploy

  const inputField = page.locator("input#db");

  // Modify the text
  // Get the current working directory
  await inputField.fill(dbPath);

  // Locate the DSN text box and modify
  const inputField2 = page.locator("input#sql");

  // Modify the text
  await inputField2.fill(`SELECT * FROM ${file};`);

  // Locate the source_name text box and modify
  const inputField3 = page.locator("input#name");

  // Modify the text
  await inputField3.fill(`${file}`);

  // add source menu item
  await page
    .locator('button[form="add-data-duckdb-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-duckdb-form"]').click();
  //  await page.waitForTimeout(5000); // Waits for 5000 milliseconds (5 seconds) -- refreshing page bc idk whats happnening forever load
  //  await page.reload();

  // TODO: infer duplicate
  //currently we dont check anything when duplicate source is added, or sometimes it overwrites.
}

export async function MotherDuck(
  page: Page,
  file: string,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#motherduck").click();

  // input the needed details for mysql
  // Locate the  SQL text box and modify
  const currentDir = process.cwd(); // Returns the directory where the script is executed
  const dbPath = path.resolve(currentDir, "tests/data/playwright.db");

  const inputField = page.locator("input#dsn");

  // Modify the text
  // Get the current working directory
  await inputField.fill(dbPath);

  // Locate the DSN text box and modify
  const inputField2 = page.locator("input#sql");

  // Modify the text
  await inputField2.fill(`SELECT * FROM ${file};`);

  // Locate the source_name text box and modify
  const inputField3 = page.locator("input#token");
  // Modify the text
  await inputField3.fill("");

  // Locate the source_name text box and modify
  const inputField4 = page.locator("input#name");
  // Modify the text
  await inputField4.fill(`${file}`);

  // add source menu item
  await page
    .locator('button[form="add-data-motherduck-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-motherduck-form"]').click();
  //  await page.waitForTimeout(5000); // Waits for 5000 milliseconds (5 seconds) -- refreshing page bc idk whats happnening forever load
  //  await page.reload();

  // TODO: infer duplicate
  //currently we dont check anything when duplicate source is added, or sometimes it overwrites.
}

export async function BigQuery(
  page: Page,
  file: string,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#bigquery").click();

  const inputField = page.locator("input#sql");
  // Modify the text
  await inputField.fill("");

  const inputField2 = page.locator("input#project_id");
  // Modify the text
  await inputField2.fill("");

  const inputField3 = page.locator("input#name");
  // Modify the text
  await inputField3.fill(`${file}`);
  // add source menu item
  await page
    .locator('button[form="add-data-bigquery-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-bigquery-form"]').click();
}

export async function Athena(
  page: Page,
  file: string,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#athena").click();

  const inputField = page.locator("input#sql");
  // Modify the text
  await inputField.fill("");
  const inputField2 = page.locator("input#name");
  // Modify the text
  await inputField2.fill(`${file}`);

  // add source menu item
  await page
    .locator('button[form="add-data-athena-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-athena-form"]').click();
}

export async function Redshift(
  page: Page,
  file: string,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#redshift").click();

  const inputField = page.locator("input#sql");
  // Modify the text
  await inputField.fill("");

  const inputField2 = page.locator("input#output-location");
  // Modify the text
  await inputField2.fill("");

  const inputField3 = page.locator("input#database");
  // Modify the text
  await inputField3.fill("");

  const inputField4 = page.locator("input#role_arn");
  // Modify the text
  await inputField4.fill("");

  const inputField5 = page.locator("input#name");
  // Modify the text
  await inputField5.fill(`${file}`);
  // add source menu item
  await page
    .locator('button[form="add-data-redshift-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-redshift-form"]').click();
}

export async function Snowflake(
  page: Page,
  file: string,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#snowflake").click();

  const inputField = page.locator("input#sql");
  // Modify the text
  await inputField.fill("");

  const inputField2 = page.locator("input#name");
  // Modify the text
  await inputField2.fill(`${file}`);

  //optional
  const inputField3 = page.locator("input#dsn");
  // Modify the text
  await inputField3.fill("");

  // add source menu item
  await page
    .locator('button[form="add-data-snowflake-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-snowflake-form"]').click();
}

export async function Salesforce(
  page: Page,
  file: string,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#salesforce").click();

  const inputField = page.locator("input#soql");
  // Modify the text
  await inputField.fill("");

  const inputField2 = page.locator("input#sobject");
  // Modify the text
  await inputField2.fill("");

  const inputField3 = page.locator("input#queryAll"); //checkbox
  // Modify the text
  await inputField3.fill("");

  const inputField4 = page.locator("input#username");
  // Modify the text
  await inputField4.fill("");

  const inputField5 = page.locator("input#password");
  // Modify the text
  await inputField5.fill("");

  const inputField6 = page.locator("input#name");
  // Modify the text
  await inputField6.fill(`${file}`);

  // add source menu item
  await page
    .locator('button[form="add-data-salesforce-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-salesforce-form"]').click();
}

export async function HTTPS(
  page: Page,
  file: string,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#https").click();

  const inputField = page.locator("input#path");
  // Modify the text
  await inputField.fill("");
  const inputField2 = page.locator("input#name");
  // Modify the text
  await inputField2.fill(`${file}`);
  // add source menu item
  await page
    .locator('button[form="add-data-https-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-https-form"]').click();
}

/// OLAP ENGINES (also need to test making a manual file, and DSN)
export async function ClickHouse(
  page: Page,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#clickhouse").click();

  const inputField = page.locator("input#host");
  // Modify the text
  await inputField.fill("");

  const inputField2 = page.locator("input#port");
  // Modify the text
  await inputField2.fill("");

  const inputField3 = page.locator("input#username");
  // Modify the text
  await inputField3.fill("");

  const inputField4 = page.locator("input#password");
  // Modify the text
  await inputField4.fill("");

  const inputField5 = page.locator("input#ssl");
  // Modify the text
  await inputField5.fill("");

  const inputField6 = page.locator("input#database");
  // Modify the text
  await inputField6.fill("");

  // add source menu item
  await page
    .locator('button[form="add-data-clickhouse-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-clickhouse-form"]').click();
}

export async function Druid(
  page: Page,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#druid").click();

  const inputField = page.locator("input#host");
  // Modify the text
  await inputField.fill("");

  const inputField2 = page.locator("input#port");
  // Modify the text
  await inputField2.fill("");

  const inputField3 = page.locator("input#username");
  // Modify the text
  await inputField3.fill("");

  const inputField4 = page.locator("input#password");
  // Modify the text
  await inputField4.fill("");

  const inputField5 = page.locator("input#ssl");
  // Modify the text
  await inputField5.fill("");

  // add source menu item
  await page
    .locator('button[form="add-data-druid-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-druid-form"]').click();
}

export async function Pinot(
  page: Page,
  //  isDuplicate = false,
  //  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#pinot").click();

  const inputField = page.locator("input#host");
  // Modify the text
  await inputField.fill("");

  const inputField2 = page.locator("input#port");
  // Modify the text
  await inputField2.fill("");

  const inputField3 = page.locator("input#username");
  // Modify the text
  await inputField3.fill("");

  const inputField4 = page.locator("input#password");
  // Modify the text
  await inputField4.fill("");

  const inputField5 = page.locator("input#ssl");
  // Modify the text
  await inputField5.fill("");

  // add source menu item
  await page
    .locator('button[form="add-data-pinot-form"]')
    .waitFor({ state: "visible" });
  await page.locator('button[form="add-data-pinot-form"]').click();
}
