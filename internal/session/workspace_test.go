package session_test

import (
	"path/filepath"
	"testing"

	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestWorkspaceStoreSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	paths := store.PathsFromDirs(dir, dir)
	_ = paths.EnsureConfig()
	profPath := paths.ConnectionsJSON()
	ps := store.NewProfileStore(profPath)
	created, err := ps.Create(store.Profile{
		Name: "Local", Engine: "postgres", Env: "local", PasswordMode: "none",
		Transport: store.Transport{Kind: "tcp", Host: "127.0.0.1", Port: 5432},
	})
	if err != nil {
		t.Fatal(err)
	}
	ws := session.NewWorkspaceStore(paths, ps)
	profileID := created.ID
	in := session.WorkspaceState{
		Version:     1,
		ActiveTabID: "tab-a",
		Tabs: []session.WorkspaceTabState{
			{
				ID: "tab-a", Kind: session.TabKindTableData, Schema: "public", Table: "orders",
				ScrollTop: 120, Pinned: true,
			},
			{ID: "q1", Kind: session.TabKindQuery, QueryText: "select 1"},
		},
	}
	if err := ws.Save(profileID, in); err != nil {
		t.Fatal(err)
	}
	out, err := ws.Load(profileID)
	if err != nil {
		t.Fatal(err)
	}
	if out.ActiveTabID != "tab-a" {
		t.Fatalf("active=%q", out.ActiveTabID)
	}
	if len(out.Tabs) != 2 || out.Tabs[0].ScrollTop != 120 || out.Tabs[1].QueryText != "select 1" {
		t.Fatalf("tabs=%+v", out.Tabs)
	}
	if _, err := filepath.Abs(paths.WorkspaceJSON(profileID)); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceStoreUnknownProfile(t *testing.T) {
	dir := t.TempDir()
	paths := store.PathsFromDirs(dir, dir)
	ps := store.NewProfileStore(paths.ConnectionsJSON())
	_, _ = ps.Load()
	ws := session.NewWorkspaceStore(paths, ps)
	_, err := ws.Load("missing")
	if err == nil {
		t.Fatal("want error")
	}
}
