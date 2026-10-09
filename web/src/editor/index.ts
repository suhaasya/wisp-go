import { closeBrackets } from "@codemirror/autocomplete";
import { history } from "@codemirror/commands";
import { bracketMatching } from "@codemirror/language";
import { MySQL, PostgreSQL, sql } from "@codemirror/lang-sql";
import { Compartment, EditorState, type Extension } from "@codemirror/state";
import {
  drawSelection,
  EditorView,
  highlightActiveLine,
  highlightActiveLineGutter,
  lineNumbers,
} from "@codemirror/view";
import {
  mergeSnippetsIntoNamespace,
  toSQLNamespace,
  type CompletionSchemaDTO,
  type SchemaCompletionState,
  type SnippetCompletionDTO,
} from "./complete";
import { readCspNonce } from "./csp";
import { wispEditorKeymap } from "./keymap";
import { runKeymapExtension, type RunKeymapBindings, type RunKeymapHandlers } from "./run-keymap";
import { applyRunError, clearRunError, runErrorExtension } from "./run-error";
import { wispEditorTheme } from "./theme";

export type SqlEngine = "postgres" | "mysql" | "mariadb";

export type MountSqlEditorOptions = {
  parent: HTMLElement;
  value: string;
  engine: SqlEngine;
  schema?: string;
  completion?: CompletionSchemaDTO;
  snippets?: SnippetCompletionDTO[];
  onChange: (value: string) => void;
  runKeys?: RunKeymapHandlers;
  runKeyBindings?: RunKeymapBindings;
};

export type SqlEditorHandle = {
  view: EditorView;
  setCompletionSchema: (data: CompletionSchemaDTO) => void;
  setRunErrorAt: (byteOffset: number) => void;
  clearRunError: () => void;
  getSelectionRange: () => { start: number; end: number };
  getCursor: () => number;
  insertText: (text: string) => void;
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
  const dialect = sqlDialect(opts.engine);
  let namespace = opts.completion ? toSQLNamespace(opts.completion) : {};
  if (opts.snippets?.length) {
    namespace = mergeSnippetsIntoNamespace(namespace, opts.snippets);
  }
  const completionState: SchemaCompletionState = {
    namespace,
    defaultSchema: opts.completion?.schema ?? opts.schema ?? "public",
  };
  const completionCompartment = new Compartment();

  function sqlExtension(): Extension {
    return sql({
      dialect,
      schema: completionState.namespace,
      defaultSchema: completionState.defaultSchema,
      upperCaseKeywords: opts.engine === "postgres",
      keywordCompletion: (label, type) => ({
        label,
        type,
        detail: type === "keyword" ? "keyword" : undefined,
      }),
    });
  }

  const baseExtensions: Extension[] = [
    lineNumbers(),
    highlightActiveLineGutter(),
    highlightActiveLine(),
    history(),
    bracketMatching(),
    closeBrackets(),
    drawSelection(),
    wispEditorTheme(),
    wispEditorKeymap(),
    runErrorExtension(),
    ...(opts.runKeys ? [runKeymapExtension(opts.runKeys, opts.runKeyBindings)] : []),
    completionCompartment.of(sqlExtension()),
    EditorView.updateListener.of((u) => {
      if (u.docChanged) {
        opts.onChange(u.state.doc.toString());
      }
    }),
    ...(nonce ? [EditorView.cspNonce.of(nonce)] : []),
  ];

  const state = EditorState.create({
    doc: opts.value,
    extensions: baseExtensions,
  });
  const view = new EditorView({ state, parent: opts.parent });

  function setCompletionSchema(data: CompletionSchemaDTO) {
    completionState.namespace = toSQLNamespace(data);
    completionState.defaultSchema = data.schema;
    view.dispatch({
      effects: completionCompartment.reconfigure(sqlExtension()),
    });
  }

  return {
    view,
    setCompletionSchema,
    setRunErrorAt: (byteOffset: number) => applyRunError(view, byteOffset),
    clearRunError: () => clearRunError(view),
    getSelectionRange: () => {
      const sel = view.state.selection.main;
      return { start: sel.from, end: sel.to };
    },
    getCursor: () => view.state.selection.main.head,
    insertText: (text: string) => {
      const sel = view.state.selection.main;
      view.dispatch({
        changes: { from: sel.from, to: sel.to, insert: text },
        selection: { anchor: sel.from + text.length },
      });
    },
    destroy: () => {
      view.destroy();
    },
  };
}
