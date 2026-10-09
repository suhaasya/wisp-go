/**
 * Playwright: 100 hash navigations, heap should stay within ±1 MB.
 * WISP_SHELL_URL=http://127.0.0.1:5173/#/connections npm --prefix tools/tabbench run shell-nav
 */
import { chromium } from "playwright";

const base = process.env.WISP_SHELL_URL ?? "http://127.0.0.1:5173/#/connections";

async function main() {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  await page.goto(base, { waitUntil: "networkidle" });

  const before = await page.evaluate(() =>
    performance.memory ? performance.memory.usedJSHeapSize : 0,
  );
  if (!before) {
    console.log("shell-nav: skipped (performance.memory unavailable)");
    await browser.close();
    return;
  }

  for (let i = 0; i < 100; i++) {
    await page.evaluate(() => {
      location.hash = i % 2 === 0 ? "#/connections" : "#/connections/new";
    });
    await page.waitForTimeout(10);
  }

  const after = await page.evaluate(() => performance.memory.usedJSHeapSize);
  const deltaMB = Math.abs(after - before) / (1024 * 1024);
  if (deltaMB > 1) {
    throw new Error(`heap delta ${deltaMB.toFixed(2)} MB exceeds 1 MB`);
  }
  console.log(`shell-nav: heap delta ${deltaMB.toFixed(2)} MB`);
  await browser.close();
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
