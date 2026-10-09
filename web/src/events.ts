import { applySettings, type WispSettings } from "./settings";
import { emitExportEvent } from "./editor/export-events";
import { emitQueryEvent } from "./editor/query-events";
import { emitSchemaInvalidated } from "./editor/schema-events";
import { handleSessionSSE } from "./session-workspace";
import { goMemoryMB, wispStopped } from "./state";

const maxBackoffMs = 20_000;
let source: EventSource | null = null;
let reconnectTimer: number | null = null;
let backoffMs = 1000;
let reconnectAttempts = 0;

function clearReconnectTimer(): void {
  if (reconnectTimer !== null) {
    window.clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
}

function scheduleReconnect(): void {
  clearReconnectTimer();
  reconnectAttempts++;
  const wait = Math.min(backoffMs, maxBackoffMs);
  reconnectTimer = window.setTimeout(() => {
    backoffMs = Math.min(backoffMs * 2, maxBackoffMs);
    openSource();
  }, wait);
}

function handleMessage(raw: string): void {
  let parsed: { type?: string; data?: { go_mb?: number } & WispSettings };
  try {
    parsed = JSON.parse(raw) as { type?: string; data?: { go_mb?: number } & WispSettings };
  } catch {
    return;
  }
  switch (parsed.type) {
    case "memory":
      if (typeof parsed.data?.go_mb === "number") {
        goMemoryMB.value = parsed.data.go_mb;
      }
      break;
    case "settings":
      if (parsed.data) {
        applySettings(parsed.data);
      }
      break;
    case "shutdown":
      wispStopped.value = true;
      disconnectEvents();
      break;
    case "session":
      if (parsed.data && typeof parsed.data === "object") {
        handleSessionSSE(parsed.data as { sid?: string; state?: string });
      }
      break;
    case "schema":
      if (parsed.data && typeof parsed.data === "object") {
        const sid = (parsed.data as { sid?: string }).sid;
        if (sid) {
          emitSchemaInvalidated(sid);
        }
      }
      break;
    case "query":
      if (parsed.data && typeof parsed.data === "object") {
        emitQueryEvent(parsed.data as import("./editor/query-events").QuerySSEPayload);
      }
      break;
    case "export":
      if (parsed.data && typeof parsed.data === "object") {
        emitExportEvent(parsed.data as import("./editor/export-events").ExportSSEPayload);
      }
      break;
    default:
      break;
  }
}

function openSource(): void {
  if (source) {
    source.close();
  }
  source = new EventSource("/api/v1/events", { withCredentials: true });
  source.onopen = () => {
    backoffMs = 1000;
    reconnectAttempts = 0;
    wispStopped.value = false;
  };
  source.onmessage = (ev) => {
    backoffMs = 1000;
    reconnectAttempts = 0;
    handleMessage(ev.data);
  };
  source.onerror = () => {
    source?.close();
    source = null;
    scheduleReconnect();
  };
}

export function connectEvents(): void {
  if (source) {
    return;
  }
  wispStopped.value = false;
  backoffMs = 1000;
  reconnectAttempts = 0;
  openSource();
}

export function disconnectEvents(): void {
  clearReconnectTimer();
  if (source) {
    source.close();
    source = null;
  }
}
