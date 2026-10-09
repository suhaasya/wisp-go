import { useEffect, useState } from "preact/hooks";
import type { ComponentChild, ComponentType } from "preact";
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

function LazyWorkspaceTabsBench() {
  const [Screen, setScreen] = useState<ComponentType | null>(null);
  useEffect(() => {
    let cancelled = false;
    import("../workspace/WorkspaceTabsBench").then((mod) => {
      if (!cancelled) {
        setScreen(() => mod.WorkspaceTabsBench);
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
  let screen: ComponentChild;
  switch (r.name) {
    case "connections":
      screen = <LazyConnectionsScreen />;
      break;
    case "connection-new":
      screen = <LazyConnectionForm />;
      break;
    case "connection-edit":
      screen = <LazyConnectionForm profileId={r.id} />;
      break;
    case "workspace":
      screen = <LazyWorkspace sessionId={r.sessionId} />;
      break;
    case "bench-grid":
      screen = <LazyGridBench />;
      break;
    case "bench-workspace-tabs":
      screen = <LazyWorkspaceTabsBench />;
      break;
    default:
      screen = null;
  }
  return <div class="flex min-h-0 flex-1 flex-col overflow-hidden">{screen}</div>;
}
