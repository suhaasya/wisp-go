import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { Menu } from "./Menu";

describe("Menu", () => {
  it("activates item on Enter and closes", () => {
    const onSelect = vi.fn();
    const onClose = vi.fn();
    render(
      <Menu
        open
        x={10}
        y={10}
        activeIndex={0}
        items={[
          { id: "a", label: "Connect" },
          { id: "b", label: "New query" },
        ]}
        onSelect={onSelect}
        onClose={onClose}
        onActiveIndexChange={vi.fn()}
      />,
    );
    fireEvent.keyDown(document, { key: "Enter" });
    expect(onSelect).toHaveBeenCalledWith("a");
    expect(onClose).toHaveBeenCalled();
  });

  it("closes on Escape", () => {
    const onClose = vi.fn();
    render(
      <Menu
        open
        x={0}
        y={0}
        activeIndex={0}
        items={[{ id: "a", label: "One" }]}
        onSelect={vi.fn()}
        onClose={onClose}
        onActiveIndexChange={vi.fn()}
      />,
    );
    fireEvent.keyDown(document, { key: "Escape" });
    expect(onClose).toHaveBeenCalled();
  });
});
