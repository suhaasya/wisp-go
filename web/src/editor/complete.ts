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

export function toSQLNamespace(data: CompletionSchemaDTO): SQLNamespace {
  const out: Record<string, SQLNamespace> = {};
  for (const table of data.tables) {
    const cols: Record<string, Completion> = {};
    for (const col of table.columns) {
      cols[col.name] = {
        label: col.name,
        type: "property",
        detail: col.type,
      };
    }
    out[table.name] = cols as unknown as SQLNamespace;
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

export function mergeCompletionSchema(
  current: CompletionSchemaDTO,
  patch: CompletionSchemaDTO,
): CompletionSchemaDTO {
  const byName = new Map(current.tables.map((t) => [t.name, t]));
  for (const t of patch.tables) {
    byName.set(t.name, t);
  }
  const tables = [...byName.values()].sort((a, b) => a.name.localeCompare(b.name));
  return { schema: current.schema || patch.schema, tables };
}
