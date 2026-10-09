import { useEffect, useState } from "preact/hooks";
import type { ColumnMeta } from "../grid/window";
import { Button } from "../ui/Button";
import { Select } from "../ui/Select";
import type { TableBrowseFilter, TableBrowseQuery, TableSortKey } from "../state";
import { emptyBrowseFilter, formatSortHint } from "./browse-query";
import { useFilterFocusShortcut } from "./sidebar/filter";

const FILTER_INPUT_ID = "table-data-filter-value";

const OPERATORS = [
  { value: "=", label: "=" },
  { value: "!=", label: "!=" },
  { value: "<", label: "<" },
  { value: ">", label: ">" },
  { value: "like", label: "LIKE" },
  { value: "ilike", label: "ILIKE" },
  { value: "in", label: "IN" },
  { value: "is null", label: "IS NULL" },
  { value: "between", label: "BETWEEN" },
];

export type FilterBarProps = {
  columns: ColumnMeta[];
  applied: TableBrowseQuery;
  sortDraft: TableSortKey[];
  onApply: (query: TableBrowseQuery) => void;
};

export function FilterBar({ columns, applied, sortDraft, onApply }: FilterBarProps) {
  const [rawMode, setRawMode] = useState(applied.rawMode);
  const [rawWhere, setRawWhere] = useState(applied.rawWhere);
  const [filter, setFilter] = useState<TableBrowseFilter>(() =>
    applied.filter.conditions.length ? applied.filter : emptyBrowseFilter(),
  );

  useEffect(() => {
    setRawMode(applied.rawMode);
    setRawWhere(applied.rawWhere);
    setFilter(applied.filter.conditions.length ? applied.filter : emptyBrowseFilter());
  }, [applied]);

  useFilterFocusShortcut(FILTER_INPUT_ID, true);

  const columnOptions = columns.map((c) => ({ value: c.name, label: c.name }));

  function updateCondition(index: number, patch: Partial<TableBrowseFilter["conditions"][0]>) {
    setFilter((f) => {
      const conditions = f.conditions.map((c, i) => (i === index ? { ...c, ...patch } : c));
      return { ...f, conditions };
    });
  }

  function addCondition() {
    setFilter((f) => ({
      ...f,
      conditions: [...f.conditions, { column: columnOptions[0]?.value ?? "", op: "=", value: "" }],
    }));
  }

  function apply() {
    onApply({
      rawMode,
      rawWhere: rawMode ? rawWhere : "",
      filter: rawMode ? emptyBrowseFilter() : filter,
      sort: sortDraft,
    });
  }

  const sortHint = formatSortHint(sortDraft);

  return (
    <div class="flex flex-none flex-wrap items-center gap-1.5 border-b border-line bg-raise px-2.5 py-1.5 text-12">
      <label class="flex items-center gap-1 text-ink-2">
        <input
          type="checkbox"
          checked={rawMode}
          onChange={(e) => setRawMode((e.target as HTMLInputElement).checked)}
        />
        Raw WHERE
      </label>
      {rawMode ? (
        <input
          class="min-w-[200px] flex-1 rounded-control border border-line bg-panel px-1.5 py-[3px] font-mono text-12 text-ink"
          value={rawWhere}
          onInput={(e) => setRawWhere((e.target as HTMLInputElement).value)}
          aria-label="Raw WHERE"
        />
      ) : (
        <>
          {filter.conditions.map((cond, i) => (
            <div key={i} class="flex items-center gap-1.5">
              {i > 0 ? (
                <Select
                  value={filter.logic}
                  options={[
                    { value: "and", label: "AND" },
                    { value: "or", label: "OR" },
                  ]}
                  onChange={(v) => setFilter((f) => ({ ...f, logic: v as "and" | "or" }))}
                />
              ) : null}
              <div class="w-[120px]">
                <Select
                  value={cond.column || columnOptions[0]?.value}
                  options={columnOptions}
                  onChange={(v) => updateCondition(i, { column: v })}
                />
              </div>
              <div class="w-[100px]">
                <Select
                  value={cond.op}
                  options={OPERATORS}
                  onChange={(v) => updateCondition(i, { op: v })}
                />
              </div>
              {cond.op !== "is null" ? (
                <input
                  id={i === 0 ? FILTER_INPUT_ID : undefined}
                  class="w-[200px] rounded-control border border-line bg-panel px-1.5 py-[3px] font-mono text-12 text-ink"
                  value={cond.value}
                  onInput={(e) => updateCondition(i, { value: (e.target as HTMLInputElement).value })}
                  aria-label="Filter value"
                />
              ) : null}
              {cond.op === "between" ? (
                <input
                  class="w-[120px] rounded-control border border-line bg-panel px-1.5 py-[3px] font-mono text-12 text-ink"
                  value={cond.value2 ?? ""}
                  onInput={(e) => updateCondition(i, { value2: (e.target as HTMLInputElement).value })}
                  aria-label="Filter value end"
                />
              ) : null}
            </div>
          ))}
          <button
            type="button"
            class="rounded-control border border-line bg-panel px-2 py-[3px] text-12 text-ink-2 hover:bg-raise"
            onClick={addCondition}
          >
            + condition
          </button>
        </>
      )}
      <Button label="Apply" size="sm" onClick={apply} />
      {sortHint ? (
        <span class="ml-1.5 text-12 text-ink-3">Sorted by {sortHint}</span>
      ) : null}
    </div>
  );
}
