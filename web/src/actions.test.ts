import { describe, expect, it } from "vitest";
import { allActionDefs, listActions, TOOLBAR_ACTION_IDS, type ActionContext } from "./actions";

const workspaceTableCtx: ActionContext = {
  routeName: "workspace",
  sessionId: "s1",
  activeTab: { id: "t1", kind: "table_data", schema: "public", table: "orders" },
  readOnly: false,
};

describe("actions registry", () => {
  it("exposes every toolbar action in the palette list", () => {
    const ids = new Set(listActions(workspaceTableCtx).map((a) => a.id));
    for (const id of TOOLBAR_ACTION_IDS) {
      expect(ids.has(id)).toBe(true);
    }
  });

  it("hides write actions when read-only", () => {
    const ctx: ActionContext = { ...workspaceTableCtx, readOnly: true };
    const ids = new Set(listActions(ctx).map((a) => a.id));
    expect(ids.has("import-data")).toBe(false);
    expect(ids.has("export-data")).toBe(true);
  });

  it("registers stable action ids", () => {
    const known = new Set(allActionDefs().map((a) => a.id));
    expect(known.has("new-query")).toBe(true);
  });
});
