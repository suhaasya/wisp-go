import { cleanup, render, screen } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { goMemoryMB, tabMemoryMB } from "../state";
import { StatusBar } from "./StatusBar";
import * as tabMemory from "./tab-memory";

describe("StatusBar tab meter", () => {
  afterEach(() => {
    cleanup();
    goMemoryMB.value = null;
    tabMemoryMB.value = null;
    vi.restoreAllMocks();
  });

  it("hides Tab meter when heap API is unavailable", () => {
    vi.spyOn(tabMemory, "tabMemorySupported").mockReturnValue(false);
    goMemoryMB.value = 12;
    tabMemoryMB.value = 5;
    render(<StatusBar />);
    expect(screen.getByText(/Go/)).toBeTruthy();
    expect(screen.queryByText(/^Tab$/)).toBeNull();
  });

  it("shows Tab meter when heap API is supported", () => {
    vi.spyOn(tabMemory, "tabMemorySupported").mockReturnValue(true);
    vi.stubGlobal("performance", { memory: { usedJSHeapSize: 4_200_000 } });
    render(<StatusBar />);
    expect(screen.getByText("Tab")).toBeTruthy();
  });
});
