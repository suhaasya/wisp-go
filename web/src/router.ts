import { signal } from "@preact/signals";

export type Route =
  | { name: "connections" }
  | { name: "connection-new" }
  | { name: "connection-edit"; id: string }
  | { name: "workspace"; sessionId: string }
  | { name: "gallery" }
  | { name: "bench-grid" }
  | { name: "bench-workspace-tabs" };

function initialRoute(): Route {
  if (typeof location === "undefined") {
    return { name: "connections" };
  }
  return parseRoute(location.hash || "#/connections");
}

export const route = signal<Route>(initialRoute());

export function parseRoute(hash: string): Route {
  const path = hash.replace(/^#/, "") || "/connections";
  if (path === "/gallery" || path.startsWith("/gallery?")) {
    return { name: "gallery" };
  }
  if (path === "/bench/grid") {
    return { name: "bench-grid" };
  }
  if (path === "/bench/workspace-tabs") {
    return { name: "bench-workspace-tabs" };
  }
  if (path === "/connections") {
    return { name: "connections" };
  }
  if (path === "/connections/new") {
    return { name: "connection-new" };
  }
  const edit = /^\/connections\/([^/]+)\/edit$/.exec(path);
  if (edit) {
    return { name: "connection-edit", id: decodeURIComponent(edit[1]!) };
  }
  const ws = /^\/w\/([^/]+)$/.exec(path);
  if (ws) {
    return { name: "workspace", sessionId: decodeURIComponent(ws[1]!) };
  }
  return { name: "connections" };
}

let listening = false;

export function syncRouteFromHash(): void {
  route.value = parseRoute(location.hash || "#/connections");
}

export function startRouter(): void {
  if (listening) {
    return;
  }
  if (!location.hash) {
    location.hash = "#/connections";
  }
  syncRouteFromHash();
  window.addEventListener("hashchange", syncRouteFromHash);
  listening = true;
}

export function stopRouter(): void {
  if (!listening) {
    return;
  }
  window.removeEventListener("hashchange", syncRouteFromHash);
  listening = false;
}

export function navigate(path: string): void {
  location.hash = path.startsWith("#") ? path : `#${path}`;
}
