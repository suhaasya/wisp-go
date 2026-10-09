import type { TableDetailsResponse } from "../schema-client";

function columnKeyLabel(
  col: string,
  primaryKey: string[],
  foreignKeys: TableDetailsResponse["foreign_keys"],
): string {
  if (primaryKey.includes(col)) {
    return "Primary";
  }
  for (const fk of foreignKeys ?? []) {
    if (fk.columns.includes(col)) {
      const refs = fk.ref_columns.join(", ");
      return `FK ${fk.ref_table}(${refs})`;
    }
  }
  return "";
}

function rowClass(mark?: string): string {
  if (mark === "mod") {
    return "border-b border-line-2 bg-modified";
  }
  if (mark === "ins") {
    return "border-b border-line-2 bg-inserted";
  }
  return "border-b border-line-2";
}

export function StructureColumnsSection({
  details,
  editable,
  rowMarks,
  onFieldChange,
}: {
  details: TableDetailsResponse;
  editable?: boolean;
  rowMarks?: Record<string, string>;
  onFieldChange?: (column: string, field: "type" | "nullable" | "default", value: string) => void;
}) {
  const pk = details.primary_key ?? [];
  const cols = details.columns ?? [];
  return (
    <div class="overflow-hidden rounded-panel border border-line bg-panel" data-testid="structure-columns">
      <h5 class="flex items-center justify-between border-b border-line-2 px-3 py-2.5 text-12-5 font-semibold text-ink">
        Columns
        <span class="font-normal text-ink-3">{cols.length}</span>
      </h5>
      <table class="w-full border-collapse text-12-5">
        <thead>
          <tr class="bg-raise text-12 text-ink-2">
            <th class="border-b border-line-2 px-3 py-1.5 text-left font-medium">Name</th>
            <th class="border-b border-line-2 px-3 py-1.5 text-left font-medium">Type</th>
            <th class="border-b border-line-2 px-3 py-1.5 text-left font-medium">Nullable</th>
            <th class="border-b border-line-2 px-3 py-1.5 text-left font-medium">Default</th>
            <th class="border-b border-line-2 px-3 py-1.5 text-left font-medium">Key</th>
          </tr>
        </thead>
        <tbody>
          {cols.map((col) => (
            <tr key={col.name} class={rowClass(rowMarks?.[col.name])} data-row-mark={rowMarks?.[col.name] ?? ""}>
              <td class="whitespace-nowrap px-3 py-1.5 font-mono text-12 text-ink">{col.name}</td>
              <td class="whitespace-nowrap px-3 py-1.5 font-mono text-12 text-ink">
                {editable && onFieldChange ? (
                  <input
                    class="input w-full min-w-[80px] font-mono text-12"
                    value={col.type}
                    onChange={(e) => onFieldChange(col.name, "type", (e.target as HTMLInputElement).value)}
                    onBlur={(e) => onFieldChange(col.name, "type", (e.target as HTMLInputElement).value)}
                  />
                ) : (
                  col.type
                )}
              </td>
              <td class="whitespace-nowrap px-3 py-1.5 text-ink">
                {editable && onFieldChange ? (
                  <select
                    class="input font-mono text-12"
                    value={col.nullable !== false ? "Yes" : "No"}
                    onChange={(e) =>
                      onFieldChange(col.name, "nullable", (e.target as HTMLSelectElement).value)
                    }
                  >
                    <option value="Yes">Yes</option>
                    <option value="No">No</option>
                  </select>
                ) : col.nullable !== false ? (
                  "Yes"
                ) : (
                  "No"
                )}
              </td>
              <td class="max-w-[200px] whitespace-nowrap px-3 py-1.5 font-mono text-12 text-ink">
                {editable && onFieldChange ? (
                  <input
                    class="input w-full font-mono text-12"
                    value={col.default ?? ""}
                    onBlur={(e) => onFieldChange(col.name, "default", (e.target as HTMLInputElement).value)}
                  />
                ) : (
                  col.default ?? ""
                )}
              </td>
              <td class="whitespace-nowrap px-3 py-1.5 text-12 text-ink-2">
                {columnKeyLabel(col.name, pk, details.foreign_keys)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
