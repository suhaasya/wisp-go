package session

import (
	"context"

	"github.com/suhaasya/wisp-go/internal/transport"
)

// DetectLocal returns databases detected on this machine (sockets and localhost ports).
func DetectLocal(ctx context.Context) []transport.LocalSuggestion {
	return transport.DetectLocal(ctx)
}
