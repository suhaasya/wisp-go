import { spawn } from "node:child_process";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { chromium } from "@playwright/test";
import { chromiumLaunchOptions } from "./launch-options.mjs";

const e2eDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.join(e2eDir, "..");
const port = Number(process.env.WISP_E2E_PORT ?? "19287");
const wispBin = path.join(repoRoot, "bin", "wisp");
const authDir = path.join(e2eDir, ".auth");
const stateFile = path.join(authDir, "storage.json");
const pidFile = path.join(repoRoot, "bench-results", "e2e-runtime", "wisp-e2e.pid");

function defaultHandoffSecretPath() {
  const home = os.homedir();
  if (process.platform === "darwin") {
    return path.join(home, "Library", "Caches", "wisp", "runtime", "handoff.secret");
  }
  if (process.platform === "win32") {
    const local = process.env.LOCALAPPDATA ?? path.join(home, "AppData", "Local");
    return path.join(local, "wisp", "runtime", "handoff.secret");
  }
  return path.join(home, ".cache", "wisp", "runtime", "handoff.secret");
}

function defaultInstanceJSONPath() {
  const home = os.homedir();
  if (process.platform === "darwin") {
    return path.join(home, "Library", "Caches", "wisp", "runtime", "instance.json");
  }
  if (process.platform === "win32") {
    const local = process.env.LOCALAPPDATA ?? path.join(home, "AppData", "Local");
    return path.join(local, "wisp", "runtime", "instance.json");
  }
  return path.join(home, ".cache", "wisp", "runtime", "instance.json");
}

async function readRunningBaseURL() {
  try {
    const raw = await fs.readFile(defaultInstanceJSONPath(), "utf8");
    const rec = JSON.parse(raw);
    if (rec?.port > 0) {
      return `http://127.0.0.1:${rec.port}`;
    }
  } catch {
    // no instance
  }
  return null;
}

async function fetchLaunchURL(baseURL) {
  const secretPath = defaultHandoffSecretPath();
  const secret = (await fs.readFile(secretPath, "utf8")).trim();
  const handoff = await fetch(`${baseURL}/internal/handoff`, {
    method: "POST",
    headers: { Authorization: `Bearer ${secret}` },
  });
  if (!handoff.ok) {
    throw new Error(`handoff failed: ${handoff.status}`);
  }
  const { url } = await handoff.json();
  return url;
}

async function mintStorageState(launchURL, baseURL) {
  await fs.mkdir(authDir, { recursive: true });
  const browser = await chromium.launch(chromiumLaunchOptions());
  const context = await browser.newContext();
  const page = await context.newPage();
  await page.goto(launchURL, { waitUntil: "load", timeout: 60_000 });
  await page.waitForURL(
    (url) => url.origin === new URL(baseURL).origin && !url.pathname.startsWith("/auth/"),
    { timeout: 60_000 },
  );
  await context.storageState({ path: stateFile });
  await browser.close();
}

function spawnWisp() {
  return spawn(wispBin, ["--no-browser", "--port", String(port), "--idle-exit", "24h"], {
    cwd: repoRoot,
    stdio: ["ignore", "pipe", "pipe"],
  });
}

function waitForStartup(proc, timeoutMs = 30_000) {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error("wisp startup timeout")), timeoutMs);
    const onData = (chunk) => {
      const text = chunk.toString();
      const m = text.match(/Wisp running at http:\/\/127\.0\.0\.1:(\d+)\//);
      if (m) {
        clearTimeout(timer);
        proc.stdout?.off("data", onData);
        resolve(`http://127.0.0.1:${m[1]}`);
      }
    };
    proc.stdout?.on("data", onData);
    proc.stderr?.on("data", (chunk) => process.stderr.write(chunk));
    proc.on("exit", (code) => {
      if (code !== 0 && code !== null) {
        clearTimeout(timer);
        reject(new Error(`wisp exited with code ${code}`));
      }
    });
  });
}

export default async function globalSetup() {
  if (process.env.WISP_LAUNCH_URL && process.env.WISP_BASE) {
    process.env.WISP_BASE = process.env.WISP_BASE;
    await mintStorageState(process.env.WISP_LAUNCH_URL, process.env.WISP_BASE);
    return;
  }

  const forceFresh = process.env.WISP_E2E_FRESH === "1";
  let baseURL = forceFresh ? null : await readRunningBaseURL();
  let spawned = null;

  if (!baseURL) {
    try {
      await fs.access(wispBin);
    } catch {
      throw new Error("bin/wisp missing; run make build");
    }
    spawned = spawnWisp();
    try {
      baseURL = await waitForStartup(spawned);
      await fs.mkdir(path.dirname(pidFile), { recursive: true });
      await fs.writeFile(pidFile, String(spawned.pid));
    } catch (err) {
      spawned.kill();
      baseURL = await readRunningBaseURL();
      if (!baseURL) {
        throw err;
      }
    }
  }

  const launchURL = await fetchLaunchURL(baseURL);
  process.env.WISP_BASE = baseURL;
  process.env.WISP_LAUNCH_URL = launchURL;
  await mintStorageState(launchURL, baseURL);
}
