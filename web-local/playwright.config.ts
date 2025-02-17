import { defineConfig, devices } from "@playwright/test";
/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
  testDir: "./tests",
  /* Don't run tests in files in parallel in CI*/
  fullyParallel: !process.env.CI,
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: !!process.env.CI,
  retries: 0,
  /* Opt out of parallel testing in CI */
  workers: process.env.CI ? 1 : 8,
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: "html",
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    /* Base URL to use in actions like `await page.goto('/')`. */
    baseURL: "http://localhost:8083",
    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: "on-first-retry",
    video: "retain-on-failure",
    launchOptions: {
      slowMo: parseInt(process.env.PLAYWRIGHT_SLOW_MO || "0"),
    },
  },
  timeout: 45 * 1000,
  /* Configure projects for major browsers */
  projects: [
    {
      name: "setup",
      testMatch: "setup.ts",
    },
    {
      name: "e2e",
      dependencies: ["setup"],
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
