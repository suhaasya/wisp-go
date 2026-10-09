import { useEffect } from "preact/hooks";
import {
  goMemoryMB,
  latencyMs,
  readOnly,
  serverVersion,
  sessionState,
  shellConnection,
  tabMemoryMB,
} from "../state";
import { MemoryMeter } from "./MemoryMeter";

const envDotClass = {
  local: "bg-env-local",
  dev: "bg-env-dev",
  staging: "bg-env-staging",
  prod: "bg-env-prod",
} as const;

const goCapMB = 30;
const tabCapMB = 10;

function sampleTabMemory(): void {
  const perf = performance as Performance & { memory?: { usedJSHeapSize: number } };
  if (perf.memory?.usedJSHeapSize) {
    tabMemoryMB.value = perf.memory.usedJSHeapSize / (1024 * 1024);
  }
}

export function StatusBar() {
  useEffect(() => {
    sampleTabMemory();
    const id = window.setInterval(sampleTabMemory, 2000);
    return () => window.clearInterval(id);
  }, []);

  const conn = shellConnection.value;
  const env = conn?.env;

  return (
    <footer class="flex h-[26px] shrink-0 items-center gap-4 border-t border-line bg-sidebar px-3 text-11-5 text-ink-2">
      <span class="inline-flex items-center gap-2">
        <span
          class={`inline-block h-2 w-2 rounded-full ${env ? envDotClass[env] : "bg-line"}`}
          aria-hidden
        />
        {sessionState.value === "reconnecting" ? (
          <span class="text-env-staging">Reconnecting…</span>
        ) : null}
        {readOnly.value ? <span>Read-only</span> : null}
        <span>v{serverVersion.value}</span>
        {latencyMs.value !== null ? <span>{latencyMs.value} ms</span> : null}
      </span>
      <span class="flex-1" />
      {goMemoryMB.value !== null ? (
        <MemoryMeter label="Go" mb={goMemoryMB.value} capMB={goCapMB} mono />
      ) : null}
      {tabMemoryMB.value !== null ? (
        <MemoryMeter label="Tab" mb={tabMemoryMB.value} capMB={tabCapMB} mono />
      ) : null}
    </footer>
  );
}
