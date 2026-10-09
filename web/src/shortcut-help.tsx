import { useEffect } from "preact/hooks";
import { cheatSheetBindings } from "./keydispatch";
import { shortcutHelpOpen } from "./state";

export function ShortcutHelpOverlay() {
  if (!shortcutHelpOpen.value) {
    return null;
  }

  const rows = cheatSheetBindings();

  function close() {
    shortcutHelpOpen.value = false;
  }

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        close();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  return (
    <div
      class="fixed inset-0 z-50 flex items-start justify-center bg-canvas/80 pt-20"
      role="presentation"
      onClick={(e) => {
        if (e.target === e.currentTarget) {
          close();
        }
      }}
    >
      <div
        class="w-[480px] max-w-[calc(100vw-24px)] overflow-hidden rounded-dialog border border-line bg-panel shadow-lg"
        role="dialog"
        aria-modal="true"
        aria-label="Keyboard shortcuts"
        tabIndex={-1}
      >
        <header class="border-b border-line px-4 py-3">
          <h2 class="text-15 font-semibold text-ink">Keyboard shortcuts</h2>
          <p class="mt-1 text-12-5 text-ink-2">
            Browser shortcuts (⌘/Ctrl+W, T, N…) stay with the browser. Wisp uses Alt for tabs and
            connections.
          </p>
        </header>
        <ul class="max-h-[360px] list-none overflow-auto p-2">
          {rows.map((row) => (
            <li
              key={row.id}
              class="flex items-center gap-2 rounded-control px-2.5 py-1.5 text-13 text-ink"
            >
              <span class="min-w-0 flex-1">{row.label}</span>
              <kbd class="shrink-0 rounded-control border border-line bg-raise px-1.5 py-0.5 font-mono text-12 text-ink-2">
                {row.chord}
              </kbd>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
