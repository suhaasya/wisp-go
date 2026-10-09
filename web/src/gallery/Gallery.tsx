import { useEffect, useState } from "preact/hooks";
import { setThemePreference, themePreference, type ThemePreference } from "../theme";
import {
  Button,
  Checkbox,
  Input,
  Menu,
  Modal,
  Segmented,
  Select,
  Tabs,
  ToastViewport,
  Tooltip,
  VirtualList,
  pushToast,
} from "../ui";

const themeOptions: ThemePreference[] = ["light", "dark", "system"];

export function Gallery() {
  const [seg, setSeg] = useState("local");
  const [tab, setTab] = useState("general");
  const [ssl, setSsl] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [menuIndex, setMenuIndex] = useState(0);
  const [menuPos, setMenuPos] = useState({ x: 0, y: 0 });

  useEffect(() => {
    if (!import.meta.env.DEV) {
      return;
    }
    const w = window as Window & { __wispGalleryStressCycles?: (n: number) => Promise<void> };
    w.__wispGalleryStressCycles = async (n: number) => {
      const { render } = await import("preact");
      for (let i = 0; i < n; i++) {
        const host = document.createElement("div");
        document.body.appendChild(host);
        await new Promise<void>((resolve) => {
          render(<Gallery />, host);
          queueMicrotask(() => {
            render(null, host);
            host.remove();
            resolve();
          });
        });
      }
    };
    return () => {
      delete w.__wispGalleryStressCycles;
    };
  }, []);

  return (
    <div class="min-h-screen bg-canvas text-ink">
      <header class="flex flex-wrap items-center justify-between gap-3 border-b border-line bg-panel px-4 py-3">
        <div>
          <h1 class="text-15 font-semibold">Component gallery</h1>
          <p class="text-12-5 text-ink-2">Dev only — tree-shaken from production builds.</p>
        </div>
        <Segmented
          aria-label="Theme"
          value={themePreference.value}
          options={themeOptions.map((v) => ({ value: v, label: v }))}
          onChange={(v) => setThemePreference(v as ThemePreference)}
        />
      </header>

      <main class="mx-auto flex max-w-4xl flex-col gap-10 p-6">
        <section aria-labelledby="sec-buttons">
          <h2 id="sec-buttons" class="mb-3 text-15 font-semibold">
            Buttons
          </h2>
          <div class="flex flex-wrap gap-2">
            <Button label="Default" />
            <Button label="Primary" variant="primary" />
            <Button label="Warning" variant="warning" />
            <Button label="Ghost" variant="ghost" />
            <Button label="Danger" variant="danger" />
            <Button label="Save" variant="primary" kbd="⌘S" size="sm" />
          </div>
        </section>

        <section aria-labelledby="sec-form">
          <h2 id="sec-form" class="mb-3 text-15 font-semibold">
            Form controls
          </h2>
          <div class="grid gap-4 rounded-card border border-line bg-panel p-4 md:grid-cols-2">
            <Input label="Host" placeholder="localhost" mono />
            <Input label="Password" type="password" />
            <Select
              label="Engine"
              value="pg"
              options={[
                { value: "pg", label: "PostgreSQL" },
                { value: "my", label: "MySQL" },
              ]}
            />
            <Checkbox label="Use SSL" checked={ssl} onChange={setSsl} />
          </div>
          <div class="mt-4">
            <Tabs
              aria-label="Connection form"
              selectedId={tab}
              tabs={[
                { id: "general", label: "General" },
                { id: "ssh", label: "SSH" },
                { id: "ssl", label: "SSL" },
              ]}
              onSelect={setTab}
            />
            <div
              id={`panel-${tab}`}
              role="tabpanel"
              aria-labelledby={`tab-${tab}`}
              class="rounded-b-card border border-t-0 border-line bg-panel p-4 text-13 text-ink-2"
            >
              {tab} pane content
            </div>
          </div>
        </section>

        <section aria-labelledby="sec-palette">
          <h2 id="sec-palette" class="mb-3 text-15 font-semibold">
            Command palette list
          </h2>
          <div class="overflow-hidden rounded-dialog border border-line bg-panel">
            <Input placeholder="Type a command…" />
            <VirtualList
              class="max-h-[330px]"
              rowHeight={34}
              items={Array.from({ length: 200 }, (_, i) => ({
                title: `Command ${i + 1}`,
                hint: "↵",
              }))}
              render={(item, i) => (
                <div
                  class={`flex items-center gap-2 px-2.5 py-1.5 text-13 ${i === 0 ? "bg-accent-soft" : ""}`}
                >
                  <span class="text-ink">{item.title}</span>
                  <small class="ml-auto text-ink-3">{item.hint}</small>
                </div>
              )}
            />
          </div>
        </section>

        <section aria-labelledby="sec-overlay">
          <h2 id="sec-overlay" class="mb-3 text-15 font-semibold">
            Overlays
          </h2>
          <div class="flex flex-wrap gap-2">
            <Button label="Review SQL" variant="primary" onClick={() => setModalOpen(true)} />
            <Button label="Toast" onClick={() => pushToast("Connection saved", "success")} />
            <Button
              label="Context menu"
              onClick={() => {
                setMenuPos({ x: 160, y: 320 });
                setMenuOpen(true);
                setMenuIndex(0);
              }}
            />
            <Tooltip label="Latency" tip="Round-trip time to the database server" />
          </div>
        </section>

        <section aria-labelledby="sec-seg">
          <h2 id="sec-seg" class="mb-3 text-15 font-semibold">
            Segmented
          </h2>
          <Segmented
            aria-label="Environment"
            value={seg}
            options={[
              { value: "local", label: "Local" },
              { value: "dev", label: "Dev" },
              { value: "prod", label: "Prod" },
            ]}
            onChange={setSeg}
          />
        </section>
      </main>

      <Modal
        open={modalOpen}
        title="Review SQL"
        description="Apply staged changes to the database."
        onClose={() => setModalOpen(false)}
        footer={
          <>
            <Button label="Cancel" onClick={() => setModalOpen(false)} />
            <Button label="Commit" variant="primary" onClick={() => setModalOpen(false)} />
          </>
        }
      >
        <pre class="rounded-control bg-raise p-3 font-mono text-12-5 text-ink">UPDATE users SET active = true;</pre>
      </Modal>

      <Menu
        open={menuOpen}
        x={menuPos.x}
        y={menuPos.y}
        activeIndex={menuIndex}
        items={[
          { id: "connect", label: "Connect" },
          { id: "query", label: "New query" },
        ]}
        onActiveIndexChange={setMenuIndex}
        onSelect={() => undefined}
        onClose={() => setMenuOpen(false)}
      />

      <ToastViewport />
    </div>
  );
}
