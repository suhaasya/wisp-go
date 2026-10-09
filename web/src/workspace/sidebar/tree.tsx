import type { ComponentChild } from "preact";
import { useCallback, useEffect, useMemo, useRef, useState } from "preact/hooks";
import { cx } from "../../ui/cx";
import { VirtualList } from "../../ui/VirtualList";
import type { ObjectKind } from "../schema-client";
import { fetchObjects } from "../schema-client";
import { filterNames } from "./filter";
import { ObjectContextMenu, type ObjectMenuAction } from "./menu";

export type SchemaGroupId = "tables" | "views" | "functions" | "sequences" | "types";

export const SCHEMA_GROUPS: {
  id: SchemaGroupId;
  label: string;
  kinds: ObjectKind[];
}[] = [
  { id: "tables", label: "Tables", kinds: ["table"] },
  { id: "views", label: "Views", kinds: ["view", "matview"] },
  { id: "functions", label: "Functions", kinds: ["function", "procedure"] },
  { id: "sequences", label: "Sequences", kinds: ["sequence"] },
  { id: "types", label: "Types", kinds: ["enum"] },
];

export type GroupLoadState = Record<SchemaGroupId, "idle" | "loading" | "ready" | "error">;

export type SidebarRow =
  | { type: "group"; id: SchemaGroupId; label: string; count: number | null; expanded: boolean }
  | { type: "loading"; groupId: SchemaGroupId }
  | { type: "error"; groupId: SchemaGroupId; message: string }
  | { type: "object"; groupId: SchemaGroupId; name: string };

export function buildSidebarRows(input: {
  expanded: Record<SchemaGroupId, boolean>;
  namesByGroup: Record<SchemaGroupId, string[]>;
  load: GroupLoadState;
  filterQuery: string;
}): SidebarRow[] {
  const rows: SidebarRow[] = [];
  for (const group of SCHEMA_GROUPS) {
    const isExpanded = input.expanded[group.id];
    const load = input.load[group.id];
    const names = input.namesByGroup[group.id] ?? [];
    const count = load === "ready" ? names.length : null;
    rows.push({
      type: "group",
      id: group.id,
      label: group.label,
      count,
      expanded: isExpanded,
    });
    if (!isExpanded) {
      continue;
    }
    if (load === "loading") {
      rows.push({ type: "loading", groupId: group.id });
      continue;
    }
    if (load === "error") {
      rows.push({ type: "error", groupId: group.id, message: "Could not load" });
      continue;
    }
    if (load === "idle") {
      continue;
    }
    const filtered = filterNames(names, input.filterQuery);
    for (const name of filtered) {
      rows.push({ type: "object", groupId: group.id, name });
    }
  }
  return rows;
}

export type OpenTableRequest = {
  schema: string;
  table: string;
  mode: "data" | "structure";
};

export type SchemaTreeProps = {
  sessionId: string;
  schema: string;
  filterQuery: string;
  expanded: Record<SchemaGroupId, boolean>;
  onExpandedChange: (next: Record<SchemaGroupId, boolean>) => void;
  selectedObject: { schema: string; name: string } | null;
  onSelectObject: (schema: string, name: string) => void;
  onOpenTable: (req: OpenTableRequest) => void;
  readOnly?: boolean;
  onCatalogDDL?: (req: { action: "truncate" | "drop" | "rename"; schema: string; table: string }) => void;
  fetchObjectsImpl?: typeof fetchObjects;
};

const ROW_HEIGHT = 24;

function groupIcon(id: SchemaGroupId): string {
  switch (id) {
    case "views":
      return "V";
    case "functions":
      return "ƒ";
    case "sequences":
      return "#";
    case "types":
      return "T";
    default:
      return "▦";
  }
}

