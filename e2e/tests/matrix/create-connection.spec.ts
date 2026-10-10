import { expect, test } from "@playwright/test";

test("new connection form exposes labelled general fields", async ({ page }) => {
  await page.goto("/#/connections/new");
  await page.getByTestId("app-chrome-ready").waitFor({ state: "attached" });
  await expect(page.getByLabel("Name")).toBeVisible();
  await expect(page.getByLabel("Host")).toBeVisible();
  await expect(page.getByLabel("Port")).toBeVisible();
});
