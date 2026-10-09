import { describe, expect, it, beforeEach } from "vitest";
import { activeWorkspaceTabId, workspaceTabs, type WorkspaceTab } from "../state";
import { closeWorkspaceTab, reorderWorkspaceTabs, activateTabAtIndex } from "./tabs";

function tab(id: string): WorkspaceTab {
  return { id, kind: "table_data", schema: "public", table: id, pendingResultLoad: false };
}

describe("workspace tab helpers", () => {
  beforeEach(() => {
    workspaceTabs.value = [tab("a"), tab("b"), tab("c")];
    activeWorkspaceTabId.value = "b";
  });

  it("closeActive removes tab and picks neighbor", () => {
    closeWorkspaceTab("b");
    expect(workspaceTabs.value.map((t) => t.id)).toEqual(["a", "c"]);
    expect(activeWorkspaceTabId.value).toBe("c");
  });

  it("reorder moves tab", () => {
    reorderWorkspaceTabs(2, 0);
    expect(workspaceTabs.value.map((t) => t.id)).toEqual(["c", "a", "b"]);
  });

  it("activateTabAtIndex selects tab", () => {
    activateTabAtIndex(0);
    expect(activeWorkspaceTabId.value).toBe("a");
  });
});
