package store

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths are Wisp config, data, cache and runtime locations for the current user.
type Paths struct {
	ConfigDir  string
	DataDir    string
	CacheDir   string
	RuntimeDir string
}

// DefaultPaths resolves paths from os.UserConfigDir and os.UserCacheDir.
func DefaultPaths() (Paths, error) {
	cfgRoot, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, fmt.Errorf("user config dir: %w", err)
	}
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return Paths{}, fmt.Errorf("user cache dir: %w", err)
	}
	return PathsFromDirs(cfgRoot, cacheRoot), nil
}

// PathsFromDirs builds paths under the given OS config and cache roots.
func PathsFromDirs(configRoot, cacheRoot string) Paths {
	baseCfg := filepath.Join(configRoot, "wisp")
	baseCache := filepath.Join(cacheRoot, "wisp")
	return Paths{
		ConfigDir:  baseCfg,
		DataDir:    filepath.Join(baseCfg, "data"),
		CacheDir:   baseCache,
		RuntimeDir: filepath.Join(baseCache, "runtime"),
	}
}

// InstanceJSON is the running instance metadata file.
func (p Paths) InstanceJSON() string {
	return filepath.Join(p.RuntimeDir, "instance.json")
}

// InstanceLock is the exclusive single-instance lock file.
func (p Paths) InstanceLock() string {
	return filepath.Join(p.RuntimeDir, "instance.lock")
}

// HandoffSecret is the shared secret for /internal/handoff.
func (p Paths) HandoffSecret() string {
	return filepath.Join(p.RuntimeDir, "handoff.secret")
}

// SettingsJSON is the persisted preferences file.
func (p Paths) SettingsJSON() string {
	return filepath.Join(p.ConfigDir, "settings.json")
}

// SecretsBin is the encrypted secrets fallback file.
func (p Paths) SecretsBin() string {
	return filepath.Join(p.DataDir, "secrets.bin")
}

// EnsureConfig creates config (and parent) directories with restrictive permissions.
func (p Paths) EnsureConfig() error {
	if err := os.MkdirAll(p.ConfigDir, 0o700); err != nil {
		return err
	}
	return nil
}

// EnsureRuntime creates runtime (and parent) directories with restrictive permissions.
func (p Paths) EnsureRuntime() error {
	if err := os.MkdirAll(p.RuntimeDir, 0o700); err != nil {
		return err
	}
	return nil
}
