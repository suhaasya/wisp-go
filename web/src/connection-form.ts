import type { ConnectionProfile, ProfileEnv } from "./connections";
import { defaultSSLForHost, type SSLFormState } from "./connection-form/ssl-helpers";
import { defaultSSHForm, type SSHFormState } from "./connection-form/ssh-helpers";

export type { SSLFormState, SSHFormState };

export type FormState = {
  id?: string;
  name: string;
  folder: string;
  engine: ConnectionProfile["engine"];
  transportKind: "tcp" | "socket" | "pipe";
  host: string;
  port: number;
  portTouched: boolean;
  user: string;
  database: string;
  password: string;
  passwordDirty: boolean;
  savePassword: boolean;
  env: ProfileEnv;
  readOnly: boolean;
  safeMode: boolean;
  ssl: SSLFormState;
  ssh: SSHFormState;
  advanced: {
    defaultSchema: string;
    timeZone: string;
    statementTimeoutSec: string;
    keepAliveSec: string;
    runOnConnect: string;
  };
};

export type FieldErrors = Partial<Record<keyof FormState | "productionSafe", string>>;

const defaultPorts: Record<ConnectionProfile["engine"], number> = {
  postgres: 5432,
  mysql: 3306,
  mariadb: 3306,
};

export function defaultFormState(): FormState {
  return {
    name: "",
    folder: "",
    engine: "postgres",
    transportKind: "tcp",
    host: "127.0.0.1",
    port: 5432,
    portTouched: false,
    user: "",
    database: "",
    password: "",
    passwordDirty: false,
    savePassword: true,
    env: "local",
    readOnly: false,
    safeMode: false,
    ssl: defaultSSLForHost("127.0.0.1"),
    ssh: defaultSSHForm(),
    advanced: {
      defaultSchema: "",
      timeZone: "",
      statementTimeoutSec: "",
      keepAliveSec: "",
      runOnConnect: "",
    },
  };
}

export function formFromParseResult(profile: Record<string, unknown>, password: string): FormState {
	const engine = (profile.engine as ConnectionProfile["engine"]) ?? "postgres";
	const transport = (profile.transport as ConnectionProfile["transport"]) ?? {
		kind: "tcp",
		host: "127.0.0.1",
		port: defaultPorts[engine],
		path: "",
	};
	const sslRaw = (profile.ssl as ConnectionProfile["ssl"]) ?? { mode: "prefer" };
	const options = (profile.options as { provider?: string }) ?? {};
	const host = transport.host ?? "";
	const advanced = (profile.advanced as Record<string, unknown>) ?? {};
	const sslDefaults = defaultSSLForHost(host, options.provider);
	return {
		name: String(profile.name ?? profile.database ?? ""),
		folder: String(profile.folder ?? ""),
		engine,
		transportKind: (transport.kind as FormState["transportKind"]) || "tcp",
		host: transport.host ?? "",
		port: transport.port || defaultPorts[engine],
		portTouched: true,
		user: String(profile.user ?? ""),
		database: String(profile.database ?? ""),
		password,
		passwordDirty: password.length > 0,
		savePassword: profile.password_mode !== "none",
		env: (profile.env as ProfileEnv) ?? "development",
		readOnly: Boolean(profile.read_only),
		safeMode: Boolean(profile.safe_mode),
		ssh: defaultSSHForm(),
		ssl: {
			mode: sslRaw.mode || sslDefaults.mode,
			caSource: (sslRaw.ca_source as SSLFormState["caSource"]) || sslDefaults.caSource,
			caFile: sslRaw.ca_file ?? "",
			clientCert: sslRaw.client_cert ?? "",
			clientKey: sslRaw.client_key ?? "",
		},
		advanced: {
			defaultSchema: String(advanced.default_schema ?? ""),
			timeZone: String(advanced.time_zone ?? ""),
			statementTimeoutSec: advanced.statement_timeout_sec ? String(advanced.statement_timeout_sec) : "",
			keepAliveSec: advanced.keep_alive_sec ? String(advanced.keep_alive_sec) : "",
			runOnConnect: String(advanced.run_on_connect ?? ""),
		},
	};
}

export function formFromProfile(p: ConnectionProfile): FormState {
  return {
    id: p.id,
    name: p.name,
    folder: p.folder,
    engine: p.engine,
    transportKind: (p.transport.kind as FormState["transportKind"]) || "tcp",
    host: p.transport.host,
    port: p.transport.port || defaultPorts[p.engine],
    portTouched: true,
    user: p.user,
    database: p.database,
    password: "",
    passwordDirty: false,
    savePassword: p.password_mode !== "none",
    env: p.env,
    readOnly: p.read_only,
    safeMode: p.safe_mode,
    ssl: sslFromProfile(p),
    ssh: sshFromProfile(p),
    advanced: {
      defaultSchema: p.advanced?.default_schema ?? "",
      timeZone: p.advanced?.time_zone ?? "",
      statementTimeoutSec: p.advanced?.statement_timeout_sec ? String(p.advanced.statement_timeout_sec) : "",
      keepAliveSec: p.advanced?.keep_alive_sec ? String(p.advanced.keep_alive_sec) : "",
      runOnConnect: p.advanced?.run_on_connect ?? "",
    },
  };
}

export function validateConnectionForm(f: FormState): FieldErrors {
  const err: FieldErrors = {};
  if (!f.name.trim()) {
    err.name = "Name is required";
  }
  if (f.transportKind === "tcp" && !f.host.trim()) {
    err.host = "Host is required";
  }
  if (f.transportKind === "tcp" && (f.port < 1 || f.port > 65535)) {
    err.port = "Port must be 1–65535";
  }
  if (!f.user.trim()) {
    err.user = "User is required";
  }
  if (!f.database.trim()) {
    err.database = "Database is required";
  }
  if (f.env === "production" && !f.safeMode) {
    err.productionSafe = "Production connections should enable Confirm every write";
  }
  return err;
}

