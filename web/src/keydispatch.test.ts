import { afterEach, describe, expect, it, vi } from "vitest";
import { handleGlobalKeydown } from "./keydispatch";
import * as state from "./state";

describe("handleGlobalKeydown shortcut help", () => {
  afterEach(() => {
    state.shortcutHelpOpen.value = false;
    vi.restoreAllMocks();
  });

  it("opens command palette from a text field", () => {
    const spy = vi.spyOn(state, "openCommandPalette").mockImplementation(() => {});
    handleGlobalKeydown({
      key: "k",
      metaKey: true,
      ctrlKey: false,
      altKey: false,
      shiftKey: false,
      target: document.createElement("input"),
      preventDefault: () => {},
    } as unknown as KeyboardEvent);
    expect(spy).toHaveBeenCalled();
  });

  it("delegates Mod-/ to toggleShortcutHelp", () => {
    const spy = vi.spyOn(state, "toggleShortcutHelp").mockImplementation(() => {});
    handleGlobalKeydown({
      key: "/",
      metaKey: true,
      ctrlKey: false,
      altKey: false,
      shiftKey: false,
      preventDefault: () => {},
    } as KeyboardEvent);
    expect(spy).toHaveBeenCalled();
  });
});
