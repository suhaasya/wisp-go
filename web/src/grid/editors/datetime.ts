const isoDateRe = /^\d{4}-\d{2}-\d{2}(T\d{2}:\d{2}(:\d{2})?)?$/;

export function validateDateTimeInput(value: string): boolean {
  if (value.trim() === "") {
    return true;
  }
  if (!isoDateRe.test(value.trim())) {
    return false;
  }
  return !Number.isNaN(Date.parse(value));
}

export function mountDateTimeEditor(
  host: HTMLElement,
  initial: string,
  onCommit: (value: string) => void,
  onCancel: () => void,
): () => void {
  const input = document.createElement("input");
  input.type = "text";
  input.placeholder = "YYYY-MM-DD or ISO datetime";
  input.className =
    "h-full w-full border border-accent bg-panel px-2 font-mono text-12-5 text-ink outline-none";
  input.value = initial;
  host.replaceChildren(input);
  input.focus();
  input.select();

  const commit = () => {
    const v = input.value.trim();
    if (v !== "" && !validateDateTimeInput(v)) {
      input.classList.add("border-env-prod");
      return;
    }
    onCommit(v);
  };

  const onKey = (e: KeyboardEvent) => {
    if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
      return;
    }
    if (e.key === "Enter") {
      e.preventDefault();
      commit();
    }
  };
  input.addEventListener("keydown", onKey);
  input.addEventListener("blur", commit, { once: true });
  return () => input.removeEventListener("keydown", onKey);
}
