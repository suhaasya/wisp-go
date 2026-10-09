import { cleanup, fireEvent, render, screen } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { Modal } from "./Modal";

describe("Modal", () => {
  afterEach(() => {
    cleanup();
  });

  it("exposes dialog semantics and closes on Escape", () => {
    const onClose = vi.fn();
    render(
      <Modal open title="Confirm action" onClose={onClose}>
        <p>Body</p>
      </Modal>,
    );
    const dialog = screen.getByRole("dialog", { name: "Confirm action" });
    expect(dialog.getAttribute("aria-modal")).toBe("true");
    dialog.dispatchEvent(new Event("cancel", { cancelable: true }));
    expect(onClose).toHaveBeenCalled();
  });
});
