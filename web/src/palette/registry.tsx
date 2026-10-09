import { listActions, actionContext, runAction, type ActionContext } from "../actions";
import { loadConnectionProfiles, type ConnectionProfile } from "../connections";
import { fetchObjects, fetchSchemas } from "../workspace/schema-client";
import { fetchSnippets, type Snippet } from "../workspace/snippets-client";
import { navigate } from "../router";
import { paletteHandlers } from "./handlers";
import { parsePaletteQuery, rankByQuery, touchRecency, type PalettePrefixMode } from "./matcher";

export type PaletteSection = "tables" | "actions" | "connections" | "snippets";

export type PaletteSelectable = {
  kind: "item";
  id: string;
  section: PaletteSection;
  title: string;
  subtitle?: string;
  shortcut?: string;
  searchText: string;
  run: () => void;
};

export type PaletteRow = { kind: "section"; label: string } | PaletteSelectable;

const SECTION_LABEL: Record<PaletteSection, string> = {
  tables: "Tables",
  actions: "Actions",
  connections: "Connections",
  snippets: "Snippets",
};

export async function loadPaletteItems(
  sessionId: string | null,
  ctx: ActionContext,
  signal?: AbortSignal,
): Promise<PaletteSelectable[]> {
  const out: PaletteSelectable[] = [];
  for (const action of listActions(ctx)) {
    out.push({
      kind: "item",
      id: `action:${action.id}`,
      section: "actions",
      title: action.title,
      shortcut: action.shortcut,
      searchText: action.title,
      run: () => {
        touchRecency(`action:${action.id}`);
        runAction(action.id, ctx);
      },
    });
  }
  try {
    const { profiles } = await loadConnectionProfiles();
    if (signal?.aborted) {
      return out;
    }
    for (const p of profiles) {
      out.push(connectionItem(p));
    }
  } catch {
    // offline
  }
  try {
    const doc = await fetchSnippets();
    if (signal?.aborted) {
      return out;
    }
    for (const s of doc.snippets) {
      out.push(snippetItem(s));
    }
  } catch {
    // ignore
  }
  if (sessionId && !sessionId.startsWith("mock-")) {
    try {
      const tables = await loadSessionTables(sessionId, signal);
      out.push(...tables);
    } catch {
      // ignore
    }
  }
  return out;
}

function connectionItem(p: ConnectionProfile): PaletteSelectable {
  const subtitle = p.database || p.engine;
  return {
    kind: "item",
    id: `conn:${p.id}`,
    section: "connections",
    title: p.name,
    subtitle,
    searchText: `${p.name} ${subtitle} ${p.folder}`,
    run: () => {
      touchRecency(`conn:${p.id}`);
      navigate(`/connections/${encodeURIComponent(p.id)}/edit`);
    },
  };
}

function snippetItem(s: Snippet): PaletteSelectable {
  return {
    kind: "item",
    id: `snippet:${s.id}`,
    section: "snippets",
    title: s.name,
    subtitle: s.tags?.join(", "),
    searchText: `${s.name} ${s.tags?.join(" ") ?? ""}`,
    run: () => {
      touchRecency(`snippet:${s.id}`);
      paletteHandlers.openSnippet?.(s.id);
    },
  };
}

async function loadSessionTables(sessionId: string, signal?: AbortSignal): Promise<PaletteSelectable[]> {
  const snap = await fetchSchemas(sessionId, { signal });
  const schemas = snap.schemas ?? snap.databases ?? [];
  const out: PaletteSelectable[] = [];
  const limit = Math.min(schemas.length, 32);
  for (let i = 0; i < limit; i++) {
    const schema = schemas[i]!;
    const names = await fetchObjects(sessionId, schema, "table", { signal });
    for (const name of names) {
      out.push({
        kind: "item",
        id: `table:${schema}.${name}`,
        section: "tables",
        title: name,
        subtitle: schema,
        searchText: `${name} ${schema}`,
        run: () => {
          touchRecency(`table:${schema}.${name}`);
          paletteHandlers.openTable?.(schema, name);
        },
      });
    }
  }
  return out;
}

export function buildPaletteRows(
  items: PaletteSelectable[],
  query: string,
  recency: Record<string, number>,
): PaletteRow[] {
  const { mode, query: q } = parsePaletteQuery(query);
  const filtered = items.filter((item) => sectionAllowed(mode, item.section));
  const ranked = rankByQuery(
    q,
    filtered.map((item) => ({ id: item.id, text: item.searchText })),
    recency,
  );
  const byId = new Map(filtered.map((i) => [i.id, i]));
  const ordered: PaletteSelectable[] = [];
  for (const m of ranked) {
    const item = byId.get(m.id);
    if (item) {
      ordered.push(item);
    }
  }
  if (q === "" && ordered.length === 0) {
    return groupWithSections(filtered);
  }
  return groupWithSections(ordered);
}

function sectionAllowed(mode: PalettePrefixMode, section: PaletteSection): boolean {
  if (mode === "all") {
    return true;
  }
  if (mode === "actions") {
    return section === "actions";
  }
  if (mode === "connections") {
    return section === "connections";
  }
  return section === "snippets";
}

function groupWithSections(items: PaletteSelectable[]): PaletteRow[] {
  const order: PaletteSection[] = ["tables", "actions", "connections", "snippets"];
  const rows: PaletteRow[] = [];
  for (const sec of order) {
    const group = items.filter((i) => i.section === sec);
    if (group.length === 0) {
      continue;
    }
    rows.push({ kind: "section", label: SECTION_LABEL[sec] });
    rows.push(...group);
  }
  return rows;
}

export function selectableRows(rows: PaletteRow[]): PaletteSelectable[] {
  return rows.filter((r): r is PaletteSelectable => r.kind === "item");
}