export function applyEngineDefaultPort(f: FormState, engine: ConnectionProfile["engine"]): FormState {
  if (f.portTouched) {
    return { ...f, engine };
  }
  return { ...f, engine, port: defaultPorts[engine] };
}

export type TestRequestBody = Record<string, unknown>;

export function buildTestRequest(f: FormState, hasStoredPassword: boolean): TestRequestBody {
  const body: TestRequestBody = {
    profile_id: f.id,
    name: f.name,
    folder: f.folder,
    engine: f.engine,
    transport: { kind: f.transportKind, host: f.host, port: f.port, path: "" },
    user: f.user,
    database: f.database,
    password_mode: f.savePassword ? "keychain" : "none",
    ssl: buildSSLPayload(f.ssl),
    env: f.env,
    read_only: f.readOnly,
    safe_mode: f.safeMode,
    options: {},
    advanced: buildAdvancedPayload(f),
  };
  if (f.ssh.enabled) {
    body.ssh = buildSSHPayload(f.ssh);
  }
  if (f.passwordDirty || (!f.id && f.password) || (f.id && !hasStoredPassword && f.password)) {
    body.password = f.password;
  }
  if (f.ssh.passphraseDirty && f.ssh.passphrase) {
    body.ssh_passphrase = f.ssh.passphrase;
  }
  if (f.ssh.passwordDirty && f.ssh.password) {
    body.ssh_password = f.ssh.password;
  }
  return body;
}

export function buildSaveRequest(f: FormState, hasStoredPassword: boolean): TestRequestBody {
  const body = buildTestRequest(f, hasStoredPassword);
  delete body.profile_id;
  return body;
}

function sslFromProfile(p: ConnectionProfile): SSLFormState {
  const base = defaultSSLForHost(p.transport.host, p.options.provider);
  return {
    mode: p.ssl.mode || base.mode,
    caSource: (p.ssl.ca_source as SSLFormState["caSource"]) || base.caSource,
    caFile: p.ssl.ca_file ?? "",
    clientCert: p.ssl.client_cert ?? "",
    clientKey: p.ssl.client_key ?? "",
  };
}

function sshFromProfile(p: ConnectionProfile): SSHFormState {
  const base = defaultSSHForm();
  if (!p.ssh) {
    return base;
  }
  return {
    enabled: Boolean(p.ssh.enabled) || Boolean(p.ssh.host),
    host: p.ssh.host ?? "",
    port: p.ssh.port || 22,
    user: p.ssh.user ?? "",
    auth: p.ssh.auth || "key",
    keyPath: p.ssh.key_path ?? "",
    configHost: p.ssh.config_host ?? "",
    passphrase: "",
    passphraseDirty: false,
    password: "",
    passwordDirty: false,
    jumps: (p.ssh.jumps ?? []).map((j) => ({
      host: j.host,
      port: j.port || 22,
      user: j.user,
      auth: j.auth || "key",
      keyPath: j.key_path ?? "",
      configHost: j.config_host ?? "",
    })),
  };
}

function buildSSHPayload(ssh: SSHFormState): Record<string, unknown> {
  const out: Record<string, unknown> = {
    enabled: true,
    host: ssh.host,
    port: ssh.port || 22,
    user: ssh.user,
    auth: ssh.auth,
    key_path: ssh.keyPath,
  };
  if (ssh.configHost.trim()) {
    out.config_host = ssh.configHost.trim();
  }
  if (ssh.jumps.length > 0) {
    out.jumps = ssh.jumps.map((j) => ({
      host: j.host,
      port: j.port || 22,
      user: j.user,
      auth: j.auth,
      key_path: j.keyPath,
      config_host: j.configHost || undefined,
    }));
  }
  return out;
}

function buildSSLPayload(ssl: SSLFormState): Record<string, string> {
  const out: Record<string, string> = { mode: ssl.mode };
  if (ssl.caSource === "file" && ssl.caFile.trim()) {
    out.ca_source = "file";
    out.ca_file = ssl.caFile.trim();
  } else if (ssl.caSource === "system") {
    out.ca_source = "system";
  }
  if (ssl.clientCert.trim()) {
    out.client_cert = ssl.clientCert.trim();
  }
  if (ssl.clientKey.trim()) {
    out.client_key = ssl.clientKey.trim();
  }
  return out;
}

function buildAdvancedPayload(f: FormState) {
  const a = f.advanced;
  const out: Record<string, unknown> = {};
  if (a.defaultSchema.trim()) {
    out.default_schema = a.defaultSchema.trim();
  }
  if (a.timeZone.trim()) {
    out.time_zone = a.timeZone.trim();
  }
  if (a.statementTimeoutSec.trim()) {
    out.statement_timeout_sec = Number(a.statementTimeoutSec);
  }
  if (a.keepAliveSec.trim()) {
    out.keep_alive_sec = Number(a.keepAliveSec);
  }
  if (a.runOnConnect.trim()) {
    out.run_on_connect = a.runOnConnect;
  }
  return out;
}

export function formatTestSuccess(latencyMs: number, server: string, tls?: string, viaSSH?: boolean): string {
  let s = `Connected in ${latencyMs} ms · ${server}`;
  if (tls) {
    s += `, ${tls}`;
  }
  if (viaSSH) {
    s += ", via SSH";
  }
  return s;
}
