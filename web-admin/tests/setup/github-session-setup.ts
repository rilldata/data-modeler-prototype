import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";
import { updateEnvVariable } from "../utils/dotenv";
import { execAsync } from "web-common/tests/utils/spawn";
import { test as setup } from "./base";
import { GITHUB_STORAGE_STATE } from "./constants";

const githubEnvVarName = "RILL_DEVTOOL_E2E_GITHUB_STORAGE_STATE_JSON";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const envFilePath = path.resolve(__dirname, "../../../.env");

setup(
  "should authenticate to GitHub and save the session",
  async ({ page }) => {
    if (
      !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL ||
      !process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD
    ) {
      throw new Error(
        "Missing environment variables required for GitHub authentication",
      );
    }

    // Log-in to GitHub
    await page.goto("https://github.com/login");
    await page.getByLabel("Username or email address").click();
    await page
      .getByLabel("Username or email address")
      .fill(process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL);
    await page.getByLabel("Password").click();
    await page
      .getByLabel("Password")
      .fill(process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_PASSWORD);
    await page.getByRole("button", { name: "Sign in", exact: true }).click();

    // Pause for 2FA if required (using Playwright's `debug` mode), then wait for the GitHub home page to load
    await page.pause();
    await page.waitForURL("https://github.com");

    // Save the browser's current session (storage state with cookies and local storage)
    await page.context().storageState({ path: GITHUB_STORAGE_STATE });
  },
);

setup("should write the GitHub session state to `.env`", () => {
  // Read the session state from file and compact it into a one-line JSON string
  const storageStateJson = fs.readFileSync(GITHUB_STORAGE_STATE, "utf8");
  const compactJson = JSON.stringify(JSON.parse(storageStateJson));

  // Update the variable in the `.env` file
  updateEnvVariable(envFilePath, githubEnvVarName, compactJson);
});

// Note: This test is illustrative and intentionally skipped.
// Overwriting the shared `cloud-e2e.env` file in GCS should be done manually and with caution.
setup.skip("should upload the .env file to GCS", async () => {
  await execAsync("gsutil cp .env gs://rill-devtool/dotenv/cloud-e2e.env");
});
