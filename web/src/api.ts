import {
  csrfToken,
  serverVersion,
  sessionEnded,
  settings,
} from "./state";

export type ApiErrorBody = {
  code: string;
  message: string;
  detail?: string;
};

export class ApiError extends Error {
  code: string;
  detail?: string;

  constructor(body: ApiErrorBody) {
    super(body.message);
    this.name = "ApiError";
    this.code = body.code;
    this.detail = body.detail;
  }
}

type HelloResponse = {
  version: string;
  csrf: string;
  settings: Record<string, unknown>;
};

let bootstrapped = false;

export async function bootstrap(): Promise<HelloResponse> {
  const res = await fetch("/api/v1/hello", { credentials: "same-origin" });
  if (res.status === 401) {
    sessionEnded.value = true;
    throw new Error("session ended");
  }
  if (!res.ok) {
    throw new Error(`hello failed: ${res.status}`);
  }
  const body = (await res.json()) as HelloResponse;
  csrfToken.value = body.csrf;
  serverVersion.value = body.version;
  settings.value = body.settings ?? {};
  bootstrapped = true;
  return body;
}

export async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  opts?: { signal?: AbortSignal },
): Promise<T> {
  if (!bootstrapped && path !== "/api/v1/hello") {
    await bootstrap();
  }
  const headers: Record<string, string> = {
    Accept: "application/json",
  };
  if (csrfToken.value) {
    headers["X-Wisp-CSRF"] = csrfToken.value;
  }
  let payload: string | undefined;
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
    payload = JSON.stringify(body);
  }
  const res = await fetch(path, {
    method,
    credentials: "same-origin",
    headers,
    body: payload,
    signal: opts?.signal,
  });
  if (res.status === 401) {
    sessionEnded.value = true;
    throw new Error("session ended");
  }
  if (!res.ok) {
    const text = await res.text();
    try {
      const parsed = JSON.parse(text) as { error?: ApiErrorBody };
      if (parsed.error) {
        throw new ApiError(parsed.error);
      }
    } catch (e) {
      if (e instanceof ApiError) {
        throw e;
      }
    }
    throw new Error(`HTTP ${res.status}`);
  }
  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}
