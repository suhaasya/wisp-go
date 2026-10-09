import { useEffect, useMemo, useState } from "preact/hooks";
import { actionContext } from "../actions";
import { currentSession, paletteOpen } from "../state";
import { Input } from "../ui/Input";
import { VirtualList } from "../ui/VirtualList";
import { loadRecency } from "./matcher";
import {
  buildPaletteRows,
  loadPaletteItems,
  selectableRows,
  type PaletteRow,
  type PaletteSelectable,
} from "./registry";

const ROW_H = 34;

export function CommandPalette() {
  const open = paletteOpen.value;
  const [query, setQuery] = useState("");
  const [items, setItems] = useState<PaletteSelectable[]>([]);
  const [cursor, setCursor] = useState(0);
  const [recency, setRecency] = useState(loadRecency);
  const inputId = "wisp-palette-input";

  useEffect(() => {
    if (!open) {
      setQuery("");
      setCursor(0);
      return;
    }
    const ac = new AbortController();
    const ctx = actionContext();
    void loadPaletteItems(currentSession.value, ctx, ac.signal).then((loaded) => {
      if (!ac.signal.aborted) {
        setItems(loaded);
        setRecency(loadRecency());
      }
    });
    const t = window.setTimeout(() => {
      const el = document.getElementById(inputId) as HTMLInputElement | null;
      el?.focus();
      el?.select();
    }, 0);
    return () => {
      ac.abort();
      window.clearTimeout(t);
    };
  }, [open]);

  const rows: PaletteRow[] = useMemo(
    () => buildPaletteRows(items, query, recency),
    [items, query, recency],
  );
  const selectable = useMemo(() => selectableRows(rows), [rows]);

  useEffect(() => {
    if (cursor >= selectable.length) {
      setCursor(Math.max(0, selectable.length - 1));
    }
  }, [selectable.length, cursor]);

  const listRows = useMemo(() => flattenRows(rows), [rows]);

  if (!open) {
    return null;
  }

  function close() {
    paletteOpen.value = false;
  }

  function activate(item: PaletteSelectable) {
    item.run();
    setRecency(loadRecency());
    close();
  }

  function onKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      e.preventDefault();
      close();
      return;
    }
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setCursor((c) => Math.min(selectable.length - 1, c + 1));
      return;
    }
    if (e.key === "ArrowUp") {
      e.preventDefault();
      setCursor((c) => Math.max(0, c - 1));
      return;
    }
    if (e.key === "Enter") {
      e.preventDefault();
      const item = selectable[cursor];
      if (item) {
        activate(item);
      }
    }
  }

  return (
    <div
      class="fixed inset-0 z-50 flex justify-center bg-canvas/80 pt-20"
      role="presentation"
      onClick={(e) => {
        if (e.target === e.currentTarget) {
          close();
        }
      }}
    >
      <div
        class="h-fit w-[560px] max-w-[calc(100vw-24px)] overflow-hidden rounded-dialog border border-line bg-panel shadow-lg"
        role="dialog"
        aria-modal="true"
        aria-label="Command palette"
        onKeyDown={onKeyDown as unknown as (e: Event) => void}
      >
        <Input
          id={inputId}
          ariaLabel="Command"
          placeholder="Type a table, action or connection"
          value={query}
          onInput={setQuery}
        />
        <VirtualList
          class="max-h-[330px]"
          rowHeight={ROW_H}
          items={listRows}
          render={(row) => {
            if (row.kind === "section") {
              return (
                <div class="flex items-center px-2.5 text-11-5 text-ink-3" style={{ height: `${ROW_H}px` }}>
                  {row.label}
                </div>
              );
            }
            const selIdx = selectable.indexOf(row);
            const cur = selIdx === cursor;
            return (
              <div
                role="option"
                aria-selected={cur}
                class={`flex cursor-default items-center gap-2 px-2.5 text-13 ${cur ? "bg-accent-soft" : "text-ink"}`}
                style={{ height: `${ROW_H}px` }}
                onMouseDown={(e) => {
                  e.preventDefault();
                  activate(row);
                }}
              >
                <span class="min-w-0 flex-1 truncate">{row.title}</span>
                {row.subtitle ? (
                  <small class="ml-auto shrink-0 text-12 text-ink-3">{row.subtitle}</small>
                ) : null}
                {row.shortcut ? (
                  <small class="shrink-0 text-12 text-ink-3">{row.shortcut}</small>
                ) : null}
              </div>
            );
          }}
        />
      </div>
    </div>
  );
}

type FlatRow = { kind: "section"; label: string } | PaletteSelectable;

function flattenRows(rows: PaletteRow[]): FlatRow[] {
  return rows as FlatRow[];
}

