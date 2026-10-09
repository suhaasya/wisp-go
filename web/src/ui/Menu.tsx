import { useEffect, useRef } from "preact/hooks";
import { cx } from "./cx";

export interface MenuItem {
  id: string;
  label: string;
  disabled?: boolean;
}

export interface MenuProps {
  open: boolean;
  x: number;
  y: number;
  items: MenuItem[];
  activeIndex: number;
  onSelect: (id: string) => void;
  onClose: () => void;
  onActiveIndexChange: (index: number) => void;
}

export function Menu({
  open,
  x,
  y,
  items,
  activeIndex,
  onSelect,
  onClose,
  onActiveIndexChange,
}: MenuProps) {
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) {
      return;
    }
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        onClose();
        return;
      }
      if (e.key === "ArrowDown") {
        e.preventDefault();
        const next = Math.min(activeIndex + 1, items.length - 1);
        onActiveIndexChange(next);
      }
      if (e.key === "ArrowUp") {
        e.preventDefault();
        const next = Math.max(activeIndex - 1, 0);
        onActiveIndexChange(next);
      }
      if (e.key === "Enter") {
        e.preventDefault();
        const item = items[activeIndex];
        if (item && !item.disabled) {
          onSelect(item.id);
          onClose();
        }
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [open, activeIndex, items, onActiveIndexChange, onClose, onSelect]);

  useEffect(() => {
    if (!open) {
      return;
    }
    const onPointerDown = (e: MouseEvent) => {
      if (!menuRef.current?.contains(e.target as Node)) {
        onClose();
      }
    };
    document.addEventListener("mousedown", onPointerDown);
    return () => document.removeEventListener("mousedown", onPointerDown);
  }, [open, onClose]);

  if (!open) {
    return null;
  }

  return (
    <div
      ref={menuRef}
      role="menu"
      class="fixed z-50 min-w-40 rounded-control border border-line bg-panel py-1 shadow-lg"
      style={{ left: `${x}px`, top: `${y}px` }}
    >
      {items.map((item, i) => (
        <button
          key={item.id}
          type="button"
          role="menuitem"
          disabled={item.disabled}
          class={cx(
            "block w-full border-0 bg-transparent px-3 py-1.5 text-left text-13 text-ink",
            i === activeIndex && "bg-accent-soft",
            item.disabled && "cursor-not-allowed opacity-50",
          )}
          onClick={() => {
            if (!item.disabled) {
              onSelect(item.id);
              onClose();
            }
          }}
        >
          {item.label}
        </button>
      ))}
    </div>
  );
}
