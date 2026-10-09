import { describe, expect, it } from "vitest";
import { shouldLoadTableResult } from "./table-load";

describe("shouldLoadTableResult", () => {
  it("skips fetch while pending restore load", () => {
    expect(
      shouldLoadTableResult({
        id: "t",
        kind: "table_data",
        schema: "public",
        table: "orders",
        pendingResultLoad: true,
      }),
    ).toBe(false);
  });

  it("loads after tab activation clears pending", () => {
    expect(
      shouldLoadTableResult({
        id: "t",
        kind: "table_data",
        schema: "public",
        table: "orders",
        pendingResultLoad: false,
      }),
    ).toBe(true);
  });
});
