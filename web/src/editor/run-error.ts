import { RangeSetBuilder, StateEffect, StateField } from "@codemirror/state";
import { Decoration, type DecorationSet, EditorView } from "@codemirror/view";

export type RunErrorRange = { from: number; to: number } | null;

export const setRunErrorEffect = StateEffect.define<RunErrorRange>();

export const runErrorField = StateField.define<DecorationSet>({
  create() {
    return Decoration.none;
  },
  update(deco, tr) {
    for (const e of tr.effects) {
      if (e.is(setRunErrorEffect)) {
        if (!e.value) {
          return Decoration.none;
        }
        const b = new RangeSetBuilder<Decoration>();
        b.add(
          e.value.from,
          e.value.to,
          Decoration.mark({ class: "cm-runError" }),
        );
        return b.finish();
      }
    }
    return deco.map(tr.changes);
  },
  provide: (f) => EditorView.decorations.from(f),
});

export function runErrorExtension() {
  return runErrorField;
}

export function applyRunError(view: EditorView, pos: number) {
  const from = Math.max(0, pos);
  const to = Math.min(view.state.doc.length, from + 1);
  view.dispatch({
    effects: setRunErrorEffect.of({ from, to }),
  });
}

export function clearRunError(view: EditorView) {
  view.dispatch({ effects: setRunErrorEffect.of(null) });
}
