import { signal } from "@preact/signals";

export type ToastTone = "default" | "success" | "error";

export type ToastItem = {
  id: string;
  message: string;
  tone?: ToastTone;
  durationMs?: number;
};

export type ThemePreference = "light" | "dark" | "system";

export type SessionState = "disconnected" | "connecting" | "connected" | "reconnecting" | "error";

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
/** Staged edit count for the active table result (LW-029). */
export const stagedChangeCount = signal(0);
export const shellConnection = signal<ShellConnection | null>(null);
export const sessionEngine = signal<string>("postgres");
export const workspaceProfileId = signal<string | null>(null);
export const paletteOpen = signal(false);
export const shortcutHelpOpen = signal(false);

export function openCommandPalette(): void {
  paletteOpen.value = true;
}

export type TableFilterCondition = {
  column: string;
  op: string;
  value: string;
  value2?: string;
};

export type TableBrowseFilter = {
  logic: "and" | "or";
  conditions: TableFilterCondition[];
};

export type TableSortKey = {
  column: string;
  desc: boolean;
};

export type TableBrowseQuery = {
  rawMode: boolean;
  rawWhere: string;
  filter: TableBrowseFilter;
  sort: TableSortKey[];
};

export type WorkspaceTab =
  | {
      id: string;
      kind: "table_data";
      schema: string;
      table: string;
      pinned?: boolean;
      browseQuery?: TableBrowseQuery;
      scrollTop?: number;
      selection?: { anchor_row: number; anchor_col: number };
      /** When true, table result is not fetched until the user activates the tab. */
      pendingResultLoad?: boolean;
    }
  | {
      id: string;
      kind: "table_structure";
      schema: string;
      table: string;
      pinned?: boolean;
    }
  | {
      id: string;
      kind: "query";
      title?: string;
      queryText?: string;
      pinned?: boolean;
    }
  | {
      id: string;
      kind: "server_tools";
      pinned?: boolean;
    };

export const workspaceTabs = signal<WorkspaceTab[]>([]);
export const workspaceHydratedFromServer = signal(false);
export const activeWorkspaceTabId = signal<string | null>(null);
export const csrfToken = signal("");
