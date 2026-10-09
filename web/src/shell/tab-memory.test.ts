import { afterEach, describe, expect, it, vi } from "vitest";
import { tabMemorySupported } from "./tab-memory";

describe("tabMemorySupported", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("is false without performance.memory", () => {
    vi.stubGlobal("performance", {});
    expect(tabMemorySupported()).toBe(false);
  });

  it("is true when usedJSHeapSize is available", () => {
    vi.stubGlobal("performance", { memory: { usedJSHeapSize: 4_000_000 } });
    expect(tabMemorySupported()).toBe(true);
  });
});
