/**
 * Playwright: 200 connection cards should keep JS heap ≤ 5 MB.
 * WISP_SHELL_URL=http://127.0.0.1:5173 npm --prefix tools/tabbench run connections-screen
 */
import { chromium } from "playwright";

const base =
  process.env.WISP_SHELL_URL ?? "http://127.0.0.1:5173/?benchConnections=200#/connections";

async function main() {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  await page.goto(base, { waitUntil: "networkidle" });
  await page.waitForSelector("[data-connection-id]", { timeout: 15_000 });

  const heap = await page.evaluate(() =>
    performance.memory ? performance.memory.usedJSHeapSize : 0,
  );
  if (!heap) {
    console.log("connections-screen: skipped (performance.memory unavailable)");
    await browser.close();
    return;
  }
  const mb = heap / (1024 * 1024);
  if (mb > 5) {
    throw new Error(`JS heap ${mb.toFixed(2)} MB exceeds 5 MB cap`);
  }
  console.log(`connections-screen: JS heap ${mb.toFixed(2)} MB (200 cards)`);
  await browser.close();
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
