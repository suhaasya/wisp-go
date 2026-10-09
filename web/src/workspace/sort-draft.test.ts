import { describe, expect, it } from "vitest";
import { toggleSortDraft } from "./sort-draft";

describe("toggleSortDraft", () => {
  it("click sets ascending primary sort", () => {
    expect(toggleSortDraft([], "created_at", false)).toEqual([
      { column: "created_at", desc: false },
    ]);
  });

  it("click primary again toggles direction", () => {
    const cur = [{ column: "created_at", desc: false }];
    expect(toggleSortDraft(cur, "created_at", false)).toEqual([
      { column: "created_at", desc: true },
    ]);
  });

  it("shift click appends secondary sort", () => {
    const cur = [{ column: "created_at", desc: true }];
    expect(toggleSortDraft(cur, "id", true)).toEqual([
      { column: "created_at", desc: true },
      { column: "id", desc: false },
    ]);
  });
});
