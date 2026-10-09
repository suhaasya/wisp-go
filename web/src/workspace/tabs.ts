import { activeWorkspaceTabId, workspaceTabs, type WorkspaceTableTab } from "../state";
import type { OpenTableRequest } from "./sidebar/tree";

export function openWorkspaceTableTab(req: OpenTableRequest): void {
  const id = `${req.schema}/${req.table}/${req.mode}`;
  const existing = workspaceTabs.value.find((t) => t.id === id);
  if (!existing) {
    const tab: WorkspaceTableTab = {
      id,
      schema: req.schema,
      table: req.table,
      mode: req.mode,
    };
    workspaceTabs.value = [...workspaceTabs.value, tab];
  }
  activeWorkspaceTabId.value = id;
}
