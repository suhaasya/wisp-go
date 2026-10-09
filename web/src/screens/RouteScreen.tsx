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

function LazyWorkspace({ sessionId }: { sessionId: string }) {
  const [Screen, setScreen] = useState<ComponentType<{ sessionId: string }> | null>(null);
  useEffect(() => {
    let cancelled = false;
    import("./workspace").then((mod) => {
      if (!cancelled) {
        setScreen(() => mod.WorkspaceScreen);
      }
    });
    return () => {
      cancelled = true;
    };
  }, []);
  if (!Screen) {
    return <div class="flex flex-1 bg-canvas" aria-busy="true" />;
  }
  return <Screen sessionId={sessionId} />;
}

function LazyGridBench() {
  const [Screen, setScreen] = useState<ComponentType | null>(null);
  useEffect(() => {
    let cancelled = false;
    import("../grid/GridBench").then((mod) => {
      if (!cancelled) {
        setScreen(() => mod.GridBench);
      }
    });
    return () => {
      cancelled = true;
    };
  }, []);
  if (!Screen) {
    return <div class="flex flex-1 bg-canvas" aria-busy="true" />;
  }
  return (
    <div class="flex min-h-0 flex-1 flex-col p-2">
      <Screen />
    </div>
  );
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
      return <LazyWorkspace sessionId={r.sessionId} />;
    case "bench-grid":
      return <LazyGridBench />;
    default:
      return null;
  }
}
