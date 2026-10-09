import { keymap } from "@codemirror/view";

export type RunKeymapHandlers = {
  onRunStatement: () => void;
  onRunAll: () => void;
  onCancel: () => void;
};

export function runKeymapExtension(handlers: RunKeymapHandlers) {
  return keymap.of([
    {
      key: "Mod-Enter",
      run: () => {
        handlers.onRunStatement();
        return true;
      },
    },
    {
      key: "Shift-Mod-Enter",
      run: () => {
        handlers.onRunAll();
        return true;
      },
    },
    {
      key: "Escape",
      run: () => {
        handlers.onCancel();
        return true;
      },
    },
  ]);
}
