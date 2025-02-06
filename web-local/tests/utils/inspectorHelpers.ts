import { expect } from "@playwright/test";
import type { Page } from "playwright";

// checking source inspector rows and column values.
export async function checkInspectorSource(
  page: Page,
  expectedRows: string, //expected row count
  expectedColumns: string, //expected column_count
  sourceColumns: Array<string>,
) {
  // check row count
  const inspectorRows = await page.getByRole("cell", { name: /rows$/i });
  const fullTextRow = await inspectorRows.textContent();

  // Extract the numeric part from the string
  const numericValueRow = fullTextRow?.match(/\d{1,3}(,\d{3})*/)?.[0];
  expect(numericValueRow).toBe(expectedRows);

  // check column count
  const inspectorCols = await page.getByRole("cell", { name: /columns$/i });
  const fullTextCol = await inspectorCols.textContent();

  // Extract the numeric part from the string
  const numericValueCol = fullTextCol?.match(/\d{1,3}(,\d{3})*/)?.[0];
  expect(numericValueCol).toBe(expectedColumns);

  // checking the  column details,
  await Promise.all([testColumnsAndChartnDiv(page, sourceColumns)]);
}

// function that opens each column and checks if not empty. Need to check with CH source to see if this is actually working. //closes div
async function testColumnsAndChartnDiv(page, expectedColumns) {
  for (const columnName of expectedColumns) {
    // Click the button by its name
    await page.getByRole("button", { name: columnName }).click();

    // Wait for the column div to be visible
    const presentationDiv = await page.waitForSelector(
      'div[role="presentation"]',
      { state: "visible" },
    );
    // Assert that the div is not empty (not falsy) IE: has contents
    expect(presentationDiv).not.toBeFalsy();

    // Close the column div by clicking the button again
    await page.getByRole("button", { name: columnName }).click();
  }
}

// checking model inspector rows and column values.
export async function checkInspectorModel(
  page: Page,
  expectedRows: string, //expected row count
  expectedColumns: string, //expected column_count
  sourceColumns: Array<string>,
) {
  const inspectorWrapper = await page.locator(".inspector-wrapper");

  // Check if the text '100,000 rows' and '16 columns' exists within the container, needed to change as UI changes a bit from source
  await expect(
    inspectorWrapper.locator(`text="${expectedRows} rows"`),
  ).toHaveCount(1);
  await expect(
    inspectorWrapper.locator(`text="${expectedColumns} columns"`),
  ).toHaveCount(1);

  // checking the column details for not empty,
  await Promise.all([testColumnsAndChartnDiv(page, sourceColumns)]);
}
