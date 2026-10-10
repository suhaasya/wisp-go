/**
 * Engine × transport matrix for Playwright. Tags: pr (CI on pull_request), nightly (scheduled).
 * @typedef {'pr' | 'nightly'} MatrixTag
 * @typedef {'postgres' | 'mysql' | 'mariadb'} Engine
 * @typedef {'tcp' | 'tls-require' | 'tls-verify-full' | 'ssh'} TransportKind
 *
 * @typedef {object} MatrixCell
 * @property {string} id
 * @property {MatrixTag[]} tags
 * @property {Engine} engine
 * @property {TransportKind} transport
 * @property {string} versionLabel
 * @property {(ctx: { schema: string; repoRoot: string }) => object} profileBody
 */

import path from "node:path";

const wispUser = "wisp";
const wispPass = "wisp";
const wispDB = "wisp_bench";

/** @param {Engine} engine @param {string} schema */
export function seedSQL(engine, schema) {
  const safe = schema.replace(/[^a-z0-9_]/gi, "");
  if (engine === "postgres") {
    return (
      `CREATE SCHEMA IF NOT EXISTS ${safe};` +
      `CREATE TABLE IF NOT EXISTS ${safe}.matrix_probe (id int PRIMARY KEY, label text NOT NULL);` +
      `TRUNCATE ${safe}.matrix_probe;` +
      `INSERT INTO ${safe}.matrix_probe (id, label) VALUES (1, 'matrix-ok');`
    );
  }
  const table = `matrix_probe_${safe}`;
  return (
    `CREATE TABLE IF NOT EXISTS ${table} (id INT PRIMARY KEY, label VARCHAR(64) NOT NULL);` +
    `TRUNCATE ${table};` +
    `INSERT INTO ${table} (id, label) VALUES (1, 'matrix-ok');`
  );
}

/** @param {object} p */
function baseProfile(p) {
  return {
    folder: "E2E Matrix",
    user: wispUser,
    database: wispDB,
    password: wispPass,
    password_mode: "store",
    env: "local",
    read_only: false,
    safe_mode: false,
    options: { allow_public_key_retrieval: true },
  };
}

