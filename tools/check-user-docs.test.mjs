import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.join(path.dirname(fileURLToPath(import.meta.url)), "..");
const userDir = path.join(root, "docs/user");
const readme = fs.readFileSync(path.join(userDir, "README.md"), "utf8");

test("user guide README links resolve to files", () => {
  const links = [...readme.matchAll(/\]\(([^)]+\.md)\)/g)].map((m) => m[1]);
  assert.ok(links.length >= 10, "expected guide index links");
  for (const rel of links) {
    const full = path.join(userDir, rel);
    assert.ok(fs.existsSync(full), `missing ${rel}`);
  }
});

test("release checklist references user docs and CHANGELOG", () => {
  const checklist = fs.readFileSync(path.join(root, "docs/release-checklist.md"), "utf8");
  assert.match(checklist, /CHANGELOG\.md/);
  assert.match(checklist, /docs\/user/);
  assert.match(checklist, /make gate/);
});
