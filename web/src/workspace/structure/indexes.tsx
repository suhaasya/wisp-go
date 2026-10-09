import type { TableDetailsResponse } from "../schema-client";

export function StructureIndexesSection({ details }: { details: TableDetailsResponse }) {
  const indexes = details.indexes ?? [];
  return (
    <div class="overflow-hidden rounded-panel border border-line bg-panel">
      <h5 class="flex items-center justify-between border-b border-line-2 px-3 py-2.5 text-12-5 font-semibold text-ink">
        Indexes
        <span class="font-normal text-ink-3">{indexes.length}</span>
      </h5>
      <table class="w-full border-collapse text-12-5">
        <thead>
          <tr class="bg-raise text-12 text-ink-2">
            <th class="border-b border-line-2 px-3 py-1.5 text-left font-medium">Name</th>
            <th class="border-b border-line-2 px-3 py-1.5 text-left font-medium">Columns</th>
            <th class="border-b border-line-2 px-3 py-1.5 text-left font-medium">Unique</th>
          </tr>
        </thead>
        <tbody>
          {indexes.map((idx) => (
            <tr key={idx.name} class="border-b border-line-2">
              <td class="whitespace-nowrap px-3 py-1.5 font-mono text-12 text-ink">{idx.name}</td>
              <td class="whitespace-nowrap px-3 py-1.5 font-mono text-12 text-ink">
                {idx.columns.join(", ")}
              </td>
              <td class="whitespace-nowrap px-3 py-1.5 text-ink">{idx.unique ? "Yes" : "No"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