/** @returns {MatrixCell[]} */
export function allMatrixCells() {
  /** @type {MatrixCell[]} */
  const cells = [];

  const pgPorts = {
    12: 54312,
    13: 54313,
    14: 54314,
    15: 54315,
    16: 54316,
    17: 54317,
  };

  for (const [ver, port] of Object.entries(pgPorts)) {
    cells.push({
      id: `pg${ver}-tcp`,
      tags: Number(ver) === 17 ? ["pr", "nightly"] : ["nightly"],
      engine: "postgres",
      transport: "tcp",
      versionLabel: ver,
      profileBody: ({ schema }) => ({
        ...baseProfile({}),
        name: `E2E PG ${ver} TCP`,
        engine: "postgres",
        transport: { kind: "tcp", host: "127.0.0.1", port, path: "" },
        ssl: { mode: "disable" },
        advanced: {
          default_schema: schema.replace(/[^a-z0-9_]/gi, ""),
          run_on_connect: seedSQL("postgres", schema),
        },
      }),
    });
  }

  cells.push({
    id: "pg-tls-require",
    tags: ["nightly"],
    engine: "postgres",
    transport: "tls-require",
    versionLabel: "16",
    profileBody: ({ schema }) => ({
      ...baseProfile({}),
      name: "E2E PG TLS require",
      engine: "postgres",
      transport: { kind: "tcp", host: "127.0.0.1", port: 54330, path: "" },
      ssl: { mode: "require" },
      advanced: {
        default_schema: schema.replace(/[^a-z0-9_]/gi, ""),
        run_on_connect: seedSQL("postgres", schema),
      },
    }),
  });

  cells.push({
    id: "pg-tls-verify-full",
    tags: ["nightly"],
    engine: "postgres",
    transport: "tls-verify-full",
    versionLabel: "16",
    profileBody: ({ schema, repoRoot }) => ({
      ...baseProfile({}),
      name: "E2E PG TLS verify-full",
      engine: "postgres",
      transport: { kind: "tcp", host: "127.0.0.1", port: 54330, path: "" },
      ssl: {
        mode: "verify-full",
        ca_source: "file",
        ca_file: path.join(repoRoot, "ci", "certs", "ca.pem"),
      },
      advanced: {
        default_schema: schema.replace(/[^a-z0-9_]/gi, ""),
        run_on_connect: seedSQL("postgres", schema),
      },
    }),
  });

  cells.push({
    id: "pg17-ssh",
    tags: ["pr", "nightly"],
    engine: "postgres",
    transport: "ssh",
    versionLabel: "17",
    profileBody: ({ schema, repoRoot }) => ({
      ...baseProfile({}),
      name: "E2E PG 17 SSH",
      engine: "postgres",
      transport: { kind: "tcp", host: "pg17", port: 5432, path: "" },
      ssl: { mode: "disable" },
      ssh: {
        enabled: true,
        host: "127.0.0.1",
        port: 2221,
        user: "wisp",
        auth: "key",
        key_path: path.join(repoRoot, "ci", "ssh", "client", "id_ed25519"),
      },
      advanced: {
        default_schema: schema.replace(/[^a-z0-9_]/gi, ""),
        run_on_connect: seedSQL("postgres", schema),
      },
    }),
  });

  const mysqlDefs = [
    { id: "mysql57", port: 33057, tags: ["nightly"], label: "5.7" },
    { id: "mysql80", port: 33080, tags: ["nightly"], label: "8.0" },
    { id: "mysql84", port: 33084, tags: ["pr", "nightly"], label: "8.4" },
  ];
  for (const m of mysqlDefs) {
    cells.push({
      id: `${m.id}-tcp`,
      tags: m.tags,
      engine: "mysql",
      transport: "tcp",
      versionLabel: m.label,
      profileBody: ({ schema }) => ({
        ...baseProfile({}),
        name: `E2E MySQL ${m.label} TCP`,
        engine: "mysql",
        transport: { kind: "tcp", host: "127.0.0.1", port: m.port, path: "" },
        ssl: { mode: "disable" },
        advanced: { run_on_connect: seedSQL("mysql", schema) },
      }),
    });
    if (m.id === "mysql84") {
      cells.push({
        id: "mysql84-ssh",
        tags: ["pr", "nightly"],
        engine: "mysql",
        transport: "ssh",
        versionLabel: "8.4",
        profileBody: ({ schema, repoRoot }) => ({
          ...baseProfile({}),
          name: "E2E MySQL 8.4 SSH",
          engine: "mysql",
          transport: { kind: "tcp", host: "mysql84", port: 3306, path: "" },
          ssl: { mode: "disable" },
          ssh: {
            enabled: true,
            host: "127.0.0.1",
            port: 2221,
            user: "wisp",
            auth: "key",
            key_path: path.join(repoRoot, "ci", "ssh", "client", "id_ed25519"),
          },
          advanced: { run_on_connect: seedSQL("mysql", schema) },
        }),
      });
    }
  }

  const mariaDefs = [
    { id: "mariadb106", port: 33106, tags: ["nightly"], label: "10.6" },
    { id: "mariadb114", port: 33114, tags: ["pr", "nightly"], label: "11.4" },
  ];
  for (const m of mariaDefs) {
    cells.push({
      id: `${m.id}-tcp`,
      tags: m.tags,
      engine: "mariadb",
      transport: "tcp",
      versionLabel: m.label,
      profileBody: ({ schema }) => ({
        ...baseProfile({}),
        name: `E2E MariaDB ${m.label} TCP`,
        engine: "mariadb",
        transport: { kind: "tcp", host: "127.0.0.1", port: m.port, path: "" },
        ssl: { mode: "disable" },
        advanced: { run_on_connect: seedSQL("mysql", schema) },
      }),
    });
    if (m.id === "mariadb114") {
      cells.push({
        id: "mariadb114-ssh",
        tags: ["pr", "nightly"],
        engine: "mariadb",
        transport: "ssh",
        versionLabel: "11.4",
        profileBody: ({ schema, repoRoot }) => ({
          ...baseProfile({}),
          name: "E2E MariaDB 11.4 SSH",
          engine: "mariadb",
          transport: { kind: "tcp", host: "mariadb114", port: 3306, path: "" },
          ssl: { mode: "disable" },
          ssh: {
            enabled: true,
            host: "127.0.0.1",
            port: 2221,
            user: "wisp",
            auth: "key",
            key_path: path.join(repoRoot, "ci", "ssh", "client", "id_ed25519"),
          },
          advanced: { run_on_connect: seedSQL("mysql", schema) },
        }),
      });
    }
  }

  return cells;
}

/** @param {MatrixTag} tag */
export function cellsForTag(tag) {
  return allMatrixCells().filter((c) => c.tags.includes(tag));
}
