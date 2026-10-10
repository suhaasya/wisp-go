import { expect, test } from "@playwright/test";

test("status bar hides Tab heap meter without performance.memory", async ({ page, browserName }) => {
  test.skip(browserName === "chromium", "Chromium exposes performance.memory in automation");
  await page.goto("/#/connections");
  await expect(page.getByText(/^Tab$/)).toHaveCount(0);
});
