import { signal } from "@preact/signals";

export type ThemePreference = "light" | "dark" | "system";
export type ResolvedTheme = "light" | "dark";

const storageKey = "wisp-theme";

export const themePreference = signal<ThemePreference>("system");

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

/** Test-only reset for module listener state. */
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
