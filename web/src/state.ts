import { signal } from "@preact/signals";

export type ToastTone = "default" | "success" | "error";

export type ToastItem = {
  id: string;
  message: string;
  tone?: ToastTone;
  durationMs?: number;
};

export type ThemePreference = "light" | "dark" | "system";

export type SessionState = "disconnected" | "connecting" | "connected" | "error";

export type ShellConnection = {
  name: string;
  subtitle: string;
  env: "local" | "dev" | "staging" | "prod" | null;
};

export const settings = signal<Record<string, unknown>>({});
export const themePreference = signal<ThemePreference>("system");
export const currentSession = signal<string | null>(null);
export const sessionState = signal<SessionState>("disconnected");
export const toasts = signal<ToastItem[]>([]);
export const goMemoryMB = signal<number | null>(null);
export const tabMemoryMB = signal<number | null>(null);
export const sessionEnded = signal(false);
export const wispStopped = signal(false);
export const serverVersion = signal("—");
export const latencyMs = signal<number | null>(null);
export const readOnly = signal(false);
export const shellConnection = signal<ShellConnection | null>(null);
export const csrfToken = signal("");
