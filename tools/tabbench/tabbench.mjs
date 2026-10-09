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

  await page.goto(`${base}/#/bench/workspace-tabs`, { waitUntil: "networkidle" });
  await page.waitForSelector('[data-testid="workspace-tabs-bench"]');
  const tabCount = await page.locator('[role="tab"]').count();
  if (tabCount !== 20) {
    throw new Error(`expected 20 workspace tabs, got ${tabCount}`);
  }
  const wsCdp = await cdp.send("Performance.getMetrics");
  const wsByName = Object.fromEntries(wsCdp.metrics.map((m) => [m.name, m.value]));
  metrics.js_heap_idle_mb = { value: (wsByName.JSHeapUsedSize ?? jsBytes) / (1024 * 1024) };
  metrics.dom_nodes = { value: wsByName.Nodes ?? metrics.dom_nodes.value };

  await page.goto(`${base}/#/bench/grid`, { waitUntil: "networkidle" });
  await page.waitForSelector('[data-testid="wisp-grid-scroll"]');

  const longTasks = [];
  await page.evaluate(() => {
    window.__wispLongTasks = [];
    new PerformanceObserver((list) => {
      for (const e of list.getEntries()) {
        window.__wispLongTasks.push(e.duration);
      }
    }).observe({ type: "longtask", buffered: true });
  });

  const grid = page.locator('[data-testid="wisp-grid-scroll"]');
  const maxScroll = await grid.evaluate((el) => el.scrollHeight - el.clientHeight);
  const steps = 40;
  for (let i = 0; i <= steps; i++) {
    await grid.evaluate((el, top) => {
      el.scrollTop = top;
    }, (maxScroll * i) / steps);
    await page.waitForTimeout(16);
  }

  const after = await cdp.send("Performance.getMetrics");
  const afterByName = Object.fromEntries(after.metrics.map((m) => [m.name, m.value]));
  const jsAfter = afterByName.JSHeapUsedSize ?? jsBytes;
  metrics.js_heap_1m_mb = { value: jsAfter / (1024 * 1024) };
  metrics.dom_nodes = { value: afterByName.Nodes ?? metrics.dom_nodes.value };

  const longest = await page.evaluate(() => {
    const tasks = window.__wispLongTasks ?? [];
    return tasks.length ? Math.max(...tasks) : 0;
  });
  metrics.long_task_ms = { value: longest };

  try {
    const { benchQueryEditorLoop } = await import("./query-editor.mjs");
    const delta = await benchQueryEditorLoop(page, cdp);
    metrics.js_heap_query_editor_delta_mb = { value: delta };
  } catch (err) {
    metrics.js_heap_query_editor_delta_mb = {
      skipped: true,
      reason: err instanceof Error ? err.message : String(err),
    };
  }

  await browser.close();
  await writeResults();
}

async function writeResults() {
  await mkdir(dirname(outPath), { recursive: true });
  await writeFile(outPath, JSON.stringify({ metrics }, null, 2) + "\n");
  console.log("tabbench: wrote", outPath);
}

await run();
