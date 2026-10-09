package server

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/suhaasya/wisp-go/internal/api"
)

// RuntimePaths are on-disk locations for the running instance.
type RuntimePaths struct {
	RuntimeDir    string
	InstanceJSON  string
	InstanceLock  string
	HandoffSecret string
}

// StartConfig configures the primary wisp instance.
type StartConfig struct {
	Paths         RuntimePaths
	Port          int
	Dev           bool
	IdleExit      time.Duration
	LogLevel      string
	Version       string
	NoBrowser     bool
	AppWindow     bool
	Launcher      Launcher
	AssetFS       fs.FS
	RSSSample     func() (float64, error)
	SetupHandlers func(api.EventBroadcaster) (*api.AppHandlers, error)
}

// RunOrHandoff starts the primary server or handoffs to an existing instance.
func RunOrHandoff(ctx context.Context, cfg StartConfig) error {
	if err := ensureRuntimeDir(cfg.Paths.RuntimeDir); err != nil {
		return err
	}
	lock, err := AcquireInstanceLock(cfg.Paths.InstanceLock)
	if err == ErrInstanceAlreadyRunning {
		return RunSecondInstance(ctx, cfg.Paths, cfg)
	}
	if err != nil {
		return err
	}
	defer func() { _ = lock.Release() }()
	return runPrimary(ctx, cfg)
}

// runPrimary starts the loopback server until ctx is canceled or idle exit fires.
func runPrimary(ctx context.Context, cfg StartConfig) error {

	secretPath := cfg.Paths.HandoffSecret
	if _, err := EnsureHandoffSecret(secretPath); err != nil {
		return err
	}

	started := time.Now()
	ln, err := ListenLoopback(cfg.Port)
	if err != nil {
		return err
	}
	tcp, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("unexpected listener addr")
	}
	port := tcp.Port
	LogStartupDuration(cfg.LogLevel, started)

	guard := NewGuard(port)
	guard.SetDev(cfg.Dev)
	version := cfg.Version
	if version == "" {
		version = "dev"
	}
	hub := NewEventHub(EventHubConfig{Dev: cfg.Dev, RSSMegabytes: cfg.RSSSample})
	root := http.NewServeMux()
	if err := registerRoutes(root, cfg, guard, secretPath, hub, version); err != nil {
		return err
	}

	srv := NewHTTPServer(root)
	idle := NewIdleLifecycle(cfg.IdleExit, func() {
		hub.BroadcastShutdown()
		_ = srv.Shutdown(context.Background())
	})

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		_ = GracefulShutdown(shutdownCtx, cfg.Paths.InstanceJSON)
	}()

	rec := InstanceRecord{PID: os.Getpid(), Port: port, Started: time.Now().UTC()}
	if err := WriteInstanceJSON(cfg.Paths.InstanceJSON, rec); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "Wisp running at http://127.0.0.1:%d/ (press Ctrl+C to stop)\n", port)

	if !cfg.NoBrowser {
		launchURL := guard.MintLaunchURL()
		l := cfg.Launcher
		if l.Run == nil {
			l = DefaultLauncher()
		}
		appBin := ""
		if cfg.AppWindow {
			appBin = FindAppBrowser(l.Platform)
		}
		if err := l.OpenURL(ctx, launchURL, cfg.AppWindow, appBin); err != nil {
			slog.Warn("open browser", "err", err)
		}
	}

	err = srv.Serve(ln)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = GracefulShutdown(shutdownCtx, cfg.Paths.InstanceJSON)
	_ = idle
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// RunSecondInstance handoffs to the running primary and opens the browser.
func RunSecondInstance(ctx context.Context, paths RuntimePaths, cfg StartConfig) error {
	if err := ensureRuntimeDir(paths.RuntimeDir); err != nil {
		return err
	}
	_, err := AcquireInstanceLock(paths.InstanceLock)
	if err == nil {
		return fmt.Errorf("second instance: lock acquired unexpectedly")
	}
	if err != ErrInstanceAlreadyRunning {
		return err
	}
	rec, err := ReadInstanceJSON(paths.InstanceJSON)
	if err != nil {
		return err
	}
	secret, err := LoadHandoffSecret(paths.HandoffSecret)
	if err != nil {
		return err
	}
	base := fmt.Sprintf("http://127.0.0.1:%d", rec.Port)
	url, err := RequestHandoffLoginURL(ctx, base, secret)
	if err != nil {
		return err
	}
	l := cfg.Launcher
	if l.Run == nil {
		l = DefaultLauncher()
	}
	appBin := ""
	if cfg.AppWindow {
		appBin = FindAppBrowser(l.Platform)
	}
	if cfg.NoBrowser {
		return nil
	}
	return l.OpenURL(ctx, url, cfg.AppWindow, appBin)
}

// RunStop sends SIGTERM to the instance in instance.json.
func RunStop(paths RuntimePaths) error {
	rec, err := ReadInstanceJSON(paths.InstanceJSON)
	if err != nil {
		return fmt.Errorf("wisp stop: %w", err)
	}
	return StopProcess(rec.PID)
}

func registerRoutes(root *http.ServeMux, cfg StartConfig, guard *Guard, secretPath string, hub *EventHub, version string) error {
	root.Handle("/internal/handoff", api.NewHandoffHandler(secretPath, guard.MintLaunchURL))
	inner := http.NewServeMux()
	if cfg.SetupHandlers == nil {
		return fmt.Errorf("SetupHandlers required")
	}
	handlers, err := cfg.SetupHandlers(hub)
	if err != nil {
		return err
	}
	guard.RegisterRoutes(inner, version, hub, handlers.Settings, handlers.Secrets, handlers.Connections, handlers.ConnectionTest, handlers.FS, handlers.HelloSettings)
	assetFS := cfg.AssetFS
	if assetFS == nil {
		assetFS = emptyFS{}
	}
	st, err := NewStatic(assetFS, cfg.Dev)
	if err != nil {
		return err
	}
	st.Register(inner)
	root.Handle("/", guard.Wrap(inner))
	return nil
}

type emptyFS struct{}

func (emptyFS) Open(name string) (fs.File, error) {
	return nil, fs.ErrNotExist
}

func ensureRuntimeDir(dir string) error {
	return os.MkdirAll(dir, 0o700)
}

// RuntimePathsFromStore copies path strings from store layout helpers.
func RuntimePathsFromStore(runtimeDir, instanceJSON, instanceLock, handoffSecret string) RuntimePaths {
	return RuntimePaths{
		RuntimeDir:    runtimeDir,
		InstanceJSON:  instanceJSON,
		InstanceLock:  instanceLock,
		HandoffSecret: handoffSecret,
	}
}
