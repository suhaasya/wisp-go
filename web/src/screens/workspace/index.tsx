import { useEffect } from "preact/hooks";
import { activeWorkspaceTabId, sessionEngine, workspaceProfileId, workspaceTabs } from "../../state";
import { refreshSessionInfo } from "../../session-workspace";
import { SchemaSidebar } from "../../workspace/sidebar";
import { openWorkspaceTableTab } from "../../workspace/tabs";

export function WorkspaceScreen({ sessionId }: { sessionId: string }) {
  useEffect(() => {
    let cancelled = false;
    const tick = () => {
      void refreshSessionInfo(sessionId).catch(() => {
        if (!cancelled) {
          // keep last known state; errors surface via toast in a later ticket
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

  const tabs = workspaceTabs.value;
  const activeId = activeWorkspaceTabId.value;

  return (
    <div class="flex min-h-0 flex-1">
      <SchemaSidebar
        sessionId={sessionId}
        profileId={workspaceProfileId.value}
        engine={sessionEngine.value}
        onOpenTable={openWorkspaceTableTab}
      />
      <div class="flex min-w-0 flex-1 flex-col bg-canvas">
        {tabs.length > 0 ? (
          <div class="flex gap-1 border-b border-line bg-panel px-2 py-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                type="button"
                class={
                  tab.id === activeId
                    ? "rounded-control bg-accent-soft px-2 py-0.5 text-12-5 font-medium text-ink"
                    : "rounded-control px-2 py-0.5 text-12-5 text-ink-2 hover:bg-raise"
                }
                onClick={() => {
                  activeWorkspaceTabId.value = tab.id;
                }}
              >
                {tab.table}
                {tab.mode === "structure" ? " (structure)" : ""}
              </button>
            ))}
          </div>
        ) : null}
        <div class="flex flex-1 flex-col p-6">
          <h2 class="text-15 font-semibold text-ink">Workspace</h2>
          <p class="text-13 text-ink-2">
            {activeId
              ? `Open tab: ${tabs.find((t) => t.id === activeId)?.table ?? activeId}`
              : `Session ${sessionId} — select a table in the sidebar`}
          </p>
        </div>
      </div>
    </div>
  );
}
