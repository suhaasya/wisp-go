import type { WindowBuffer } from "./buffer";
import {
  columnWindow,
  isEnumType,
  isNumericType,
  rowNumberWidth,
  totalBodyWidth,
  type GridColumn,
} from "./columns";
import { cellInSelection, rowIsSelected, type SelectionState } from "./selection";
import { DEFAULT_WINDOW_ROWS, windowOffsetForRow, type RowTotal } from "./window";

export type RowPoolOptions = {
  scrollEl: HTMLElement;
  headerEl: HTMLElement;
  bodyEl: HTMLElement;
  spacerEl: HTMLElement;
  rowHeight: number;
  totalRows: number;
  total: RowTotal;
  buffer: WindowBuffer;
  columns: GridColumn[];
  getSelection: () => SelectionState;
  onCellClick: (row: number, col: number, extend: boolean) => void;
  onRowClick: (row: number, extend: boolean) => void;
  onHorizontalScroll?: (scrollLeft: number, viewportWidth: number) => void;
};

export type RowPool = {
  setColumns: (columns: GridColumn[]) => void;
  setTotalRows: (n: number, total: RowTotal) => void;
  scheduleLayout: () => void;
  layoutNow: () => void;
  rowElementCount: () => number;
  destroy: () => void;
};

const POOL_EXTRA = 10;

