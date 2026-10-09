import type { FormState } from "../../connection-form";
import { Button } from "../../ui/Button";
import { Input } from "../../ui/Input";

export function AdvancedTab({
  form,
  onChange,
  onBrowseFile,
}: {
  form: FormState;
  onChange: (next: FormState) => void;
  onBrowseFile?: () => void;
}) {
  const a = form.advanced;
  const setAdv = (patch: Partial<FormState["advanced"]>) =>
    onChange({ ...form, advanced: { ...a, ...patch } });

  return (
    <div class="grid grid-cols-6 gap-x-3.5 gap-y-3 px-5 py-4" role="tabpanel">
      <div class="col-span-3">
        <Input
          label="Default schema"
          value={a.defaultSchema}
          onInput={(defaultSchema) => setAdv({ defaultSchema })}
        />
      </div>
      <div class="col-span-3">
        <Input label="Session time zone" value={a.timeZone} onInput={(timeZone) => setAdv({ timeZone })} />
      </div>
      <div class="col-span-2">
        <Input
          label="Statement timeout (seconds)"
          value={a.statementTimeoutSec}
          onInput={(statementTimeoutSec) => setAdv({ statementTimeoutSec })}
        />
      </div>
      <div class="col-span-2">
        <Input
          label="Keep-alive (seconds)"
          value={a.keepAliveSec}
          onInput={(keepAliveSec) => setAdv({ keepAliveSec })}
        />
      </div>
      <div class="col-span-6">
        <Input
          label="Run on connect"
          mono
          value={a.runOnConnect}
          onInput={(runOnConnect) => setAdv({ runOnConnect })}
        />
      </div>
      {onBrowseFile ? (
        <div class="col-span-6">
          <Button label="Browse file…" size="sm" onClick={onBrowseFile} />
        </div>
      ) : null}
    </div>
  );
}
