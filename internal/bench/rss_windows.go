//go:build windows

package bench

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// RSSMegabytes returns current process working set in megabytes.
func RSSMegabytes() (float64, error) {
	h := windows.CurrentProcess()
	var mem windows.PROCESS_MEMORY_COUNTERS
	mem.CB = uint32(unsafe.Sizeof(mem))
	if err := windows.GetProcessMemoryInfo(h, &mem, mem.CB); err != nil {
		return 0, err
	}
	return float64(mem.WorkingSetSize) / (1024 * 1024), nil
}
