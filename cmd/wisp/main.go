// Command wisp is the Wisp database client launcher and CLI.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/suhaasya/wisp-go/internal/server"
	"github.com/suhaasya/wisp-go/internal/store"
)

var version = "dev"

func main() {
	mode, cfg, err := ParseArgs(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wisp: %v\n", err)
		os.Exit(2)
	}

	switch mode {
	case ModeVersion:
		fmt.Println(version)
		return
	case ModeUpdate:
		fmt.Fprintln(os.Stderr, "wisp: --update is not implemented yet (LW-047)")
		os.Exit(1)
	case ModeStop:
		paths, err := store.DefaultPaths()
		if err != nil {
			fmt.Fprintf(os.Stderr, "wisp stop: %v\n", err)
			os.Exit(1)
		}
		if err := server.RunStop(runtimePaths(paths)); err != nil {
			fmt.Fprintf(os.Stderr, "wisp stop: %v\n", err)
			os.Exit(1)
		}
		return
	case ModeRun:
		runServer(cfg)
	default:
		os.Exit(2)
	}
}

func runServer(cfg RunConfig) {
	paths, err := store.DefaultPaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wisp: %v\n", err)
		os.Exit(1)
	}

	setLogLevel(cfg.LogLevel)

	startCfg := server.StartConfig{
		Paths:     runtimePaths(paths),
		Port:      cfg.Port,
		Dev:       cfg.Dev,
		IdleExit:  cfg.IdleExit,
		LogLevel:  cfg.LogLevel,
		NoBrowser: cfg.NoBrowser,
		AppWindow: cfg.AppWindow,
		WebRoot:   server.DefaultWebRoot(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := server.RunOrHandoff(ctx, startCfg); err != nil {
		fmt.Fprintf(os.Stderr, "wisp: %v\n", err)
		os.Exit(1)
	}
}

func runtimePaths(p store.Paths) server.RuntimePaths {
	return server.RuntimePathsFromStore(p.RuntimeDir, p.InstanceJSON(), p.InstanceLock(), p.HandoffSecret())
}

func setLogLevel(level string) {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})))
}
