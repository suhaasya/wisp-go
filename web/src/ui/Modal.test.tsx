import { fireEvent, render, screen } from "@testing-library/preact";
import { useState } from "preact/hooks";
import { describe, expect, it } from "vitest";
import { Button } from "./Button";
import { Modal } from "./Modal";

function ModalHarness() {
  const [open, setOpen] = useState(false);
  return (
    <div>
      <Button label="Open review" onClick={() => setOpen(true)} />
      <Modal open={open} title="Review SQL" onClose={() => setOpen(false)}>
        <p>SELECT 1;</p>
      </Modal>
    </div>
  );
}

describe("Modal", () => {
  it("opens dialog content when trigger clicked", () => {
    render(<ModalHarness />);
    fireEvent.click(screen.getByRole("button", { name: "Open review" }));
    expect(screen.getByText("Review SQL")).toBeTruthy();
    expect(screen.getByText("SELECT 1;")).toBeTruthy();
  });

});
