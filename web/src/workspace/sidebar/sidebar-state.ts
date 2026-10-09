import type { SchemaGroupId } from "./tree";

export type SidebarUiState = {
  width: number;
  expanded: Record<SchemaGroupId, boolean>;
  selectedDatabase: string;
  selectedSchema: string;
};

const STORAGE_VERSION = 1;
const MIN_WIDTH = 180;
const MAX_WIDTH = 420;
const DEFAULT_WIDTH = 240;

export function sidebarStorageKey(profileId: string): string {
  return `wisp.sidebar.v${STORAGE_VERSION}.${profileId}`;
}

export function clampSidebarWidth(px: number): number {
  return Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, px));
}

export function defaultSidebarUiState(): SidebarUiState {
  return {
    width: DEFAULT_WIDTH,
    expanded: {
      tables: true,
      views: false,
      functions: false,
      sequences: false,
      types: false,
    },
    selectedDatabase: "",
    selectedSchema: "",
  };
}

export function loadSidebarUiState(profileId: string): SidebarUiState {
  if (!profileId || typeof localStorage === "undefined") {
    return defaultSidebarUiState();
  }
  try {
    const raw = localStorage.getItem(sidebarStorageKey(profileId));
    if (!raw) {
      return defaultSidebarUiState();
    }
    const parsed = JSON.parse(raw) as Partial<SidebarUiState>;
    const base = defaultSidebarUiState();
    return {
      width: clampSidebarWidth(typeof parsed.width === "number" ? parsed.width : base.width),
      expanded: { ...base.expanded, ...parsed.expanded },
      selectedDatabase: parsed.selectedDatabase ?? base.selectedDatabase,
      selectedSchema: parsed.selectedSchema ?? base.selectedSchema,
    };
  } catch {
    return defaultSidebarUiState();
  }
}

export function saveSidebarUiState(profileId: string, state: SidebarUiState): void {
  if (!profileId || typeof localStorage === "undefined") {
    return;
  }
  try {
    localStorage.setItem(
      sidebarStorageKey(profileId),
      JSON.stringify({
        ...state,
        width: clampSidebarWidth(state.width),
      }),
    );
  } catch {
    // quota or private mode
  }
}
