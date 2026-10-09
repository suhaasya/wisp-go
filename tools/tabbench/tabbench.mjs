import { writeFile, mkdir } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(fileURLToPath(import.meta.url), "..", "..", "..");
const outPath = join(root, "bench-results", "tab.json");

const skipAll = process.env.WISP_SKIP_TAB === "1";

/** @type {Record<string, { value?: number, skipped?: boolean, reason?: string }>} */
const metrics = {};

function skip(metric, reason) {
  metrics[metric] = { skipped: true, reason };
}

async function run() {
  if (skipAll) {
    for (const m of [
      "js_heap_idle_mb",
      "js_heap_1m_mb",
      "dom_nodes",
      "tab_total_mb",
      "long_task_ms",
    ]) {
      skip(m, "WISP_SKIP_TAB=1");
    }
    await writeResults();
    return;
  }

  let chromium;
  try {
    ({ chromium } = await import("playwright"));
  } catch {
    for (const m of [
      "js_heap_idle_mb",
      "js_heap_1m_mb",
      "dom_nodes",
      "tab_total_mb",
      "long_task_ms",
    ]) {
      skip(m, "playwright not installed (npm ci in tools/tabbench)");
    }
    await writeResults();
    return;
  }

  const base = process.env.WISP_BASE;
  const launchURL = process.env.WISP_LAUNCH_URL;
  if (!base || !launchURL) {
    for (const m of [
      "js_heap_idle_mb",
      "js_heap_1m_mb",
      "dom_nodes",
      "tab_total_mb",
      "long_task_ms",
    ]) {
      skip(m, "set WISP_BASE and WISP_LAUNCH_URL (run wisp first)");
    }
    await writeResults();
    return;
  }

  const browser = await chromium.launch({
    headless: true,
    args: ["--disable-extensions", "--no-first-run"],
  });
  const context = await browser.newContext();
  const page = await context.newPage();

  await page.goto(launchURL, { waitUntil: "networkidle" });
  await page.waitForURL(base + "/");

  const cdp = await context.newCDPSession(page);
  const perf = await cdp.send("Performance.getMetrics");
  const byName = Object.fromEntries(perf.metrics.map((m) => [m.name, m.value]));
  const jsBytes = byName.JSHeapUsedSize ?? 0;
  metrics.js_heap_idle_mb = { value: jsBytes / (1024 * 1024) };
  metrics.dom_nodes = { value: byName.Nodes ?? 0 };
  metrics.tab_total_mb = { value: (byName.JSHeapUsedSize ?? 0) / (1024 * 1024) };

  skip("js_heap_1m_mb", "open_table_scroll_1m not implemented (M1)");
  skip("long_task_ms", "grid scroll long-task observer not implemented (M1)");

  await browser.close();
  await writeResults();
}

async function writeResults() {
  await mkdir(dirname(outPath), { recursive: true });
  await writeFile(outPath, JSON.stringify({ metrics }, null, 2) + "\n");
  console.log("tabbench: wrote", outPath);
}

await run();
