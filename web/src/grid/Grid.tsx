import { useEffect, useRef } from "preact/hooks";
import { useSignal } from "@preact/signals";
import { settings } from "../state";
import { createWindowBuffer } from "./buffer";
import { buildColumns, clampColumnWidth, type GridColumn } from "./columns";
import { copySelection, formatSelectionSync, type CopyFormat } from "./copy";
import { createGridHeader } from "./header";
import { createRowPool, type RowPool } from "./pool";
import { openCellEditor, revertCell, setCellDefault, setCellNull, type EditContext } from "./edit";
import { Menu } from "../ui/Menu";
import { fetchResultWindow, postResultChange } from "./results-client";
import {
  emptySelection,
  extendCellSelection,
  selectCell,
  toggleRowSelection,
  type SelectionState,
} from "./selection";

export type GridCellAccess = {
  getCell: (row: number, col: number) => unknown;
  isTruncated: (row: number, col: number) => boolean;
  invalidateWindows: () => void;
};
import type { TableSortKey } from "../state";
import { toggleSortDraft } from "../workspace/sort-draft";
import { DEFAULT_WINDOW_ROWS, type ColumnMeta, type RowTotal, type WindowPayload } from "./window";

export type GridProps = {
  sessionId: string;
  rid: string;
  columns: ColumnMeta[];
  total: RowTotal;
  schema: string;
  table: string;
  sortDraft?: TableSortKey[];
  onSortDraftChange?: (sort: TableSortKey[]) => void;
  initialScrollTop?: number;
  onScrollTopChange?: (top: number) => void;
  onSelectionChange?: (state: SelectionState) => void;
  onGridReady?: (access: GridCellAccess) => void;
  onToggleDetail?: () => void;
  fetchWindow?: (offset: number, signal: AbortSignal) => Promise<WindowPayload>;
  editable?: boolean;
  primaryKey?: string[];
  onStagedCount?: (count: number) => void;
  rowTone?: (row: number) => "blocked" | "long_running" | null;
};

function rowHeightFromSettings(): number {
  const rh = settings.value.row_height;
  return rh === "comfortable" ? 32 : 26;
}

