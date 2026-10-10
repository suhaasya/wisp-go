/** @param {import('@playwright/test').Page} page */
async function waitAppChromeReady(page) {
  await page.getByTestId("app-chrome-ready").waitFor({ state: "attached", timeout: 30_000 });
}

/** @param {import('@playwright/test').Page} page @param {string} profileName */
export async function connectFromConnectionsScreen(page, profileName) {
  await page.goto("/#/connections");
  await waitAppChromeReady(page);
  const card = page.getByRole("button", { name: new RegExp(profileName.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")) });
  await card.waitFor({ state: "visible", timeout: 30_000 });
  await card.focus();
  await page.keyboard.press("Enter");

  const trust = page.getByRole("button", { name: "Trust and connect" });
  if (await trust.isVisible().catch(() => false)) {
    await trust.click();
  }

  await page.waitForURL(/#\/w\//, { timeout: 120_000 });
}

/**
 * @param {import('@playwright/test').Page} page
 * @param {object} opts
 * @param {'postgres' | 'mysql' | 'mariadb'} opts.engine
 * @param {string} opts.schema
 * @param {string} opts.tableName
 */
export async function openMatrixTableInSidebar(page, { engine, schema, tableName }) {
  const tree = page.getByRole("tree");
  await tree.waitFor({ state: "visible", timeout: 60_000 });

  const group = page.getByRole("treeitem", { name: /Tables/i });
  if ((await group.getAttribute("aria-expanded")) !== "true") {
    await group.click();
  }

  const row = page.locator(`[data-object-row="${tableName}"]`);
  await row.waitFor({ state: "visible", timeout: 60_000 });
  await row.click();
  await page.getByRole("grid").waitFor({ state: "visible", timeout: 60_000 });
  await page.getByText("matrix-ok").waitFor({ state: "visible", timeout: 60_000 });
}
