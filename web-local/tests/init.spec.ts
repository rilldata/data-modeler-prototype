import { EXAMPLES } from "@rilldata/web-common/features/welcome/constants";
import { expect } from "playwright/test";
import { test } from "./utils/test";

test.describe("Project initialization", () => {
  EXAMPLES.forEach((example) => {
    test.describe(`Example project: ${example.title}`, () => {
      test.use({ includeRillYaml: false });
      test("should initialize new project", async ({ page }) => {
        await page.getByRole("button", { name: example.title }).click();

        await page.waitForURL(`**/files/dashboards/${example.firstFile}`);

        await expect(
          page.getByRole("heading", { name: example.firstFile }),
        ).toBeVisible();
      });
    });
  });
});
