import { describe, expect, it, beforeEach } from "vitest";
import {
  clampSidebarWidth,
  defaultSidebarUiState,
  loadSidebarUiState,
  saveSidebarUiState,
  sidebarStorageKey,
} from "./sidebar-state";

describe("sidebar localStorage state", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("clamps width between 180 and 420", () => {
    expect(clampSidebarWidth(100)).toBe(180);
    expect(clampSidebarWidth(500)).toBe(420);
    expect(clampSidebarWidth(240)).toBe(240);
  });

  it("round-trips expanded groups", () => {
    const key = sidebarStorageKey("profile-1");
    const state = defaultSidebarUiState();
    state.expanded.views = true;
    saveSidebarUiState("profile-1", state);
    const loaded = loadSidebarUiState("profile-1");
    expect(loaded.expanded.views).toBe(true);
    expect(localStorage.getItem(key)).toContain("views");
  });
});
