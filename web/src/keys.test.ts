import { describe, expect, it } from "vitest";
import {
  DEFAULT_KEYBINDINGS,
  eventMatchesBinding,
  findKeybindingConflicts,
  formatBinding,
  formatBindingDisplay,
  isBrowserReservedBinding,
  parseBinding,
  resolveKeybindings,
} from "./keys";

describe("parseBinding / formatBinding", () => {
  it("round-trips Mod-K", () => {
    const p = parseBinding("Mod-K");
    expect(p).not.toBeNull();
    expect(formatBinding(p!)).toBe("Mod-K");
  });

  it("round-trips Shift-Mod-Enter", () => {
    const p = parseBinding("Shift-Mod-Enter");
    expect(p).not.toBeNull();
    expect(formatBinding(p!)).toBe("Shift-Mod-Enter");
  });

  it("round-trips Alt-1", () => {
    const p = parseBinding("Alt-1");
    expect(p).not.toBeNull();
    expect(formatBinding(p!)).toBe("Alt-1");
  });

  it("rejects empty input", () => {
    expect(parseBinding("")).toBeNull();
    expect(parseBinding("   ")).toBeNull();
  });
});

describe("isBrowserReservedBinding", () => {
  it("flags Mod-W", () => {
    expect(isBrowserReservedBinding("Mod-W")).toBe(true);
  });

  it("allows Mod-K", () => {
    expect(isBrowserReservedBinding("Mod-K")).toBe(false);
  });
});

describe("resolveKeybindings", () => {
  it("returns defaults when overrides empty", () => {
    const resolved = resolveKeybindings({});
    expect(resolved["command-palette"]).toBe(DEFAULT_KEYBINDINGS["command-palette"]);
  });

  it("applies a valid override", () => {
    const resolved = resolveKeybindings({ "command-palette": "Alt-P" });
    expect(resolved["command-palette"]).toBe("Alt-P");
  });

  it("ignores reserved override chords", () => {
    const resolved = resolveKeybindings({ "command-palette": "Mod-W" });
    expect(resolved["command-palette"]).toBe(DEFAULT_KEYBINDINGS["command-palette"]);
  });
});

describe("findKeybindingConflicts", () => {
  it("detects duplicate chords", () => {
    const resolved = resolveKeybindings({
      "command-palette": "Alt-P",
      "new-connection": "Alt-P",
    });
    const conflicts = findKeybindingConflicts(resolved);
    expect(conflicts).toHaveLength(1);
    expect(conflicts[0]!.chord).toBe("Alt-P");
    expect(conflicts[0]!.ids.sort()).toEqual(["command-palette", "new-connection"]);
  });
});

describe("formatBindingDisplay", () => {
  it("shows ⌘ on mac", () => {
    expect(formatBindingDisplay("Mod-K", { mac: true })).toBe("⌘K");
  });

  it("shows Ctrl on non-mac", () => {
    expect(formatBindingDisplay("Mod-K", { mac: false })).toBe("Ctrl+K");
  });
});

describe("eventMatchesBinding", () => {
  it("matches Mod-K with metaKey", () => {
    const e = { metaKey: true, ctrlKey: false, altKey: false, shiftKey: false, key: "k" } as KeyboardEvent;
    expect(eventMatchesBinding(e, "Mod-K")).toBe(true);
  });

  it("matches Alt-T without mod", () => {
    const e = { metaKey: false, ctrlKey: false, altKey: true, shiftKey: false, key: "t" } as KeyboardEvent;
    expect(eventMatchesBinding(e, "Alt-T")).toBe(true);
  });

  it("matches Alt-T on macOS when Option produces a special character", () => {
    const e = {
      metaKey: false,
      ctrlKey: false,
      altKey: true,
      shiftKey: false,
      key: "†",
      code: "KeyT",
    } as KeyboardEvent;
    expect(eventMatchesBinding(e, "Alt-T")).toBe(true);
  });
});
