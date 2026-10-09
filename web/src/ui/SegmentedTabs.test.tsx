import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { Segmented } from "./Segmented";
import { Tabs } from "./Tabs";

describe("Segmented", () => {
  it("updates selection with aria-pressed", () => {
    const onChange = vi.fn();
    render(
      <Segmented
        aria-label="Environment"
        value="local"
        options={[
          { value: "local", label: "Local" },
          { value: "prod", label: "Prod" },
        ]}
        onChange={onChange}
      />,
    );
    expect(screen.getByRole("button", { name: "Local" }).getAttribute("aria-pressed")).toBe("true");
    fireEvent.click(screen.getByRole("button", { name: "Prod" }));
    expect(onChange).toHaveBeenCalledWith("prod");
  });
});

describe("Tabs", () => {
  it("marks selected tab with aria-selected", () => {
    const onSelect = vi.fn();
    render(
      <Tabs
        aria-label="Connection"
        selectedId="general"
        tabs={[
          { id: "general", label: "General" },
          { id: "ssh", label: "SSH" },
        ]}
        onSelect={onSelect}
      />,
    );
    expect(screen.getByRole("tab", { name: "General" }).getAttribute("aria-selected")).toBe("true");
    fireEvent.click(screen.getByRole("tab", { name: "SSH" }));
    expect(onSelect).toHaveBeenCalledWith("ssh");
  });
});
