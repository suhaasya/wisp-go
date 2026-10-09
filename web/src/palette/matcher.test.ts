import { describe, expect, it } from "vitest";
import { fuzzyScore, rankByQuery } from "./matcher";

describe("fuzzyScore", () => {
  it("ranks exact prefix above subsequence match", () => {
    const orders = fuzzyScore("ord", "orders");
    const customers = fuzzyScore("ord", "customers");
    expect(orders).toBeGreaterThan(customers);
  });

  it("returns zero when query letters are out of order", () => {
    expect(fuzzyScore("dro", "orders")).toBe(0);
  });
});

describe("rankByQuery", () => {
  it("orders items by score descending", () => {
    const items = [
      { id: "a", text: "customers" },
      { id: "b", text: "orders" },
      { id: "c", text: "order_items" },
    ];
    const ranked = rankByQuery("ord", items, {});
    expect(ranked.map((x) => x.id)).toEqual(["c", "b"]);
    expect(ranked.some((x) => x.id === "a")).toBe(false);
  });

  it("boosts recency for ties", () => {
    const items = [
      { id: "a", text: "orders_alpha" },
      { id: "b", text: "orders_beta" },
    ];
    const ranked = rankByQuery("orders", items, { b: 100, a: 1 });
    expect(ranked[0]?.id).toBe("b");
  });
});
