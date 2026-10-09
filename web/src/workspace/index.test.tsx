import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { WorkspaceTabsBench } from "./WorkspaceTabsBench";

describe("WorkspaceTabsBench", () => {
  it("renders 20 tabs and one active pane", () => {
    render(<WorkspaceTabsBench />);
    expect(screen.getByTestId("workspace-tab-count").textContent).toBe("20");
    expect(document.querySelectorAll('[role="tab"]').length).toBe(20);
    expect(screen.getAllByTestId("workspace-tab-pane").length).toBe(1);
  });
});
