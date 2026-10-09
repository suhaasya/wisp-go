//go:build !windows

package update

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReplaceExecutable atomically replaces exePath with newBytes.
func ReplaceExecutable(exePath string, newBytes []byte) error {
	dir := filepath.Dir(exePath)
	tmp, err := os.CreateTemp(dir, ".wisp-update-*")
	if err != nil {
		return fmt.Errorf("update: temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(newBytes); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("update: write: %w", err)
	}
	if err := tmp.Chmod(0o755); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("update: chmod: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("update: close: %w", err)
	}
	if err := os.Rename(tmpName, exePath); err != nil {
		return fmt.Errorf("update: rename: %w", err)
	}
	return nil
}
