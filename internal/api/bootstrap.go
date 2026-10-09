package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/suhaasya/wisp-go/internal/store"
)

// AppHandlers are HTTP handlers wired from the store layer.
type AppHandlers struct {
	Settings      http.Handler
	Secrets       http.Handler
	HelloSettings SettingsSnapshot
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
	snap := func() (map[string]any, error) { return settingsStore.DocumentJSON() }
	return &AppHandlers{
		Settings:      NewSettingsHandler(settingsStore, hub),
		Secrets:       NewSecretsHandler(secretsStore),
		HelloSettings: snap,
	}, nil
}
