import { request } from "./api";
import {
  currentSession,
  latencyMs,
  readOnly,
  sessionEngine,
  sessionState,
  shellConnection,
  type ShellConnection,
} from "./state";

export type SessionInfo = {
  state: string;
  engine: string;
  version: string;
  database: string;
  user: string;
  read_only: boolean;
  safe_mode: boolean;
  env: string;
};

function mapEnv(env: string): ShellConnection["env"] {
  switch (env) {
    case "local":
      return "local";
    case "development":
      return "dev";
    case "staging":
      return "staging";
    case "production":
      return "prod";
    default:
      return null;
  }
}

export function applySessionInfo(sid: string, info: SessionInfo): void {
  currentSession.value = sid;
  readOnly.value = info.read_only;
  sessionEngine.value = info.engine || "postgres";
  const env = mapEnv(info.env);
  shellConnection.value = {
    name: info.database,
    subtitle: `${info.user}@${info.engine}`,
    env,
  };
  switch (info.state) {
    case "ready":
    case "busy":
      sessionState.value = "connected";
      break;
    case "connecting":
      sessionState.value = "connecting";
      break;
    case "reconnecting":
      sessionState.value = "reconnecting";
      break;
    case "failed":
      sessionState.value = "error";
      break;
    default:
      sessionState.value = "disconnected";
  }
}

export async function refreshSessionInfo(sid: string): Promise<void> {
  const info = await request<SessionInfo>("GET", `/api/v1/sessions/${encodeURIComponent(sid)}/info`);
  applySessionInfo(sid, info);
  if (latencyMs.value === null) {
    latencyMs.value = 1;
  }
}

export function handleSessionSSE(data: { sid?: string; state?: string }): void {
  const sid = currentSession.value;
  if (!sid || data.sid !== sid || !data.state) {
    return;
  }
  applySessionInfo(sid, {
    state: data.state,
    engine: shellConnection.value?.subtitle.split("@")[1] ?? "",
    version: "",
    database: shellConnection.value?.name ?? "",
    user: shellConnection.value?.subtitle.split("@")[0] ?? "",
    read_only: readOnly.value,
    safe_mode: false,
    env: shellConnection.value?.env ?? "",
  });
}
