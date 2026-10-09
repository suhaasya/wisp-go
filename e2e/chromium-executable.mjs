import fs from "node:fs";
import os from "node:os";
import path from "node:path";

/** Fallback when chromium_headless_shell is not installed (uses full Chromium binary). */
export function chromiumExecutablePath() {
  const cache = path.join(os.homedir(), "Library", "Caches", "ms-playwright");
  let entries = [];
  try {
    entries = fs.readdirSync(cache);
  } catch {
    return undefined;
  }
  const dir = entries.find((e) => e.startsWith("chromium-") && !e.includes("headless"));
  if (!dir) {
    return undefined;
  }
  if (process.platform === "darwin") {
    const exe = path.join(
      cache,
      dir,
      "chrome-mac",
      "Chromium.app",
      "Contents",
      "MacOS",
      "Chromium",
    );
    return fs.existsSync(exe) ? exe : undefined;
  }
  if (process.platform === "linux") {
    const exe = path.join(cache, dir, "chrome-linux", "chrome");
    return fs.existsSync(exe) ? exe : undefined;
  }
  if (process.platform === "win32") {
    const exe = path.join(cache, dir, "chrome-win", "chrome.exe");
    return fs.existsSync(exe) ? exe : undefined;
  }
  return undefined;
}
