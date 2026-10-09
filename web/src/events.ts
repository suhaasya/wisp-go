import { applySettings, type WispSettings } from "./settings";
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
  if (backoffMs >= maxBackoffMs) {
    wispStopped.value = true;
    return;
  }
  reconnectTimer = window.setTimeout(() => {
    backoffMs = Math.min(backoffMs * 2, maxBackoffMs);
    openSource();
  }, backoffMs);
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
    default:
      break;
  }
}

function openSource(): void {
  if (source) {
    source.close();
  }
  source = new EventSource("/api/v1/events", { withCredentials: true });
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
