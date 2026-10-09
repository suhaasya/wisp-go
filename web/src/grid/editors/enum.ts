export function mountEnumEditor(
  host: HTMLElement,
  initial: string,
  values: string[],
  onCommit: (value: string) => void,
  onCancel: () => void,
): () => void {
  const select = document.createElement("select");
  select.className =
    "h-full w-full border border-accent bg-panel px-1 font-mono text-12-5 text-ink outline-none";
  for (const v of values) {
    const opt = document.createElement("option");
    opt.value = v;
    opt.textContent = v;
    if (v === initial) {
      opt.selected = true;
    }
    select.appendChild(opt);
  }
  host.replaceChildren(select);
  select.focus();

  const commit = () => onCommit(select.value);
  const onKey = (e: KeyboardEvent) => {
    if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    }
    if (e.key === "Enter") {
      e.preventDefault();
      commit();
    }
  };
  select.addEventListener("keydown", onKey);
  select.addEventListener("change", commit);
  select.addEventListener("blur", commit, { once: true });
  return () => select.removeEventListener("keydown", onKey);
}
