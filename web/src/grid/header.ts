import type { TableSortKey } from "../state";
import { sortRank } from "../workspace/sort-draft";
import {
  clampColumnWidth,
  columnWindow,
  rowNumberWidth,
  totalBodyWidth,
  type GridColumn,
} from "./columns";

export type HeaderCallbacks = {
  onResize: (colId: string, width: number) => void;
  onReorder: (fromId: string, toId: string) => void;
  onHide: (colId: string) => void;
  onSortClick: (colId: string, shiftKey: boolean) => void;
  sortDraft: TableSortKey[];
};

export type GridHeader = {
  setColumns: (columns: GridColumn[]) => void;
  setSortDraft: (sort: TableSortKey[]) => void;
  setHorizontalScroll: (scrollLeft: number, viewportWidth: number) => void;
  destroy: () => void;
};

function columnWindowKey(
  columns: GridColumn[],
  scrollLeft: number,
  viewportWidth: number,
): string {
  const { start, visible, leadWidth } = columnWindow(columns, scrollLeft, viewportWidth);
  return `${start}:${leadWidth}:${visible.map((c) => c.id).join("\t")}`;
}

export function createGridHeader(
  root: HTMLElement,
  columns: GridColumn[],
  callbacks: HeaderCallbacks,
): GridHeader {
  const row = document.createElement("div");
  row.className =
    "sticky top-0 z-20 flex flex-nowrap border-b border-line bg-sidebar text-12-5 font-medium text-ink-2";
  row.setAttribute("role", "row");
  root.appendChild(row);

  let cols = columns;
  let sortDraft = callbacks.sortDraft;
  let scrollLeft = 0;
  let viewportWidth = 800;
  let lastWindowKey = "";
  let dragCol: string | null = null;
  let resizeCol: string | null = null;
  let resizeStartX = 0;
  let resizeStartW = 0;

  function render(): void {
    row.replaceChildren();
    row.style.width = `${totalBodyWidth(cols)}px`;

    const rn = document.createElement("div");
    rn.className =
      "sticky left-0 z-30 shrink-0 border-r border-line-2 bg-sidebar text-right text-11-5 text-ink-3";
    rn.style.width = `${rowNumberWidth()}px`;
    rn.style.minWidth = `${rowNumberWidth()}px`;
    rn.textContent = "#";
    row.appendChild(rn);

    const { visible, leadWidth } = columnWindow(cols, scrollLeft, viewportWidth);
    if (leadWidth > 0) {
      const lead = document.createElement("div");
      lead.className = "shrink-0";
      lead.style.width = `${leadWidth}px`;
      lead.setAttribute("aria-hidden", "true");
      row.appendChild(lead);
    }

    for (const col of visible) {
      const th = document.createElement("div");
      th.className =
        "relative shrink-0 border-r border-line-2 px-2.5 py-1.5 select-none";
      th.style.width = `${col.width}px`;
      th.style.minWidth = `${col.width}px`;
      th.dataset.colId = col.id;
      th.draggable = true;

      const rank = sortRank(sortDraft, col.meta.name);
      const name = document.createElement("div");
      name.className = "truncate font-sans text-12-5 font-medium text-ink-2";
      if (rank >= 0) {
        const arrow = sortDraft[rank]?.desc ? " ↓" : " ↑";
        name.textContent = col.meta.name + arrow + (rank > 0 ? ` ${rank + 1}` : "");
      } else {
        name.textContent = col.meta.name;
      }
      th.appendChild(name);
      th.addEventListener("click", (e) => {
        if (resizeCol) {
          return;
        }
        callbacks.onSortClick(col.meta.name, (e as MouseEvent).shiftKey);
      });
      const small = document.createElement("div");
      small.className = "mt-0.5 truncate font-mono text-11-5 font-normal text-ink-3";
      small.textContent = col.meta.type;
      th.appendChild(small);

      const grip = document.createElement("div");
      grip.className = "absolute right-0 top-0 h-full w-1.5 cursor-col-resize";
      th.appendChild(grip);

      th.addEventListener("dragstart", (e) => {
        dragCol = col.id;
        e.dataTransfer?.setData("text/plain", col.id);
      });
      th.addEventListener("dragover", (e) => {
        e.preventDefault();
      });
      th.addEventListener("drop", (e) => {
        e.preventDefault();
        const from = dragCol ?? e.dataTransfer?.getData("text/plain");
        if (from && from !== col.id) {
          callbacks.onReorder(from, col.id);
        }
        dragCol = null;
      });
      th.addEventListener("contextmenu", (e) => {
        e.preventDefault();
        callbacks.onHide(col.id);
      });

      grip.addEventListener("pointerdown", (e) => {
        e.preventDefault();
        resizeCol = col.id;
        resizeStartX = e.clientX;
        resizeStartW = col.width;
        (e.target as HTMLElement).setPointerCapture(e.pointerId);
      });
      grip.addEventListener("pointermove", (e) => {
        if (resizeCol !== col.id) {
          return;
        }
        const dw = e.clientX - resizeStartX;
        callbacks.onResize(col.id, clampColumnWidth(resizeStartW + dw));
      });
      grip.addEventListener("pointerup", () => {
        resizeCol = null;
      });

      row.appendChild(th);
    }
  }

  render();

  return {
    setColumns(next: GridColumn[]) {
      cols = next;
      lastWindowKey = "";
      render();
    },
    setSortDraft(next: TableSortKey[]) {
      sortDraft = next;
      render();
    },
    setHorizontalScroll(left: number, width: number) {
      scrollLeft = left;
      viewportWidth = width;
      const key = columnWindowKey(cols, left, width);
      if (key === lastWindowKey) {
        return;
      }
      lastWindowKey = key;
      render();
    },
    destroy() {
      row.remove();
    },
  };
}
