import { describe, expect, it } from "vitest";
import { overlayAt } from "./overlay";
import type { WindowPayload } from "./window";

const base: WindowPayload = {
  rid: "r",
  offset: 0,
  total: { kind: "exact", value: 10 },
  columns: [],
  data: [],
  overlay: { mod: [[1, 2]], ins: [3], del: [0] },
};

describe("overlayAt", () => {
  it("maps mod del ins by local row and column", () => {
    expect(overlayAt(base, 0, 0)).toBe("del");
    expect(overlayAt(base, 1, 2)).toBe("mod");
    expect(overlayAt(base, 3, 0)).toBe("ins");
    expect(overlayAt(base, 2, 2)).toBeNull();
  });
});
