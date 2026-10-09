package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
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
	Paths     RuntimePaths
	Port      int
	Dev       bool
	IdleExit  time.Duration
	LogLevel  string
	NoBrowser bool
	AppWindow bool
	Launcher  Launcher
	WebRoot   string
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
	ln, port, err := listenLoopback(cfg.Port)
	if err != nil {
		return err
	}
	LogStartupDuration(cfg.LogLevel, started)

	mux := http.NewServeMux()
	registerRoutes(mux, cfg, port, secretPath)

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	idle := NewIdleLifecycle(cfg.IdleExit, func() {
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
		launchURL, err := mintLaunchURL(port)
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

func listenLoopback(port int) (net.Listener, int, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	lc := net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", addr)
	if err != nil {
		return nil, 0, err
	}
	tcp, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return nil, 0, fmt.Errorf("unexpected listener addr")
	}
	return ln, tcp.Port, nil
}

func registerRoutes(mux *http.ServeMux, cfg StartConfig, port int, secretPath string) {
	mux.Handle("/internal/handoff", api.NewHandoffHandler(secretPath, func() string {
		url, err := mintLaunchURL(port)
		if err != nil {
			return fmt.Sprintf("http://127.0.0.1:%d/", port)
		}
		return url
	}))
	if cfg.Dev && cfg.WebRoot != "" {
		mux.Handle("/", http.FileServer(http.Dir(cfg.WebRoot)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Wisp"))
		})
	}
}

func mintLaunchURL(port int) (string, error) {
	tok, err := randomBytes(32)
	if err != nil {
		return "", err
	}
	enc := base64.RawURLEncoding.EncodeToString(tok)
	return fmt.Sprintf("http://127.0.0.1:%d/auth/exchange?t=%s", port, enc), nil
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

func DefaultWebRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "web"
	}
	return filepath.Join(wd, "web")
}
