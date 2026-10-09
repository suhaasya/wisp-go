import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { Button } from "./Button";

describe("Button", () => {
  it("renders label and fires onClick", () => {
    const onClick = vi.fn();
    render(<Button label="Connect" onClick={onClick} />);
    fireEvent.click(screen.getByRole("button", { name: "Connect" }));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it("shows kbd hint in accessible name", () => {
    render(<Button label="Save" kbd="⌘S" />);
    expect(screen.getByRole("button", { name: /Save/ })).toBeTruthy();
    expect(screen.getByText("⌘S")).toBeTruthy();
  });

  it("respects disabled", () => {
    const onClick = vi.fn();
    render(<Button label="Run" disabled onClick={onClick} />);
    const btn = screen.getByRole("button", { name: "Run" });
    expect(btn).toHaveProperty("disabled", true);
    fireEvent.click(btn);
    expect(onClick).not.toHaveBeenCalled();
  });
});
