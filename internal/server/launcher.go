package server

import (
	"context"
	"os"
	"os/exec"
	"runtime"
)

// Platform identifies the OS for browser launching.
type Platform string

const (
	PlatformDarwin  Platform = "darwin"
	PlatformLinux   Platform = "linux"
	PlatformWindows Platform = "windows"
)

// CommandRunner runs a subprocess (injectable in tests).
type CommandRunner func(ctx context.Context, name string, args ...string) error

// Launcher opens URLs in the system or app browser.
type Launcher struct {
	Run      CommandRunner
	Platform Platform
}

// DefaultLauncher uses exec.Command for the current OS.
func DefaultLauncher() Launcher {
	p := Platform(runtime.GOOS)
	return Launcher{
		Platform: p,
		Run: func(ctx context.Context, name string, args ...string) error {
			cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // G204: browser path from OS defaults or user --app
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	}
}

// OpenURL opens url, using an app window when appMode and appBinary are set.
func (l Launcher) OpenURL(ctx context.Context, url string, appMode bool, appBinary string) error {
	if appMode && appBinary != "" {
		return l.Run(ctx, appBinary, "--app="+url)
	}
	switch l.Platform {
	case PlatformDarwin:
		return l.Run(ctx, "open", url)
	case PlatformLinux:
		return l.Run(ctx, "xdg-open", url)
	case PlatformWindows:
		return l.Run(ctx, "rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return l.Run(ctx, "xdg-open", url)
	}
}

// FindAppBrowser returns Chrome/Edge/Chromium path for --app mode, or "".
func FindAppBrowser(platform Platform) string {
	switch platform {
	case PlatformDarwin:
		candidates := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
	case PlatformLinux:
		candidates := []string{
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/usr/bin/microsoft-edge",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
	case PlatformWindows:
		candidates := []string{
			os.ExpandEnv("${ProgramFiles}\\Google\\Chrome\\Application\\chrome.exe"),
			os.ExpandEnv("${ProgramFiles(x86)}\\Google\\Chrome\\Application\\chrome.exe"),
			os.ExpandEnv("${ProgramFiles}\\Microsoft\\Edge\\Application\\msedge.exe"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
	}
	return ""
}
