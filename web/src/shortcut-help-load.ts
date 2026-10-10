/** Shared dynamic import so the shell and key handler agree when the chunk is ready. */
export function loadShortcutHelp() {
  return import("./shortcut-help");
}
