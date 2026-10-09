import type { FormState } from "../../connection-form";
import { sslDisableWarning, sslModes } from "../../connection-form/ssl-helpers";
import { Button } from "../../ui/Button";
import { Input } from "../../ui/Input";
import { Segmented } from "../../ui/Segmented";

export function SslTab({
  form,
  onChange,
  onBrowseFile,
}: {
  form: FormState;
  onChange: (next: FormState) => void;
  onBrowseFile: (field: "caFile" | "clientCert" | "clientKey") => void;
}) {
  const ssl = form.ssl;
  const setSsl = (patch: Partial<FormState["ssl"]>) =>
    onChange({ ...form, ssl: { ...ssl, ...patch } });

  const warn = sslDisableWarning(form.host, ssl.mode);

  return (
    <div class="grid grid-cols-6 gap-x-3.5 gap-y-3 px-5 py-4" role="tabpanel">
      {warn ? (
        <div class="col-span-6 rounded-control border border-env-prod/30 bg-env-prod/10 px-3 py-2 text-12-5 text-env-prod">
          TLS is disabled while connecting to a remote host. Traffic is not encrypted.
        </div>
      ) : null}
      <div class="col-span-6">
        <span class="mb-1 block text-12-5 font-medium text-ink-2">SSL mode</span>
        <Segmented
          aria-label="SSL mode"
          options={sslModes}
          value={ssl.mode}
          onChange={(mode) => setSsl({ mode })}
        />
      </div>
      <div class="col-span-6">
        <span class="mb-1 block text-12-5 font-medium text-ink-2">CA certificate</span>
        <div class="mb-2">
          <Segmented
            aria-label="CA source"
            options={[
              { value: "system", label: "System trust store" },
              { value: "file", label: "CA file" },
            ]}
            value={ssl.caSource}
            onChange={(v) => setSsl({ caSource: v as FormState["ssl"]["caSource"] })}
          />
        </div>
        {ssl.caSource === "file" ? (
          <div class="flex gap-2">
            <div class="min-w-0 flex-1">
              <Input
                mono
                ariaLabel="CA certificate file"
                placeholder="/path/to/ca.pem"
                value={ssl.caFile}
                onInput={(caFile) => setSsl({ caFile })}
              />
            </div>
            <Button label="Browse…" size="sm" onClick={() => onBrowseFile("caFile")} />
          </div>
        ) : (
          <Input
            mono
            ariaLabel="CA certificate"
            value="Use system trust store"
            disabled
            onInput={() => {}}
          />
        )}
      </div>
      <div class="col-span-3">
        <Input
          label="Client certificate"
          mono
          placeholder="Optional .crt or .pem"
          value={ssl.clientCert}
          onInput={(clientCert) => setSsl({ clientCert })}
        />
        <div class="mt-1">
          <Button label="Browse…" size="sm" onClick={() => onBrowseFile("clientCert")} />
        </div>
      </div>
      <div class="col-span-3">
        <Input
          label="Client key"
          mono
          placeholder="Optional .key"
          value={ssl.clientKey}
          onInput={(clientKey) => setSsl({ clientKey })}
        />
        <div class="mt-1">
          <Button label="Browse…" size="sm" onClick={() => onBrowseFile("clientKey")} />
        </div>
      </div>
    </div>
  );
}
