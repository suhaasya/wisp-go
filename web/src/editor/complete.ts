import type { Completion } from "@codemirror/autocomplete";
import type { SQLNamespace } from "@codemirror/lang-sql";

export type CompletionTableDTO = {
  name: string;
  columns: { name: string; type: string }[];
};

export type CompletionSchemaDTO = {
  schema: string;
  tables: CompletionTableDTO[];
};

export type SnippetCompletionDTO = {
  name: string;
  tags?: string[];
};

export function toSQLNamespace(data: CompletionSchemaDTO): SQLNamespace {
  const normalized = normalizeCompletionSchema(data);
  const out: Record<string, SQLNamespace> = {};
  for (const table of normalized.tables) {
    const columns = Array.isArray(table.columns) ? table.columns : [];
    const cols: Completion[] = columns.map((col) => ({
      label: col.name,
      type: "property",
      detail: col.type,
    }));
    out[table.name] = cols as unknown as SQLNamespace;
  }
  return out as SQLNamespace;
}

export function mergeSnippetsIntoNamespace(
  ns: SQLNamespace,
  snippets: SnippetCompletionDTO[],
): SQLNamespace {
  const out = { ...(ns as Record<string, SQLNamespace>) };
  for (const sn of snippets) {
    const tag = sn.tags?.length ? sn.tags.join(", ") : "snippet";
    out[sn.name] = {
      self: {
        label: sn.name,
        type: "keyword",
        detail: tag,
      },
      children: {},
    } as unknown as SQLNamespace;
  }
  return out as SQLNamespace;
}

export type SchemaCompletionState = {
  namespace: SQLNamespace;
  defaultSchema: string;
};

export function referencedTableNames(sql: string): string[] {
  const found = new Set<string>();
  const re = /\b(?:from|join)\s+(?:ONLY\s+)?("([^"]+)"|`([^`]+)`|([a-zA-Z_][\w$]*))/gi;
  let m: RegExpExecArray | null;
  while ((m = re.exec(sql)) !== null) {
    const name = (m[2] ?? m[3] ?? m[4] ?? "").trim();
    if (name) {
      found.add(name);
    }
  }
  return [...found];
}

export function missingTableNames(loaded: Set<string>, referenced: string[]): string[] {
  return referenced.filter((t) => !loaded.has(t));
}

/** Tables referenced in SQL that still need column metadata for autocomplete. */
export function tablesNeedingColumnHydration(
  tables: CompletionTableDTO[],
  referenced: string[],
): string[] {
  const byName = new Map(tables.map((t) => [t.name, t]));
  const out: string[] = [];
  for (const name of referenced) {
    const row = byName.get(name);
    if (!row || row.columns.length === 0) {
      out.push(name);
    }
  }
  return out;
}

export function normalizeCompletionSchema(
  data: Partial<CompletionSchemaDTO> | null | undefined,
): CompletionSchemaDTO {
  const tables = Array.isArray(data?.tables) ? data.tables : [];
  return { schema: data?.schema ?? "", tables };
}

export function mergeCompletionSchema(
  current: CompletionSchemaDTO,
  patch: CompletionSchemaDTO,
): CompletionSchemaDTO {
  const base = normalizeCompletionSchema(current);
  const delta = normalizeCompletionSchema(patch);
  const byName = new Map(base.tables.map((t) => [t.name, t]));
  for (const t of delta.tables) {
    byName.set(t.name, t);
  }
  const tables = [...byName.values()].sort((a, b) => a.name.localeCompare(b.name));
  return { schema: base.schema || delta.schema, tables };
}
