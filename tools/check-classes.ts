import { readdir, readFile } from "node:fs/promises";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

/** Tailwind arbitrary values that embed raw colour literals. */
const arbitraryColorPattern =
  /-\[[^\]`"']*(?:#[0-9a-fA-F]{3,8}|(?:rgb|rgba|hsl|hsla)\()[^\]`"']*\]/g;

export type ClassViolation = {
  file: string;
  line: number;
  match: string;
};

export function findArbitraryColorClasses(source: string, file = "<string>"): ClassViolation[] {
  const violations: ClassViolation[] = [];
  const lines = source.split("\n");
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]!;
    let m: RegExpExecArray | null;
    arbitraryColorPattern.lastIndex = 0;
    while ((m = arbitraryColorPattern.exec(line)) !== null) {
      violations.push({ file, line: i + 1, match: m[0] });
    }
  }
  return violations;
}

async function walkSrc(dir: string): Promise<string[]> {
  const out: string[] = [];
  const entries = await readdir(dir, { withFileTypes: true });
  for (const ent of entries) {
    const p = join(dir, ent.name);
    if (ent.isDirectory()) {
      out.push(...(await walkSrc(p)));
    } else if (ent.isFile() && (p.endsWith(".ts") || p.endsWith(".tsx"))) {
      out.push(p);
    }
  }
  return out;
}

export async function scanWebSrc(webSrcDir: string): Promise<ClassViolation[]> {
  const files = await walkSrc(webSrcDir);
  const all: ClassViolation[] = [];
  for (const file of files) {
    const text = await readFile(file, "utf8");
    all.push(...findArbitraryColorClasses(text, file));
  }
  return all;
}

async function main(): Promise<void> {
  const webSrc = join(import.meta.dirname, "..", "web", "src");
  const violations = await scanWebSrc(webSrc);
  if (violations.length === 0) {
    return;
  }
  for (const v of violations) {
    console.error(`${v.file}:${v.line}: arbitrary colour class ${v.match}`);
  }
  process.exit(1);
}

const cliEntry = process.argv[1] ? fileURLToPath(import.meta.url) === process.argv[1] : false;
if (cliEntry) {
  main().catch((err) => {
    console.error(err);
    process.exit(1);
  });
}
