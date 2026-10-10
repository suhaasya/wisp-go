import { afterEach, describe, expect, it, vi } from "vitest";
import { openCommandPalette, paletteOpen } from "./state";

describe("openCommandPalette", () => {
  afterEach(() => {
    paletteOpen.value = false;
    vi.restoreAllMocks();
  });

  it("sets paletteOpen after the lazy chunk loads", async () => {
    paletteOpen.value = false;
    let resolveLoad!: () => void;
    const loadGate = new Promise<void>((r) => {
      resolveLoad = r;
    });
    vi.spyOn(await import("./palette-load"), "loadCommandPalette").mockImplementation(() =>
      loadGate.then(() => import("./palette/index")),
    );

    openCommandPalette();
    expect(paletteOpen.value).toBe(false);
    resolveLoad();
    await vi.waitFor(() => expect(paletteOpen.value).toBe(true));
  });
});
