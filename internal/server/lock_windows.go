//go:build windows

package server

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

const errLockViolation = syscall.Errno(33) // ERROR_LOCK_VIOLATION

// AcquireInstanceLock takes an exclusive non-blocking lock on path.
func AcquireInstanceLock(path string) (*InstanceLock, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600) //nolint:gosec // G304: runtime lock under user cache
	if err != nil {
		return nil, err
	}
	h := windows.Handle(f.Fd())
	var ol windows.Overlapped
	err = windows.LockFileEx(h, windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 0xFFFFFFFF, 0xFFFFFFFF, &ol)
	if err != nil {
		_ = f.Close()
		if errors.Is(err, errLockViolation) {
			return nil, ErrInstanceAlreadyRunning
		}
		return nil, fmt.Errorf("LockFileEx: %w", err)
	}
	return &InstanceLock{release: func() error {
		_ = windows.UnlockFileEx(h, 0, 0xFFFFFFFF, 0xFFFFFFFF, &ol)
		return f.Close()
	}}, nil
}
