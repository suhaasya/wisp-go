import type { WindowOverlay, WindowPayload } from "./window";

export type CellOverlayKind = "mod" | "del" | "ins";

export function overlayAt(
  w: WindowPayload | undefined,
  localRow: number,
  col: number,
): CellOverlayKind | null {
  if (!w?.overlay) {
    return null;
  }
  const ov = normalizeOverlay(w.overlay);
  if (ov.del.includes(localRow)) {
    return "del";
  }
  if (ov.ins.includes(localRow)) {
    return "ins";
  }
  for (const m of ov.mod) {
    if (m[0] === localRow && m[1] === col) {
      return "mod";
    }
  }
  return null;
}

export function normalizeOverlay(ov: WindowOverlay): WindowOverlay {
  return {
    mod: ov.mod ?? [],
    ins: ov.ins ?? [],
    del: ov.del ?? [],
  };
}
