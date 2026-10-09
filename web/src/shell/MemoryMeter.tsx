import { cx } from "../ui/cx";

export interface MemoryMeterProps {
  label: string;
  mb: number;
  capMB: number;
  mono?: boolean;
}

export function MemoryMeter({ label, mb, capMB, mono }: MemoryMeterProps) {
  const pct = Math.min(100, Math.max(0, (mb / capMB) * 100));
  return (
    <span class={cx("inline-flex items-center gap-1.5 text-11-5 text-ink-2", mono && "font-mono")}>
      <span>{label}</span>
      <span class="text-ink">{mb.toFixed(0)} MB</span>
      <span class="h-1.5 w-[54px] overflow-hidden rounded-sm bg-line">
        <span class="block h-full bg-env-local" style={{ width: `${pct}%` }} />
      </span>
    </span>
  );
}
