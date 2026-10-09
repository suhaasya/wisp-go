//go:build darwin

package bench

import "golang.org/x/sys/unix"

// RSSMegabytes returns current process RSS in megabytes (Darwin getrusage).
func RSSMegabytes() (float64, error) {
	var info unix.Rusage
	if err := unix.Getrusage(unix.RUSAGE_SELF, &info); err != nil {
		return 0, err
	}
	return float64(info.Maxrss) / (1024 * 1024), nil
}
