import {
  applyEngineDefaultPort,
  type FieldErrors,
  type FormState,
} from "../../connection-form";
import type { ProfileEnv } from "../../connections";
import { Checkbox } from "../../ui/Checkbox";
import { Input } from "../../ui/Input";
import { Segmented } from "../../ui/Segmented";
import { cx } from "../../ui/cx";

const engines = [
  { value: "postgres", label: "PostgreSQL" },
  { value: "mysql", label: "MySQL" },
  { value: "mariadb", label: "MariaDB" },
];

const envs: ProfileEnv[] = ["local", "development", "staging", "production"];

export function GeneralTab({
  form,
  errors,
  onChange,
}: {
  form: FormState;
  errors: FieldErrors;
  onChange: (next: FormState) => void;
}) {
  const set = (patch: Partial<FormState>) => onChange({ ...form, ...patch });

  return (
    <div class="grid grid-cols-6 gap-x-3.5 gap-y-3 px-5 py-4" role="tabpanel">
      <div class="col-span-6">
        <span class="mb-1 block text-12-5 font-medium text-ink-2">Database type</span>
        <Segmented
          aria-label="Database type"
          options={engines}
          value={form.engine}
          onChange={(v) => onChange(applyEngineDefaultPort(form, v as FormState["engine"]))}
        />
      </div>
      <div class="col-span-3">
        <Input label="Name" value={form.name} onInput={(name) => set({ name })} />
        {errors.name ? <p class="mt-1 text-12-5 text-env-prod">{errors.name}</p> : null}
      </div>
      <div class="col-span-3">
        <Input label="Folder" value={form.folder} onInput={(folder) => set({ folder })} />
      </div>
      <div class="col-span-6">
        <span class="mb-1 block text-12-5 font-medium text-ink-2">Connect using</span>
        <Segmented
          aria-label="Connect using"
          options={[
            { value: "tcp", label: "Host and port" },
            { value: "socket", label: "Socket" },
            { value: "pipe", label: "Named pipe" },
          ]}
          value={form.transportKind}
          onChange={(transportKind) => set({ transportKind: transportKind as FormState["transportKind"] })}
        />
        {form.transportKind !== "tcp" ? (
          <p class="mt-1 text-12-5 text-ink-3">Socket and named pipe support ships in LW-043.</p>
        ) : null}
      </div>
      <div class="col-span-4">
        <Input
          label="Host"
          value={form.host}
          disabled={form.transportKind !== "tcp"}
          onInput={(host) => set({ host })}
        />
        {errors.host ? <p class="mt-1 text-12-5 text-env-prod">{errors.host}</p> : null}
      </div>
      <div class="col-span-2">
        <Input
          label="Port"
          value={String(form.port)}
          disabled={form.transportKind !== "tcp"}
          onInput={(v) => {
            const port = Number.parseInt(v, 10);
            set({ port: Number.isNaN(port) ? 0 : port, portTouched: true });
          }}
        />
        {errors.port ? <p class="mt-1 text-12-5 text-env-prod">{errors.port}</p> : null}
      </div>
      <div class="col-span-2">
        <Input label="User" value={form.user} onInput={(user) => set({ user })} />
        {errors.user ? <p class="mt-1 text-12-5 text-env-prod">{errors.user}</p> : null}
      </div>
      <div class="col-span-2">
        <Input
          label="Password"
          type="password"
          value={form.password}
          placeholder={form.id ? "Leave blank to keep saved password" : ""}
          onInput={(password) => set({ password, passwordDirty: true })}
        />
      </div>
      <div class="col-span-2">
        <Input label="Database" value={form.database} onInput={(database) => set({ database })} />
        {errors.database ? <p class="mt-1 text-12-5 text-env-prod">{errors.database}</p> : null}
      </div>
      <div class="col-span-6">
        <Checkbox
          label="Save password in keychain"
          checked={form.savePassword}
          onChange={(savePassword) => set({ savePassword })}
        />
      </div>
      <div class="col-span-6">
        <span class="mb-1 block text-12-5 font-medium text-ink-2">Environment</span>
        <div class="flex flex-wrap gap-1.5">
          {envs.map((env) => (
            <button
              key={env}
              type="button"
              aria-pressed={form.env === env}
              onClick={() => set({ env })}
              class={cx(
                "flex items-center gap-1.5 rounded-full border bg-panel px-2 py-0.5 text-12-5",
                form.env === env
                  ? cx("font-medium text-ink", envSelectedStyle(env))
                  : "border-line text-ink-2",
              )}
            >
              <span class={cx("h-2 w-2 rounded-full", envDot(env))} />
              {env === "development" ? "Development" : env.charAt(0).toUpperCase() + env.slice(1)}
            </button>
          ))}
        </div>
        {errors.productionSafe ? (
          <p class="mt-1 text-12-5 text-env-prod">{errors.productionSafe}</p>
        ) : null}
      </div>
      <div class="col-span-6 flex flex-col gap-2">
        <Checkbox label="Read-only" checked={form.readOnly} onChange={(readOnly) => set({ readOnly })} />
        <Checkbox
          label="Confirm every write (safe mode)"
          checked={form.safeMode}
          onChange={(safeMode) => set({ safeMode })}
        />
      </div>
    </div>
  );
}

function envSelectedStyle(env: ProfileEnv): string {
  switch (env) {
    case "local":
      return "border-env-local ring-1 ring-inset ring-env-local";
    case "development":
      return "border-env-dev ring-1 ring-inset ring-env-dev";
    case "staging":
      return "border-env-staging ring-1 ring-inset ring-env-staging";
    case "production":
      return "border-env-prod ring-1 ring-inset ring-env-prod";
  }
}

function envDot(env: ProfileEnv): string {
  switch (env) {
    case "local":
      return "bg-env-local";
    case "development":
      return "bg-env-dev";
    case "staging":
      return "bg-env-staging";
    case "production":
      return "bg-env-prod";
  }
}
