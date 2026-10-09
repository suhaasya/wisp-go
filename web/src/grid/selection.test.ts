import { describe, expect, it } from "vitest";
import { primaryRow, selectCell, selectRow } from "./selection";

describe("primaryRow", () => {
  it("returns min row for row selection", () => {
    expect(primaryRow(selectRow(5))).toBe(5);
  });

  it("returns anchor row for cell selection", () => {
    expect(primaryRow(selectCell(12, 2))).toBe(12);
  });
});
