import { Segmented } from "../../ui/Segmented";
import { openWorkspaceTableTab } from "../tabs";
import type { OpenTableRequest } from "../sidebar/tree";

export type TableViewMode = "data" | "structure";

export function TableModeToolbar({
  schema,
  table,
  mode,
}: {
  schema: string;
  table: string;
  mode: TableViewMode;
}) {
  return (
    <div class="toolbar flex shrink-0 items-center gap-2 border-b border-line px-2.5 py-2">
      <Segmented
        aria-label="Table view"
        value={mode}
        options={[
          { value: "data", label: "Data" },
          { value: "structure", label: "Structure" },
        ]}
        onChange={(next) => {
          if (next === mode) {
            return;
          }
          const req: OpenTableRequest = { schema, table, mode: next as TableViewMode };
          openWorkspaceTableTab(req);
        }}
      />
    </div>
  );
}
