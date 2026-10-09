import { describe, expect, it } from "vitest";
import { filterNames } from "./filter";

describe("filterNames", () => {
  it("returns all names when query is empty", () => {
    const names = ["orders", "customers", "products"];
    expect(filterNames(names, "")).toEqual(names);
    expect(filterNames(names, "   ")).toEqual(names);
  });

  it("filters case-insensitively by substring", () => {
    const names = ["Orders", "order_items", "customers"];
    expect(filterNames(names, "ord")).toEqual(["Orders", "order_items"]);
  });
});
