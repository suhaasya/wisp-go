import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";

describe("focus-visible ring in app.css", () => {
  it("uses 2px accent outline with 1px offset", () => {
    const css = readFileSync(join(import.meta.dirname, "..", "web", "src", "styles", "app.css"), "utf8");
    expect(css).toContain("outline: 2px solid var(--color-accent)");
    expect(css).toContain("outline-offset: 1px");
  });
});
