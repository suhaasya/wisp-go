import { describe, expect, it } from "vitest";
import { buildColumns, defaultColumnWidthPx } from "./columns";

describe("defaultColumnWidthPx", () => {
  it("uses 12vw of the viewport clamped between 60 and 480", () => {
    expect(defaultColumnWidthPx(1000)).toBe(120);
    expect(defaultColumnWidthPx(5000)).toBe(480);
    expect(defaultColumnWidthPx(400)).toBe(60);
  });
});

describe("buildColumns", () => {
  it("assigns the default width from the viewport", () => {
    const cols = buildColumns([{ name: "id", type: "int" }], 800);
    expect(cols[0]!.width).toBe(96);
  });
});
