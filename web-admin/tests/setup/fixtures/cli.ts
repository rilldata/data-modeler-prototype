import { type Page } from "@playwright/test";
import {
  execAsync,
  spawnAndMatch,
  type SpawnAndMatchResult,
} from "web-common/tests/utils/spawn";

export async function cliLogin(page: Page) {
  // Run the login command and capture the verification URL
  const { process, match }: SpawnAndMatchResult = await spawnAndMatch(
    "rill",
    ["login"],
    /Open this URL in your browser to confirm the login: (.*)\n/,
  );

  const verificationUrl = match[1];

  // Manually navigate to the verification URL
  await page.goto(verificationUrl);

  // Click the confirm button
  await page.getByRole("button", { name: /confirm/i }).click();

  // Wait for the process to complete
  await new Promise((resolve, reject) => {
    process.on("close", (code) => {
      if (code === 0) resolve(null);
      else reject(new Error(`Process exited with code ${code}`));
    });
  });
}

export async function cliLogout() {
  await execAsync("rill logout");
}
