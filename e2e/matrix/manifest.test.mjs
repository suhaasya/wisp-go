import test from "node:test";
import assert from "node:assert/strict";
import { allMatrixCells, cellsForTag } from "./manifest.mjs";

test("PR matrix includes pg17, mysql84, mariadb114 TCP and SSH", () => {
  const pr = cellsForTag("pr");
  const ids = new Set(pr.map((c) => c.id));
  for (const want of ["pg17-tcp", "pg17-ssh", "mysql84-tcp", "mysql84-ssh", "mariadb114-tcp", "mariadb114-ssh"]) {
    assert.ok(ids.has(want), `missing ${want}`);
  }
});

test("nightly matrix includes postgres 12 through 17", () => {
  const nightly = cellsForTag("nightly");
  for (let v = 12; v <= 17; v++) {
    assert.ok(nightly.some((c) => c.id === `pg${v}-tcp`), `missing pg${v}-tcp`);
  }
});

test("all cells have unique ids", () => {
  const ids = allMatrixCells().map((c) => c.id);
  assert.equal(ids.length, new Set(ids).size);
});
