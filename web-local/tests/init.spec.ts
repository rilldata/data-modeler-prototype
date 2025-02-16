import { test } from "./utils/test";
import { expect } from "playwright/test";
import { EXAMPLES } from "@rilldata/web-common/features/welcome/constants";

test.describe("Example project initialization", () => {
  EXAMPLES.forEach((example) => {
    test.describe(`Example project: ${example.title}`, () => {
      test.use({ projectInit: { empty: true } });
      test("should initialize new project", async ({ page }) => {
        await page.getByRole("link", { name: example.title }).click();

        await page.waitForURL(`**/files/dashboards/${example.firstFile}`);

        await expect(
          page.getByRole("heading", { name: example.firstFile }),
        ).toBeVisible();
      });
    });
  });

  test.describe("Empty project", () => {
    test.use({ projectInit: { empty: true } });
    test("should initialize new project", async ({ page }) => {
      await page.getByRole("link", { name: "Empty Project" }).click();

      await expect(page.getByText("Getting started")).toBeVisible();

      await page.getByRole("link", { name: "rill.yaml" }).click();

      await expect(
        page.getByRole("heading", { name: "rill.yaml" }),
      ).toBeVisible();
    });
  });
});
