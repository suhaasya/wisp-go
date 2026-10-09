package main

import (
	"flag"
	"fmt"
	"time"
)

// RunMode is the top-level wisp command.
type RunMode int

const (
	ModeRun RunMode = iota
	ModeStop
	ModeVersion
	ModeUpdate
)

// RunConfig holds flags for starting the server.
type RunConfig struct {
	Port      int
	NoBrowser bool
	AppWindow bool
	Dev       bool
	IdleExit  time.Duration
	LogLevel  string
}

// ParseArgs interprets argv (including program name) into a mode and run config.
func ParseArgs(args []string) (RunMode, RunConfig, error) {
	if len(args) < 1 {
		return ModeRun, RunConfig{}, fmt.Errorf("missing program name")
	}
	cfg := RunConfig{
		IdleExit: 10 * time.Minute,
		LogLevel: "info",
	}
	if len(args) == 1 {
		return ModeRun, cfg, nil
	}
	switch args[1] {
	case "stop":
		return ModeStop, cfg, nil
	case "version":
		return ModeVersion, cfg, nil
	case "update":
		return ModeUpdate, cfg, nil
	}

	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(flag.CommandLine.Output())
	showVersion := fs.Bool("version", false, "print version and exit")
	port := fs.Int("port", 0, "listen port (0 = random)")
	noBrowser := fs.Bool("no-browser", false, "do not open a browser")
	appWindow := fs.Bool("app", false, "open in Chrome/Edge app window when available")
	dev := fs.Bool("dev", false, "serve web/ from disk")
	idleExit := fs.Duration("idle-exit", 10*time.Minute, "exit after last tab disconnect")
	logLevel := fs.String("log-level", "info", "log level")
	update := fs.Bool("update", false, "download and install update (stub)")
	if err := fs.Parse(args[1:]); err != nil {
		return ModeRun, cfg, err
	}
	if *showVersion {
		return ModeVersion, cfg, nil
	}
	if *update {
		return ModeUpdate, cfg, nil
	}
	cfg.Port = *port
	cfg.NoBrowser = *noBrowser
	cfg.AppWindow = *appWindow
	cfg.Dev = *dev
	cfg.IdleExit = *idleExit
	cfg.LogLevel = *logLevel
	return ModeRun, cfg, nil
}
