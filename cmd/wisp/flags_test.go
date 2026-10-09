package main

import (
	"testing"
	"time"
)

func TestParseArgs_runDefaults(t *testing.T) {
	t.Parallel()
	mode, cfg, err := ParseArgs([]string{"wisp"})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if mode != ModeRun {
		t.Fatalf("mode = %v, want ModeRun", mode)
	}
	if cfg.Port != 0 {
		t.Fatalf("Port = %d, want 0", cfg.Port)
	}
	if cfg.IdleExit != 10*time.Minute {
		t.Fatalf("IdleExit = %v, want 10m", cfg.IdleExit)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want info", cfg.LogLevel)
	}
}

func TestParseArgs_stopSubcommand(t *testing.T) {
	t.Parallel()
	mode, _, err := ParseArgs([]string{"wisp", "stop"})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if mode != ModeStop {
		t.Fatalf("mode = %v, want ModeStop", mode)
	}
}

func TestParseArgs_versionFlag(t *testing.T) {
	t.Parallel()
	mode, _, err := ParseArgs([]string{"wisp", "--version"})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if mode != ModeVersion {
		t.Fatalf("mode = %v, want ModeVersion", mode)
	}
}

func TestParseArgs_runFlags(t *testing.T) {
	t.Parallel()
	mode, cfg, err := ParseArgs([]string{
		"wisp", "--port", "4242", "--no-browser", "--app", "--dev",
		"--idle-exit", "5m", "--log-level", "debug",
	})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if mode != ModeRun {
		t.Fatalf("mode = %v, want ModeRun", mode)
	}
	if cfg.Port != 4242 || !cfg.NoBrowser || !cfg.AppWindow || !cfg.Dev {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	if cfg.IdleExit != 5*time.Minute || cfg.LogLevel != "debug" {
		t.Fatalf("unexpected timing/log: %+v", cfg)
	}
}
