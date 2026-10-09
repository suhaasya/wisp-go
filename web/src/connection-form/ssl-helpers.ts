export type SSLFormState = {
  mode: string;
  caSource: "system" | "file";
  caFile: string;
  clientCert: string;
  clientKey: string;
};

export function isLocalHost(host: string): boolean {
  const h = host.trim().toLowerCase();
  return h === "127.0.0.1" || h === "localhost" || h === "::1" || h === "";
}

export function defaultSSLForHost(host: string, provider?: string): SSLFormState {
  let mode = "require";
  if (provider) {
    mode = "verify-full";
  } else if (isLocalHost(host)) {
    mode = "disable";
  }
  return {
    mode,
    caSource: "system",
    caFile: "",
    clientCert: "",
    clientKey: "",
  };
}

export function sslDisableWarning(host: string, mode: string): boolean {
  return mode === "disable" && !isLocalHost(host);
}

export const sslModes = [
  { value: "disable", label: "Disable" },
  { value: "prefer", label: "Prefer" },
  { value: "require", label: "Require" },
  { value: "verify-ca", label: "Verify CA" },
  { value: "verify-full", label: "Verify full" },
];
