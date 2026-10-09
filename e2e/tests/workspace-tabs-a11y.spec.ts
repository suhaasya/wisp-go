import { expect, test } from "@playwright/test";

test("workspace bench tabs expose tablist semantics", async ({ page }) => {
  await page.goto("/#/bench/workspace-tabs");
  const tablist = page.getByRole("tablist");
  await expect(tablist).toBeVisible();
  const tabs = page.getByRole("tab");
  await expect(tabs).toHaveCount(20);
  await expect(tabs.first()).toHaveAttribute("aria-selected", "true");
  await tabs.nth(1).click();
  await expect(tabs.nth(1)).toHaveAttribute("aria-selected", "true");
});
