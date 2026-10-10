import { expect, test } from "@playwright/test";

test("workspace bench tabs expose tablist semantics", async ({ page }) => {
  await page.goto("/#/bench/workspace-tabs");
  const tablist = page.getByRole("tablist");
  await expect(tablist).toBeVisible();
  const tabs = page.getByRole("tab");
  await expect(tabs).toHaveCount(20);
  await expect(tabs.first()).toHaveAttribute("aria-selected", "true");
  await tabs.first().focus();
  await page.keyboard.press("ArrowRight");
  await expect(tabs.nth(1)).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(tabs.nth(1)).toHaveAttribute("aria-selected", "true");
});
