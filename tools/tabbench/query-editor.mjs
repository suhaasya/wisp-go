/**
 * Playwright: open/close query editor 10×; heap delta should stay ≤ 1 MB.
 * @param {import('playwright').Page} page
 * @param {import('playwright').CDPSession} cdp
 */
export async function benchQueryEditorLoop(page, cdp) {
  const base = process.env.WISP_BASE;
  if (!base) {
    throw new Error("WISP_BASE required");
  }
  await page.goto(`${base}/#/bench/query-editor`, { waitUntil: "networkidle" });
  await page.waitForSelector('[data-testid="query-editor-bench"]');

  const heapBefore = await readJSHeap(cdp);

  for (let i = 0; i < 10; i++) {
    await page.getByTestId("query-editor-toggle").click();
    await page.waitForSelector(".cm-editor");
    await page.getByTestId("query-editor-toggle").click();
    await page.waitForSelector("text=Editor closed");
  }

  const heapAfter = await readJSHeap(cdp);
  const deltaMB = (heapAfter - heapBefore) / (1024 * 1024);
  if (deltaMB > 1) {
    throw new Error(`query editor heap delta ${deltaMB.toFixed(2)} MB exceeds 1 MB`);
  }
  console.log(`query-editor: heap delta ${deltaMB.toFixed(2)} MB after 10 open/close cycles`);
  return deltaMB;
}

async function readJSHeap(cdp) {
  const perf = await cdp.send("Performance.getMetrics");
  const byName = Object.fromEntries(perf.metrics.map((m) => [m.name, m.value]));
  return byName.JSHeapUsedSize ?? 0;
}
