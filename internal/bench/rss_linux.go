//go:build linux

package bench

import (
	"os"
	"strconv"
	"strings"
)

// RSSMegabytes returns current process RSS in megabytes.
func RSSMegabytes() (float64, error) {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0, err
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				break
			}
			kb, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				return 0, err
			}
			return kb / 1024, nil
		}
	}
	return 0, os.ErrNotExist
}
