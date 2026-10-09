package ssh_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/transport/ssh"
)

func TestKeyAuthWrongPassphrase(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "id_ed25519")
	cmd := exec.CommandContext(context.Background(), "ssh-keygen", "-t", "ed25519", "-f", keyPath, "-N", "secret", "-q")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("ssh-keygen: %v (%s)", err, out)
	}
	_, err := ssh.AuthMethods(context.Background(), ssh.AuthInput{
		Method:     "key",
		KeyPath:    keyPath,
		Passphrase: []byte("wrong"),
	})
	if err == nil {
		t.Fatal("expected error for wrong passphrase")
	}
	low := strings.ToLower(err.Error())
	if !strings.Contains(low, "passphrase") && !strings.Contains(low, "password incorrect") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKeyAuthMissingFile(t *testing.T) {
	_, err := ssh.AuthMethods(context.Background(), ssh.AuthInput{
		Method:  "key",
		KeyPath: filepath.Join(t.TempDir(), "missing"),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !os.IsNotExist(err) && !strings.Contains(err.Error(), "read key") {
		t.Fatalf("unexpected: %v", err)
	}
}