export function SchemaTree({
  sessionId,
  schema,
  filterQuery,
  expanded,
  onExpandedChange,
  selectedObject,
  onSelectObject,
  onOpenTable,
  readOnly = false,
  onCatalogDDL,
  fetchObjectsImpl = fetchObjects,
}: SchemaTreeProps) {
  const [namesByGroup, setNamesByGroup] = useState<Record<SchemaGroupId, string[]>>({
    tables: [],
    views: [],
    functions: [],
    sequences: [],
    types: [],
  });
  const [load, setLoad] = useState<GroupLoadState>({
    tables: "idle",
    views: "idle",
    functions: "idle",
    sequences: "idle",
    types: "idle",
  });
  const abortRef = useRef<AbortController | null>(null);
  const [menu, setMenu] = useState<{
    x: number;
    y: number;
    name: string;
  } | null>(null);
  const loadGroup = useCallback(
    async (groupId: SchemaGroupId, refresh = false) => {
      const group = SCHEMA_GROUPS.find((g) => g.id === groupId);
      if (!group || !schema) {
        return;
      }
      abortRef.current?.abort();
      const ac = new AbortController();
      abortRef.current = ac;
      setLoad((prev) => ({ ...prev, [groupId]: "loading" }));
      try {
        const chunks = await Promise.all(
          group.kinds.map((kind) =>
            fetchObjectsImpl(sessionId, schema, kind, { refresh, signal: ac.signal }),
          ),
        );
        const merged = [...new Set(chunks.flat())].sort((a, b) => a.localeCompare(b));
        setNamesByGroup((prev) => ({ ...prev, [groupId]: merged }));
        setLoad((prev) => ({ ...prev, [groupId]: "ready" }));
      } catch (err) {
        if (ac.signal.aborted) {
          return;
        }
        setLoad((prev) => ({ ...prev, [groupId]: "error" }));
      }
    },
    [fetchObjectsImpl, schema, sessionId],
  );

  useEffect(() => {
    return () => abortRef.current?.abort();
  }, []);

  useEffect(() => {
    setNamesByGroup({ tables: [], views: [], functions: [], sequences: [], types: [] });
    setLoad({ tables: "idle", views: "idle", functions: "idle", sequences: "idle", types: "idle" });
    for (const group of SCHEMA_GROUPS) {
      if (expanded[group.id]) {
        void loadGroup(group.id);
      }
    }
  }, [schema, sessionId]); // eslint-disable-line react-hooks/exhaustive-deps -- reset on schema change

  const toggleGroup = (id: SchemaGroupId) => {
    const next = !expanded[id];
    onExpandedChange({ ...expanded, [id]: next });
    if (next && load[id] === "idle") {
      void loadGroup(id);
    }
  };

  const rows = useMemo(
    () => buildSidebarRows({ expanded, namesByGroup, load, filterQuery }),
    [expanded, namesByGroup, load, filterQuery],
  );

  const handleMenuAction = (action: ObjectMenuAction, objectName: string) => {
    switch (action) {
      case "open-data":
        onOpenTable({ schema, table: objectName, mode: "data" });
        break;
      case "open-structure":
        onOpenTable({ schema, table: objectName, mode: "structure" });
        break;
      case "copy-name":
        void navigator.clipboard.writeText(objectName);
        break;
      case "copy-select":
        void navigator.clipboard.writeText(`SELECT * FROM ${quoteTable(schema, objectName)};`);
        break;
      case "refresh":
        for (const g of SCHEMA_GROUPS) {
          if (namesByGroup[g.id].includes(objectName)) {
            void loadGroup(g.id, true);
            break;
          }
        }
        break;
      case "truncate":
      case "drop":
      case "rename":
        onCatalogDDL?.({ action, schema, table: objectName });
        break;
    }
  };

  const renderRow = (row: SidebarRow, index: number): ComponentChild => {
    if (row.type === "group") {
      return (
        <button
          type="button"
          role="treeitem"
          aria-expanded={row.expanded}
          class="flex w-full items-center gap-1.5 px-1.5 py-1 text-12-5 font-medium text-ink-2"
          onClick={() => toggleGroup(row.id)}
        >
          <span class="w-3 text-ink-3">{row.expanded ? "▾" : "▸"}</span>
          <span>{row.label}</span>
          {row.count !== null ? <small class="ml-auto font-normal text-ink-3">{row.count}</small> : null}
        </button>
      );
    }
    if (row.type === "loading") {
      return <div class="px-6 py-1 text-12-5 text-ink-3">Loading…</div>;
    }
    if (row.type === "error") {
      return (
        <div class="flex items-center gap-2 px-4 py-1 text-12-5 text-env-prod">
          <span>{row.message}</span>
          <button
            type="button"
            class="text-accent underline"
            onClick={() => void loadGroup(row.groupId)}
          >
            Retry
          </button>
        </div>
      );
    }
    const selected =
      selectedObject?.schema === schema && selectedObject?.name === row.name;
    return (
      <button
        type="button"
        role="treeitem"
        data-object-row={row.name}
        class={cx(
          "flex w-full items-center gap-2 rounded-[5px] py-1 pl-5 pr-1.5 text-12-5 text-ink",
          selected ? "bg-accent-soft font-medium" : "hover:bg-raise",
        )}
        onClick={() => {
          onSelectObject(schema, row.name);
          onOpenTable({ schema, table: row.name, mode: "data" });
        }}
        onContextMenu={(e) => {
          e.preventDefault();
          setMenu({ x: e.clientX, y: e.clientY, name: row.name });
        }}
      >
        <span class="flex h-4 w-4 shrink-0 items-center justify-center rounded text-10 text-ink-3">
          {groupIcon(row.groupId)}
        </span>
        <span class="truncate">{row.name}</span>
      </button>
    );
  };

  const focusableObjectIndex = useRef(0);

  const onTreeKeyDown = (e: KeyboardEvent) => {
    const objectIndexes = rows
      .map((r, i) => (r.type === "object" ? i : -1))
      .filter((i) => i >= 0);
    if (objectIndexes.length === 0) {
      return;
    }
    if (e.key === "ArrowDown") {
      e.preventDefault();
      focusableObjectIndex.current = Math.min(
        focusableObjectIndex.current + 1,
        objectIndexes.length - 1,
      );
      const row = rows[objectIndexes[focusableObjectIndex.current]!];
      if (row?.type === "object") {
        onSelectObject(schema, row.name);
      }
    }
    if (e.key === "ArrowUp") {
      e.preventDefault();
      focusableObjectIndex.current = Math.max(focusableObjectIndex.current - 1, 0);
      const row = rows[objectIndexes[focusableObjectIndex.current]!];
      if (row?.type === "object") {
        onSelectObject(schema, row.name);
      }
    }
    if (e.key === "Enter") {
      e.preventDefault();
      const row = rows[objectIndexes[focusableObjectIndex.current]!];
      if (row?.type === "object") {
        onOpenTable({ schema, table: row.name, mode: "data" });
      }
    }
  };

  return (
    <div
      class="flex min-h-0 flex-1 flex-col outline-none"
      role="tree"
      aria-label="Schema objects"
      tabIndex={0}
      onKeyDown={onTreeKeyDown}
    >
      <VirtualList
        class="tree min-h-0 flex-1 px-1.5 py-1.5 text-12-5"
        items={rows}
        rowHeight={ROW_HEIGHT}
        render={(row, index) => renderRow(row, index)}
      />
      <ObjectContextMenu
        open={menu !== null}
        x={menu?.x ?? 0}
        y={menu?.y ?? 0}
        readOnly={readOnly}
        tablesOnly={menu ? namesByGroup.tables.includes(menu.name) : false}
        onClose={() => setMenu(null)}
        onAction={(action) => {
          if (menu) {
            handleMenuAction(action, menu.name);
          }
          setMenu(null);
        }}
      />
    </div>
  );
}

function quoteTable(schema: string, table: string): string {
  if (/^[a-z_][a-z0-9_]*$/i.test(schema) && /^[a-z_][a-z0-9_]*$/i.test(table)) {
    return `${schema}.${table}`;
  }
  return `"${schema.replace(/"/g, '""')}"."${table.replace(/"/g, '""')}"`;
}
