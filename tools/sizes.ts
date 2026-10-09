import { createGzip } from "node:zlib";
import { readFile, writeFile } from "node:fs/promises";
import { join } from "node:path";

const distDir = join(import.meta.dirname, "..", "web", "dist");
const manifestPath = join(distDir, ".vite", "manifest.json");
const outPath = join(distDir, "sizes.json");

const maxInitialJSGzip = 15 * 1024;

type ManifestEntry = {
  file: string;
  css?: string[];
  isEntry?: boolean;
  imports?: string[];
};

async function gzipSize(bytes: Buffer): Promise<number> {
  return new Promise((resolve, reject) => {
    const gz = createGzip();
    const chunks: Buffer[] = [];
    gz.on("data", (c) => chunks.push(c));
    gz.on("error", reject);
    gz.on("end", () => resolve(Buffer.concat(chunks).length));
    gz.end(bytes);
  });
}

const raw = await readFile(manifestPath, "utf8");
const manifest = JSON.parse(raw) as Record<string, ManifestEntry>;

let initialJS = 0;
const lazyChunks: Record<string, number> = {};
let initialCSS = 0;

for (const entry of Object.values(manifest)) {
  if (!entry.isEntry) {
    continue;
  }
  const jsPath = join(distDir, entry.file);
  initialJS += await gzipSize(await readFile(jsPath));
  if (entry.css) {
    for (const cssFile of entry.css) {
      initialCSS += await gzipSize(await readFile(join(distDir, cssFile)));
    }
  }
}

for (const entry of Object.values(manifest)) {
  if (!entry.isEntry && entry.file.includes("editor")) {
    lazyChunks[entry.file] = await gzipSize(await readFile(join(distDir, entry.file)));
  }
}

const report = {
  initialJSGzip: initialJS,
  initialCSSGzip: initialCSS,
  lazyChunksGzip: lazyChunks,
};

await writeFile(outPath, JSON.stringify(report, null, 2) + "\n");

if (initialJS > maxInitialJSGzip) {
  console.error(`initial JS gzip ${initialJS} exceeds ${maxInitialJSGzip} byte budget`);
  process.exit(1);
}

console.log(`wrote ${outPath} (initial JS gzip ${initialJS} bytes)`);
