package driver

import (
	"crypto/tls"

	"github.com/suhaasya/wisp-go/internal/transport"
)

// ClientTLSConfig builds a client tls.Config from connection TLS params.
func ClientTLSConfig(p TLSParams) (*tls.Config, error) {
	return transport.BuildTLSConfig(p)
}

// FormatTLSLine formats negotiated TLS for ServerInfo and the test line.
func FormatTLSLine(st *tls.ConnectionState) string {
	if st == nil {
		return ""
	}
	return transport.FormatNegotiated(transport.DescribeNegotiated(st))
}
