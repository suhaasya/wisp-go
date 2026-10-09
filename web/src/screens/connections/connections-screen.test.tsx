import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/preact";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { ConnectionProfile } from "../../connections";
import { handleGlobalKeydown } from "../../keydispatch";
import * as router from "../../router";
import { route } from "../../router";
import { ConnectionsScreen } from "./index";

const sample: ConnectionProfile[] = [
  {
    id: "a1",
    name: "shop_prod",
    folder: "Acme backend",
    engine: "postgres",
    transport: { kind: "tcp", host: "db-prod.internal", port: 5432, path: "" },
    user: "u",
    database: "d",
    password_mode: "keychain",
    ssl: { mode: "disable" },
    env: "production",
    read_only: false,
    safe_mode: false,
    options: {},
    has_password: true,
  },
  {
    id: "a2",
    name: "habit-app",
    folder: "Side projects",
    engine: "postgres",
    transport: { kind: "tcp", host: "db.supabase.co", port: 5432, path: "" },
    user: "u",
    database: "d",
    password_mode: "keychain",
    ssl: { mode: "disable" },
    env: "development",
    read_only: false,
    safe_mode: false,
    options: {},
    has_password: false,
  },
];

describe("ConnectionsScreen", () => {
  beforeEach(() => {
    location.hash = "#/connections";
  });

  afterEach(() => {
    cleanup();
  });

  it("filters cards when searching", () => {
    render(<ConnectionsScreen initialProfiles={sample} initialFolders={["Acme backend", "Side projects"]} />);
    expect(screen.getByText("shop_prod")).toBeTruthy();
    expect(screen.getByText("habit-app")).toBeTruthy();
    fireEvent.input(screen.getByLabelText("Search connections"), { target: { value: "habit" } });
    expect(screen.queryByText("shop_prod")).toBeNull();
    expect(screen.getByText("habit-app")).toBeTruthy();
  });

  it("deletes after confirm in modal", async () => {
    render(<ConnectionsScreen initialProfiles={sample} initialFolders={["Acme backend", "Side projects"]} />);
    const card = screen.getByRole("button", { name: /shop_prod/i });
    fireEvent.contextMenu(card);
    fireEvent.click(await screen.findByText("Delete…"));
    const dialog = await screen.findByRole("dialog");
    fireEvent.click(within(dialog).getByRole("button", { name: "Delete" }));
    await waitFor(() => {
      expect(screen.queryByRole("button", { name: /shop_prod/i })).toBeNull();
    });
    expect(screen.getByText("habit-app")).toBeTruthy();
  });

  it("Alt+N navigates to new connection form", () => {
    route.value = { name: "connections" };
    render(<ConnectionsScreen initialProfiles={sample} />);
    handleGlobalKeydown({
      key: "n",
      altKey: true,
      metaKey: false,
      ctrlKey: false,
      shiftKey: false,
      preventDefault: () => {},
    } as KeyboardEvent);
    expect(location.hash).toBe("#/connections/new");
  });

  it("Enter on focused card opens mock session", async () => {
    const nav = vi.spyOn(router, "navigate").mockImplementation((path: string) => {
      location.hash = path.startsWith("#") ? path : `#${path}`;
    });
    const onConnect = vi.fn(async (id: string) => `sess-${id}`);
    render(
      <ConnectionsScreen
        initialProfiles={sample}
        onConnect={onConnect}
      />,
    );
    const card = screen.getByRole("button", { name: /shop_prod/i });
    fireEvent.keyDown(card, { key: "Enter" });
    await waitFor(() => expect(onConnect).toHaveBeenCalledWith("a1"));
    await waitFor(() => expect(location.hash).toBe("#/w/sess-a1"));
    nav.mockRestore();
  });
});
