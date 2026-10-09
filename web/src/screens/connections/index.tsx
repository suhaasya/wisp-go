import { useCallback, useEffect, useMemo, useRef, useState } from "preact/hooks";
import {
  countByEnv,
  countByFolder,
  deleteConnectionProfile,
  filterConnections,
  useMockConnectionData,
  duplicateConnectionProfile,
  groupByFolder,
  loadConnectionProfiles,
  makeMockProfiles,
  moveConnectionToFolder,
  openWorkspaceSession,
  type ConnectionProfile,
  type ProfileEnv,
} from "../../connections";
import { navigate } from "../../router";
import { Button } from "../../ui/Button";
import { Input } from "../../ui/Input";
import { Menu } from "../../ui/Menu";
import { Modal } from "../../ui/Modal";
import { pushToast } from "../../ui/Toast";
import { ConnectionCard } from "./card";
import { ConnectionsEmpty } from "./empty";
import { ConnectionsRail } from "./rail";
import { useConnectionPaste } from "./paste";

export type ConnectionsScreenProps = {
  initialProfiles?: ConnectionProfile[];
  initialFolders?: string[];
  onConnect?: (profileId: string) => Promise<string>;
};

export function ConnectionsScreen({
  initialProfiles,
  initialFolders,
  onConnect,
}: ConnectionsScreenProps = {}) {
  useConnectionPaste(initialProfiles === undefined);
  const [profiles, setProfiles] = useState<ConnectionProfile[]>(initialProfiles ?? []);
  const [folders, setFolders] = useState<string[]>(initialFolders ?? []);
  const [search, setSearch] = useState("");
  const [activeFolder, setActiveFolder] = useState<string | "all">("all");
  const [activeEnv, setActiveEnv] = useState<ProfileEnv | "all">("all");
  const [focusedId, setFocusedId] = useState<string | null>(null);
  const [menu, setMenu] = useState<{ x: number; y: number; profileId: string } | null>(null);
  const [menuIndex, setMenuIndex] = useState(0);
  const [deleteTarget, setDeleteTarget] = useState<ConnectionProfile | null>(null);
  const [moveTarget, setMoveTarget] = useState<ConnectionProfile | null>(null);
  const [moveFolder, setMoveFolder] = useState("");
  const sectionRef = useRef<HTMLElement>(null);

  useEffect(() => {
    if (initialProfiles !== undefined) {
      return;
    }
    if (import.meta.env.DEV) {
      const bench = new URLSearchParams(location.search).get("benchConnections");
      if (bench === "200") {
        setProfiles(makeMockProfiles(200));
        setFolders(["Acme backend", "Side projects", "Local Docker"]);
        return;
      }
    }
    let cancelled = false;
    loadConnectionProfiles().then((doc) => {
      if (!cancelled) {
        setProfiles(doc.profiles);
        setFolders(doc.folders);
        if (doc.warning) {
          pushToast(doc.warning, "error");
        }
      }
    });
    return () => {
      cancelled = true;
    };
  }, [initialProfiles !== undefined]);

  const filtered = useMemo(
    () =>
      filterConnections(profiles, {
        query: search,
        folder: activeFolder,
        env: activeEnv,
      }),
    [profiles, search, activeFolder, activeEnv],
  );

  const groups = useMemo(() => groupByFolder(filtered, folders), [filtered, folders]);
  const flatIds = useMemo(() => groups.flatMap((g) => g.profiles.map((p) => p.id)), [groups]);

  const folderCounts = useMemo(() => countByFolder(profiles), [profiles]);
  const envCounts = useMemo(() => countByEnv(profiles), [profiles]);

  const doConnect = useCallback(
    async (profileId: string) => {
      try {
        if (onConnect) {
          const sid = await onConnect(profileId);
          navigate(`/w/${encodeURIComponent(sid)}`);
        } else {
          await openWorkspaceSession(profileId);
        }
      } catch {
        pushToast("Could not open session", "error");
      }
    },
    [onConnect],
  );

  const openMenu = (profileId: string, e: MouseEvent) => {
    setMenu({ x: e.clientX, y: e.clientY, profileId });
    setMenuIndex(0);
  };

  const profileById = (id: string) => profiles.find((p) => p.id === id);

  const handleMenuSelect = async (action: string) => {
    const id = menu?.profileId;
    setMenu(null);
    if (!id) {
      return;
    }
    const profile = profileById(id);
    if (!profile) {
      return;
    }
    switch (action) {
      case "edit":
        navigate(`/connections/${encodeURIComponent(id)}/edit`);
        break;
      case "duplicate": {
        try {
          if (useMockConnectionData()) {
            setProfiles((prev) => [
              ...prev,
              {
                ...profile,
                id: `${profile.id}-copy`,
                name: `${profile.name} (copy)`,
              },
            ]);
          } else {
            const dup = await duplicateConnectionProfile(id);
            setProfiles((prev) => [...prev, dup]);
          }
          pushToast("Connection duplicated", "success");
        } catch {
          pushToast("Duplicate failed", "error");
        }
        break;
      }
      case "move":
        setMoveTarget(profile);
        setMoveFolder(profile.folder);
        break;
      case "delete":
        setDeleteTarget(profile);
        break;
    }
  };

  const confirmDelete = (id: string) => {
    setDeleteTarget(null);
    if (!profileById(id)) {
      return;
    }
    setProfiles((prev) => prev.filter((x) => x.id !== id));
    void deleteConnectionProfile(id)
      .then(() => pushToast("Connection deleted", "success"))
      .catch(() => pushToast("Delete failed", "error"));
  };

  const confirmMove = async () => {
    const p = moveTarget;
    if (!p) {
      return;
    }
    const folder = moveFolder.trim();
    try {
      await moveConnectionToFolder(p, folder);
      setProfiles((prev) => prev.map((x) => (x.id === p.id ? { ...x, folder } : x)));
      if (folder && !folders.includes(folder)) {
        setFolders((f) => [...f, folder]);
      }
      setMoveTarget(null);
      pushToast("Moved to folder", "success");
    } catch {
      pushToast("Move failed", "error");
    }
  };

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (!flatIds.length || menu || deleteTarget || moveTarget) {
        return;
      }
      const idx = focusedId ? flatIds.indexOf(focusedId) : -1;
      if (e.key === "ArrowRight" || e.key === "ArrowDown") {
        e.preventDefault();
        const next = flatIds[Math.min(idx + 1, flatIds.length - 1)] ?? flatIds[0]!;
        setFocusedId(next);
        focusCard(next);
      }
      if (e.key === "ArrowLeft" || e.key === "ArrowUp") {
        e.preventDefault();
        const next = flatIds[Math.max(idx - 1, 0)]!;
        setFocusedId(next);
        focusCard(next);
      }
      if (e.key === "Enter") {
        const active = document.activeElement as HTMLElement | null;
        const id = active?.getAttribute("data-connection-id") ?? focusedId;
        if (id) {
          e.preventDefault();
          void doConnect(id);
        }
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [flatIds, focusedId, menu, deleteTarget, moveTarget, doConnect]);

  const focusCard = (id: string) => {
    const el = sectionRef.current?.querySelector(`[data-connection-id="${CSS.escape(id)}"]`) as
      | HTMLButtonElement
      | undefined;
    el?.focus();
  };

  const menuItems = [
    { id: "edit", label: "Edit" },
    { id: "duplicate", label: "Duplicate" },
    { id: "move", label: "Move to folder…" },
    { id: "delete", label: "Delete…" },
  ];

  return (
    <section ref={sectionRef} class="flex min-h-0 flex-1">
      <ConnectionsRail
        folders={folders}
        folderCounts={folderCounts}
        envCounts={envCounts}
        activeFolder={activeFolder}
        activeEnv={activeEnv}
        onFolder={setActiveFolder}
        onEnv={setActiveEnv}
      />
      <div class="min-h-0 flex-1 overflow-auto px-7 py-[22px]">
        <div class="mb-1.5 flex flex-wrap items-center gap-2.5">
          <div class="max-w-[360px] min-w-[200px] flex-1">
            <Input
              ariaLabel="Search connections"
              placeholder="Search connections"
              value={search}
              onInput={setSearch}
            />
          </div>
          <Button
            label="New connection"
            variant="primary"
            kbd="Alt+N"
            onClick={() => navigate("/connections/new")}
          />
          <Button label="Settings" variant="ghost" onClick={() => navigate("/settings")} />
          <span title="Coming in v1.1">
            <Button label="Import" disabled />
          </span>
        </div>
        <p class="mb-[18px] text-12-5 text-ink-3">
          Paste a URL anywhere to create a connection, e.g.{" "}
          <code class="font-mono text-11-5 text-ink-2">
            postgres://app@db.example.com:5432/shop?sslmode=require
          </code>
        </p>

        {filtered.length === 0 ? (
          <ConnectionsEmpty
            filtered={profiles.length > 0}
            onNewConnection={() => navigate("/connections/new")}
          />
        ) : (
          groups.map((group) => (
            <div key={group.folder} class="mb-4">
              <h3 class="mb-2.5 mt-[18px] text-13 font-semibold text-ink">{group.folder}</h3>
              <div class="grid grid-cols-[repeat(auto-fill,minmax(250px,1fr))] gap-3">
                {group.profiles.map((profile) => (
                  <ConnectionCard
                    key={profile.id}
                    profile={profile}
                    tabIndex={focusedId === profile.id || (!focusedId && profile.id === flatIds[0]) ? 0 : -1}
                    focused={focusedId === profile.id}
                    onFocus={() => setFocusedId(profile.id)}
                    onConnect={() => void doConnect(profile.id)}
                    onContextMenu={(e) => openMenu(profile.id, e)}
                  />
                ))}
              </div>
            </div>
          ))
        )}
      </div>

      <Menu
        open={menu !== null}
        x={menu?.x ?? 0}
        y={menu?.y ?? 0}
        items={menuItems}
        activeIndex={menuIndex}
        onActiveIndexChange={setMenuIndex}
        onSelect={handleMenuSelect}
        onClose={() => setMenu(null)}
      />

      <Modal
        open={deleteTarget !== null}
        title="Delete connection?"
        description={
          deleteTarget
            ? `Remove “${deleteTarget.name}” and its saved secrets. This cannot be undone.`
            : undefined
        }
        onClose={() => setDeleteTarget(null)}
        footer={
          <div class="flex justify-end gap-2">
            <Button label="Cancel" onClick={() => setDeleteTarget(null)} />
            <Button
              label="Delete"
              variant="danger"
              onClick={() => {
                if (deleteTarget) {
                  void confirmDelete(deleteTarget.id);
                }
              }}
            />
          </div>
        }
      />

      <Modal
        open={moveTarget !== null}
        title="Move to folder"
        onClose={() => setMoveTarget(null)}
        footer={
          <div class="flex justify-end gap-2">
            <Button label="Cancel" onClick={() => setMoveTarget(null)} />
            <Button label="Move" variant="primary" onClick={() => void confirmMove()} />
          </div>
        }
      >
        <Input label="Folder name" value={moveFolder} onInput={setMoveFolder} />
        <div class="mt-3 flex flex-wrap gap-1">
          {folders.map((f) => (
            <button
              key={f}
              type="button"
              class="rounded-[5px] border border-line bg-raise px-2 py-0.5 text-12-5 text-ink-2 hover:border-ink-3"
              onClick={() => setMoveFolder(f)}
            >
              {f}
            </button>
          ))}
        </div>
      </Modal>
    </section>
  );
}
