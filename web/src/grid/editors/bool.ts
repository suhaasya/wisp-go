export function mountBoolEditor(
  host: HTMLElement,
  initial: boolean | null,
  onCommit: (value: boolean) => void,
  onCancel: () => void,
): () => void {
  const wrap = document.createElement("div");
  wrap.className = "flex h-full items-center px-2";
  const btn = document.createElement("button");
  btn.type = "button";
  btn.className = "rounded border border-line bg-raise px-2 py-0.5 text-12-5 text-ink";
  btn.textContent = initial === null ? "NULL" : initial ? "true" : "false";
  wrap.appendChild(btn);
  host.replaceChildren(wrap);
  btn.focus();

  const toggle = () => {
    const next = initial === null ? true : !initial;
    onCommit(next);
  };
  btn.addEventListener("click", toggle);
  const onKey = (e: KeyboardEvent) => {
    if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    }
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      toggle();
    }
  };
  btn.addEventListener("keydown", onKey);
  return () => {
    btn.removeEventListener("keydown", onKey);
  };
}
