import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  initTheme,
  resetThemeForTests,
  setThemePreference,
  themePreference,
  type ThemePreference,
} from "./theme";

const storageKey = "wisp-theme";

function mockMatchMedia(dark: boolean) {
  const listeners = new Set<(e: MediaQueryListEvent) => void>();
  const mql = {
    matches: dark,
    media: "(prefers-color-scheme: dark)",
    addEventListener: (_: string, fn: (e: MediaQueryListEvent) => void) => {
      listeners.add(fn);
    },
    removeEventListener: (_: string, fn: (e: MediaQueryListEvent) => void) => {
      listeners.delete(fn);
    },
    dispatchChange(darkNext: boolean) {
      mql.matches = darkNext;
      const ev = { matches: darkNext } as MediaQueryListEvent;
      for (const fn of listeners) {
        fn(ev);
      }
    },
  };
  vi.stubGlobal("matchMedia", (query: string) => {
    if (query !== "(prefers-color-scheme: dark)") {
      throw new Error(`unexpected query ${query}`);
    }
    return mql as unknown as MediaQueryList;
  });
  return mql;
}

describe("theme", () => {
  beforeEach(() => {
    resetThemeForTests();
    document.documentElement.removeAttribute("data-theme");
    localStorage.clear();
    themePreference.value = "system";
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("applies light and dark on documentElement", () => {
    mockMatchMedia(false);
    initTheme();
    setThemePreference("dark");
    expect(document.documentElement.getAttribute("data-theme")).toBe("dark");
    setThemePreference("light");
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
  });

  it("system follows prefers-color-scheme and updates on change", () => {
    const mql = mockMatchMedia(true);
    initTheme();
    setThemePreference("system");
    expect(document.documentElement.getAttribute("data-theme")).toBe("dark");
    mql.dispatchChange(false);
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
  });

  it("persists preference in localStorage", () => {
    mockMatchMedia(false);
    initTheme();
    setThemePreference("dark");
    expect(localStorage.getItem(storageKey)).toBe("dark");
    const allowed: ThemePreference[] = ["light", "dark", "system"];
    expect(allowed).toContain(localStorage.getItem(storageKey) as ThemePreference);
  });
});
