#!/usr/bin/env node
/**
 * Generate docs/compatibility.md from Playwright JSON report (e2e/matrix-report.json).
 * Usage: node tools/gen-compatibility.mjs [report.json]
 */
import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = path.join(path.dirname(fileURLToPath(import.meta.url)), "..");
const reportPath = process.argv[2] ?? path.join(repoRoot, "e2e", "matrix-report.json");
const outPath = path.join(repoRoot, "docs", "compatibility.md");

function flattenSuites(suites, out = []) {
  for (const s of suites ?? []) {
    for (const spec of s.specs ?? []) {
      out.push({
        title: spec.title,
        ok: spec.ok,
        tags: (spec.tests ?? []).flatMap((t) => t.annotations?.map((a) => a.description) ?? []),
      });
    }
    flattenSuites(s.suites, out);
  }
  return out;
}

async function main() {
  let report;
  try {
    report = JSON.parse(await fs.readFile(reportPath, "utf8"));
  } catch {
    report = { suites: [] };
  }
  const rows = flattenSuites(report.suites);
  const date = new Date().toISOString().slice(0, 10);
  const lines = [
    "# Compatibility matrix (generated)",
    "",
    `Last updated: ${date} (from \`${path.relative(repoRoot, reportPath)}\`).`,
    "",
    "| Cell | Result |",
    "| --- | --- |",
  ];
  if (rows.length === 0) {
    lines.push("| _no report yet_ | Run \`WISP_MATRIX_TAG=nightly make e2e-matrix\` in CI |");
  } else {
    for (const r of rows) {
      const cell = r.title.replace(/^connect and browse: /, "");
      lines.push(`| ${cell} | ${r.ok ? "pass" : "fail"} |`);
    }
  }
  lines.push("", "## Notes", "", "- PR CI runs cells tagged `pr` (pg17, mysql84, mariadb114 × TCP + SSH).", "- Nightly runs all cells in `e2e/matrix/manifest.mjs`.", "- Socket transports: follow-up (not in CI yet).", "");
  await fs.writeFile(outPath, lines.join("\n"));
  console.log(`wrote ${outPath}`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
