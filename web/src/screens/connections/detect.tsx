import { useEffect, useState } from "preact/hooks";
import { request } from "../../api";
import { defaultFormState, type FormState } from "../../connection-form";
import type { ConnectionProfile } from "../../connections";
import { navigate } from "../../router";
import { Button } from "../../ui/Button";
import { stashImportDraft } from "./paste";

const DISMISS_KEY = "wisp-dismiss-detect";

export type LocalSuggestion = {
  id: string;
  engine: string;
  label: string;
  user: string;
  database: string;
  kind: string;
  host: string;
  port: number;
  path: string;
};

type DetectResponse = { suggestions: LocalSuggestion[] };

function loadDismissed(): Set<string> {
  try {
    const raw = localStorage.getItem(DISMISS_KEY);
    if (!raw) {
      return new Set();
    }
    return new Set(JSON.parse(raw) as string[]);
  } catch {
    return new Set();
  }
}

function saveDismissed(set: Set<string>): void {
  localStorage.setItem(DISMISS_KEY, JSON.stringify([...set]));
}

function formFromSuggestion(s: LocalSuggestion): FormState {
  const base = defaultFormState();
  const engine = s.engine as ConnectionProfile["engine"];
  const transportKind =
    s.kind === "pipe" ? "pipe" : s.kind === "socket" ? "socket" : "tcp";
  return {
    ...base,
    name: s.label,
    engine,
    transportKind,
    host: s.host || "127.0.0.1",
    port: s.port || base.port,
    socketPath: s.path,
    user: s.user,
    database: s.database,
    env: "local",
    ssl: { ...base.ssl, mode: transportKind === "tcp" ? base.ssl.mode : "disable" },
  };
}

export function LocalDetectRow() {
  const [items, setItems] = useState<LocalSuggestion[]>([]);
  const [dismissed, setDismissed] = useState(loadDismissed);

  useEffect(() => {
    const ac = new AbortController();
    void request<DetectResponse>("GET", "/api/v1/detect/local", undefined, { signal: ac.signal })
      .then((doc) => {
        setItems(doc.suggestions ?? []);
      })
      .catch(() => {});
    return () => ac.abort();
  }, []);

  const visible = items.filter((s) => !dismissed.has(s.id));
  if (visible.length === 0) {
    return null;
  }

  function dismiss(id: string) {
    const next = new Set(dismissed);
    next.add(id);
    setDismissed(next);
    saveDismissed(next);
  }

  function connect(s: LocalSuggestion) {
    stashImportDraft(formFromSuggestion(s));
    navigate("/connections/new");
  }

  return (
    <div class="mb-4 rounded-dialog border border-line bg-panel px-4 py-3" role="region" aria-label="Detected databases">
      <div class="mb-2 flex items-center justify-between gap-2">
        <h2 class="text-13 font-semibold text-ink">Local databases detected</h2>
        <span class="text-12 text-ink-3">One-click to pre-fill a new connection</span>
      </div>
      <ul class="list-none space-y-2">
        {visible.map((s) => (
          <li key={s.id} class="flex flex-wrap items-center gap-2 text-13">
            <span class="min-w-0 flex-1 truncate font-medium text-ink">{s.label}</span>
            <span class="font-mono text-12 text-ink-2">
              {s.kind === "tcp" ? `${s.host}:${s.port}` : s.path || s.host}
            </span>
            <Button label="Connect" size="sm" variant="primary" onClick={() => connect(s)} />
            <Button label="Dismiss" size="sm" variant="ghost" onClick={() => dismiss(s.id)} />
          </li>
        ))}
      </ul>
    </div>
  );
}
