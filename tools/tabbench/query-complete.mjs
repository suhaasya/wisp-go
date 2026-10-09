/**
 * Playwright: SQL editor shows table completion after FROM and column after alias dot.
 * @param {import('playwright').Page} page
 */
export async function benchQueryComplete(page) {
  const base = process.env.WISP_BASE;
  if (!base) {
    throw new Error("WISP_BASE required");
  }
  await page.goto(`${base}/#/bench/query-complete`, { waitUntil: "networkidle" });
  await page.waitForSelector('[data-testid="query-complete-bench"] .cm-editor');

  await page.locator(".cm-content").click();
  await page.keyboard.type("users u");
  await page.keyboard.press("Enter");
  await page.locator(".cm-content").click();
  await page.keyboard.press("End");
  await page.keyboard.type(" WHERE u.");
  await page.waitForSelector(".cm-tooltip-autocomplete", { timeout: 3000 });
  const hasId = await page.locator(".cm-tooltip-autocomplete").getByText("id").count();
  if (hasId < 1) {
    throw new Error("expected column completion for alias u");
  }

  await page.locator(".cm-content").click();
  await page.keyboard.press("Meta+a");
  await page.keyboard.type("SELECT * FROM ");
  const t0 = Date.now();
  await page.keyboard.type("u");
  await page.waitForSelector(".cm-tooltip-autocomplete", { timeout: 3000 });
  const elapsed = Date.now() - t0;
  const hasUsers = await page.locator(".cm-tooltip-autocomplete").getByText("users").count();
  if (hasUsers < 1) {
    throw new Error("expected users table in completion");
  }
  if (elapsed > 30) {
    throw new Error(`completion took ${elapsed.toFixed(1)}ms, want <= 30ms`);
  }
  console.log(`query-complete: completion popup ${elapsed.toFixed(1)}ms after FROM`);
}
