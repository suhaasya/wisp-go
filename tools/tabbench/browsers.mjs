/**
 * Cross-browser smoke for bench routes (no CDP metrics on Firefox/WebKit).
 * WISP_BASE=http://127.0.0.1:8787 WISP_LAUNCH_URL=... node tools/tabbench/browsers.mjs
 */
import { chromium, firefox, webkit } from "playwright";

const base = process.env.WISP_BASE;
const launchURL = process.env.WISP_LAUNCH_URL;

const engines = [
  { name: "chromium", type: chromium },
  { name: "firefox", type: firefox },
  { name: "webkit", type: webkit },
];

async function runEngine({ name, type }) {
  if (!base || !launchURL) {
    console.log(`${name}: skipped (set WISP_BASE and WISP_LAUNCH_URL)`);
    return;
  }
  const browser = await type.launch({ headless: true });
  const page = await browser.newPage();
  await page.goto(launchURL, { waitUntil: "networkidle" });
  await page.goto(`${base}/#/bench/workspace-tabs`, { waitUntil: "networkidle" });
  await page.waitForSelector('[data-testid="workspace-tabs-bench"]');
  const tabs = await page.locator('[role="tab"]').count();
  if (tabs !== 20) {
    throw new Error(`${name}: expected 20 tabs, got ${tabs}`);
  }
  const heap = await page.evaluate(() =>
    performance.memory ? performance.memory.usedJSHeapSize : null,
  );
  console.log(
    `${name}: ok (tabs=${tabs}, performance.memory=${heap === null ? "unavailable" : `${(heap / (1024 * 1024)).toFixed(2)} MB`})`,
  );
  await browser.close();
}

for (const engine of engines) {
  await runEngine(engine);
}
