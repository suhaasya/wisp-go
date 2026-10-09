import { request } from "./api";
import { settings as settingsSignal, type ThemePreference } from "./state";
import { setThemePreference } from "./theme";

export type WispSettings = {
  version?: number;
  theme?: ThemePreference;
  density?: "compact" | "comfortable";
  window_rows?: number;
  windows_in_memory?: number;
  result_cap_mb?: number;
  keep_alive_sec?: number;
  idle_exit_min?: number;
  go_mem_limit_mb?: number;
  confirm_writes_on_production?: boolean;
  keybindings?: Record<string, string>;
  telemetry?: boolean;
  [key: string]: unknown;
};

let cached: WispSettings = {};

export function getSettings(): WispSettings {
  return cached;
}

export function applySettings(doc: WispSettings): void {
  const density = doc.density === "comfortable" ? "comfortable" : "compact";
  cached = {
    ...doc,
    density,
    row_height: density === "comfortable" ? "comfortable" : "compact",
  };
  settingsSignal.value = cached;
  const theme = doc.theme;
  if (theme === "light" || theme === "dark" || theme === "system") {
    setThemePreference(theme);
  }
  document.documentElement.setAttribute("data-density", density);
}

export async function loadSettings(): Promise<WispSettings> {
  const doc = await request<WispSettings>("GET", "/api/v1/settings");
  applySettings(doc);
  return doc;
}

export async function saveSettings(doc: WispSettings): Promise<WispSettings> {
  const out = await request<WispSettings>("PUT", "/api/v1/settings", doc);
  applySettings(out);
  return out;
}
