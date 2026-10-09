package main

import (
	"context"
	"fmt"
	"os"

	"github.com/suhaasya/wisp-go/internal/update"
)

func runUpdate() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("update: executable path: %w", err)
	}
	_, err = update.Run(context.Background(), update.Options{
		CurrentVersion: version,
		ExecPath:       exe,
		PublicKeyPEM:   update.DefaultPublicKey,
		Out:            os.Stderr,
	})
	return err
}
