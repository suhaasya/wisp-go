import { request } from "./api";
import { navigate } from "./router";
import { MOCK_CONNECTIONS, makeMockProfiles } from "./connections/mock-data";

export type ProfileEnv = "local" | "development" | "staging" | "production";

export type ConnectionProfile = {
  id: string;
  name: string;
  folder: string;
  engine: "postgres" | "mysql" | "mariadb";
  transport: { kind: string; host: string; port: number; path: string };
  user: string;
  database: string;
  password_mode: string;
  ssl: {
    mode: string;
    ca_source?: string;
    ca_file?: string;
    client_cert?: string;
    client_key?: string;
  };
  ssh?: {
    enabled?: boolean;
    host: string;
    port: number;
    user: string;
    auth?: string;
    key_path: string;
    config_host?: string;
    jumps?: {
      host: string;
      port: number;
      user: string;
      auth?: string;
      key_path?: string;
      config_host?: string;
    }[];
  };
  env: ProfileEnv;
  read_only: boolean;
  safe_mode: boolean;
  options: {
    allow_public_key_retrieval?: boolean;
    provider?: string;
    extra?: Record<string, string>;
  };
  advanced?: {
    default_schema?: string;
    time_zone?: string;
    statement_timeout_sec?: number;
    keep_alive_sec?: number;
    run_on_connect?: string;
  };
  last_used?: string | null;
  has_password: boolean;
};

export type ConnectionsListResponse = {
  version: number;
  folders: string[];
  connections: ConnectionProfile[];
  warning?: string;
};

export type ConnectionFilters = {
  query: string;
  folder: string | "all";
  env: ProfileEnv | "all";
};

export function useMockConnectionData(): boolean {
  if (import.meta.env.VITE_WISP_LIVE_CONNECTIONS === "1") {
    return false;
  }
  return import.meta.env.MODE === "test" || import.meta.env.DEV;
}

export async function loadConnectionProfiles(): Promise<{
  profiles: ConnectionProfile[];
  folders: string[];
  warning?: string;
}> {
  if (useMockConnectionData()) {
    return { profiles: [...MOCK_CONNECTIONS], folders: mockFolders() };
  }
  const doc = await request<ConnectionsListResponse>("GET", "/api/v1/connections");
  return {
    profiles: doc.connections,
    folders: doc.folders ?? [],
    warning: doc.warning,
  };
}

function mockFolders(): string[] {
  const set = new Set<string>();
  for (const p of MOCK_CONNECTIONS) {
    if (p.folder) {
      set.add(p.folder);
    }
  }
  return [...set];
}

export { makeMockProfiles, MOCK_CONNECTIONS };

export function formatHost(p: ConnectionProfile): string {
  if (p.transport.kind === "socket" || p.transport.kind === "pipe") {
    return p.transport.path || p.transport.host;
  }
  const host = p.transport.host;
  const port = p.transport.port;
  if (port > 0) {
    return `${host}:${port}`;
  }
  return host;
}

export function deriveTags(p: ConnectionProfile): string[] {
  const tags: string[] = [];
  if (p.ssh?.host) {
    tags.push(`SSH via ${p.ssh.host}`);
  }
  if (p.ssl.mode && p.ssl.mode !== "disable") {
    tags.push(`TLS ${p.ssl.mode}`);
  }
  if (p.read_only) {
    tags.push("Read-only");
  }
  if (p.engine === "mysql") {
    tags.push("MySQL");
  }
  if (p.engine === "mariadb") {
    tags.push("MariaDB");
  }
  return tags;
}

export function filterConnections(profiles: ConnectionProfile[], f: ConnectionFilters): ConnectionProfile[] {
  const q = f.query.trim().toLowerCase();
  return profiles.filter((p) => {
    if (f.folder !== "all" && p.folder !== f.folder) {
      return false;
    }
    if (f.env !== "all" && p.env !== f.env) {
      return false;
    }
    if (!q) {
      return true;
    }
    const hay = [p.name, formatHost(p), ...deriveTags(p)].join(" ").toLowerCase();
    return hay.includes(q);
  });
}

export type FolderGroup = { folder: string; profiles: ConnectionProfile[] };

