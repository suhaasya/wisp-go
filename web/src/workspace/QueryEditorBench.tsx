import { useState } from "preact/hooks";
import { QueryTab } from "./QueryTab";

const benchTab = {
  id: "bench/query",
  kind: "query" as const,
  title: "Query",
  queryText: "SELECT 1;\nSELECT 2;",
};

/** Memory gate: mount/unmount CodeMirror query editor repeatedly. */
export function QueryEditorBench() {
  const [open, setOpen] = useState(false);

  return (
    <div class="flex min-h-0 flex-1 flex-col bg-canvas" data-testid="query-editor-bench">
      <button
        type="button"
        data-testid="query-editor-toggle"
        class="m-4 w-fit rounded-control border border-line bg-panel px-3 py-1.5 text-13 text-ink"
        onClick={() => setOpen((v) => !v)}
      >
        {open ? "Close editor" : "Open editor"}
      </button>
      {open ? (
        <QueryTab sessionId="bench-session" tab={benchTab} />
      ) : (
        <p class="px-4 text-13 text-ink-2">Editor closed</p>
      )}
    </div>
  );
}
