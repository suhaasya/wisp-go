import { cleanup, fireEvent, render, screen } from "@testing-library/preact";
import { afterEach, describe, expect, it } from "vitest";
import { paletteOpen } from "../state";
import { CommandPalette } from "./index";

describe("CommandPalette", () => {
  afterEach(() => {
    cleanup();
    paletteOpen.value = false;
  });
  it("opens dialog when paletteOpen signal is true", () => {
    paletteOpen.value = true;
    render(<CommandPalette />);
    expect(screen.getByRole("dialog", { name: "Command palette" })).toBeTruthy();
    expect(screen.getByLabelText("Command")).toBeTruthy();
    paletteOpen.value = false;
  });

  it("closes on Escape", () => {
    paletteOpen.value = true;
    render(<CommandPalette />);
    fireEvent.keyDown(screen.getByRole("dialog", { name: "Command palette" }), {
      key: "Escape",
    });
    expect(paletteOpen.value).toBe(false);
  });
});
