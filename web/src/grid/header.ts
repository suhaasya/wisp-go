import { clampColumnWidth, rowNumberWidth, type GridColumn } from "./columns";

export type HeaderCallbacks = {
  onResize: (colId: string, width: number) => void;
  onReorder: (fromId: string, toId: string) => void;
  onHide: (colId: string) => void;
};

export type GridHeader = {
  setColumns: (columns: GridColumn[]) => void;
  destroy: () => void;
};

export function createGridHeader(
  root: HTMLElement,
  columns: GridColumn[],
  callbacks: HeaderCallbacks,
): GridHeader {
  const row = document.createElement("div");
  row.className =
    "sticky top-0 z-20 flex border-b border-line bg-sidebar text-12-5 font-medium text-ink-2";
  row.setAttribute("role", "row");
  root.appendChild(row);

  let cols = columns;
  let dragCol: string | null = null;
  let resizeCol: string | null = null;
  let resizeStartX = 0;
  let resizeStartW = 0;

  function render(): void {
    row.replaceChildren();
    const rn = document.createElement("div");
    rn.className =
      "sticky left-0 z-30 shrink-0 border-r border-line-2 bg-sidebar text-right text-11-5 text-ink-3";
    rn.style.width = `${rowNumberWidth()}px`;
    rn.style.minWidth = `${rowNumberWidth()}px`;
    rn.textContent = "#";
    row.appendChild(rn);

    for (const col of cols) {
      if (col.hidden) {
        continue;
      }
      const th = document.createElement("div");
      th.className =
        "relative shrink-0 border-r border-line-2 px-2.5 py-1.5 select-none whitespace-nowrap";
      th.style.width = `${col.width}px`;
      th.style.minWidth = `${col.width}px`;
      th.dataset.colId = col.id;
      th.draggable = true;

      const name = document.createElement("span");
      name.textContent = col.meta.name;
      th.appendChild(name);
      const small = document.createElement("small");
      small.className = "mt-0.5 block font-mono text-11-5 font-normal text-ink-3";
      small.textContent = col.meta.type;
      th.appendChild(small);

      const grip = document.createElement("div");
      grip.className = "absolute right-0 top-0 h-full w-1.5 cursor-col-resize";
      grip.dataset.resize = col.id;
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
      render();
    },
    destroy() {
      row.remove();
    },
  };
}