export function createRowPool(opts: RowPoolOptions): RowPool {
  let columns = opts.columns;
  let totalRows = opts.totalRows;
  let total = opts.total;
  let raf = 0;

  const rows: HTMLDivElement[] = [];
  let poolSize = 0;

  function ensurePoolSize(viewportRows: number): void {
    const need = Math.ceil(viewportRows) + POOL_EXTRA;
    if (need <= poolSize) {
      return;
    }
    for (let i = poolSize; i < need; i++) {
      const row = document.createElement("div");
      row.className = "absolute left-0 flex flex-nowrap border-b border-line-2";
      row.style.height = `${opts.rowHeight}px`;
      row.dataset.poolIndex = String(i);
      opts.bodyEl.appendChild(row);
      rows.push(row);
    }
    poolSize = need;
  }

  function truncatedAt(buffer: WindowBuffer, row: number, col: number): boolean {
    const w = buffer.getWindow(windowOffsetForRow(row, DEFAULT_WINDOW_ROWS));
    if (!w?.truncated) {
      return false;
    }
    const idx = row - w.offset;
    return w.truncated[col]?.includes(idx) ?? false;
  }

  function renderCell(
    cell: HTMLDivElement,
    row: number,
    colIndex: number,
    col: GridColumn,
  ): void {
    cell.replaceChildren();
    cell.className =
      "shrink-0 overflow-hidden text-ellipsis whitespace-nowrap border-r border-line-2 px-2.5 font-mono text-12-5 text-ink";
    cell.style.width = `${col.width}px`;
    cell.style.minWidth = `${col.width}px`;
    if (isNumericType(col.meta.type)) {
      cell.classList.add("text-right");
    }

    if (!opts.buffer.isRowLoaded(row)) {
      const bar = document.createElement("div");
      bar.className = "mx-2 mt-[11px] h-1 rounded bg-line";
      cell.appendChild(bar);
      return;
    }

    const raw = opts.buffer.cellValue(row, colIndex);
    if (raw === null) {
      cell.classList.add("italic", "text-null-text");
      cell.textContent = "NULL";
      return;
    }
    if (raw === undefined) {
      const bar = document.createElement("div");
      bar.className = "mx-2 mt-[11px] h-1 rounded bg-line";
      cell.appendChild(bar);
      return;
    }

    const text = String(raw);
    if (isEnumType(col.meta.type)) {
      const pill = document.createElement("span");
      pill.className =
        "inline-block rounded-full border border-line bg-raise px-1.5 py-0.5 font-sans text-11-5 font-medium text-ink-2";
      pill.textContent = text;
      cell.appendChild(pill);
      return;
    }

    cell.textContent = text;
    if (truncatedAt(opts.buffer, row, colIndex)) {
      const icon = document.createElement("span");
      icon.className = "ml-1 text-ink-3";
      icon.textContent = "…";
      icon.title = "Truncated";
      cell.appendChild(icon);
    }
  }

  function layout(): void {
    const scrollTop = opts.scrollEl.scrollTop;
    const scrollLeft = opts.scrollEl.scrollLeft;
    const viewH = opts.scrollEl.clientHeight;
    const viewW = opts.scrollEl.clientWidth;
    const headerH = opts.headerEl.offsetHeight;
    const bodyScroll = Math.max(0, scrollTop - headerH);
    const firstRow = Math.floor(bodyScroll / opts.rowHeight);
    const visibleRows = Math.ceil(viewH / opts.rowHeight) + 1;
    ensurePoolSize(visibleRows);

    opts.spacerEl.style.height = `${totalRows * opts.rowHeight}px`;
    opts.bodyEl.style.width = `${totalBodyWidth(columns)}px`;
    opts.bodyEl.style.top = `${headerH}px`;

    const { active: activeCols, visible, leadWidth } = columnWindow(columns, scrollLeft, viewW);
    opts.onHorizontalScroll?.(scrollLeft, viewW);
    const sel = opts.getSelection();

    for (let i = 0; i < poolSize; i++) {
      const rowIndex = firstRow + i;
      const rowEl = rows[i]!;
      if (rowIndex >= totalRows) {
        rowEl.style.display = "none";
        continue;
      }
      rowEl.style.display = "flex";
      rowEl.style.transform = `translateY(${rowIndex * opts.rowHeight}px)`;
      rowEl.style.width = `${totalBodyWidth(columns)}px`;
      rowEl.dataset.row = String(rowIndex);

      const selected = rowIsSelected(sel, rowIndex);
      rowEl.className = selected
        ? "absolute left-0 flex w-full border-b border-line-2 bg-accent-soft"
        : "absolute left-0 flex w-full border-b border-line-2 hover:bg-raise";

      rowEl.replaceChildren();

      const rn = document.createElement("div");
      rn.className =
        "sticky left-0 z-10 shrink-0 cursor-pointer border-r border-line-2 bg-sidebar text-right font-sans text-11-5 text-ink-3";
      rn.style.width = `${rowNumberWidth()}px`;
      rn.style.minWidth = `${rowNumberWidth()}px`;
      rn.style.lineHeight = `${opts.rowHeight}px`;
      rn.textContent = String(rowIndex + 1);
      rn.addEventListener("click", (e) => {
        opts.onRowClick(rowIndex, e.shiftKey);
      });
      rowEl.appendChild(rn);

      if (leadWidth > 0) {
        const lead = document.createElement("div");
        lead.className = "shrink-0";
        lead.style.width = `${leadWidth}px`;
        lead.setAttribute("aria-hidden", "true");
        rowEl.appendChild(lead);
      }

      for (let vi = 0; vi < visible.length; vi++) {
        const col = visible[vi]!;
        const colIndex = activeCols.indexOf(col);
        const cell = document.createElement("div");
        if (cellInSelection(sel, rowIndex, colIndex)) {
          cell.classList.add("bg-accent-soft");
        }
        cell.addEventListener("click", (e) => {
          opts.onCellClick(rowIndex, colIndex, e.shiftKey);
        });
        renderCell(cell, rowIndex, colIndex, col);
        rowEl.appendChild(cell);
      }
    }

    void opts.buffer.ensureViewport(firstRow, firstRow + visibleRows + POOL_EXTRA).catch(() => {});
  }

  function scheduleLayout(): void {
    if (raf) {
      return;
    }
    raf = requestAnimationFrame(() => {
      raf = 0;
      layout();
    });
  }

  scheduleLayout();

  return {
    setColumns(next) {
      columns = next;
      scheduleLayout();
    },
    setTotalRows(n, t) {
      totalRows = n;
      total = t;
      scheduleLayout();
    },
    scheduleLayout,
    layoutNow: layout,
    rowElementCount: () => poolSize,
    destroy() {
      if (raf) {
        cancelAnimationFrame(raf);
      }
      for (const r of rows) {
        r.remove();
      }
      rows.length = 0;
      poolSize = 0;
    },
  };
}
