import { useEffect, useState } from "preact/hooks";
import type { ComponentType } from "preact";
import { route, startRouter, stopRouter } from "../router";
import { RouteScreen } from "../screens/RouteScreen";
import { handleGlobalKeydown } from "../keydispatch";
import { sessionEnded, wispStopped } from "../state";
import { ToastViewport } from "../ui/Toast";
import { StatusBar } from "./StatusBar";
import { TitleBar } from "./TitleBar";

function SessionEndedScreen() {
  return (
    <div class="flex flex-1 flex-col items-center justify-center gap-3 p-8 text-center">
      <h1 class="text-15 font-semibold text-ink">Session ended</h1>
      <p class="max-w-md text-13 text-ink-2">Relaunch Wisp from your terminal to open a new session.</p>
    </div>
  );
}

function StoppedBanner() {
  return (
    <div
      class="border-b border-env-prod bg-deleted px-4 py-2 text-center text-13 text-ink"
      role="alert"
    >
      Wisp stopped. Run <span class="font-mono">wisp</span> to continue.
    </div>
  );
}

export function AppShell() {
  const [Gallery, setGallery] = useState<ComponentType | null>(null);
  const [Palette, setPalette] = useState<ComponentType | null>(null);
  const [ShortcutHelp, setShortcutHelp] = useState<ComponentType | null>(null);

  useEffect(() => {
    startRouter();
    return () => stopRouter();
  }, []);

  useEffect(() => {
    void import("../palette/index").then((mod) => setPalette(() => mod.CommandPalette));
    void import("../shortcut-help").then((mod) => setShortcutHelp(() => mod.ShortcutHelpOverlay));
  }, []);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => handleGlobalKeydown(e);
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  useEffect(() => {
    if (!import.meta.env.DEV || route.value.name !== "gallery") {
      setGallery(null);
      return;
    }
    let cancelled = false;
    import("../gallery/Gallery").then((mod) => {
      if (!cancelled) {
        setGallery(() => mod.Gallery);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [route.value.name]);

  if (sessionEnded.value) {
    return (
      <div class="flex h-dvh min-h-0 flex-col overflow-hidden bg-canvas text-ink">
        <SessionEndedScreen />
      </div>
    );
  }

  if (import.meta.env.DEV && route.value.name === "gallery" && Gallery) {
    return <Gallery />;
  }

  return (
    <div class="flex h-dvh min-h-0 flex-col overflow-hidden bg-canvas text-ink">
      {wispStopped.value ? <StoppedBanner /> : null}
      <TitleBar />
      <main class="flex min-h-0 flex-1 flex-col overflow-hidden">
        <RouteScreen route={route.value} />
      </main>
      <StatusBar />
      <ToastViewport />
      {Palette ? <Palette /> : null}
      {ShortcutHelp ? <ShortcutHelp /> : null}
    </div>
  );
}
