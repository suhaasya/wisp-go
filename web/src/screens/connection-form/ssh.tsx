import type { FormState } from "../../connection-form";
import { sshAuthMethods } from "../../connection-form/ssh-helpers";
import { Button } from "../../ui/Button";
import { Checkbox } from "../../ui/Checkbox";
import { Input } from "../../ui/Input";
import { Segmented } from "../../ui/Segmented";

export function SshTab({
  form,
  onChange,
  onBrowseKey,
}: {
  form: FormState;
  onChange: (next: FormState) => void;
  onBrowseKey: () => void;
}) {
  const ssh = form.ssh;
  const setSsh = (patch: Partial<FormState["ssh"]>) => onChange({ ...form, ssh: { ...ssh, ...patch } });

  const setJump = (i: number, patch: Partial<FormState["ssh"]["jumps"][0]>) => {
    const jumps = ssh.jumps.map((j, idx) => (idx === i ? { ...j, ...patch } : j));
    setSsh({ jumps });
  };

  return (
    <div class="grid grid-cols-6 gap-x-3.5 gap-y-3 px-5 py-4" role="tabpanel">
      <div class="col-span-6">
        <Checkbox
          label="Connect through an SSH tunnel"
          checked={ssh.enabled}
          onChange={(enabled) => setSsh({ enabled })}
        />
      </div>
      <div class="col-span-4">
        <Input label="SSH host" mono value={ssh.host} onInput={(host) => setSsh({ host })} />
      </div>
      <div class="col-span-2">
        <Input
          label="Port"
          mono
          value={String(ssh.port)}
          onInput={(v) => setSsh({ port: Number(v) || 22 })}
        />
      </div>
      <div class="col-span-3">
        <Input label="SSH user" mono value={ssh.user} onInput={(user) => setSsh({ user })} />
      </div>
      <div class="col-span-3">
        <span class="mb-1 block text-12-5 font-medium text-ink-2">Authentication</span>
        <Segmented
          aria-label="SSH authentication"
          options={sshAuthMethods}
          value={ssh.auth}
          onChange={(auth) => setSsh({ auth })}
        />
      </div>
      {ssh.auth === "config" ? (
        <div class="col-span-6">
          <Input
            label="~/.ssh/config Host alias"
            mono
            value={ssh.configHost}
            onInput={(configHost) => setSsh({ configHost })}
          />
        </div>
      ) : null}
      {ssh.auth === "key" ? (
        <>
          <div class="col-span-4">
            <Input label="Private key" mono value={ssh.keyPath} onInput={(keyPath) => setSsh({ keyPath })} />
          </div>
          <div class="col-span-2 flex items-end">
            <Button label="Browse…" size="sm" onClick={onBrowseKey} />
          </div>
          <div class="col-span-6">
            <Input
              label="Passphrase"
              type="password"
              value={ssh.passphrase}
              onInput={(passphrase) => setSsh({ passphrase, passphraseDirty: true })}
            />
          </div>
        </>
      ) : null}
      {ssh.auth === "password" ? (
        <div class="col-span-6">
          <Input
            label="SSH password"
            type="password"
            value={ssh.password}
            onInput={(password) => setSsh({ password, passwordDirty: true })}
          />
        </div>
      ) : null}
      {ssh.jumps.map((j, i) => (
        <div class="col-span-6 grid grid-cols-6 gap-x-3.5 gap-y-2 rounded-control border border-line p-3" key={i}>
          <div class="col-span-6 flex items-center justify-between">
            <span class="text-12-5 font-medium text-ink-2">Jump host {i + 1}</span>
            <Button
              label="Remove"
              size="sm"
              onClick={() => setSsh({ jumps: ssh.jumps.filter((_, idx) => idx !== i) })}
            />
          </div>
          <div class="col-span-4">
            <Input label="Host" mono value={j.host} onInput={(host) => setJump(i, { host })} />
          </div>
          <div class="col-span-2">
            <Input
              label="Port"
              mono
              value={String(j.port)}
              onInput={(v) => setJump(i, { port: Number(v) || 22 })}
            />
          </div>
          <div class="col-span-3">
            <Input label="User" mono value={j.user} onInput={(user) => setJump(i, { user })} />
          </div>
          <div class="col-span-3">
            <Input label="Key path" mono value={j.keyPath} onInput={(keyPath) => setJump(i, { keyPath })} />
          </div>
        </div>
      ))}
      <div class="col-span-6">
        <Button
          label="Add jump host"
          size="sm"
          disabled={ssh.jumps.length >= 2}
          onClick={() =>
            setSsh({
              jumps: [
                ...ssh.jumps,
                { host: "", port: 22, user: "", auth: "key", keyPath: "", configHost: "" },
              ],
            })
          }
        />
      </div>
    </div>
  );
}
