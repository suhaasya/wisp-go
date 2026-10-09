import { useEffect } from "preact/hooks";
import { toasts, type ToastItem, type ToastTone } from "../state";
import { cx } from "./cx";

export type { ToastItem, ToastTone };
export { toasts } from "../state";

let toastSeq = 0;

export function pushToast(message: string, tone: ToastTone = "default", durationMs = 4000): void {
  const id = `toast-${++toastSeq}`;
  toasts.value = [...toasts.value, { id, message, tone, durationMs }];
}

export function dismissToast(id: string): void {
  toasts.value = toasts.value.filter((t) => t.id !== id);
}

function ToastEntry({ item }: { item: ToastItem }) {
  useEffect(() => {
    const ms = item.durationMs ?? 4000;
    const timer = window.setTimeout(() => dismissToast(item.id), ms);
    return () => window.clearTimeout(timer);
  }, [item.id, item.durationMs]);

  const toneClass =
    item.tone === "error"
      ? "border-env-prod bg-deleted text-ink"
      : item.tone === "success"
        ? "border-env-local bg-inserted text-ink"
        : "border-line bg-panel text-ink";

  return (
    <div
      role="status"
      class={cx("rounded-control border px-3 py-2 text-13 shadow-lg", toneClass)}
    >
      {item.message}
    </div>
  );
}

export function ToastViewport() {
  if (toasts.value.length === 0) {
    return null;
  }
  return (
    <div
      class="pointer-events-none fixed right-4 bottom-4 z-50 flex max-w-sm flex-col gap-2"
      aria-live="polite"
    >
      {toasts.value.map((item) => (
        <ToastEntry key={item.id} item={item} />
      ))}
    </div>
  );
}
