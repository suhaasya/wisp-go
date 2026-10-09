package ssh_test

import (
	"bufio"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/transport/ssh"
)

func TestParseConfigResolveHost(t *testing.T) {
	const raw = `
Host bastion
  HostName jump.example.com
  User deploy
  Port 2222
  IdentityFile ~/.ssh/id_ed25519
  ProxyJump hop1,hop2
`
	blocks, err := ssh.ParseConfig(bufio.NewReader(strings.NewReader(raw)))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := ssh.ResolveHost(blocks, "bastion")
	if !ok {
		t.Fatal("not found")
	}
	if got.HostName != "jump.example.com" || got.User != "deploy" || got.Port != 2222 {
		t.Fatalf("fields: %+v", got)
	}
	if got.IdentityFile != "~/.ssh/id_ed25519" {
		t.Fatalf("identity: %q", got.IdentityFile)
	}
	if got.ProxyJump != "hop1,hop2" {
		t.Fatalf("proxyjump: %q", got.ProxyJump)
	}
}