export function Grid(props: GridProps) {
  const hostRef = useRef<HTMLDivElement>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const headerMountRef = useRef<HTMLDivElement>(null);
  const bodyMountRef = useRef<HTMLDivElement>(null);
  const spacerRef = useRef<HTMLDivElement>(null);

  const columns = useSignal<GridColumn[]>(buildColumns(props.columns));
  const selection = useSignal<SelectionState>(emptySelection());
  const statusRows = useSignal({ first: 0, last: 0 });
  const lastFetch = useSignal<string>("—");
  const cellMenu = useSignal<{ x: number; y: number; row: number; col: number } | null>(null);
  const menuIndex = useSignal(0);
  const totalRows = props.total.value;

  const headerApiRef = useRef<ReturnType<typeof createGridHeader> | null>(null);
  const bufferRef = useRef<ReturnType<typeof createWindowBuffer> | null>(null);
  const poolRef = useRef<RowPool | null>(null);
  const onToggleDetailRef = useRef(props.onToggleDetail);
  onToggleDetailRef.current = props.onToggleDetail;

  useEffect(() => {
    headerApiRef.current?.setSortDraft(props.sortDraft ?? []);
  }, [props.sortDraft]);

  useEffect(() => {
    const scrollEl = scrollRef.current;
    const headerMount = headerMountRef.current;
    const bodyMount = bodyMountRef.current;
    const spacerEl = spacerRef.current;
    if (!scrollEl || !headerMount || !bodyMount || !spacerEl) {
      return;
    }

    const rowHeight = rowHeightFromSettings();
    const tableName = `${props.schema}.${props.table}`;

    let pool: RowPool | null = null;
    let editorCleanup: (() => void) | null = null;

    const header = createGridHeader(headerMount, columns.value, {
      sortDraft: props.sortDraft ?? [],
      onSortClick(colId, shiftKey) {
        if (!props.onSortDraftChange) {
          return;
        }
        props.onSortDraftChange(toggleSortDraft(props.sortDraft ?? [], colId, shiftKey));
      },
      onResize(colId, width) {
        columns.value = columns.value.map((c) =>
          c.id === colId ? { ...c, width: clampColumnWidth(width) } : c,
        );
        header.setColumns(columns.value);
        pool?.setColumns(columns.value);
      },
      onReorder(fromId, toId) {
        const cols = [...columns.value];
        const from = cols.findIndex((c) => c.id === fromId);
        const to = cols.findIndex((c) => c.id === toId);
        if (from < 0 || to < 0) {
          return;
        }
        const [moved] = cols.splice(from, 1);
        cols.splice(to, 0, moved!);
        columns.value = cols;
        header.setColumns(columns.value);
        pool?.setColumns(columns.value);
      },
      onHide(colId) {
        columns.value = columns.value.map((c) => (c.id === colId ? { ...c, hidden: true } : c));
        header.setColumns(columns.value);
        pool?.setColumns(columns.value);
      },
    });

    bufferRef.current = createWindowBuffer({
      windowRows: DEFAULT_WINDOW_ROWS,
      maxWindows: 3,
      onUpdate: () => {
        pool?.resetViewportFetch();
        pool?.scheduleLayout();
      },
      fetch: (offset, signal) => {
        if (props.fetchWindow) {
          return props.fetchWindow(offset, signal);
        }
        return fetchResultWindow(props.sessionId, props.rid, offset, DEFAULT_WINDOW_ROWS, signal);
      },
    });
    const buffer = bufferRef.current;

    const editCtx = (): EditContext | null => {
      if (!props.editable || !props.primaryKey?.length) {
        return null;
      }
      return {
        sessionId: props.sessionId,
        rid: props.rid,
        primaryKey: props.primaryKey,
        columns: columns.value,
        getCell: (row, col) => buffer.cellValue(row, col),
        postChange: (body) => postResultChange(props.sessionId, props.rid, body),
        onStagedCount: (n) => props.onStagedCount?.(n),
        invalidateWindows: () => {
          buffer.invalidateAll();
          pool?.scheduleLayout();
        },
      };
    };

    const openEditor = (row: number, col: number, cellEl: HTMLDivElement) => {
      const ctx = editCtx();
      if (!ctx) {
        return;
      }
      editorCleanup?.();
      const onTab = () => {
        const nextCol = col + 1;
        if (nextCol >= columns.value.length) {
          return;
        }
        selection.value = selectCell(row, nextCol);
        props.onSelectionChange?.(selection.value);
        requestAnimationFrame(() => {
          const nextEl = scrollEl.querySelector(
            `[data-row="${row}"] [data-col-index="${nextCol}"]`,
          ) as HTMLDivElement | null;
          if (nextEl) {
            openEditor(row, nextCol, nextEl);
          }
        });
      };
      editorCleanup = openCellEditor(ctx, row, col, cellEl, undefined, onTab);
    };

    headerApiRef.current = header;

    poolRef.current = pool = createRowPool({
      scrollEl,
      headerEl: headerMount,
      bodyEl: bodyMount,
      spacerEl,
      rowHeight,
      totalRows,
      total: props.total,
      buffer,
      columns: columns.value,
      onHorizontalScroll: (left, w) => header.setHorizontalScroll(left, w),
      getSelection: () => selection.value,
      onCellClick(row, col, extend) {
        selection.value = extend
          ? extendCellSelection(selection.value, row, col)
          : selectCell(row, col);
        props.onSelectionChange?.(selection.value);
        pool?.scheduleLayout();
      },
      onRowClick(row, extend) {
        selection.value = toggleRowSelection(selection.value, row, extend);
        props.onSelectionChange?.(selection.value);
        pool?.scheduleLayout();
      },
      overlayKind: (row, col) => buffer.overlayKind(row, col),
      rowTone: props.rowTone,
      onCellActivate: (row, col, cellEl) => openEditor(row, col, cellEl),
    });

    props.onGridReady?.({
      getCell: (row, col) => buffer.cellValue(row, col),
      isTruncated: (row, col) => buffer.isTruncated(row, col),
      invalidateWindows: () => {
        buffer.invalidateAll();
        pool?.scheduleLayout();
      },
    });

    const onScroll = () => {
      props.onScrollTopChange?.(scrollEl.scrollTop);
      pool?.scheduleLayout();
      const headerH = headerMount.offsetHeight;
      const first = Math.floor(Math.max(0, scrollEl.scrollTop - headerH) / rowHeight);
      const visible = Math.ceil(scrollEl.clientHeight / rowHeight);
      statusRows.value = {
        first: totalRows === 0 ? 0 : Math.min(totalRows, first + 1),
        last: Math.min(totalRows, first + visible),
      };
      const t = buffer.lastFetchTime();
      if (t) {
        lastFetch.value = t.toLocaleTimeString();
      }
    };

    if (props.initialScrollTop && props.initialScrollTop > 0) {
      scrollEl.scrollTop = props.initialScrollTop;
    }

    scrollEl.addEventListener("scroll", onScroll, { passive: true });
    onScroll();

    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === " " && e.target === hostRef.current) {
        e.preventDefault();
        onToggleDetailRef.current?.();
        return;
      }
      if (e.key === "Enter" && props.editable && selection.value.kind === "cell") {
        const { row, col } = selection.value.focus;
        const cellEl = scrollEl.querySelector(
          `[data-row="${row}"] [data-col-index="${col}"]`,
        ) as HTMLDivElement | null;
        if (cellEl) {
          openEditor(row, col, cellEl);
          e.preventDefault();
        }
      }
      if ((e.metaKey || e.ctrlKey) && e.key === "c") {
        const getCell = (row: number, colIndex: number) => {
          const v = buffer.cellValue(row, colIndex);
          if (v === null || v === undefined) {
            return null;
          }
          return String(v);
        };
        void copySelection(selection.value, columns.value, buffer, "tsv", tableName, getCell);
        e.preventDefault();
      }
    };
    hostRef.current?.addEventListener("keydown", onKeyDown);

    const onContextMenu = (e: MouseEvent) => {
      if (!props.editable) {
        return;
      }
      const cell = (e.target as HTMLElement).closest("[data-col-index]");
      if (!cell) {
        return;
      }
      const rowEl = cell.closest("[data-row]");
      if (!rowEl) {
        return;
      }
      const row = Number(rowEl.getAttribute("data-row"));
      const col = Number(cell.getAttribute("data-col-index"));
      if (Number.isNaN(row) || Number.isNaN(col)) {
        return;
      }
      e.preventDefault();
      cellMenu.value = { x: e.clientX, y: e.clientY, row, col };
      menuIndex.value = 0;
    };
    hostRef.current?.addEventListener("contextmenu", onContextMenu);

    return () => {
      scrollEl.removeEventListener("scroll", onScroll);
      hostRef.current?.removeEventListener("keydown", onKeyDown);
      hostRef.current?.removeEventListener("contextmenu", onContextMenu);
      editorCleanup?.();
      poolRef.current = null;
      pool?.destroy();
      header.destroy();
      headerApiRef.current = null;
      buffer.destroy();
    };
  }, [
    props.sessionId,
    props.rid,
    props.schema,
    props.table,
    props.fetchWindow,
    props.onSortDraftChange,
    props.onSelectionChange,
    props.onGridReady,
    props.editable,
    props.primaryKey,
    props.onStagedCount,
  ]);

  const totalLabel =
    props.total.kind === "estimate"
      ? `~${props.total.value.toLocaleString()}`
      : props.total.value.toLocaleString();

  const menu = cellMenu.value;
  const menuItems = [
    { id: "null", label: "Set NULL" },
    { id: "default", label: "Set DEFAULT" },
    { id: "revert", label: "Revert" },
  ];

  return (
    <div ref={hostRef} class="flex min-h-0 min-w-0 flex-1 flex-col outline-none" tabIndex={0}>
      <Menu
        open={menu !== null}
        x={menu?.x ?? 0}
        y={menu?.y ?? 0}
        items={menuItems}
        activeIndex={menuIndex.value}
        onActiveIndexChange={(i) => {
          menuIndex.value = i;
        }}
        onClose={() => {
          cellMenu.value = null;
        }}
        onSelect={(id) => {
          const m = cellMenu.value;
          cellMenu.value = null;
          if (!m) {
            return;
          }
          const ctx = {
            sessionId: props.sessionId,
            rid: props.rid,
            primaryKey: props.primaryKey ?? [],
            columns: columns.value,
            getCell: (row, col) => bufferRef.current?.cellValue(row, col),
            postChange: (body) => postResultChange(props.sessionId, props.rid, body),
            onStagedCount: (n) => props.onStagedCount?.(n),
            invalidateWindows: () => {
              bufferRef.current?.invalidateAll();
              poolRef.current?.scheduleLayout();
            },
          } satisfies EditContext;
          if (id === "null") {
            void setCellNull(ctx, m.row, m.col);
          } else if (id === "default") {
            void setCellDefault(ctx, m.row, m.col);
          } else {
            void revertCell(ctx, m.row, m.col);
          }
        }}
      />
      <div
        ref={scrollRef}
        data-testid="wisp-grid-scroll"
        class="gridwrap min-h-0 min-w-0 flex-1 overflow-auto bg-panel"
        role="grid"
        aria-label={props.table ? `${props.schema}.${props.table} data` : "Query results"}
        aria-rowcount={totalRows}
        aria-colcount={columns.value.length}
      >
        <div class="relative min-h-full min-w-min">
          <div ref={headerMountRef} class="min-w-min" />
          <div ref={spacerRef} aria-hidden />
          <div ref={bodyMountRef} class="absolute left-0 top-0 min-w-min" />
        </div>
      </div>
      <div class="flex shrink-0 items-center justify-between border-t border-line bg-sidebar px-3 py-1 text-11-5 text-ink-2">
        <span>
          Rows {statusRows.value.first}–{statusRows.value.last} of {totalLabel}
        </span>
        <span>Last fetch {lastFetch.value}</span>
      </div>
    </div>
  );
}

export function gridSelectionToClipboard(
  state: SelectionState,
  columns: GridColumn[],
  getCell: (row: number, col: number) => string | null,
  format: CopyFormat,
  tableName: string,
): string {
  return formatSelectionSync(state, columns, getCell, format, tableName);
}
