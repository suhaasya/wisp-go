import { render, screen, fireEvent } from "@testing-library/preact";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { workspaceTabs, activeWorkspaceTabId } from "../../state";
import { TableModeToolbar } from "./mode-toolbar";

describe("TableModeToolbar", () => {
  beforeEach(() => {
    workspaceTabs.value = [];
    activeWorkspaceTabId.value = null;
  });

  it("switches to structure tab when Structure is pressed", () => {
    render(<TableModeToolbar schema="public" table="orders" mode="data" />);
    fireEvent.click(screen.getByRole("button", { name: "Structure" }));
    expect(workspaceTabs.value.some((t) => t.id === "public/orders/structure")).toBe(true);
    expect(activeWorkspaceTabId.value).toBe("public/orders/structure");
  });
});
