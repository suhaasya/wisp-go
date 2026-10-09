import type { TableDetailsResponse } from "../schema-client";

function MetaTable({
  title,
  count,
  headers,
  rows,
}: {
  title: string;
  count: number;
  headers: string[];
  rows: string[][];
}) {
  if (count === 0) {
    return null;
  }
  return (
    <div class="overflow-hidden rounded-panel border border-line bg-panel">
      <h5 class="flex items-center justify-between border-b border-line-2 px-3 py-2.5 text-12-5 font-semibold text-ink">
        {title}
        <span class="font-normal text-ink-3">{count}</span>
      </h5>
      <table class="w-full border-collapse text-12-5">
        <thead>
          <tr class="bg-raise text-12 text-ink-2">
            {headers.map((h) => (
              <th key={h} class="border-b border-line-2 px-3 py-1.5 text-left font-medium">
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} class="border-b border-line-2">
              {row.map((cell, j) => (
                <td
                  key={j}
                  class="whitespace-nowrap px-3 py-1.5 font-mono text-12 text-ink"
                >
                  {cell}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export function StructureForeignKeysSection({ details }: { details: TableDetailsResponse }) {
  const fks = details.foreign_keys ?? [];
  const rows = fks.map((fk) => [
    fk.name,
    fk.columns.join(", "),
    `${fk.ref_schema}.${fk.ref_table}(${fk.ref_columns.join(", ")})`,
  ]);
  return (
    <MetaTable
      title="Foreign keys"
      count={fks.length}
      headers={["Name", "Columns", "References"]}
      rows={rows}
    />
  );
}

export function StructureChecksSection({ details }: { details: TableDetailsResponse }) {
  const checks = details.checks ?? [];
  const rows = checks.map((c) => [c.name, c.expression]);
  return (
    <MetaTable title="Checks" count={checks.length} headers={["Name", "Expression"]} rows={rows} />
  );
}

export function StructureTriggersSection({ details }: { details: TableDetailsResponse }) {
  const triggers = details.triggers ?? [];
  const rows = triggers.map((tr) => [tr.name, tr.timing, tr.events, tr.enabled ? "Yes" : "No"]);
  return (
    <MetaTable
      title="Triggers"
      count={triggers.length}
      headers={["Name", "Timing", "Events", "Enabled"]}
      rows={rows}
    />
  );
}
