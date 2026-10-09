import { effect } from "@preact/signals";
import { useEffect } from "preact/hooks";
import {
  activeWorkspaceTabId,
  currentSession,
  sessionEngine,
  workspaceHydratedFromServer,
  workspaceProfileId,
  workspaceTabs,
} from "../state";
import { refreshSessionInfo } from "../session-workspace";
import { DataTab } from "../grid/DataTab";
import { StructureTab } from "./structure";
import { QueryTab } from "./QueryTab";
import { SchemaSidebar } from "./sidebar";
import {
  activateTabAtIndex,
  activateWorkspaceTab,
  closeActiveTab,
  closeWorkspaceTab,
  newQueryTab,
  openWorkspaceTableTab,
  reorderWorkspaceTabs,
  tabTitle,
  togglePinTab,
} from "./tabs";
import { applyWorkspaceSnapshot, fetchWorkspace, scheduleWorkspaceSave } from "./workspace-client";

export type WorkspaceShellProps = {
  sessionId: string;
};

function isAppDisplayMode(): boolean {
  return (
    typeof window !== "undefined" &&
    window.matchMedia("(display-mode: standalone)").matches
  );
}

export function WorkspaceShell({ sessionId }: WorkspaceShellProps) {
  useEffect(() => {
    currentSession.value = sessionId;
  }, [sessionId]);

  useEffect(() => {
    let cancelled = false;
    const tick = () => {
      void refreshSessionInfo(sessionId).catch(() => {
        if (!cancelled) {
          // keep last known state
        }
      });
    };
    tick();
    const id = window.setInterval(tick, 60_000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, [sessionId]);

  const profileId = workspaceProfileId.value;

  useEffect(() => {
    if (!profileId) {
      return;
    }
    let cancelled = false;
    void fetchWorkspace(profileId)
      .then((snap) => {
        if (cancelled) {
          return;
        }
        applyWorkspaceSnapshot(snap);
        workspaceHydratedFromServer.value = true;
      })
      .catch(() => {
        workspaceHydratedFromServer.value = true;
      });
    return () => {
      cancelled = true;
    };
  }, [profileId]);

  useEffect(() => {
    if (!profileId) {
      return;
    }
    return effect(() => {
      workspaceTabs.value;
      activeWorkspaceTabId.value;
      if (workspaceHydratedFromServer.value) {
        scheduleWorkspaceSave(profileId);
      }
    });
  }, [profileId]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (!e.altKey || e.metaKey || e.ctrlKey) {
        if (isAppDisplayMode() && e.ctrlKey && e.key === "Tab") {
          e.preventDefault();
          cycleTabs(e.shiftKey ? -1 : 1);
        }
        return;
      }
      const k = e.key.toLowerCase();
      if (k === "t") {
        e.preventDefault();
        const tab = newQueryTab();
        workspaceTabs.value = [...workspaceTabs.value, tab];
        activateWorkspaceTab(tab.id);
        return;
      }
      if (k === "w") {
        e.preventDefault();
        closeActiveTab();
        return;
      }
      if (k >= "1" && k <= "9") {
        e.preventDefault();
        activateTabAtIndex(Number(k) - 1);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  const tabs = workspaceTabs.value;
  const activeId = activeWorkspaceTabId.value;
  const activeTab = tabs.find((t) => t.id === activeId);

  return (
    <div class="flex min-h-0 flex-1 overflow-hidden">
      <SchemaSidebar
        sessionId={sessionId}
        profileId={profileId}
        engine={sessionEngine.value}
        onOpenTable={openWorkspaceTableTab}
      />
      <div class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden bg-canvas">
        {tabs.length > 0 ? (
          <div
            class="flex h-[34px] shrink-0 overflow-x-auto border-b border-line bg-sidebar"
            role="tablist"
          >
            {tabs.map((tab, index) => {
              const active = tab.id === activeId;
              return (
                <div
                  key={tab.id}
                  role="tab"
                  aria-selected={active}
                  draggable
                  class={
                    active
                      ? "flex shrink-0 cursor-pointer items-center gap-2 border-r border-t-2 border-t-accent border-line bg-canvas px-3.5 text-12-5 font-medium text-ink"
                      : "flex shrink-0 cursor-pointer items-center gap-2 border-r border-line px-3.5 text-12-5 text-ink-2 hover:bg-raise"
                  }
                  onClick={() => activateWorkspaceTab(tab.id)}
                  onDragStart={(e) => {
                    e.dataTransfer?.setData("text/plain", String(index));
                  }}
                  onDragOver={(e) => e.preventDefault()}
                  onDrop={(e) => {
                    e.preventDefault();
                    const from = Number(e.dataTransfer?.getData("text/plain"));
                    if (!Number.isNaN(from)) {
                      reorderWorkspaceTabs(from, index);
                    }
                  }}
                >
                  <span class="max-w-[140px] truncate">{tabTitle(tab)}</span>
                  <button
                    type="button"
                    class={`text-11 ${tab.pinned ? "text-accent" : "text-ink-3 hover:text-ink"}`}
                    aria-label={tab.pinned ? "Unpin tab" : "Pin tab"}
                    aria-pressed={tab.pinned}
                    onClick={(e) => {
                      e.stopPropagation();
                      togglePinTab(tab.id);
                    }}
                  >
                    P
                  </button>
                  <button
                    type="button"
                    class="text-ink-3 hover:text-ink"
                    aria-label="Close tab"
                    onClick={(e) => {
                      e.stopPropagation();
                      closeWorkspaceTab(tab.id);
                    }}
                  >
                    ×
                  </button>
                </div>
              );
            })}
          </div>
        ) : null}
        <div class="flex min-h-0 flex-1 flex-col overflow-hidden">
          {activeTab?.kind === "table_data" ? (
            <DataTab key={activeTab.id} sessionId={sessionId} tab={activeTab} />
          ) : activeTab?.kind === "table_structure" ? (
            <StructureTab
              key={activeTab.id}
              sessionId={sessionId}
              schema={activeTab.schema}
              table={activeTab.table}
            />
          ) : activeTab?.kind === "query" ? (
            <QueryTab key={activeTab.id} sessionId={sessionId} tab={activeTab} />
          ) : activeTab ? (
            <div class="flex flex-1 flex-col p-6">
              <h2 class="text-15 font-semibold text-ink">{tabTitle(activeTab)}</h2>
              <p class="text-13 text-ink-2">Coming soon.</p>
            </div>
          ) : (
            <div class="flex flex-1 flex-col p-6">
              <h2 class="text-15 font-semibold text-ink">Workspace</h2>
              <p class="text-13 text-ink-2">Select a table in the sidebar or press Alt+T for a query tab.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function cycleTabs(delta: number): void {
  const tabs = workspaceTabs.value;
  if (tabs.length === 0) {
    return;
  }
  const cur = tabs.findIndex((t) => t.id === activeWorkspaceTabId.value);
  const next = (cur + delta + tabs.length) % tabs.length;
  activateWorkspaceTab(tabs[next]!.id);
}

export { openWorkspaceTableTab };
