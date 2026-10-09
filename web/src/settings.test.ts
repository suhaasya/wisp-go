import { beforeEach, describe, expect, it } from "vitest";
import { themePreference } from "./state";
import { applySettings, getSettings } from "./settings";

describe("applySettings", () => {
  beforeEach(() => {
    document.documentElement.removeAttribute("data-density");
    document.documentElement.removeAttribute("data-theme");
  });

  it("applies theme and density", () => {
    applySettings({ theme: "dark", density: "comfortable" });
    expect(themePreference.value).toBe("dark");
    expect(document.documentElement.getAttribute("data-density")).toBe("comfortable");
    expect(document.documentElement.getAttribute("data-theme")).toBe("dark");
  });

  it("mirrors density to row_height for the grid", () => {
    applySettings({ density: "comfortable" });
    expect((getSettings() as { row_height?: string }).row_height).toBe("comfortable");
  });
});
