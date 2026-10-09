import { useEffect, useState } from "preact/hooks";
import type { ComponentType } from "preact";
import type { Route } from "../router";

function LazyConnectionsScreen() {
  const [Screen, setScreen] = useState<ComponentType | null>(null);
  useEffect(() => {
    let cancelled = false;
    import("./connections").then((mod) => {
      if (!cancelled) {
        setScreen(() => mod.ConnectionsScreen);
      }
    });
    return () => {
      cancelled = true;
    };
  }, []);
  if (!Screen) {
    return <div class="flex flex-1 bg-canvas" aria-busy="true" />;
  }
  return <Screen />;
}

export function RouteScreen({ route: r }: { route: Route }) {
  switch (r.name) {
    case "connections":
      return <LazyConnectionsScreen />;
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
