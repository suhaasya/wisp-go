import { describe, expect, it } from "vitest";
import {
  deriveTags,
  filterConnections,
  formatHost,
  groupByFolder,
  type ConnectionProfile,
} from "./connections";

const acmeProd: ConnectionProfile = {
  id: "p1",
  name: "shop_prod",
  folder: "Acme backend",
  engine: "postgres",
  transport: { kind: "tcp", host: "db-prod-1.internal", port: 5432, path: "" },
  user: "app",
  database: "shop",
  password_mode: "keychain",
  ssl: { mode: "verify-full" },
  ssh: { host: "bastion.acme.io", port: 22, user: "ubuntu", key_path: "" },
  env: "production",
  read_only: true,
  safe_mode: false,
  options: {},
  has_password: true,
};

const sideDev: ConnectionProfile = {
  id: "p2",
  name: "habit-app",
  folder: "Side projects",
  engine: "postgres",
  transport: { kind: "tcp", host: "db.qwerty.supabase.co", port: 5432, path: "" },
  user: "app",
  database: "habit",
  password_mode: "keychain",
  ssl: { mode: "require" },
  env: "development",
  read_only: false,
  safe_mode: false,
  options: {},
  has_password: false,
};

const fixtures = [acmeProd, sideDev];

describe("filterConnections", () => {
  it("filters by search query on name and host", () => {
    const out = filterConnections(fixtures, {
      query: "habit",
      folder: "all",
      env: "all",
    });
    expect(out.map((p) => p.id)).toEqual(["p2"]);
  });

  it("combines folder and env filters with search", () => {
    const out = filterConnections(fixtures, {
      query: "db",
      folder: "Acme backend",
      env: "production",
    });
    expect(out.map((p) => p.id)).toEqual(["p1"]);
  });

  it("filters by env only", () => {
    const out = filterConnections(fixtures, {
      query: "",
      folder: "all",
      env: "development",
    });
    expect(out.map((p) => p.id)).toEqual(["p2"]);
  });
});

describe("groupByFolder", () => {
  it("groups profiles under folder names in stable order", () => {
    const groups = groupByFolder(fixtures, ["Acme backend", "Side projects"]);
    expect(groups.map((g) => g.folder)).toEqual(["Acme backend", "Side projects"]);
    expect(groups[0]!.profiles[0]!.id).toBe("p1");
    expect(groups[1]!.profiles[0]!.id).toBe("p2");
  });
});

describe("formatHost", () => {
  it("shows host:port for tcp", () => {
    expect(formatHost(acmeProd)).toBe("db-prod-1.internal:5432");
  });
});

describe("deriveTags", () => {
  it("includes ssh tls and read-only labels", () => {
    const tags = deriveTags(acmeProd);
    expect(tags).toContain("SSH via bastion.acme.io");
    expect(tags).toContain("TLS verify-full");
    expect(tags).toContain("Read-only");
  });
});
