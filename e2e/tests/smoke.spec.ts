import { expect, test } from "@playwright/test";

test("bench workspace tabs load", async ({ page }) => {
  await page.goto("/#/bench/workspace-tabs");
  await expect(page.getByTestId("workspace-tabs-bench")).toBeVisible();
  await expect(page.getByRole("tab")).toHaveCount(20);
});

test("bench grid scroll region is a grid", async ({ page }) => {
  await page.goto("/#/bench/grid");
  await expect(page.getByTestId("wisp-grid-scroll")).toBeVisible();
  await expect(page.getByRole("grid")).toBeVisible();
});
