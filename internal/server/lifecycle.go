package server

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ShutdownHooks are invoked during graceful shutdown (LW-021 will wire real handlers).
var ShutdownHooks = struct {
	CloseSessions func(context.Context) error
	FlushJournal  func(context.Context) error
}{
	CloseSessions: func(context.Context) error { return nil },
	FlushJournal:  func(context.Context) error { return nil },
}

// SSEActivityHooks notify idle lifecycle of tab connections (LW-009 will wire SSE).
var SSEActivityHooks = struct {
	OnConnected    func()
	OnDisconnected func()
}{}

// GracefulShutdown runs hooks, removes instance metadata, and bounds wait at 2s.
func GracefulShutdown(ctx context.Context, instanceJSON string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if ShutdownHooks.CloseSessions != nil {
		_ = ShutdownHooks.CloseSessions(ctx)
	}
	if ShutdownHooks.FlushJournal != nil {
		_ = ShutdownHooks.FlushJournal(ctx)
	}
	_ = RemoveInstanceJSON(instanceJSON)
	return nil
}

// WaitForSignal blocks until SIGINT or SIGTERM.
func WaitForSignal() os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return <-ch
}

// StopProcess sends SIGTERM to pid and waits up to 2s for exit.
func StopProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return proc.Kill()
}

// IdleLifecycle shuts down when no SSE client has been connected for idleExit.
type IdleLifecycle struct {
	idleExit time.Duration
	onIdle   func()
	done     chan struct{}
	once     sync.Once

	mu    sync.Mutex
	timer *time.Timer
}

// NewIdleLifecycle creates an idle shutdown tracker.
func NewIdleLifecycle(idleExit time.Duration, onIdle func()) *IdleLifecycle {
	lc := &IdleLifecycle{
		idleExit: idleExit,
		onIdle:   onIdle,
		done:     make(chan struct{}),
	}
	SSEActivityHooks.OnConnected = lc.NotifySSEConnected
	SSEActivityHooks.OnDisconnected = lc.NotifySSEDisconnected
	return lc
}

// Done is closed after idle shutdown fires.
func (l *IdleLifecycle) Done() <-chan struct{} {
	return l.done
}

// NotifySSEConnected marks an active tab connection.
func (l *IdleLifecycle) NotifySSEConnected() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.timer != nil {
		l.timer.Stop()
		l.timer = nil
	}
}

// NotifySSEDisconnected starts the idle exit timer.
func (l *IdleLifecycle) NotifySSEDisconnected() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.timer != nil {
		l.timer.Stop()
	}
	l.timer = time.AfterFunc(l.idleExit, l.fireIdle)
}

func (l *IdleLifecycle) fireIdle() {
	l.once.Do(func() {
		if l.onIdle != nil {
			l.onIdle()
		}
		close(l.done)
	})
}

// LogStartupDuration logs listen latency at debug level.
func LogStartupDuration(logLevel string, started time.Time) {
	if logLevel != "debug" {
		return
	}
	elapsed := time.Since(started)
	slog.Debug("startup", "listen_ms", elapsed.Milliseconds())
}
