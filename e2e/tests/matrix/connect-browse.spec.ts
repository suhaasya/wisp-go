import { expect, test } from "@playwright/test";
import { createConnectionProfile, deleteConnectionProfile } from "../../matrix/api.mjs";
import { cellsForTag } from "../../matrix/manifest.mjs";
import { matrixDockerReady } from "../../matrix/probe.mjs";
import { repoRoot } from "../../matrix/paths.mjs";
import { connectFromConnectionsScreen, openMatrixTableInSidebar } from "../../matrix/ui.mjs";

const matrixTag = (process.env.WISP_MATRIX_TAG ?? "pr") as "pr" | "nightly";
const cells = cellsForTag(matrixTag);

test.describe.configure({ mode: "serial" });

test.beforeAll(async () => {
  const ports = [...new Set(cells.flatMap((c) => {
    const body = c.profileBody({ schema: "e2e_probe", repoRoot });
    const t = body.transport as { port?: number };
    const ssh = body.ssh as { port?: number } | undefined;
    return [t.port ?? 0, ssh?.port ?? 0].filter((p) => p > 0);
  }))];
  const ok = await matrixDockerReady(ports);
  test.skip(!ok, "matrix docker not running (docker compose -f ci/docker-compose.yml up -d --wait …)");
});

for (const cell of cells) {
  test(`connect and browse: ${cell.id}`, async ({ page }, testInfo) => {
    const schemaToken = `e2e_${cell.id.replace(/[^a-z0-9]/gi, "_")}_${testInfo.parallelIndex}`;
    const safeSchema = schemaToken.replace(/[^a-z0-9_]/gi, "");
    const profilePayload = cell.profileBody({ schema: schemaToken, repoRoot });
    const profileName = profilePayload.name as string;

    let profileId: string | undefined;
    try {
      const created = await createConnectionProfile(page, profilePayload);
      profileId = created.id;

      await connectFromConnectionsScreen(page, profileName);

      const tableName = cell.engine === "postgres" ? "matrix_probe" : `matrix_probe_${safeSchema}`;
      await openMatrixTableInSidebar(page, {
        engine: cell.engine,
        schema: safeSchema,
        tableName,
      });

      await expect(page.getByText("matrix-ok")).toBeVisible();

      testInfo.annotations.push({ type: "matrix-cell", description: cell.id });
    } finally {
      if (profileId) {
        await deleteConnectionProfile(page, profileId).catch(() => {});
      }
    }
  });
}
