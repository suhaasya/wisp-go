//go:build unix

package server

import (
	"fmt"
	"os"
	"syscall"
)

// AcquireInstanceLock takes an exclusive non-blocking flock on path.
func AcquireInstanceLock(path string) (*InstanceLock, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600) //nolint:gosec // G304: runtime lock under user cache
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if err == syscall.EWOULDBLOCK {
			return nil, ErrInstanceAlreadyRunning
		}
		return nil, fmt.Errorf("flock: %w", err)
	}
	return &InstanceLock{release: func() error {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		return f.Close()
	}}, nil
}
