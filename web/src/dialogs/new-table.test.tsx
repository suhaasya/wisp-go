import { fireEvent, render, screen, waitFor } from "@testing-library/preact";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { NewTableDialog } from "./new-table";

vi.mock("../workspace/ddl-client", () => ({
  previewDDL: vi.fn().mockResolvedValue({ sql: "CREATE TABLE demo (id bigint);" }),
  executeDDL: vi.fn(),
}));

describe("NewTableDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows live DDL preview from wizard", async () => {
    render(
      <NewTableDialog
        open
        sessionId="s1"
        schema="public"
        engine="postgres"
        onClose={() => {}}
        onCreated={() => {}}
      />,
    );
    fireEvent.input(screen.getByLabelText("Table name"), { target: { value: "demo" } });
    await waitFor(() => {
      expect(screen.getByTestId("new-table-preview").textContent).toContain("CREATE TABLE");
    });
  });
});
