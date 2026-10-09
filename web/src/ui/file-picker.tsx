import { useEffect, useState } from "preact/hooks";
import { request } from "../api";
import { Button } from "./Button";
import { Input } from "./Input";
import { Modal } from "./Modal";

type FSEntry = { name: string; dir: boolean };

type FSListResponse = { path: string; entries: FSEntry[] };

export type FilePickerProps = {
  open: boolean;
  onClose: () => void;
  onSelect: (path: string) => void;
  filter?: (name: string) => boolean;
};

export function FilePicker({ open, onClose, onSelect, filter }: FilePickerProps) {
  const [path, setPath] = useState("~");
  const [filterText, setFilterText] = useState("");
  const [entries, setEntries] = useState<FSEntry[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) {
      return;
    }
    let cancelled = false;
    const ac = new AbortController();
    request<FSListResponse>("GET", `/api/v1/fs/list?path=${encodeURIComponent(path)}`, undefined, {
      signal: ac.signal,
    })
      .then((doc) => {
        if (!cancelled) {
          setEntries(doc.entries);
          setPath(doc.path);
          setError("");
        }
      })
      .catch((e) => {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "Could not list files");
        }
      });
    return () => {
      cancelled = true;
      ac.abort();
    };
  }, [open, path]);

  const crumbs = breadcrumbParts(path);
  const visible = entries.filter((e) => {
    if (filter && !e.dir && !filter(e.name)) {
      return false;
    }
    if (!filterText.trim()) {
      return true;
    }
    return e.name.toLowerCase().includes(filterText.trim().toLowerCase());
  });

  return (
    <Modal
      open={open}
      title="Choose file"
      onClose={onClose}
      footer={
        <div class="flex justify-end">
          <Button label="Cancel" onClick={onClose} />
        </div>
      }
    >
      <div class="mb-2 flex flex-wrap gap-1 text-12-5 text-ink-2">
        {crumbs.map((c) => (
          <button
            key={c.path}
            type="button"
            class="border-0 bg-transparent text-accent hover:underline"
            onClick={() => setPath(c.path)}
          >
            {c.label}
          </button>
        ))}
      </div>
      <Input ariaLabel="Filter files" placeholder="Filter" value={filterText} onInput={setFilterText} />
      {error ? <p class="mt-2 text-12-5 text-env-prod">{error}</p> : null}
      <ul class="mt-2 max-h-48 overflow-auto rounded-control border border-line">
        {visible.map((e) => (
          <li key={e.name}>
            <button
              type="button"
              class="flex w-full border-0 border-b border-line-2 bg-panel px-3 py-1.5 text-left text-13 hover:bg-raise"
              onClick={() => {
                const next = joinPath(path, e.name);
                if (e.dir) {
                  setPath(next);
                } else {
                  onSelect(next);
                  onClose();
                }
              }}
            >
              <span class="text-ink-3">{e.dir ? "Dir" : "File"}</span> {e.name}
            </button>
          </li>
        ))}
      </ul>
    </Modal>
  );
}

function breadcrumbParts(path: string): { label: string; path: string }[] {
  if (path === "~") {
    return [{ label: "Home", path: "~" }];
  }
  const rest = path.startsWith("~/") ? path.slice(2) : path;
  const parts = rest.split("/").filter(Boolean);
  const out: { label: string; path: string }[] = [{ label: "Home", path: "~" }];
  let cur = "~";
  for (const p of parts) {
    cur = cur === "~" ? `~/${p}` : `${cur}/${p}`;
    out.push({ label: p, path: cur });
  }
  return out;
}

function joinPath(base: string, name: string): string {
  if (base === "~") {
    return `~/${name}`;
  }
  return `${base}/${name}`;
}
