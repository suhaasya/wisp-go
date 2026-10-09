import { EditorState } from "@codemirror/state";
import { EditorView } from "@codemirror/view";
import { describe, expect, it } from "vitest";
import { applyRunError, runErrorField, runErrorExtension } from "./run-error";

describe("run error decoration", () => {
  it("marks the byte offset in the document", () => {
    const parent = document.createElement("div");
    const view = new EditorView({
      state: EditorState.create({
        doc: "SELECT bad\nFROM t",
        extensions: [runErrorExtension()],
      }),
      parent,
    });
    applyRunError(view, 7);
    let marked = false;
    view.state.field(runErrorField).between(7, 8, () => {
      marked = true;
    });
    expect(marked).toBe(true);
    view.destroy();
  });
});
