import type { Page } from "@playwright/test";

/** Lazy palette + shortcut help chunks (AppShell) must load before global shortcuts are reliable. */
export async function waitAppChromeReady(page: Page): Promise<void> {
  await page.getByTestId("app-chrome-ready").waitFor({ state: "attached", timeout: 30_000 });
}
