import { devices, type PlaywrightTestConfig } from "@playwright/test";
import { ADMIN_STORAGE_STATE } from "./tests/setup/constants";

const config: PlaywrightTestConfig = {
  webServer: {
    command: "npm run build && npm run preview",
    port: 3000,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
  retries: 0,
  reporter: "html",
  use: {
    baseURL: "http://localhost:3000",
    ...devices["Desktop Chrome"],
    trace: "retain-on-failure",
    video: "retain-on-failure",
  },
  testDir: "tests",
  projects: [
    /* ...(process.env.CI
      ? [] // skip in CI
      : [
          {
            // Whenever the GitHub session expires, run this project manually to re-authenticate.
            // This process captures the browser’s current storage state (i.e. cookies and local storage)
            // and updates the `RILL_DEVTOOL_E2E_GITHUB_STORAGE_STATE_JSON` environment variable.
            // Afterwards, manually deploy the updated `.env` file to GCS.
            name: "setup-github-session",
            testMatch: "github-session-setup.ts",
          },
        ]),*/
    {
      name: "setup",
      testMatch: "setup.ts",
      ...(process.env.E2E_NO_TEARDOWN || process.env.CI
        ? undefined
        : { teardown: "teardown" }),
    },
    ...(process.env.CI
      ? [] // skip in CI, since the GitHub Action uses an ephemeral runner
      : [
        {
          name: "teardown",
          testMatch: "teardown.ts",
          use: {
            storageState: ADMIN_STORAGE_STATE,
          },
        },
      ]),
    {
      name: "e2e",
      dependencies: process.env.E2E_NO_SETUP_OR_TEARDOWN ? [] : ["setup"],
      testIgnore: "/setup",
      use: {
        storageState: ADMIN_STORAGE_STATE,
      },
    },
  ],
};

export default config;
