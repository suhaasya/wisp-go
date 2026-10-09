import type { Route } from "../router";

export function RouteScreen({ route: r }: { route: Route }) {
  switch (r.name) {
    case "connections":
      return (
        <section class="p-6">
          <h2 class="text-15 font-semibold text-ink">Connections</h2>
          <p class="mt-2 text-13 text-ink-2">Connection list (LW-010+).</p>
        </section>
      );
    case "connection-new":
      return (
        <section class="p-6">
          <h2 class="text-15 font-semibold text-ink">New connection</h2>
        </section>
      );
    case "connection-edit":
      return (
        <section class="p-6">
          <h2 class="text-15 font-semibold text-ink">Edit connection</h2>
          <p class="text-13 text-ink-2">{r.id}</p>
        </section>
      );
    case "workspace":
      return (
        <section class="p-6">
          <h2 class="text-15 font-semibold text-ink">Workspace</h2>
          <p class="text-13 text-ink-2">Session {r.sessionId}</p>
        </section>
      );
    default:
      return null;
  }
}
