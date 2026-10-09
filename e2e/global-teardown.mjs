import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.join(fileURLToPath(import.meta.url), "..");
const repoRoot = path.join(root, "..");
const pidFile = path.join(repoRoot, "bench-results", "e2e-runtime", "wisp-e2e.pid");

export default async function globalTeardown() {
  if (process.env.WISP_LAUNCH_URL && process.env.WISP_E2E_NO_TEARDOWN) {
    return;
  }
  try {
    const pid = Number(await fs.readFile(pidFile, "utf8"));
    if (pid > 0) {
      process.kill(pid);
    }
    await fs.unlink(pidFile);
  } catch {
    // spawned externally or already stopped
  }
}
