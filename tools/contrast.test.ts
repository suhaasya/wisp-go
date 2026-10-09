import { describe, expect, it } from "vitest";
import { darkTokens, lightTokens } from "./design-tokens";
import { contrastRatio } from "./wcag-contrast";

const minBodyText = 4.5;

describe("design token contrast (mockup)", () => {
  it("light: ink and ink-2 on panel and canvas meet body text contrast", () => {
    expect(contrastRatio(lightTokens.ink, lightTokens.panel)).toBeGreaterThanOrEqual(minBodyText);
    expect(contrastRatio(lightTokens.ink2, lightTokens.panel)).toBeGreaterThanOrEqual(minBodyText);
    expect(contrastRatio(lightTokens.ink, lightTokens.canvas)).toBeGreaterThanOrEqual(minBodyText);
    expect(contrastRatio(lightTokens.ink2, lightTokens.canvas)).toBeGreaterThanOrEqual(minBodyText);
  });

  it("dark: ink and ink-2 on panel and canvas meet body text contrast", () => {
    expect(contrastRatio(darkTokens.ink, darkTokens.panel)).toBeGreaterThanOrEqual(minBodyText);
    expect(contrastRatio(darkTokens.ink2, darkTokens.panel)).toBeGreaterThanOrEqual(minBodyText);
    expect(contrastRatio(darkTokens.ink, darkTokens.canvas)).toBeGreaterThanOrEqual(minBodyText);
    expect(contrastRatio(darkTokens.ink2, darkTokens.canvas)).toBeGreaterThanOrEqual(minBodyText);
  });
});
