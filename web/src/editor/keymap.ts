import { defaultKeymap, historyKeymap } from "@codemirror/commands";
import { searchKeymap } from "@codemirror/search";
import { keymap } from "@codemirror/view";

export function wispEditorKeymap() {
  return keymap.of([...defaultKeymap, ...historyKeymap, ...searchKeymap]);
}
