import { describe, expect, it } from "vitest";
import { buildSidebarRows, type GroupLoadState, SCHEMA_GROUPS } from "./tree";

describe("buildSidebarRows", () => {
  it("includes group headers with counts when names are loaded", () => {
    const expanded = { tables: true, views: false, functions: false, sequences: false, types: false };
    const namesByGroup = { tables: ["a", "b"], views: [], functions: [], sequences: [], types: [] };
    const load: GroupLoadState = {
      tables: "ready",
      views: "idle",
      functions: "idle",
      sequences: "idle",
      types: "idle",
    };
    const rows = buildSidebarRows({
      expanded,
      namesByGroup,
      load,
      filterQuery: "",
    });
    const tableGroup = rows.find((r) => r.type === "group" && r.id === "tables");
    expect(tableGroup?.type === "group" && tableGroup.count).toBe(2);
    expect(rows.some((r) => r.type === "object" && r.name === "a")).toBe(true);
    expect(rows.some((r) => r.type === "object" && r.name === "b")).toBe(true);
  });

  it("shows loading row for expanded group that is loading", () => {
    const expanded = { tables: true, views: false, functions: false, sequences: false, types: false };
    const load: GroupLoadState = {
      tables: "loading",
      views: "idle",
      functions: "idle",
      sequences: "idle",
      types: "idle",
    };
    const rows = buildSidebarRows({
      expanded,
      namesByGroup: { tables: [], views: [], functions: [], sequences: [], types: [] },
      load,
      filterQuery: "",
    });
    expect(rows.some((r) => r.type === "loading" && r.groupId === "tables")).toBe(true);
  });

  it("applies filter to object rows only", () => {
    const expanded = Object.fromEntries(SCHEMA_GROUPS.map((g) => [g.id, g.id === "tables"])) as Record<
      (typeof SCHEMA_GROUPS)[number]["id"],
      boolean
    >;
    const rows = buildSidebarRows({
      expanded,
      namesByGroup: {
        tables: ["orders", "customers"],
        views: [],
        functions: [],
        sequences: [],
        types: [],
      },
      load: {
        tables: "ready",
        views: "idle",
        functions: "idle",
        sequences: "idle",
        types: "idle",
      },
      filterQuery: "ord",
    });
    const objects = rows.filter((r) => r.type === "object");
    expect(objects.map((r) => (r.type === "object" ? r.name : ""))).toEqual(["orders"]);
  });
});
