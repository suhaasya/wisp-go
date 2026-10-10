import { useEffect } from "preact/hooks";
import { activeWorkspaceTabId, workspaceTabs } from "../state";
import { handleHorizontalTablistKeydown } from "../ui/tablist-keys";
import { activateWorkspaceTab, tabTitle } from "./tabs";

/** Memory gate: 20 workspace tabs with only the active pane mounted. */
export function WorkspaceTabsBench() {
  useEffect(() => {
    workspaceTabs.value = Array.from({ length: 20 }, (_, i) => ({
      id: `bench/t${i}/data`,
      kind: "table_data" as const,
      schema: "public",
      table: `table_${i}`,
      pendingResultLoad: true,
    }));
    activeWorkspaceTabId.value = workspaceTabs.value[0]!.id;
  }, []);

  const tabs = workspaceTabs.value;
  const activeId = activeWorkspaceTabId.value;

  return (
    <div class="flex min-h-0 flex-1 flex-col" data-testid="workspace-tabs-bench">
      <div
        class="flex h-[34px] shrink-0 border-b border-line bg-sidebar"
        role="tablist"
        onKeyDown={(e) =>
          handleHorizontalTablistKeydown(e, (index) => {
            const tab = tabs[index];
            if (tab) {
              activateWorkspaceTab(tab.id);
            }
          })
        }
      >
        {tabs.map((tab) => (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={tab.id === activeId}
            tabIndex={tab.id === activeId ? 0 : -1}
            class="border-r border-line px-3 text-12-5 text-ink-2"
            onClick={() => activateWorkspaceTab(tab.id)}
          >
            {tabTitle(tab)}
          </button>
        ))}
      </div>
      <div class="flex flex-1 items-center justify-center text-13 text-ink-2" data-testid="workspace-tab-pane">
        Active: {activeId}
      </div>
      <span data-testid="workspace-tab-count" class="sr-only">
        {tabs.length}
      </span>
    </div>
  );
}
