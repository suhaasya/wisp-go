package session_test

import (
	"os"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestJournalCrashSimulationRestoresWorkspace(t *testing.T) {
	root := t.TempDir()
	paths := store.PathsFromDirs(root, root)
	profiles := store.NewProfileStore(root + "/connections.json")
	created, err := profiles.Create(store.Profile{
		ID: "p1", Name: "local", Engine: "postgres", Env: "local",
		Transport: store.Transport{Kind: "tcp", Host: "127.0.0.1", Port: 5432},
		User:      "u", Database: "d", PasswordMode: "none",
	})
	if err != nil {
		t.Fatal(err)
	}
	ws := session.NewWorkspaceStore(paths, profiles)
	j := session.NewJournalStore(paths, ws)
	want := session.WorkspaceState{
		Version:     1,
		ActiveTabID: "q1",
		Tabs: []session.WorkspaceTabState{
			{ID: "q1", Kind: session.TabKindQuery, QueryText: "SELECT 1"},
		},
	}
	if err := j.Save(created.ID, session.JournalDocument{Version: 1, Workspace: want}); err != nil {
		t.Fatal(err)
	}
	path := paths.JournalJSON(created.ID)
	if err := os.WriteFile(path+".tmp", []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := j.Restore(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Workspace.Tabs) != 1 || got.Workspace.Tabs[0].QueryText != "SELECT 1" {
		t.Fatalf("workspace=%+v", got.Workspace)
	}
	loaded, err := ws.Load(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Tabs[0].QueryText != "SELECT 1" {
		t.Fatalf("persisted=%+v", loaded)
	}
}

func TestJournalDebounceFlush(t *testing.T) {
	root := t.TempDir()
	paths := store.PathsFromDirs(root, root)
	profiles := store.NewProfileStore(root + "/connections.json")
	created, err := profiles.Create(store.Profile{
		ID: "p1", Name: "x", Engine: "postgres", Env: "local",
		Transport: store.Transport{Kind: "tcp", Host: "127.0.0.1", Port: 5432},
		User:      "u", Database: "d", PasswordMode: "none",
	})
	if err != nil {
		t.Fatal(err)
	}
	ws := session.NewWorkspaceStore(paths, profiles)
	j := session.NewJournalStore(paths, ws)
	done := make(chan struct{}, 1)
	j.ScheduleFlush(created.ID, func() (session.JournalDocument, error) {
		doc := session.JournalDocument{
			Version: 1,
			Workspace: session.WorkspaceState{
				Version: 1,
				Tabs:    []session.WorkspaceTabState{{ID: "t1", Kind: session.TabKindQuery, QueryText: "x"}},
			},
		}
		select {
		case done <- struct{}{}:
		default:
		}
		return doc, nil
	})
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("debounced flush did not run")
	}
	time.Sleep(20 * time.Millisecond)
	_ = j.Clear(created.ID)
}
