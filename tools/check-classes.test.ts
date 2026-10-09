import { describe, expect, it } from "vitest";
import { findArbitraryColorClasses } from "./check-classes";

describe("findArbitraryColorClasses", () => {
  it("flags hex arbitrary colour utilities", () => {
    const hits = findArbitraryColorClasses('class="bg-panel text-[#1D222B]"');
    expect(hits.length).toBeGreaterThan(0);
    expect(hits[0]?.match).toContain("#1D222B");
  });

  it("allows token utilities only", () => {
    const hits = findArbitraryColorClasses('class="bg-canvas text-ink-2 border-line"');
    expect(hits).toHaveLength(0);
  });
});
