import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { ConfirmDestructiveDialog } from "./confirm-destructive";

describe("ConfirmDestructiveDialog", () => {
  it("disables run until typed name matches", () => {
    const onConfirm = vi.fn();
    render(
      <ConfirmDestructiveDialog
        open
        title="Drop"
        subtitle="prod"
        sql="DROP TABLE t"
        confirmName="orders"
        needsTypedConfirm
        onCancel={() => {}}
        onConfirm={(name) => onConfirm(name)}
      />,
    );
    const btn = screen.getByTestId("confirm-destructive-run") as HTMLButtonElement;
    expect(btn.disabled).toBe(true);
    fireEvent.input(screen.getByLabelText("Object name confirmation"), {
      target: { value: "orders" },
    });
    expect(btn.disabled).toBe(false);
    fireEvent.click(btn);
    expect(onConfirm).toHaveBeenCalled();
  });
});
