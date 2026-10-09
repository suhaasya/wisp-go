import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { goMemoryMB, wispStopped } from "./state";

type Listener = (ev: { data: string }) => void;

class FakeEventSource {
  static instances: FakeEventSource[] = [];
  url: string;
  onmessage: Listener | null = null;
  onerror: (() => void) | null = null;
  onopen: (() => void) | null = null;

  constructor(url: string) {
    this.url = url;
    FakeEventSource.instances.push(this);
  }

  close() {}
}

describe("events", () => {
  beforeEach(() => {
    goMemoryMB.value = null;
    wispStopped.value = false;
    FakeEventSource.instances = [];
    vi.stubGlobal("EventSource", FakeEventSource);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("updates goMemoryMB on memory event", async () => {
    const { connectEvents, disconnectEvents } = await import("./events");
    connectEvents();
    const es = FakeEventSource.instances[0]!;
    es.onmessage?.({
      data: JSON.stringify({ type: "memory", data: { go_mb: 22 } }),
    });
    expect(goMemoryMB.value).toBe(22);
    disconnectEvents();
  });

  it("applies settings from SSE", async () => {
    const { connectEvents, disconnectEvents } = await import("./events");
    const { themePreference } = await import("./state");
    connectEvents();
    const es = FakeEventSource.instances[0]!;
    es.onmessage?.({
      data: JSON.stringify({ type: "settings", data: { theme: "dark", density: "compact" } }),
    });
    expect(themePreference.value).toBe("dark");
    disconnectEvents();
  });

  it("does not set wispStopped when SSE errors repeat", async () => {
    vi.useFakeTimers();
    const { connectEvents, disconnectEvents } = await import("./events");
    connectEvents();
    const es = FakeEventSource.instances[0]!;
    for (let i = 0; i < 20; i++) {
      es.onerror?.();
      await vi.advanceTimersByTimeAsync(25_000);
    }
    expect(wispStopped.value).toBe(false);
    disconnectEvents();
    vi.useRealTimers();
  });

  it("sets wispStopped on shutdown event", async () => {
    const { connectEvents, disconnectEvents } = await import("./events");
    connectEvents();
    const es = FakeEventSource.instances[0]!;
    es.onmessage?.({ data: JSON.stringify({ type: "shutdown", data: {} }) });
    expect(wispStopped.value).toBe(true);
    disconnectEvents();
  });
});
