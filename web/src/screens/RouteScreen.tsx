import { useEffect, useState } from "preact/hooks";
import type { ComponentType } from "preact";
import type { Route } from "../router";

function LazyConnectionForm({ profileId }: { profileId?: string }) {
  const [Screen, setScreen] = useState<ComponentType<{ profileId?: string }> | null>(null);
  useEffect(() => {
    let cancelled = false;
    import("./connection-form").then((mod) => {
      if (!cancelled) {
        setScreen(() => mod.ConnectionForm);
      }
    });
    return () => {
      cancelled = true;
    };
  }, []);
  if (!Screen) {
    return <div class="flex flex-1 bg-canvas" aria-busy="true" />;
  }
  return <Screen profileId={profileId} />;
}

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
      return <LazyConnectionForm />;
    case "connection-edit":
      return <LazyConnectionForm profileId={r.id} />;
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
