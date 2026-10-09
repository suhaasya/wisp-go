import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { ShortcutsPane } from "./shortcuts";

describe("ShortcutsPane", () => {
  it("shows conflict alert when two actions share a chord", () => {
    render(
      <ShortcutsPane
        doc={{
          keybindings: {
            "command-palette": "Alt-P",
            "new-connection": "Alt-P",
          },
        }}
        onPatch={() => {}}
      />,
    );
    expect(screen.getByRole("alert").textContent).toContain("Conflicting shortcuts");
  });
});
