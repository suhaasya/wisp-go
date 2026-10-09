import { chromiumExecutablePath } from "./chromium-executable.mjs";

/** Prefer system Chrome/Edge on desktop OSes; bundled Chromium in CI/Linux. */
export function chromiumLaunchOptions() {
  if (process.env.PW_CHROMIUM_CHANNEL) {
    return { channel: process.env.PW_CHROMIUM_CHANNEL };
  }
  if (process.platform === "darwin") {
    return { channel: "chrome" };
  }
  if (process.platform === "win32") {
    return { channel: "msedge" };
  }
  const exe = chromiumExecutablePath();
  if (exe) {
    return { executablePath: exe, args: ["--headless=new"] };
  }
  return {};
}
