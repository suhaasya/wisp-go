//go:build windows

package update

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReplaceExecutable replaces exePath with newBytes (rename swap on Windows).
func ReplaceExecutable(exePath string, newBytes []byte) error {
	dir := filepath.Dir(exePath)
	tmp, err := os.CreateTemp(dir, ".wisp-update-*.exe")
	if err != nil {
		return fmt.Errorf("update: temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(newBytes); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("update: write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("update: close: %w", err)
	}
	oldName := exePath + ".old"
	_ = os.Remove(oldName)
	if err := os.Rename(exePath, oldName); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("update: backup: %w", err)
	}
	if err := os.Rename(tmpName, exePath); err != nil {
		_ = os.Rename(oldName, exePath)
		_ = os.Remove(tmpName)
		return fmt.Errorf("update: rename: %w", err)
	}
	_ = os.Remove(oldName)
	return nil
}
