package session

import (
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/store"
)

// TLSParamsFromProfile maps a saved profile to driver TLS settings.
func TLSParamsFromProfile(p store.Profile) driver.TLSParams {
	host := strings.TrimSpace(p.Transport.Host)
	return driver.TLSParams{
		Mode:       p.SSL.Mode,
		ServerName: host,
		CASource:   p.SSL.CASource,
		CAFile:     p.SSL.CAFile,
		ClientCert: p.SSL.ClientCert,
		ClientKey:  p.SSL.ClientKey,
	}
}
