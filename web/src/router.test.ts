import { beforeEach, describe, expect, it, vi } from "vitest";
import { parseRoute, route, type Route } from "./router";

describe("router", () => {
  beforeEach(() => {
    route.value = parseRoute("#/connections");
  });

  it("parses connection routes", () => {
    expect(parseRoute("#/connections")).toEqual({ name: "connections" } satisfies Route);
    expect(parseRoute("#/connections/new")).toEqual({ name: "connection-new" } satisfies Route);
    expect(parseRoute("#/connections/pg-local/edit")).toEqual({
      name: "connection-edit",
      id: "pg-local",
    } satisfies Route);
    expect(parseRoute("#/w/s1")).toEqual({ name: "workspace", sessionId: "s1" } satisfies Route);
    expect(parseRoute("#/bench/grid")).toEqual({ name: "bench-grid" } satisfies Route);
  });

  it("syncRouteFromHash updates route signal", async () => {
    const { syncRouteFromHash, stopRouter } = await import("./router");
    location.hash = "#/connections/new";
    syncRouteFromHash();
    expect(route.value.name).toBe("connection-new");
    stopRouter();
  });
});
