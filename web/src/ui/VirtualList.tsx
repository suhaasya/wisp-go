import type { ComponentChild } from "preact";
import { useEffect, useRef, useState } from "preact/hooks";

export interface VirtualListProps<T> {
  items: T[];
  rowHeight: number;
  class?: string;
  render: (item: T, index: number) => ComponentChild;
}

const overscan = 3;

export function VirtualList<T>({ items, rowHeight, class: className, render }: VirtualListProps<T>) {
  const scrollerRef = useRef<HTMLDivElement>(null);
  const [range, setRange] = useState({ start: 0, end: 0 });

  useEffect(() => {
    const el = scrollerRef.current;
    if (!el) {
      return;
    }
    const update = () => {
      const height = el.clientHeight || 0;
      const start = Math.max(0, Math.floor(el.scrollTop / rowHeight) - overscan);
      const visible = Math.ceil(height / rowHeight) + overscan * 2;
      const end = Math.min(items.length, start + visible);
      setRange({ start, end });
    };
    update();
    el.addEventListener("scroll", update, { passive: true });
    window.addEventListener("resize", update);
    return () => {
      el.removeEventListener("scroll", update);
      window.removeEventListener("resize", update);
    };
  }, [items.length, rowHeight]);

  const totalHeight = items.length * rowHeight;
  const slice = items.slice(range.start, range.end);

  return (
    <div ref={scrollerRef} class={className} style={{ overflow: "auto", position: "relative" }}>
      <div style={{ height: `${totalHeight}px`, position: "relative" }}>
        {slice.map((item, i) => {
          const index = range.start + i;
          return (
            <div
              key={index}
              style={{
                position: "absolute",
                top: `${index * rowHeight}px`,
                left: "0",
                right: "0",
                height: `${rowHeight}px`,
              }}
            >
              {render(item, index)}
            </div>
          );
        })}
      </div>
    </div>
  );
}
