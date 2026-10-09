package session_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

type memBroadcast struct {
	types []string
}

func (m *memBroadcast) Broadcast(typ string, _ any) {
	m.types = append(m.types, typ)
}

func TestInvalidateMetadataBroadcastsSchema(t *testing.T) {
	bc := &memBroadcast{}
	mgr := session.NewManager(store.NewProfileStore(t.TempDir()+"/c.json"), nil, bc)
	s := session.NewTestSession(store.Profile{Engine: "postgres"})
	mgr.AdoptSession(s)
	s.InvalidateMetadata()
	found := false
	for _, typ := range bc.types {
		if typ == "schema" {
			found = true
		}
	}
	if !found {
		t.Fatalf("broadcasts=%v want schema", bc.types)
	}
}
