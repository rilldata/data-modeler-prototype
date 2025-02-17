import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Homepage", () => {
  test("Authenticated user should get redirected to the organization page", async ({
    page,
  }) => {
    await page.goto("/");
    await expect(page.getByRole("link", { name: "e2e" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Projects" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Users" })).toBeVisible();
  });

  test("Unauthenticated user should be redirected to login", async ({
    anonPage,
  }) => {
    await anonPage.goto("/");
    await expect(anonPage.getByText("Log in to Rill")).toBeVisible();
  });
});
