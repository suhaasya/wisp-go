import {
  workspaceTabs,
  type TableBrowseFilter,
  type TableBrowseQuery,
  type TableSortKey,
  type WorkspaceTableTab,
} from "../state";

export type { TableBrowseFilter, TableBrowseQuery, TableFilterCondition, TableSortKey } from "../state";

export const emptyBrowseFilter = (): TableBrowseFilter => ({
  logic: "and",
  conditions: [{ column: "", op: "=", value: "" }],
});

export const emptyBrowseQuery = (): TableBrowseQuery => ({
  rawMode: false,
  rawWhere: "",
  filter: emptyBrowseFilter(),
  sort: [],
});

export function browseQueryKey(q: TableBrowseQuery | undefined): string {
  if (!q) {
    return "";
  }
  return JSON.stringify(q);
}

export function appliedBrowseQuery(tab: WorkspaceTableTab): TableBrowseQuery {
  return tab.browseQuery ?? emptyBrowseQuery();
}

export function updateTableTabBrowseQuery(tabId: string, query: TableBrowseQuery): void {
  workspaceTabs.value = workspaceTabs.value.map((t) =>
    t.id === tabId ? { ...t, browseQuery: query } : t,
  );
}

export function formatSortHint(sort: TableSortKey[]): string {
  if (sort.length === 0) {
    return "";
  }
  return sort.map((s) => `${s.column} ${s.desc ? "desc" : "asc"}`).join(", ");
}
