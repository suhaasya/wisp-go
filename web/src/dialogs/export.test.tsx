import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { selectCell } from "../grid/selection";
import { ExportDialog } from "./export";

describe("ExportDialog", () => {
  it("runs download export with chosen format", () => {
    const onDownload = vi.fn();
    render(
      <ExportDialog
        open
        sessionId="s1"
        rid="r1"
        tableLabel="public.orders"
        selection={selectCell(0, 0)}
        pageOffset={0}
        pageLimit={150}
        onClose={() => {}}
        onDownload={onDownload}
        onClipboard={() => {}}
        onFileExport={() => {}}
      />,
    );
    fireEvent.change(screen.getByLabelText("Format"), { target: { value: "json" } });
    fireEvent.click(screen.getByRole("button", { name: "Export" }));
    expect(onDownload).toHaveBeenCalledWith(
      expect.objectContaining({ format: "json", scope: "selection", rid: "r1" }),
    );
  });
});
