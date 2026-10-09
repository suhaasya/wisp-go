import { useEffect } from "preact/hooks";
import { refreshSessionInfo } from "../../session-workspace";

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

  return (
    <section class="p-6">
      <h2 class="text-15 font-semibold text-ink">Workspace</h2>
      <p class="text-13 text-ink-2">Session {sessionId}</p>
    </section>
  );
}
