import { closeBrackets } from "@codemirror/autocomplete";
import { history } from "@codemirror/commands";
import { bracketMatching } from "@codemirror/language";
import { MySQL, PostgreSQL, sql } from "@codemirror/lang-sql";
import { EditorState } from "@codemirror/state";
import {
  drawSelection,
  EditorView,
  highlightActiveLine,
  highlightActiveLineGutter,
  lineNumbers,
} from "@codemirror/view";
import { readCspNonce } from "./csp";
import { wispEditorKeymap } from "./keymap";
import { wispEditorTheme } from "./theme";

export type SqlEngine = "postgres" | "mysql" | "mariadb";

export type MountSqlEditorOptions = {
  parent: HTMLElement;
  value: string;
  engine: SqlEngine;
  onChange: (value: string) => void;
};

export type SqlEditorHandle = {
  view: EditorView;
  destroy: () => void;
};

function sqlDialect(engine: SqlEngine) {
  if (engine === "mysql" || engine === "mariadb") {
    return MySQL;
  }
  return PostgreSQL;
}

export function mountSqlEditor(opts: MountSqlEditorOptions): SqlEditorHandle {
  const nonce = readCspNonce();
  const state = EditorState.create({
    doc: opts.value,
    extensions: [
      lineNumbers(),
      highlightActiveLineGutter(),
      highlightActiveLine(),
      history(),
      bracketMatching(),
      closeBrackets(),
      sql({ dialect: sqlDialect(opts.engine) }),
      drawSelection(),
      wispEditorTheme(),
      wispEditorKeymap(),
      EditorView.updateListener.of((u) => {
        if (u.docChanged) {
          opts.onChange(u.state.doc.toString());
        }
      }),
      ...(nonce ? [EditorView.cspNonce.of(nonce)] : []),
    ],
  });
  const view = new EditorView({ state, parent: opts.parent });
  return {
    view,
    destroy: () => {
      view.destroy();
    },
  };
}
