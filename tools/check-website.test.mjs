import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.join(path.dirname(fileURLToPath(import.meta.url)), "..");
const html = fs.readFileSync(path.join(root, "website/index.html"), "utf8");

/** Deal-breaker caps from internal/bench/budgets.go (independent literal check). */
const goCaps = ["25 MB", "10 MB", "30 MB", "45 MB", "4 MB", "80 MB", "300 ms"];
const tabCaps = ["10 MB", "20 MB", "4,000", "100 KB", "280 KB", "25 KB", "70 MB"];

test("landing page lists Go process deal-breaker caps", () => {
  for (const cap of goCaps) {
    assert.ok(html.includes(cap), `missing Go cap ${cap}`);
  }
});

test("landing page lists browser tab deal-breaker caps", () => {
  for (const cap of tabCaps) {
    assert.ok(html.includes(cap), `missing tab cap ${cap}`);
  }
});

test("landing page has install and download CTAs", () => {
  assert.match(html, /install\.sh/);
  assert.match(html, /releases\/latest/);
});
