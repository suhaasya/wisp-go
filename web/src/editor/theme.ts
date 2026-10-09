import { HighlightStyle, syntaxHighlighting } from "@codemirror/language";
import { tags as t } from "@lezer/highlight";
import { EditorView } from "@codemirror/view";

export function wispEditorTheme() {
  return [
    EditorView.theme({
      "&": {
        backgroundColor: "var(--color-panel)",
        color: "var(--color-ink)",
        fontSize: "13px",
        lineHeight: "21px",
        fontFamily: "var(--font-mono)",
        height: "100%",
      },
      ".cm-scroller": {
        overflow: "auto",
        fontFamily: "inherit",
      },
      ".cm-gutters": {
        backgroundColor: "var(--color-sidebar)",
        color: "var(--color-ink-3)",
        borderRight: "1px solid var(--color-line-2)",
        fontFamily: "var(--font-mono)",
        fontSize: "13px",
      },
      ".cm-activeLineGutter": {
        backgroundColor: "var(--color-raise)",
      },
      ".cm-activeLine": {
        backgroundColor: "var(--color-raise)",
      },
      ".cm-content": {
        padding: "10px 14px",
        caretColor: "var(--color-accent)",
      },
      ".cm-gutter.cm-lineNumbers .cm-gutterElement": {
        padding: "0 10px 0 14px",
        minWidth: "2.5em",
      },
      "&.cm-focused .cm-cursor": {
        borderLeftColor: "var(--color-accent)",
      },
      "&.cm-focused .cm-selectionBackground, .cm-selectionBackground": {
        backgroundColor: "var(--color-accent-soft) !important",
      },
      ".cm-tooltip.cm-tooltip-autocomplete": {
        backgroundColor: "var(--color-panel)",
        border: "1px solid var(--color-line)",
        borderRadius: "8px",
        minWidth: "250px",
      },
      ".cm-tooltip.cm-tooltip-autocomplete > ul": {
        fontFamily: "var(--font-mono)",
        fontSize: "12.5px",
      },
      ".cm-tooltip.cm-tooltip-autocomplete > ul > li[aria-selected]": {
        backgroundColor: "var(--color-accent-soft)",
        color: "var(--color-ink)",
      },
      ".cm-completionDetail": {
        color: "var(--color-ink-3)",
        fontStyle: "normal",
        marginLeft: "8px",
      },
    }),
    syntaxHighlighting(
      HighlightStyle.define([
        { tag: t.keyword, color: "var(--color-syn-kw)", fontWeight: "500" },
        { tag: [t.string, t.special(t.string)], color: "var(--color-syn-str)" },
        { tag: t.number, color: "var(--color-syn-num)" },
        { tag: [t.function(t.variableName), t.function(t.propertyName)], color: "var(--color-syn-fn)" },
        { tag: [t.comment, t.lineComment, t.blockComment], color: "var(--color-syn-cm)", fontStyle: "italic" },
      ]),
    ),
  ];
}
