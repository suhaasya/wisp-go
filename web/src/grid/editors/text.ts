export function mountTextEditor(
  host: HTMLElement,
  initial: string,
  onCommit: (value: string) => void,
  onCancel: () => void,
  onTab?: () => void,
): () => void {
  const input = document.createElement("input");
  input.type = "text";
  input.className =
    "h-full w-full border border-accent bg-panel px-2 font-mono text-12-5 text-ink outline-none";
  input.value = initial;
  host.replaceChildren(input);
  input.focus();
  input.select();

  const onKey = (e: KeyboardEvent) => {
    if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
      return;
    }
    if (e.key === "Enter") {
      e.preventDefault();
      onCommit(input.value);
      return;
    }
    if (e.key === "Tab") {
      e.preventDefault();
      onCommit(input.value);
      onTab?.();
      return;
    }
    if ((e.metaKey || e.ctrlKey) && e.key === "Backspace") {
      e.preventDefault();
      onCommit("");
    }
  };
  input.addEventListener("keydown", onKey);
  input.addEventListener("blur", () => onCommit(input.value), { once: true });

  return () => {
    input.removeEventListener("keydown", onKey);
  };
}
