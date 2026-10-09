import { useEffect, useRef } from "preact/hooks";
import type { ComponentChildren } from "preact";
import { cx } from "./cx";

export interface ModalProps {
  open: boolean;
  title: string;
  description?: string;
  onClose: () => void;
  children?: ComponentChildren;
  footer?: ComponentChildren;
}

export function Modal({ open, title, description, onClose, children, footer }: ModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const triggerRef = useRef<HTMLElement | null>(null);
  const titleId = "wisp-modal-title";

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) {
      return;
    }
    if (open && !dialog.open) {
      triggerRef.current = document.activeElement as HTMLElement | null;
      dialog.showModal();
    }
    if (!open && dialog.open) {
      dialog.close();
    }
  }, [open]);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) {
      return;
    }
    const onCancel = (e: Event) => {
      e.preventDefault();
      onClose();
    };
    const onCloseDialog = () => {
      triggerRef.current?.focus();
    };
    dialog.addEventListener("cancel", onCancel);
    dialog.addEventListener("close", onCloseDialog);
    return () => {
      dialog.removeEventListener("cancel", onCancel);
      dialog.removeEventListener("close", onCloseDialog);
    };
  }, [onClose]);

  return (
    <dialog
      ref={dialogRef}
      aria-labelledby={titleId}
      aria-modal="true"
      class={cx(
        "fixed inset-0 z-50 m-0 max-h-none max-w-none border-0 bg-transparent p-0",
        "backdrop:bg-ink/30",
      )}
      onClick={(e) => {
        if (e.target === dialogRef.current) {
          onClose();
        }
      }}
    >
      <div
        class="mx-auto mt-20 w-full max-w-[620px] overflow-hidden rounded-dialog border border-line bg-panel shadow-lg"
        role="document"
      >
        <header class="border-b border-line-2 px-[18px] py-3.5">
          <h2 id={titleId} class="text-15 font-semibold text-ink">
            {title}
          </h2>
          {description ? <p class="mt-1 text-12-5 text-ink-2">{description}</p> : null}
        </header>
        {children ? <div class="px-[18px] py-3.5">{children}</div> : null}
        {footer ? (
          <footer class="flex items-center gap-2 border-t border-line-2 px-[18px] py-3">{footer}</footer>
        ) : null}
      </div>
    </dialog>
  );
}
