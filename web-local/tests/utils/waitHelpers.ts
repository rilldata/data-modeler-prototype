import { expect } from "@playwright/test";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
import type { Page } from "playwright";
import { getFileNavEntry } from "./commonHelpers";

export async function waitForFileNavEntry(
  page: Page,
  filePath: string,
  navigated: boolean,
) {
  await page.getByLabel(`${filePath} Nav Entry`).waitFor();

  if (navigated) {
    await page.waitForURL(new RegExp(`/files${filePath}`));
  }
}

export async function gotoNavEntry(page: Page, filePath: string) {
  await page.getByLabel(`${filePath} Nav Entry`).click();
  return waitForFileNavEntry(page, filePath, true);
}

export async function fileNotPresent(page: Page, filePath: string) {
  await asyncWait(100);
  await expect(getFileNavEntry(page, filePath)).toBeHidden();
}
