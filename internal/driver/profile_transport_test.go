package driver_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
)

func TestApplyProfileTransport_postgresSocket(t *testing.T) {
	t.Parallel()
	var p driver.ConnectParams
	err := driver.ApplyProfileTransport(&p, "postgres", "socket", "", "/var/run/postgresql/.s.PGSQL.5432", 0)
	if err != nil {
		t.Fatal(err)
	}
	if p.Host != "/var/run/postgresql" || p.Port != 5432 {
		t.Fatalf("host=%q port=%d", p.Host, p.Port)
	}
}
