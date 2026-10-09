package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

// AppHandlers are HTTP handlers wired from the store layer.
type AppHandlers struct {
	Settings        http.Handler
	Secrets         http.Handler
	Connections     http.Handler
	ConnectionTest  http.Handler
	ConnectionParse http.Handler
	HostKeyAccept   http.Handler
	FS              http.Handler
	Sessions        http.Handler
	Workspace       http.Handler
	HelloSettings   SettingsSnapshot
	CloseSessions   func(context.Context) error
}

// NewAppHandlers loads settings and secrets stores and returns API handlers.
func NewAppHandlers(settingsPath string, paths store.Paths, probe store.KeyringProbe, hub EventBroadcaster) (*AppHandlers, error) {
	settingsStore := store.NewSettingsStore(settingsPath)
	if _, err := settingsStore.Load(); err != nil {
		slog.Warn("settings load", "err", err)
	}
	if probe == nil {
		probe = store.DefaultKeyringProbe
	}
	secretsStore, err := store.OpenSecretsStore(paths, probe)
	if err != nil {
		return nil, fmt.Errorf("secrets: %w", err)
	}
	profileStore := store.NewProfileStore(paths.ConnectionsJSON())
	if _, err := profileStore.Load(); err != nil {
		slog.Warn("connections load", "err", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	snap := func() (map[string]any, error) { return settingsStore.DocumentJSON() }
	sessionMgr := session.NewManager(profileStore, secretsStore, hub)
	workspaceStore := session.NewWorkspaceStore(paths, profileStore)
	return &AppHandlers{
		Settings:        NewSettingsHandler(settingsStore, hub),
		Secrets:         NewSecretsHandler(secretsStore),
		Connections:     NewConnectionsHandler(profileStore, secretsStore),
		ConnectionTest:  NewConnectionTestHandler(secretsStore),
		ConnectionParse: NewConnectionParseHandler(),
		HostKeyAccept:   NewHostKeyAcceptHandler(),
		FS:              NewFSHandler(home),
		Sessions:        NewSessionsHandler(sessionMgr),
		Workspace:       NewWorkspaceHandler(workspaceStore),
		HelloSettings:   snap,
		CloseSessions:   sessionMgr.CloseAll,
	}, nil
}
