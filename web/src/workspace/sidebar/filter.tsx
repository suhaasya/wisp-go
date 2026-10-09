import { useEffect } from "preact/hooks";
import { Input } from "../../ui/Input";
import { Select } from "../../ui/Select";

export function filterNames(names: readonly string[], query: string): string[] {
  const q = query.trim().toLowerCase();
  if (!q) {
    return [...names];
  }
  return names.filter((n) => n.toLowerCase().includes(q));
}

export type SchemaSelectorProps = {
  engine: string;
  databases: string[];
  schemas: string[];
  database: string;
  schema: string;
  filter: string;
  filterInputId?: string;
  onDatabaseChange: (db: string) => void;
  onSchemaChange: (schema: string) => void;
  onFilterChange: (value: string) => void;
};

export function SchemaSelector({
  engine,
  databases,
  schemas,
  database,
  schema,
  filter,
  filterInputId = "schema-sidebar-filter",
  onDatabaseChange,
  onSchemaChange,
  onFilterChange,
}: SchemaSelectorProps) {
  const isPostgres = engine === "postgres";
  const dbOptions = databases.map((d) => ({ value: d, label: d }));
  const schemaOptions = schemas.map((s) => ({ value: s, label: s }));

  return (
    <div class="flex shrink-0 flex-col gap-2 border-b border-line px-2.5 py-2.5">
      <div class="flex items-center justify-between rounded-control border border-line bg-panel px-2 py-1 text-12-5">
        <span class="min-w-0 truncate text-ink">
          <span class="font-medium">{database || "—"}</span>
          {isPostgres && schema ? <span class="text-ink-3"> / {schema}</span> : null}
        </span>
        <span class="text-ink-3" aria-hidden="true">
          ▾
        </span>
      </div>
      {isPostgres && dbOptions.length > 1 ? (
        <Select
          label="Database"
          value={database}
          options={dbOptions}
          onChange={onDatabaseChange}
        />
      ) : null}
      {isPostgres && schemaOptions.length > 0 ? (
        <Select label="Schema" value={schema} options={schemaOptions} onChange={onSchemaChange} />
      ) : null}
      {!isPostgres && dbOptions.length > 1 ? (
        <Select label="Database" value={database} options={dbOptions} onChange={onDatabaseChange} />
      ) : null}
      <Input
        id={filterInputId}
        ariaLabel="Filter objects"
        placeholder="Filter…"
        value={filter}
        onInput={onFilterChange}
      />
    </div>
  );
}

export function useFilterFocusShortcut(filterInputId: string, enabled: boolean): void {
  useEffect(() => {
    if (!enabled) {
      return;
    }
    const onKeyDown = (e: KeyboardEvent) => {
      if (!(e.metaKey || e.ctrlKey) || e.key.toLowerCase() !== "f") {
        return;
      }
      const target = e.target;
      if (
        target instanceof HTMLElement &&
        (target.tagName === "INPUT" ||
          target.tagName === "TEXTAREA" ||
          target.isContentEditable)
      ) {
        return;
      }
      e.preventDefault();
      document.getElementById(filterInputId)?.focus();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [filterInputId, enabled]);
}
