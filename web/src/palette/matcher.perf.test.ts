import { describe, expect, it } from "vitest";
import { rankByQuery } from "./matcher";

describe("rankByQuery performance", () => {
  it("ranks 10k entries under 16ms per keystroke", () => {
    const items = Array.from({ length: 10_000 }, (_, i) => ({
      id: `t${i}`,
      text: `schema_table_${i}`,
    }));
    const start = performance.now();
    for (let k = 0; k < 5; k++) {
      rankByQuery(`table_${1000 + k}`, items, {});
    }
    const elapsed = performance.now() - start;
    expect(elapsed / 5).toBeLessThan(16);
  });
});
