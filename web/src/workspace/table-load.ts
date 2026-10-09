import type { WorkspaceTab } from "../state";

export function shouldLoadTableResult(tab: WorkspaceTab): boolean {
  return tab.kind === "table_data" && tab.pendingResultLoad !== true;
}
