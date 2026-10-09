import { describe, expect, it, vi } from "vitest";
import { createWindowBuffer, offsetsNeededForViewport } from "./buffer";
import { DEFAULT_WINDOW_ROWS, type WindowPayload } from "./window";

describe("offsetsNeededForViewport", () => {
  const w = DEFAULT_WINDOW_ROWS;

  it("requests windows that cover the viewport rows", () => {
    const { required } = offsetsNeededForViewport(100, 200, w, []);
    expect(required).toEqual([0, 150]);
  });

  it("prefetches the previous window when within half a window of the top edge", () => {
    const { prefetch } = offsetsNeededForViewport(160, 200, w, [150]);
    expect(prefetch).toContain(0);
  });

  it("prefetches the next window when within half a window of the bottom edge", () => {
    const { prefetch } = offsetsNeededForViewport(0, 80, w, [0]);
    expect(prefetch).toContain(150);
  });
});

function fakeWindow(offset: number): WindowPayload {
  return {
    rid: "r1",
    offset,
    total: { kind: "exact", value: 10_000 },
    columns: [{ name: "id", type: "int8" }],
    data: [[offset]],
  };
}

describe("createWindowBuffer", () => {
  it("keeps at most three windows and evicts the least recently used", async () => {
    const fetches: number[] = [];
    const buf = createWindowBuffer({
      windowRows: 150,
      maxWindows: 3,
      fetch: async (offset) => {
        fetches.push(offset);
        return fakeWindow(offset);
      },
    });

    await buf.ensureViewport(0, 50);
    await buf.ensureViewport(150, 200);
    await buf.ensureViewport(300, 350);
    await buf.ensureViewport(450, 500);

    expect(buf.cachedOffsets().length).toBeLessThanOrEqual(3);
    expect(buf.getWindow(450)).toBeDefined();
    buf.destroy();
  });

  it("aborts an in-flight fetch when a newer ensureViewport supersedes it", async () => {
    let aborted = false;
    const buf = createWindowBuffer({
      windowRows: 150,
      maxWindows: 3,
      fetch: (offset, signal) => {
        if (offset === 0) {
          return new Promise((_resolve, reject) => {
            signal.addEventListener("abort", () => {
              aborted = true;
              reject(new DOMException("aborted", "AbortError"));
            });
          });
        }
        return Promise.resolve(fakeWindow(offset));
      },
    });

    const first = buf.ensureViewport(0, 50);
    const second = buf.ensureViewport(300, 350);
    await expect(first).rejects.toBeDefined();
    await second;
    expect(aborted).toBe(true);
    buf.destroy();
  });

  it("records last fetch time after a successful load", async () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-01T00:00:00Z"));
    const buf = createWindowBuffer({
      windowRows: 150,
      maxWindows: 3,
      fetch: async (offset) => fakeWindow(offset),
    });
    await buf.ensureViewport(0, 50);
    expect(buf.lastFetchTime()?.toISOString()).toBe("2026-01-01T00:00:00.000Z");
    buf.destroy();
    vi.useRealTimers();
  });
});
