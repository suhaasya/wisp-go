package store_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/store"
	"github.com/zalando/go-keyring"
)

func TestFileSecretsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.bin")
	master := []byte("correct-horse-battery")
	want := []byte("hunter2")

	st := store.NewFileSecrets(path)
	if err := st.Unlock(master); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if err := st.Set("prof1", store.SecretKindDBPassword, want); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok, err := st.Get("prof1", store.SecretKindDBPassword)
	if err != nil || !ok {
		t.Fatalf("Get: ok=%v err=%v", ok, err)
	}
	if string(got) != string(want) {
		t.Fatalf("Get = %q want %q", got, want)
	}
	st.Close()

	st2 := store.NewFileSecrets(path)
	if err := st2.Unlock(master); err != nil {
		t.Fatalf("Unlock reload: %v", err)
	}
	got, ok, err = st2.Get("prof1", store.SecretKindDBPassword)
	if err != nil || !ok {
		t.Fatalf("Get reload: ok=%v err=%v", ok, err)
	}
	if string(got) != string(want) {
		t.Fatalf("Get reload = %q want %q", got, want)
	}
}

func TestFileSecretsWrongMasterPasswordFailsCleanly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.bin")
	master := []byte("correct-horse-battery")
	wrong := []byte("wrong-password")

	st := store.NewFileSecrets(path)
	if err := st.Unlock(master); err != nil {
		t.Fatalf("Unlock(correct): %v", err)
	}
	if err := st.Set("prof1", store.SecretKindDBPassword, []byte("hunter2")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	st.Close()

	st2 := store.NewFileSecrets(path)
	err := st2.Unlock(wrong)
	if err == nil {
		t.Fatal("Unlock(wrong) expected error")
	}
	if !store.IsUnlockError(err) {
		t.Fatalf("expected unlock error, got %T: %v", err, err)
	}
	_, ok, err := st2.Get("prof1", store.SecretKindDBPassword)
	if err != nil {
		t.Fatalf("Get after wrong unlock: %v", err)
	}
	if ok {
		t.Fatal("Get should miss when not unlocked")
	}
}

func TestKeyringSecretsRoundTripAndDeleteAll(t *testing.T) {
	keyring.MockInit()
	t.Cleanup(func() { keyring.MockInit() })

	sec := store.NewKeyringSecrets()
	want := []byte("ssh-secret-value")
	if err := sec.Set("prof-a", store.SecretKindSSHPassword, want); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok, err := sec.Get("prof-a", store.SecretKindSSHPassword)
	if err != nil || !ok || string(got) != string(want) {
		t.Fatalf("Get: ok=%v err=%v got=%q", ok, err, got)
	}
	if err := sec.DeleteAll("prof-a"); err != nil {
		t.Fatalf("DeleteAll: %v", err)
	}
	_, ok, err = sec.Get("prof-a", store.SecretKindSSHPassword)
	if err != nil || ok {
		t.Fatalf("after DeleteAll: ok=%v err=%v", ok, err)
	}
}

func TestSecretsSetNoPlaintextOnDisk(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "config")
	dataDir := filepath.Join(dir, "data")
	_ = os.MkdirAll(cfgDir, 0o700)
	_ = os.MkdirAll(dataDir, 0o700)
	paths := store.Paths{ConfigDir: cfgDir, DataDir: dataDir}
	needle := "super-secret-db-pass-xyzzy"

	keyring.MockInit()
	t.Cleanup(func() { keyring.MockInit() })

	sec, backend, err := store.OpenSecrets(paths, store.FakeKeyringProbe(store.KeyringProbeAvailable))
	if err != nil {
		t.Fatalf("OpenSecrets: %v", err)
	}
	if backend != store.SecretsBackendKeyring {
		t.Fatalf("backend = %q", backend)
	}
	if err := sec.Set("p1", store.SecretKindDBPassword, []byte(needle)); err != nil {
		t.Fatalf("Set: %v", err)
	}
	scanDirsForSubstring(t, []string{cfgDir, dataDir}, needle)

	secFile := store.NewFileSecrets(filepath.Join(dataDir, "secrets.bin"))
	if err := secFile.Unlock([]byte("master-pass-for-test")); err != nil {
		t.Fatalf("file Unlock: %v", err)
	}
	if err := secFile.Set("p2", store.SecretKindDBPassword, []byte(needle)); err != nil {
		t.Fatalf("file Set: %v", err)
	}
	scanDirsForSubstring(t, []string{cfgDir, dataDir}, needle)
}

func TestOpenSecretsProbeSelectsBackend(t *testing.T) {
	dir := t.TempDir()
	paths := store.Paths{ConfigDir: dir, DataDir: filepath.Join(dir, "data")}

	sec, backend, err := store.OpenSecrets(paths, store.FakeKeyringProbe(store.KeyringProbeAvailable))
	if err != nil || backend != store.SecretsBackendKeyring {
		t.Fatalf("available: backend=%q err=%v", backend, err)
	}
	if sec == nil {
		t.Fatal("nil secrets")
	}

	sec, backend, err = store.OpenSecrets(paths, store.FakeKeyringProbe(store.KeyringProbeUnavailable))
	if err != nil || backend != store.SecretsBackendFile {
		t.Fatalf("unavailable: backend=%q err=%v", backend, err)
	}
	if _, ok := sec.(*store.FileSecrets); !ok {
		t.Fatalf("want *FileSecrets, got %T", sec)
	}

	_, backend, err = store.OpenSecrets(paths, store.FakeKeyringProbe(store.KeyringProbeLocked))
	if err == nil || backend != "" {
		t.Fatalf("locked: backend=%q err=%v", backend, err)
	}
	if !store.IsSecretServiceLocked(err) {
		t.Fatalf("want ErrSecretServiceLocked, got %v", err)
	}
}

func scanDirsForSubstring(t *testing.T, roots []string, needle string) {
	t.Helper()
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(data), needle) {
				t.Errorf("plaintext %q found in %s", needle, path)
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
}
