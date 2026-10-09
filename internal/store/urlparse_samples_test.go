package store_test

import (
	"fmt"
	"testing"

	"github.com/suhaasya/wisp-go/internal/store"
)

// Additional samples to reach PRD breadth (40+ total with urlparse_test.go).
func TestParseConnectionURLManySamples(t *testing.T) {
	samples := []string{
		"postgres://u@a/b",
		"postgres://u@a:5433/b",
		"postgresql://u:p@127.0.0.1:5432/d",
		"postgres://u@localhost/d?sslmode=verify-full",
		"postgres://u@h/d?sslmode=verify-ca",
		"postgres://u@h/d?sslrootcert=/tmp/ca.pem",
		"postgres://u@h/d?application_name=myapp",
		"postgres://u@h/d?options=-c%20search_path%3Dpublic",
		"mysql://u@h/d",
		"mysql://u@h:3307/d",
		"mysql://u:p@h/d?tls=true",
		"mariadb://u@h/d?ssl-mode=DISABLED",
		"host=h user=u dbname=d",
		"host=h port=5432 dbname=d user=u",
		"host=h dbname=d user=u password=p sslmode=disable",
		`DATABASE_URL=postgres://u@h/d`,
		"export DATABASE_URL=mysql://u@h/d",
		"postgres://u@db.aws.connect.psdb.cloud/d",
		"postgres://u@instance.region.sql.cloud.google.com/d",
		"postgres://u:enc%20pass@h/db name",
		"postgres://u@h:5432/",
		"postgres://u@/d",
		"postgresql://:@h/d",
		"mysql://:@h/d",
		"postgres://u@[2001:db8::1]:5432/d",
	}
	for i, in := range samples {
		t.Run(fmt.Sprintf("sample_%d", i), func(t *testing.T) {
			if _, err := store.ParseConnectionURL(in); err != nil {
				t.Fatalf("input %q: %v", in, err)
			}
		})
	}
}
