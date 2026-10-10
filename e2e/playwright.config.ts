import { defineConfig, devices } from "@playwright/test";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { chromiumLaunchOptions } from "./launch-options.mjs";

const e2eDir = path.dirname(fileURLToPath(import.meta.url));
/** Set WISP_BASE to your running instance (e.g. http://127.0.0.1:8787). */
const baseURL = process.env.WISP_BASE ?? "http://127.0.0.1:8787";
const chromiumLaunch = { launchOptions: chromiumLaunchOptions() };

export default defineConfig({
  testDir: "./tests",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: "list",
  use: {
    baseURL,
    storageState: path.join(e2eDir, ".auth", "storage.json"),
    trace: "on-first-retry",
  },
  projects: [
    { name: "chromium", testIgnore: /matrix\//, use: { ...devices["Desktop Chrome"], ...chromiumLaunch } },
    { name: "firefox", testIgnore: /matrix\//, use: { ...devices["Desktop Firefox"] } },
    { name: "webkit", testIgnore: /matrix\//, use: { ...devices["Desktop Safari"] } },
    {
      name: "matrix",
      testMatch: /matrix\/.*\.spec\.ts/,
      timeout: 180_000,
      use: { ...devices["Desktop Chrome"], ...chromiumLaunch },
    },
  ],
  globalSetup: path.join(e2eDir, "global-setup.mjs"),
  globalTeardown: path.join(e2eDir, "global-teardown.mjs"),
});
