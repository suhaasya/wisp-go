import { describe, expect, it } from "vitest";
import { parsePaletteQuery } from "./matcher";

describe("parsePaletteQuery", () => {
  it("detects action prefix", () => {
    expect(parsePaletteQuery("> exp")).toEqual({ mode: "actions", query: "exp" });
  });
  it("detects connection prefix", () => {
    expect(parsePaletteQuery("@ prod")).toEqual({ mode: "connections", query: "prod" });
  });
  it("detects snippet prefix", () => {
    expect(parsePaletteQuery("# daily")).toEqual({ mode: "snippets", query: "daily" });
  });
});
