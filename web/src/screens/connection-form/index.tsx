import { useEffect, useRef, useState } from "preact/hooks";
import { ApiError, request } from "../../api";
import {
  buildSaveRequest,
  buildTestRequest,
  defaultFormState,
  formFromProfile,
  validateConnectionForm,
  type FormState,
} from "../../connection-form";
import type { ConnectionProfile } from "../../connections";
import { openWorkspaceSession } from "../../connections";
import { navigate } from "../../router";
import { Button } from "../../ui/Button";
import { FilePicker } from "../../ui/file-picker";
import { Input } from "../../ui/Input";
import { Tabs } from "../../ui/Tabs";
import { pushToast } from "../../ui/Toast";
import { AdvancedTab } from "./advanced";
import { GeneralTab } from "./general";
import { SshTab } from "./ssh";
import { SslTab } from "./ssl";
import { TestResultLine } from "./test-result";
import { HostKeyDialog, type HostKeyPrompt } from "../../dialogs/host-key";
import { looksLikeConnectionURL, parseConnectionURL, takeImportDraft } from "../connections/paste";

export type ConnectionFormProps = {
  profileId?: string;
};

export function ConnectionForm({ profileId }: ConnectionFormProps) {
  const [tab, setTab] = useState("general");
  const [form, setForm] = useState<FormState>(defaultFormState());
  const [hasStoredPassword, setHasStoredPassword] = useState(false);
  const [testState, setTestState] = useState<{
    running?: boolean;
    ok?: { latencyMs: number; server: string; tls?: string; viaSSH?: boolean };
    error?: string;
  }>({});
  const [hostKeyPrompt, setHostKeyPrompt] = useState<HostKeyPrompt | null>(null);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [pickerField, setPickerField] = useState<
    "caFile" | "clientCert" | "clientKey" | "socketPath" | null
  >(null);
  const testAbort = useRef<AbortController | null>(null);
  const [pasteURL, setPasteURL] = useState("");

  useEffect(() => {
    if (!profileId) {
      const draft = takeImportDraft();
      if (draft) {
        setForm(draft);
      }
      return;
    }
    let cancelled = false;
    request<{ connections: ConnectionProfile[] }>("GET", "/api/v1/connections").then((doc) => {
      if (cancelled) {
        return;
      }
      const p = doc.connections.find((c) => c.id === profileId);
      if (p) {
        setForm(formFromProfile(p));
        setHasStoredPassword(p.has_password);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [profileId]);

  const errors = validateConnectionForm(form);
  const invalid = Object.keys(errors).length > 0;

  const runTest = async () => {
    const v = validateConnectionForm(form);
    if (Object.keys(v).length > 0) {
      setTestState({ error: Object.values(v)[0] });
      return;
    }
    testAbort.current?.abort();
    const ac = new AbortController();
    testAbort.current = ac;
    setTestState({ running: true });
    setHostKeyPrompt(null);
    try {
      const body = buildTestRequest(form, hasStoredPassword);
      const out = await request<{
        ok: boolean;
        latency_ms: number;
        server: string;
        tls?: string;
        via_ssh?: boolean;
      }>("POST", "/api/v1/connections/test", body, { signal: ac.signal });
      setTestState({
        ok: {
          latencyMs: out.latency_ms,
          server: out.server,
          tls: out.tls,
          viaSSH: out.via_ssh,
        },
      });
    } catch (e) {
      if (ac.signal.aborted) {
        return;
      }
      if (e instanceof ApiError && e.code === "host_key_needed" && e.hostKey) {
        setHostKeyPrompt(e.hostKey);
        setTestState({});
        return;
      }
      const msg = e instanceof ApiError ? e.message : "Test failed";
      setTestState({ error: msg });
    }
  };

  const trustHostKey = async () => {
    if (!hostKeyPrompt) {
      return;
    }
    try {
      await request("POST", "/api/v1/hostkeys/accept", {
        host: hostKeyPrompt.host,
        port: hostKeyPrompt.port,
        fingerprint: hostKeyPrompt.fingerprint,
      });
      setHostKeyPrompt(null);
      await runTest();
    } catch (e) {
      pushToast(e instanceof ApiError ? e.message : "Could not trust host key", "error");
    }
  };

  const save = async (andConnect: boolean) => {
    const v = validateConnectionForm(form);
    if (Object.keys(v).length > 0) {
      pushToast(Object.values(v)[0]!, "error");
      return;
    }
    const body = buildSaveRequest(form, hasStoredPassword);
    try {
      if (profileId) {
        await request("PUT", `/api/v1/connections/${encodeURIComponent(profileId)}`, body);
      } else {
        const created = await request<ConnectionProfile>("POST", "/api/v1/connections", body);
        if (andConnect) {
          await openWorkspaceSession(created.id);
          return;
        }
      }
      pushToast("Connection saved", "success");
      navigate("/connections");
      if (andConnect && profileId) {
        await openWorkspaceSession(profileId);
      }
    } catch (e) {
      pushToast(e instanceof ApiError ? e.message : "Save failed", "error");
    }
  };

  useEffect(() => {
    return () => testAbort.current?.abort();
  }, []);

  const title = profileId ? "Edit connection" : "New connection";

  const applyPaste = async () => {
    if (!looksLikeConnectionURL(pasteURL)) {
      pushToast("Not a connection URL", "error");
      return;
    }
    try {
      const { form: parsed, warnings } = await parseConnectionURL(pasteURL);
      setForm(parsed);
      for (const w of warnings) {
        pushToast(w, "default");
      }
    } catch {
      pushToast("Could not parse connection URL", "error");
    }
  };

  return (
    <div class="grid flex-1 place-items-center overflow-auto bg-canvas p-5">
      <div class="w-full max-w-[640px] overflow-hidden rounded-dialog border border-line bg-panel shadow-lg">
        <header class="border-b border-line-2 px-5 pt-4 pb-3">
          <h2 class="mb-3 text-15 font-semibold text-ink">{title}</h2>
          <div class="flex gap-2">
            <div class="flex-1">
              <Input
                mono
                ariaLabel="Paste connection URL"
                placeholder="Paste connection URL"
                value={pasteURL}
                onInput={setPasteURL}
              />
            </div>
            <Button label="Apply" onClick={() => void applyPaste()} />
          </div>
        </header>
        <Tabs
          aria-label="Connection settings"
          selectedId={tab}
          onSelect={setTab}
          tabs={[
            { id: "general", label: "General" },
            { id: "ssh", label: "SSH tunnel" },
            { id: "ssl", label: "SSL / TLS" },
            { id: "advanced", label: "Advanced" },
          ]}
        />
        {tab === "general" ? (
          <GeneralTab
            form={form}
            errors={errors}
            onChange={setForm}
            onBrowseSocketPath={() => {
              setPickerField("socketPath");
              setPickerOpen(true);
            }}
          />
        ) : null}
        {tab === "ssl" ? (
          <SslTab
            form={form}
            onChange={setForm}
            onBrowseFile={(field) => {
              setPickerField(field);
              setPickerOpen(true);
            }}
          />
        ) : null}
        {tab === "advanced" ? (
          <AdvancedTab form={form} onChange={setForm} onBrowseFile={() => setPickerOpen(true)} />
        ) : null}
        {tab === "ssh" ? (
          <SshTab
            form={form}
            onChange={setForm}
            onBrowseKey={() => {
              setPickerField(null);
              setPickerOpen(true);
            }}
          />
        ) : null}
        <footer class="flex items-center gap-2 border-t border-line-2 bg-raise px-5 py-3">
          <TestResultLine
            running={testState.running}
            ok={testState.ok}
            error={testState.error}
          />
          <Button label="Test" onClick={() => void runTest()} disabled={testState.running} />
          <Button label="Save" disabled={invalid} onClick={() => void save(false)} />
          <Button
            label="Save and connect"
            variant="primary"
            disabled={invalid}
            onClick={() => void save(true)}
          />
        </footer>
      </div>
      <FilePicker
        open={pickerOpen}
        onClose={() => {
          setPickerOpen(false);
          setPickerField(null);
        }}
        onSelect={(path) => {
          if (pickerField === "socketPath") {
            setForm((f) => ({ ...f, socketPath: path }));
          } else if (pickerField) {
            setForm((f) => ({ ...f, ssl: { ...f.ssl, [pickerField]: path } }));
          } else {
            setForm((f) => ({ ...f, ssh: { ...f.ssh, keyPath: path } }));
          }
          setPickerOpen(false);
          setPickerField(null);
        }}
        filter={(name) => !pickerField || name.includes(".") || name.length > 0}
      />
      <HostKeyDialog
        open={hostKeyPrompt !== null}
        prompt={hostKeyPrompt}
        onTrust={() => void trustHostKey()}
        onCancel={() => setHostKeyPrompt(null)}
      />
    </div>
  );
}
