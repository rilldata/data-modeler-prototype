import { test as base, type Page } from "@playwright/test";
import { ADMIN_STORAGE_STATE, VIEWER_STORAGE_STATE } from "./constants";
import { cliLogin, cliLogout } from "./fixtures/cli";
import path from "path";
import { fileURLToPath } from "url";

type MyFixtures = {
  adminPage: Page;
  viewerPage: Page;
  anonPage: Page;
  cli: void;
  embedPage: Page;
};

export const test = base.extend<MyFixtures>({
  // Note: the `e2e` project uses the admin auth file by default, so it's likely that
  // this fixture won't be used often.
  adminPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: ADMIN_STORAGE_STATE,
    });
    const adminPage = await context.newPage();
    await use(adminPage);
    await context.close();
  },

  viewerPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: VIEWER_STORAGE_STATE,
    });
    const viewerPage = await context.newPage();
    await use(viewerPage);
    await context.close();
  },

  anonPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: { cookies: [], origins: [] },
    });
    const anonPage = await context.newPage();
    await use(anonPage);
    await context.close();
  },

  cli: async ({ page }, use) => {
    await cliLogin(page);
    await use();
    await cliLogout();
  },

  embedPage: async ({ browser }, use) => {
    const __dirname = path.dirname(fileURLToPath(import.meta.url));
    const filePath = "file://" + path.resolve(__dirname, "..", "embed.html");

    const context = await browser.newContext();
    const embedPage = await context.newPage();
    await embedPage.goto(filePath);
    await embedPage.waitForTimeout(500);

    await use(embedPage);

    await context.close();
  },
});
