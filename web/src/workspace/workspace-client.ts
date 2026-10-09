import { request } from "../api";
import type { WorkspaceTab } from "../state";
import { workspaceTabs, activeWorkspaceTabId } from "../state";
import type { TableBrowseQuery } from "../state";

export type WorkspaceSnapshot = {
  version: number;
  active_tab_id?: string;
  tabs: WorkspaceTabDTO[];
};

type WorkspaceTabDTO = {
  id: string;
  kind: string;
  schema?: string;
  table?: string;
  pinned?: boolean;
  query_text?: string;
  scroll_top?: number;
  browse?: {
    raw_mode?: boolean;
    raw_where?: string;
    filter?: TableBrowseQuery["filter"];
    sort?: TableBrowseQuery["sort"];
  };
  selection?: { anchor_row: number; anchor_col: number };
};

let saveTimer: number | undefined;

export async function fetchWorkspace(profileId: string): Promise<WorkspaceSnapshot> {
  return request<WorkspaceSnapshot>(
    "GET",
    `/api/v1/workspace/${encodeURIComponent(profileId)}`,
  );
}

export function scheduleWorkspaceSave(profileId: string): void {
  if (!profileId) {
    return;
  }
  if (saveTimer !== undefined) {
    window.clearTimeout(saveTimer);
  }
  saveTimer = window.setTimeout(() => {
    saveTimer = undefined;
    void persistWorkspace(profileId).catch(() => {});
  }, 1000);
}

export async function persistWorkspace(profileId: string): Promise<void> {
  const body = snapshotFromSignals();
  await request("PUT", `/api/v1/workspace/${encodeURIComponent(profileId)}`, body);
}

function snapshotFromSignals(): WorkspaceSnapshot {
  return {
    version: 1,
    active_tab_id: activeWorkspaceTabId.value ?? undefined,
    tabs: workspaceTabs.value.map(tabToDTO),
  };
}

function tabToDTO(tab: WorkspaceTab): WorkspaceTabDTO {
  const base: WorkspaceTabDTO = { id: tab.id, kind: tab.kind, pinned: tab.pinned };
  switch (tab.kind) {
    case "table_data":
    case "table_structure":
      base.schema = tab.schema;
      base.table = tab.table;
      if (tab.kind === "table_data") {
        if (tab.scrollTop) {
          base.scroll_top = tab.scrollTop;
        }
        if (tab.selection) {
          base.selection = tab.selection;
        }
        if (tab.browseQuery) {
          base.browse = {
            raw_mode: tab.browseQuery.rawMode,
            raw_where: tab.browseQuery.rawWhere,
            filter: tab.browseQuery.filter,
            sort: tab.browseQuery.sort,
          };
        }
      }
      break;
    case "query":
      base.query_text = tab.queryText;
      break;
    default:
      break;
  }
  return base;
}

export function applyWorkspaceSnapshot(snap: WorkspaceSnapshot): void {
  workspaceTabs.value = snap.tabs.map(dtoToTab);
  activeWorkspaceTabId.value = snap.active_tab_id ?? workspaceTabs.value[0]?.id ?? null;
}

function dtoToTab(d: WorkspaceTabDTO): WorkspaceTab {
  switch (d.kind) {
    case "table_data":
      return {
        id: d.id,
        kind: "table_data",
        schema: d.schema ?? "",
        table: d.table ?? "",
        pinned: d.pinned,
        scrollTop: d.scroll_top,
        selection: d.selection,
        browseQuery: d.browse
          ? {
              rawMode: d.browse.raw_mode ?? false,
              rawWhere: d.browse.raw_where ?? "",
              filter: d.browse.filter ?? { logic: "and", conditions: [] },
              sort: d.browse.sort ?? [],
            }
          : undefined,
        pendingResultLoad: true,
      };
    case "table_structure":
      return {
        id: d.id,
        kind: "table_structure",
        schema: d.schema ?? "",
        table: d.table ?? "",
        pinned: d.pinned,
      };
    case "query":
      return {
        id: d.id,
        kind: "query",
        title: d.id.startsWith("query/") ? "Query" : d.id,
        queryText: d.query_text ?? "",
        pinned: d.pinned,
      };
    case "server_tools":
      return { id: d.id, kind: "server_tools", pinned: d.pinned };
    default:
      return {
        id: d.id,
        kind: "query",
        title: d.id,
        queryText: "",
      };
  }
}
