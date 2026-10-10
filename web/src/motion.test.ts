import { afterEach, describe, expect, it, vi } from "vitest";
import { prefersReducedMotion } from "./motion";

describe("prefersReducedMotion", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("is true when the user prefers reduced motion", () => {
    vi.stubGlobal("matchMedia", (query: string) => ({
      matches: query === "(prefers-reduced-motion: reduce)",
      media: query,
    }));
    expect(prefersReducedMotion()).toBe(true);
  });

  it("is false when motion is not reduced", () => {
    vi.stubGlobal("matchMedia", () => ({
      matches: false,
      media: "",
    }));
    expect(prefersReducedMotion()).toBe(false);
  });
});
