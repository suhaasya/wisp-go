import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { startProcessRefresh } from "./server-tools-refresh";

describe("startProcessRefresh", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("does not tick while hidden", () => {
    const tick = vi.fn();
    const stop = startProcessRefresh({
      intervalMs: 2000,
      isHidden: () => true,
      tick,
    });
    vi.advanceTimersByTime(6000);
    expect(tick).not.toHaveBeenCalled();
    stop();
  });

  it("ticks on interval when visible", () => {
    const tick = vi.fn();
    const stop = startProcessRefresh({
      intervalMs: 2000,
      isHidden: () => false,
      tick,
    });
    vi.advanceTimersByTime(2000);
    expect(tick).toHaveBeenCalledTimes(1);
    vi.advanceTimersByTime(2000);
    expect(tick).toHaveBeenCalledTimes(2);
    stop();
  });
});
