import { describe, expect, it } from "vitest";
import { looksLikeConnectionURL } from "./paste";

describe("looksLikeConnectionURL", () => {
	it("accepts postgres URLs", () => {
		expect(looksLikeConnectionURL("postgres://a@b/c")).toBe(true);
	});
	it("accepts env lines", () => {
		expect(looksLikeConnectionURL('DATABASE_URL="postgres://a@b/c"')).toBe(true);
	});
	it("rejects plain text", () => {
		expect(looksLikeConnectionURL("hello world")).toBe(false);
	});
});
