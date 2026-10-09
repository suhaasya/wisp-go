import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { Segmented } from "./Segmented";

describe("Segmented", () => {
  it("shows selected option with accent background", () => {
    render(
      <Segmented
        options={[
          { value: "a", label: "A" },
          { value: "b", label: "B" },
        ]}
        value="b"
        onChange={() => {}}
      />,
    );
    const selected = screen.getByRole("button", { name: "B" });
    const other = screen.getByRole("button", { name: "A" });
    expect(selected.className).toContain("bg-accent-soft");
    expect(selected.className).not.toContain("bg-transparent");
    expect(other.className).toContain("bg-transparent");
    expect(selected.getAttribute("aria-pressed")).toBe("true");
  });

  it("calls onChange when an option is clicked", () => {
    const onChange = vi.fn();
    render(
      <Segmented
        options={[
          { value: "postgres", label: "PostgreSQL" },
          { value: "mysql", label: "MySQL" },
        ]}
        value="postgres"
        onChange={onChange}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "MySQL" }));
    expect(onChange).toHaveBeenCalledWith("mysql");
  });
});
