export function validateJSONText(value: string): boolean {
  if (value.trim() === "") {
    return true;
  }
  try {
    JSON.parse(value);
    return true;
  } catch {
    return false;
  }
}

export function mountJSONEditor(
  host: HTMLElement,
  initial: string,
  onCommit: (value: string) => void,
  onCancel: () => void,
): () => void {
  const ta = document.createElement("textarea");
  ta.className =
    "min-h-[52px] w-full resize-none border border-accent bg-panel px-2 py-1 font-mono text-12-5 text-ink outline-none";
  ta.value = initial;
  host.replaceChildren(ta);
  ta.focus();

  const commit = () => {
    if (!validateJSONText(ta.value)) {
      ta.classList.add("border-env-prod");
      return;
    }
    onCommit(ta.value);
  };
  const onKey = (e: KeyboardEvent) => {
    if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
      return;
    }
    if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      commit();
    }
  };
  ta.addEventListener("keydown", onKey);
  ta.addEventListener("blur", commit, { once: true });
  return () => ta.removeEventListener("keydown", onKey);
}
