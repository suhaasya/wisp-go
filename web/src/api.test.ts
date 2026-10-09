import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { csrfToken, sessionEnded } from "./state";

describe("api", () => {
  beforeEach(() => {
    sessionEnded.value = false;
    csrfToken.value = "";
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("request sends X-Wisp-CSRF after bootstrap", async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ version: "dev", csrf: "abc", settings: {} }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    const { bootstrap, request } = await import("./api");
    await bootstrap();
    await request("GET", "/api/v1/hello");
    const second = fetchMock.mock.calls[1];
    const init = second?.[1] as RequestInit;
    expect(new Headers(init.headers).get("X-Wisp-CSRF")).toBe("abc");
  });

  it("maps error JSON to ApiError", async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ version: "dev", csrf: "abc", settings: {} }), {
        status: 200,
      }),
    );
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ error: { code: "x", message: "bad", detail: "d" } }), {
        status: 400,
        headers: { "Content-Type": "application/json" },
      }),
    );
    const { bootstrap, request, ApiError } = await import("./api");
    await bootstrap();
    await expect(request("POST", "/api/v1/x", {})).rejects.toBeInstanceOf(ApiError);
  });

  it("401 sets sessionEnded", async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ version: "dev", csrf: "abc", settings: {} }), {
        status: 200,
      }),
    );
    fetchMock.mockResolvedValueOnce(new Response("unauthorized", { status: 401 }));
    const { bootstrap, request } = await import("./api");
    await bootstrap();
    await expect(request("GET", "/api/v1/hello")).rejects.toThrow();
    expect(sessionEnded.value).toBe(true);
  });

  it("honours AbortSignal", async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ version: "dev", csrf: "abc", settings: {} }), {
        status: 200,
      }),
    );
    fetchMock.mockImplementation((_url, init) => {
      if (init?.signal?.aborted) {
        return Promise.reject(new DOMException("Aborted", "AbortError"));
      }
      return Promise.resolve(new Response("{}", { status: 200 }));
    });
    const { bootstrap, request } = await import("./api");
    await bootstrap();
    const ac = new AbortController();
    ac.abort();
    await expect(request("GET", "/api/v1/hello", undefined, { signal: ac.signal })).rejects.toMatchObject({
      name: "AbortError",
    });
  });
});
