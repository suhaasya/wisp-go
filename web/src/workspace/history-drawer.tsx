import { useEffect, useMemo, useState } from "preact/hooks";
import { Button } from "../ui/Button";
import { fetchHistory } from "./history-client";

export type HistoryEntry = {
  ts: number;
  sql: string;
  ms?: number;
  rows?: number;
  status: string;
};

type HistoryDrawerProps = {
  open: boolean;
  profileId: string;
  onClose: () => void;
  onInsert: (sql: string) => void;
  onRun: (sql: string) => void;
};

function dayKey(ts: number): string {
  const d = new Date(ts);
  return d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

function previewLines(sql: string): string {
  const lines = sql.split("\n").slice(0, 2);
  return lines.join("\n");
}

export function HistoryDrawer(props: HistoryDrawerProps) {
  const [query, setQuery] = useState("");
  const [entries, setEntries] = useState<HistoryEntry[]>([]);

  useEffect(() => {
    if (!props.open || !props.profileId) {
      return;
    }
    const ac = new AbortController();
    const timer = window.setTimeout(() => {
      void fetchHistory(props.profileId, query, ac.signal)
        .then(setEntries)
        .catch(() => setEntries([]));
    }, 200);
    return () => {
      ac.abort();
      window.clearTimeout(timer);
    };
  }, [props.open, props.profileId, query]);

  const grouped = useMemo(() => {
    const map = new Map<string, HistoryEntry[]>();
    for (const e of entries) {
      const key = dayKey(e.ts);
      const list = map.get(key) ?? [];
      list.push(e);
      map.set(key, list);
    }
    return [...map.entries()];
  }, [entries]);

  if (!props.open) {
    return null;
  }

  return (
    <div
      class="absolute inset-y-0 right-0 z-20 flex w-[360px] flex-col border-l border-line bg-panel shadow-lg"
      data-testid="history-drawer"
    >
      <div class="flex items-center gap-2 border-b border-line px-3 py-2">
        <input
          class="input flex-1 font-mono text-12"
          placeholder="Search history"
          value={query}
          onInput={(e) => setQuery((e.target as HTMLInputElement).value)}
          aria-label="Search history"
        />
        <Button label="Close" size="sm" onClick={props.onClose} />
      </div>
      <div class="min-h-0 flex-1 overflow-auto p-2">
        {grouped.length === 0 ? (
          <p class="text-12 text-ink-3">No history entries.</p>
        ) : (
          grouped.map(([day, list]) => (
            <section key={day} class="mb-3">
              <h3 class="mb-1 text-11-5 font-medium uppercase tracking-wide text-ink-3">{day}</h3>
              <ul class="list-none">
                {list.map((e) => (
                  <li key={`${e.ts}-${e.sql.slice(0, 20)}`}>
                    <button
                      type="button"
                      class="mb-1 w-full rounded border border-line bg-canvas px-2 py-1.5 text-left hover:bg-raise"
                      onClick={() => props.onInsert(e.sql)}
                      onDblClick={() => props.onRun(e.sql)}
                    >
                      <pre class="line-clamp-2 whitespace-pre-wrap font-mono text-12 text-ink-2">
                        {previewLines(e.sql)}
                      </pre>
                      <span class="text-11-5 text-ink-3">
                        {e.status}
                        {e.ms !== undefined ? ` · ${e.ms} ms` : ""}
                        {e.rows !== undefined && e.rows > 0 ? ` · ${e.rows} rows` : ""}
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            </section>
          ))
        )}
      </div>
    </div>
  );
}
