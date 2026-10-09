import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { ImportDialog } from "./import";

describe("ImportDialog", () => {
  it("shows step labels and disables Next without preview", () => {
    render(
      <ImportDialog
        open
        sessionId="s1"
        schema="public"
        defaultTable="orders"
        onClose={() => {}}
      />,
    );
    expect(screen.getByText("1. File")).toBeTruthy();
    const next = screen.getByRole("button", { name: "Next" }) as HTMLButtonElement;
    expect(next.disabled).toBe(true);
    fireEvent.click(screen.getByRole("button", { name: "Choose file…" }));
    expect(screen.getByText("Choose file")).toBeTruthy();
  });
});
