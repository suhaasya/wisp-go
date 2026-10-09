import { themePreference, type ThemePreference } from "./state";

export type { ThemePreference };
export type ResolvedTheme = "light" | "dark";

export { themePreference };

const storageKey = "wisp-theme";

let systemListenerAttached = false;
let mediaQuery: MediaQueryList | null = null;

function onSystemSchemeChange(): void {
  if (themePreference.value === "system") {
    applyResolvedTheme(resolveTheme("system"));
  }
}

function readStoredPreference(): ThemePreference {
  const raw = localStorage.getItem(storageKey);
  if (raw === "light" || raw === "dark" || raw === "system") {
    return raw;
  }
  return "system";
}

function systemPrefersDark(): boolean {
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

export function resolveTheme(preference: ThemePreference): ResolvedTheme {
  if (preference === "system") {
    return systemPrefersDark() ? "dark" : "light";
  }
  return preference;
}

export function applyResolvedTheme(resolved: ResolvedTheme): void {
  document.documentElement.setAttribute("data-theme", resolved);
}

function attachSystemListener(): void {
  if (systemListenerAttached && mediaQuery) {
    return;
  }
  mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
  mediaQuery.addEventListener("change", onSystemSchemeChange);
  systemListenerAttached = true;
}

export function resetThemeForTests(): void {
  if (mediaQuery && systemListenerAttached) {
    mediaQuery.removeEventListener("change", onSystemSchemeChange);
  }
  systemListenerAttached = false;
  mediaQuery = null;
}

export function setThemePreference(next: ThemePreference): void {
  themePreference.value = next;
  localStorage.setItem(storageKey, next);
  applyResolvedTheme(resolveTheme(next));
  attachSystemListener();
}

export function initTheme(): void {
  themePreference.value = readStoredPreference();
  applyResolvedTheme(resolveTheme(themePreference.value));
  attachSystemListener();
}
