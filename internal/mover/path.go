package mover

import "os"

// openImportFile opens path for streaming import (path must be resolved by the server).
func openImportFile(path string) (*os.File, error) {
	return os.Open(path) //nolint:gosec // path is home-scoped by api.FSHandler before import
}
