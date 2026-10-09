package store_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/store"
)

func sampleProfile() store.Profile {
	return store.Profile{
		Name:         "Staging DB",
		Folder:       "Work",
		Engine:       "postgres",
		Transport:    store.Transport{Kind: "tcp", Host: "db.example.com", Port: 5432},
		User:         "app",
		Database:     "appdb",
		PasswordMode: "keychain",
		Env:          "staging",
		ReadOnly:     true,
		SafeMode:     false,
	}
}

func TestProfileStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "connections.json")
	st := store.NewProfileStore(path)
	if _, err := st.Load(); err != nil {
		t.Fatal(err)
	}
	want := sampleProfile()
	created, err := st.Create(want)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" {
		t.Fatal("expected generated id")
	}

	st2 := store.NewProfileStore(path)
	doc, err := st2.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Connections) != 1 {
		t.Fatalf("connections = %d", len(doc.Connections))
	}
	got := doc.Connections[0]
	if got.ID != created.ID || got.Name != want.Name || got.Engine != want.Engine ||
		got.Transport.Host != want.Transport.Host || got.Env != want.Env || got.ReadOnly != want.ReadOnly {
		t.Fatalf("got = %+v want id=%s name=%q", got, created.ID, want.Name)
	}
}

func TestProfileStoreCorruptFileRecovery(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "connections.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	st := store.NewProfileStore(path)
	doc, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if doc.Warning == "" {
		t.Fatal("expected warning")
	}
	if len(doc.Connections) != 0 {
		t.Fatal("expected empty connections")
	}
	entries, _ := os.ReadDir(dir)
	var foundBak bool
	for _, e := range entries {
		if strings.Contains(e.Name(), ".bak-") {
			foundBak = true
		}
	}
	if !foundBak {
		t.Fatal("expected backup file")
	}
}

func TestProfileValidateRejectsBadEngine(t *testing.T) {
	p := sampleProfile()
	p.Engine = "oracle"
	if err := store.ValidateProfile(p); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewUUIDv7Format(t *testing.T) {
	id := store.NewUUIDv7()
	if len(id) != 36 || id[14] != '7' {
		t.Fatalf("id = %q", id)
	}
	_ = time.Now()
}
