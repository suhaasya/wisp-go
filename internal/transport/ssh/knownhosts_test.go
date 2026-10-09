package ssh_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"os"
	"path/filepath"
	"testing"

	gossh "golang.org/x/crypto/ssh"

	"github.com/suhaasya/wisp-go/internal/transport/ssh"
)

func testHostKey(t *testing.T) gossh.PublicKey {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := gossh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	return signer.PublicKey()
}

func TestHostKeyCallbackUnknown(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "known_hosts")
	cb, err := ssh.HostKeyCallback(path)
	if err != nil {
		t.Fatal(err)
	}
	key := testHostKey(t)
	err = cb("bastion.test:22", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}, key)
	var need *ssh.HostKeyNeeded
	if !errorsAsHostKeyNeeded(err, &need) {
		t.Fatalf("want HostKeyNeeded, got %v", err)
	}
	if need.Fingerprint == "" || need.KeyType == "" {
		t.Fatalf("missing fingerprint/type: %+v", need)
	}
}

func TestHostKeyCallbackMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "known_hosts")
	key := testHostKey(t)
	if err := ssh.AppendKnownHost(path, "bastion.test", 22, key); err != nil {
		t.Fatal(err)
	}
	cb, err := ssh.HostKeyCallback(path)
	if err != nil {
		t.Fatal(err)
	}
	other := testHostKey(t)
	err = cb("bastion.test:22", &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 22}, other)
	var mismatch *ssh.HostKeyMismatch
	if !errorsAsHostKeyMismatch(err, &mismatch) {
		t.Fatalf("want HostKeyMismatch, got %v", err)
	}
}

func TestAppendKnownHostCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ssh", "known_hosts")
	key := testHostKey(t)
	if err := ssh.AppendKnownHost(path, "host", 22, key); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("empty known_hosts")
	}
}

func errorsAsHostKeyNeeded(err error, target **ssh.HostKeyNeeded) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*ssh.HostKeyNeeded); ok {
		*target = e
		return true
	}
	return false
}

func errorsAsHostKeyMismatch(err error, target **ssh.HostKeyMismatch) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*ssh.HostKeyMismatch); ok {
		*target = e
		return true
	}
	return false
}
