import { expect, test } from "@playwright/test";

test("shortcut help dialog opens and closes with Escape", async ({ page }) => {
  await page.goto("/#/connections");
  const mod = process.platform === "darwin" ? "Meta" : "Control";
  await page.keyboard.press(`${mod}+/`);
  const dialog = page.getByRole("dialog", { name: "Keyboard shortcuts" });
  await expect(dialog).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(dialog).toBeHidden();
});
