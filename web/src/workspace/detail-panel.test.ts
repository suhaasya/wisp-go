import { describe, expect, it } from "vitest";
import { formatPreviewValue } from "./detail-panel";

describe("formatPreviewValue", () => {
  it("shows NULL for nullish", () => {
    expect(formatPreviewValue(null)).toBe("NULL");
  });
});
