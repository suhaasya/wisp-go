import { keymap } from "@codemirror/view";

export type RunKeymapHandlers = {
  onRunStatement: () => void;
  onRunAll: () => void;
  onCancel: () => void;
};

export type RunKeymapBindings = {
  runStatement?: string;
  runAll?: string;
  cancel?: string;
};

export function runKeymapExtension(handlers: RunKeymapHandlers, bindings?: RunKeymapBindings) {
  const runStatement = bindings?.runStatement ?? "Mod-Enter";
  const runAll = bindings?.runAll ?? "Shift-Mod-Enter";
  const cancel = bindings?.cancel ?? "Escape";
  return keymap.of([
    {
      key: runStatement,
      run: () => {
        handlers.onRunStatement();
        return true;
      },
    },
    {
      key: runAll,
      run: () => {
        handlers.onRunAll();
        return true;
      },
    },
    {
      key: cancel,
      run: () => {
        handlers.onCancel();
        return true;
      },
    },
  ]);
}
