import { expect, test } from "@playwright/test";
import { waitAppChromeReady } from "./helpers";

test("command palette opens and closes with keyboard", async ({ page }) => {
  await page.goto("/#/connections");
  await waitAppChromeReady(page);
  const mod = process.platform === "darwin" ? "Meta" : "Control";
  await page.keyboard.press(`${mod}+k`);
  const dialog = page.getByRole("dialog", { name: "Command palette" });
  await expect(dialog).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(dialog).toBeHidden();
});
