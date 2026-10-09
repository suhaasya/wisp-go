import { beforeEach, describe, expect, it } from "vitest";
import { themePreference } from "./state";
import { applySettings } from "./settings";

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
});
