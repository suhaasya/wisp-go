/**
 * Playwright checks for dev gallery (#/gallery).
 * Run: WISP_GALLERY_URL=http://127.0.0.1:5173 npm --prefix tools/tabbench run gallery
 */
import { chromium } from "playwright";

const url = process.env.WISP_GALLERY_URL ?? "http://127.0.0.1:5173/#/gallery";

async function main() {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();

  await page.goto(url, { waitUntil: "networkidle" });
  await page.getByRole("heading", { name: "Component gallery" }).waitFor();

  for (const theme of ["light", "dark"]) {
    await page.getByRole("button", { name: theme, exact: true }).click();
    await page.waitForFunction(
      (t) => document.documentElement.getAttribute("data-theme") === t,
      theme,
    );
    await page.getByRole("heading", { name: "Buttons" }).waitFor();
  }

  const cdp = await page.context().newCDPSession(page);
  const before = await cdp.send("Performance.getMetrics");
  const nodesBefore = before.metrics.find((m) => m.name === "Nodes")?.value ?? 0;

  await page.evaluate(async () => {
    const fn = window.__wispGalleryStressCycles;
    if (typeof fn !== "function") {
      throw new Error("missing __wispGalleryStressCycles (dev gallery only)");
    }
    await fn(1000);
  });

  const after = await cdp.send("Performance.getMetrics");
  const nodesAfter = after.metrics.find((m) => m.name === "Nodes")?.value ?? 0;
  const nodeDelta = nodesAfter - nodesBefore;
  if (nodeDelta > 50) {
    throw new Error(`gallery stress: DOM Nodes grew by ${nodeDelta} (before ${nodesBefore}, after ${nodesAfter})`);
  }

  console.log(`gallery: smoke OK; 1000 mount cycles node delta ${nodeDelta}`);
  await browser.close();
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
