import { formatTestSuccess } from "../../connection-form";

export type TestResultProps = {
  ok?: { latencyMs: number; server: string; tls?: string; viaSSH?: boolean };
  error?: string;
  running?: boolean;
};

export function TestResultLine({ ok, error, running }: TestResultProps) {
  if (running) {
    return <p class="flex-1 text-12-5 text-ink-2">Testing connection…</p>;
  }
  if (error) {
    return <p class="flex-1 text-12-5 text-env-prod">{error}</p>;
  }
  if (ok) {
    return (
      <p class="flex-1 text-12-5 text-env-local">
        {formatTestSuccess(ok.latencyMs, ok.server, ok.tls, ok.viaSSH)}
      </p>
    );
  }
  return <p class="flex-1 text-12-5 text-ink-3">Test the connection before saving.</p>;
}
