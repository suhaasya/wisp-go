/** True when Chromium exposes performance.memory for the tab heap meter. */
export function tabMemorySupported(): boolean {
  if (typeof performance === "undefined") {
    return false;
  }
  const perf = performance as Performance & { memory?: { usedJSHeapSize: number } };
  return typeof perf.memory?.usedJSHeapSize === "number" && perf.memory.usedJSHeapSize > 0;
}