export function groupByFolder(profiles: ConnectionProfile[], folderOrder: string[]): FolderGroup[] {
  const byFolder = new Map<string, ConnectionProfile[]>();
  for (const p of profiles) {
    const key = p.folder || "Other";
    const list = byFolder.get(key) ?? [];
    list.push(p);
    byFolder.set(key, list);
  }
  const order = [...folderOrder];
  for (const key of byFolder.keys()) {
    if (!order.includes(key)) {
      order.push(key);
    }
  }
  if (byFolder.has("Other") && !order.includes("Other")) {
    order.push("Other");
  }
  return order
    .filter((folder) => byFolder.has(folder))
    .map((folder) => ({ folder, profiles: byFolder.get(folder)! }));
}

export function countByFolder(profiles: ConnectionProfile[]): Map<string, number> {
  const m = new Map<string, number>();
  m.set("all", profiles.length);
  for (const p of profiles) {
    const key = p.folder || "Other";
    m.set(key, (m.get(key) ?? 0) + 1);
  }
  return m;
}

export function countByEnv(profiles: ConnectionProfile[]): Map<ProfileEnv, number> {
  const m = new Map<ProfileEnv, number>();
  for (const p of profiles) {
    m.set(p.env, (m.get(p.env) ?? 0) + 1);
  }
  return m;
}

const envLabels: Record<ProfileEnv, string> = {
  local: "Local",
  development: "Development",
  staging: "Staging",
  production: "Production",
};

export function envLabel(env: ProfileEnv): string {
  return envLabels[env];
}

export function envBorderClass(env: ProfileEnv): string {
  switch (env) {
    case "local":
      return "before:bg-env-local";
    case "development":
      return "before:bg-env-dev";
    case "staging":
      return "before:bg-env-staging";
    case "production":
      return "before:bg-env-prod";
  }
}

export function engineBadgeClass(engine: ConnectionProfile["engine"]): string {
  switch (engine) {
    case "postgres":
      return "bg-badge-pg";
    case "mysql":
      return "bg-badge-my";
    case "mariadb":
      return "bg-badge-mdb";
  }
}

export function engineAbbrev(engine: ConnectionProfile["engine"]): string {
  switch (engine) {
    case "postgres":
      return "PG";
    case "mysql":
      return "MY";
    case "mariadb":
      return "MDB";
  }
}

export function formatLastUsed(iso: string | null | undefined, nowMs = Date.now()): string | null {
  if (!iso) {
    return null;
  }
  const t = Date.parse(iso);
  if (Number.isNaN(t)) {
    return null;
  }
  const diff = nowMs - t;
  const min = Math.floor(diff / 60_000);
  if (min < 60) {
    return min <= 1 ? "1 min ago" : `${min} min ago`;
  }
  const hr = Math.floor(min / 60);
  if (hr < 24) {
    return hr === 1 ? "1 h ago" : `${hr} h ago`;
  }
  const day = Math.floor(hr / 24);
  if (day === 1) {
    return "yesterday";
  }
  return `${day} days ago`;
}

export async function connectProfile(profileId: string): Promise<string> {
  if (useMockConnectionData()) {
    return `mock-${profileId}`;
  }
  const out = await request<{ sid: string }>("POST", "/api/v1/sessions", { profile_id: profileId });
  return out.sid;
}

export async function openWorkspaceSession(profileId: string): Promise<void> {
  const sid = await connectProfile(profileId);
  navigate(`/w/${encodeURIComponent(sid)}`);
}

export async function deleteConnectionProfile(id: string): Promise<void> {
  if (useMockConnectionData()) {
    return;
  }
  await request("DELETE", `/api/v1/connections/${encodeURIComponent(id)}`);
}

export async function duplicateConnectionProfile(id: string): Promise<ConnectionProfile> {
  if (useMockConnectionData()) {
    const src = MOCK_CONNECTIONS.find((p) => p.id === id);
    if (!src) {
      throw new Error("not found");
    }
    return {
      ...src,
      id: `mock-dup-${id}`,
      name: `${src.name} (copy)`,
    };
  }
  return request<ConnectionProfile>("POST", `/api/v1/connections/${encodeURIComponent(id)}/duplicate`);
}

export async function moveConnectionToFolder(
  profile: ConnectionProfile,
  folder: string,
): Promise<void> {
  if (useMockConnectionData()) {
    return;
  }
  await request("PUT", `/api/v1/connections/${encodeURIComponent(profile.id)}`, {
    ...profile,
    folder,
  });
}
