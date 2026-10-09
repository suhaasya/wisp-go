import {
  activeWorkspaceTabId,
  workspaceTabs,
  type TableBrowseQuery,
  type WorkspaceTab,
} from "../state";
import { browseQueryKey } from "./browse-query";
import type { OpenTableRequest } from "./sidebar/tree";

export function tabTitle(tab: WorkspaceTab): string {
  switch (tab.kind) {
    case "table_data":
      return tab.table;
    case "table_structure":
      return `${tab.table} (structure)`;
    case "query":
      return tab.title ?? "Query";
    case "server_tools":
      return "Server tools";
  }
}

export function newQueryTab(): WorkspaceTab {
  const n = workspaceTabs.value.filter((t) => t.kind === "query").length + 1;
  return {
    id: `query/${Date.now()}`,
    kind: "query",
    title: `Query ${n}`,
    queryText: "",
    pinned: false,
  };
}

export function openFilteredTableTab(
  schema: string,
  table: string,
  browseQuery: TableBrowseQuery,
): void {
  const id = `${schema}/${table}/data/filtered/${browseQueryKey(browseQuery)}`;
  const tab: WorkspaceTab = {
    id,
    kind: "table_data",
    schema,
    table,
    pinned: false,
    browseQuery,
    pendingResultLoad: false,
  };
  workspaceTabs.value = [...workspaceTabs.value.filter((t) => t.id !== id), tab];
  activateWorkspaceTab(id);
}

export function openWorkspaceTableTab(req: OpenTableRequest): void {
  const kind = req.mode === "structure" ? "table_structure" : "table_data";
  const id = `${req.schema}/${req.table}/${req.mode}`;
  const existing = workspaceTabs.value.find((t) => t.id === id);
  if (!existing) {
    const tab: WorkspaceTab =
      kind === "table_structure"
        ? { id, kind, schema: req.schema, table: req.table, pinned: false }
        : {
            id,
            kind: "table_data",
            schema: req.schema,
            table: req.table,
            pinned: false,
            pendingResultLoad: false,
          };
    workspaceTabs.value = [...workspaceTabs.value, tab];
  }
  activateWorkspaceTab(id);
}

export function activateWorkspaceTab(id: string): void {
  workspaceTabs.value = workspaceTabs.value.map((t) =>
    t.id === id && t.kind === "table_data" ? { ...t, pendingResultLoad: false } : t,
  );
  activeWorkspaceTabId.value = id;
}

export function closeWorkspaceTab(id: string): void {
  const tabs = workspaceTabs.value.filter((t) => t.id !== id);
  workspaceTabs.value = tabs;
  if (activeWorkspaceTabId.value === id) {
    activeWorkspaceTabId.value = tabs.length ? tabs[tabs.length - 1]!.id : null;
  }
}

export function reorderWorkspaceTabs(fromIndex: number, toIndex: number): void {
  if (fromIndex === toIndex || fromIndex < 0 || toIndex < 0) {
    return;
  }
  const tabs = [...workspaceTabs.value];
  if (fromIndex >= tabs.length || toIndex >= tabs.length) {
    return;
  }
  const [moved] = tabs.splice(fromIndex, 1);
  tabs.splice(toIndex, 0, moved!);
  workspaceTabs.value = tabs;
}

export function togglePinTab(id: string): void {
  workspaceTabs.value = workspaceTabs.value.map((t) =>
    t.id === id ? { ...t, pinned: !t.pinned } : t,
  );
  sortTabsPinnedFirst();
}

function sortTabsPinnedFirst(): void {
  const pinned = workspaceTabs.value.filter((t) => t.pinned);
  const rest = workspaceTabs.value.filter((t) => !t.pinned);
  workspaceTabs.value = [...pinned, ...rest];
}

export function activateTabAtIndex(index: number): void {
  const tab = workspaceTabs.value[index];
  if (tab) {
    activateWorkspaceTab(tab.id);
  }
}

export function closeActiveTab(): void {
  const id = activeWorkspaceTabId.value;
  if (id) {
    closeWorkspaceTab(id);
  }
}

export type TableDataTab = Extract<WorkspaceTab, { kind: "table_data" }>;

export function isTableDataTab(tab: WorkspaceTab): tab is TableDataTab {
  return tab.kind === "table_data";
}

export function updateTableDataTabState(
  id: string,
  patch: Partial<Pick<TableDataTab, "browseQuery" | "scrollTop" | "selection">>,
): void {
  workspaceTabs.value = workspaceTabs.value.map((t) =>
    t.id === id && t.kind === "table_data" ? { ...t, ...patch } : t,
  );
}

export { updateTableTabBrowseQuery } from "./browse-query";
