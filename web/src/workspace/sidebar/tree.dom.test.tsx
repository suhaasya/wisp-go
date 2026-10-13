import { render, waitFor } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { SchemaTree } from "./tree";
import type { ObjectKind } from "../schema-client";

describe("SchemaTree DOM budget", () => {
  it("renders a bounded number of nodes for 10k table names", async () => {
    const names = Array.from({ length: 10_000 }, (_, i) => `t_${i}`);
    const fetchObjectsImpl = vi.fn(async (_sid: string, _schema: string, kind: ObjectKind) => {
      if (kind === "table") {
        return names;
      }
      return [];
    });

    const expanded = {
      tables: true,
      views: false,
      functions: false,
      sequences: false,
      types: false,
    };

    const { container } = render(
      <div class="flex h-[400px] flex-col">
        <SchemaTree
          sessionId="s1"
          schema="public"
          filterQuery=""
          expanded={expanded}
          onExpandedChange={() => {}}
          selectedObject={null}
          onSelectObject={() => {}}
          onOpenTable={() => {}}
          fetchObjectsImpl={fetchObjectsImpl}
        />
      </div>,
    );

    await waitFor(() => {
      expect(fetchObjectsImpl).toHaveBeenCalled();
      expect(container.querySelectorAll("[data-object-row]").length).toBeGreaterThan(0);
    });

    const objectRows = container.querySelectorAll("[data-object-row]");
    expect(objectRows.length).toBeLessThan(80);
    expect(objectRows.length).toBeGreaterThan(0);

    const allNodes = container.querySelectorAll("*");
    expect(allNodes.length).toBeLessThan(2000);
  });

  it("loads every expanded group when more than one is open on first mount", async () => {
    const fetchObjectsImpl = vi.fn(async (_sid: string, _schema: string, kind: ObjectKind) => {
      if (kind === "table") {
        return ["user"];
      }
      if (kind === "sequence") {
        return [];
      }
      return [];
    });

    const expanded = {
      tables: true,
      views: false,
      functions: false,
      sequences: true,
      types: false,
    };

    const { container } = render(
      <SchemaTree
        sessionId="s1"
        schema="public"
        filterQuery=""
        expanded={expanded}
        onExpandedChange={() => {}}
        selectedObject={null}
        onSelectObject={() => {}}
        onOpenTable={() => {}}
        fetchObjectsImpl={fetchObjectsImpl}
      />,
    );

    await waitFor(() => {
      expect(container.querySelector('[data-object-row="user"]')).toBeTruthy();
    });

    expect(container.textContent).not.toContain("Loading…");
    expect(fetchObjectsImpl).toHaveBeenCalledWith("s1", "public", "table", expect.any(Object));
    expect(fetchObjectsImpl).toHaveBeenCalledWith("s1", "public", "sequence", expect.any(Object));
  });
});
