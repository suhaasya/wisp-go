package store_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/store"
)

func TestParseConnectionURLTable(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		engine   string
		host     string
		port     int
		user     string
		db       string
		password string
		sslMode  string
		provider string
		warnSub  string
	}{
		{
			name:     "postgres_basic",
			in:       "postgres://app:secret@db.example.com:5432/shop?sslmode=require",
			engine:   "postgres",
			host:     "db.example.com",
			port:     5432,
			user:     "app",
			db:       "shop",
			password: "secret",
			sslMode:  "require",
		},
		{
			name:   "postgresql_alias",
			in:     "postgresql://app@localhost/mydb",
			engine: "postgres",
			host:   "localhost",
			port:   5432,
			user:   "app",
			db:     "mydb",
		},
		{
			name:     "postgres_ipv6",
			in:       "postgres://app:pass@[::1]:5432/db",
			engine:   "postgres",
			host:     "::1",
			port:     5432,
			user:     "app",
			db:       "db",
			password: "pass",
		},
		{
			name:     "postgres_multi_host",
			in:       "postgres://app:pass@host1:5432,host2:5432/mydb",
			engine:   "postgres",
			host:     "host1",
			port:     5432,
			user:     "app",
			db:       "mydb",
			password: "pass",
			warnSub:  "first host",
		},
		{
			name:     "postgres_percent_decode",
			in:       "postgres://app:p%40ss%2Fx@h.example.com/db",
			engine:   "postgres",
			host:     "h.example.com",
			port:     5432,
			user:     "app",
			db:       "db",
			password: "p@ss/x",
		},
		{
			name:    "postgres_connect_timeout",
			in:      "postgres://app@h/db?connect_timeout=15&sslmode=disable",
			engine:  "postgres",
			host:    "h",
			port:    5432,
			user:    "app",
			db:      "db",
			sslMode: "disable",
		},
		{
			name:     "libpq_string",
			in:       "host=127.0.0.1 port=5433 dbname=shop user=app password=s3cr3t sslmode=verify-full",
			engine:   "postgres",
			host:     "127.0.0.1",
			port:     5433,
			user:     "app",
			db:       "shop",
			password: "s3cr3t",
			sslMode:  "verify-full",
		},
		{
			name:     "libpq_quoted",
			in:       `host='db.example.com' user='app' dbname='my db' password='x y'`,
			engine:   "postgres",
			host:     "db.example.com",
			port:     5432,
			user:     "app",
			db:       "my db",
			password: "x y",
		},
		{
			name:     "mysql_url",
			in:       "mysql://root:pass@10.0.0.5:3306/billing?ssl-mode=REQUIRED",
			engine:   "mysql",
			host:     "10.0.0.5",
			port:     3306,
			user:     "root",
			db:       "billing",
			password: "pass",
			sslMode:  "require",
		},
		{
			name:   "mariadb_url",
			in:     "mariadb://dev@maria.internal/app",
			engine: "mariadb",
			host:   "maria.internal",
			port:   3306,
			user:   "dev",
			db:     "app",
		},
		{
			name:     "env_line_double_quote",
			in:       `DATABASE_URL="postgres://app:pw@db.qwerty.supabase.co:5432/postgres"`,
			engine:   "postgres",
			host:     "db.qwerty.supabase.co",
			port:     5432,
			user:     "app",
			db:       "postgres",
			password: "pw",
			provider: "supabase",
		},
		{
			name:     "env_export_single_quote",
			in:       `export DATABASE_URL='postgres://u:p@ep-cool.neon.tech/neondb'`,
			engine:   "postgres",
			host:     "ep-cool.neon.tech",
			port:     5432,
			user:     "u",
			db:       "neondb",
			password: "p",
			provider: "neon",
		},
		{
			name:     "provider_rds",
			in:       "postgres://app:pw@mydb.abc123.eu-west-1.rds.amazonaws.com:5432/prod",
			engine:   "postgres",
			host:     "mydb.abc123.eu-west-1.rds.amazonaws.com",
			port:     5432,
			user:     "app",
			db:       "prod",
			password: "pw",
			provider: "rds",
			sslMode:  "require",
		},
		{
			name:     "provider_planetscale",
			in:       "mysql://x:y@aws.connect.psdb.cloud/main?ssl-mode=required",
			engine:   "mysql",
			host:     "aws.connect.psdb.cloud",
			port:     3306,
			user:     "x",
			db:       "main",
			password: "y",
			provider: "planetscale",
		},
		{
			name:     "provider_azure",
			in:       "postgres://app@myserver.postgres.database.azure.com:5432/app",
			engine:   "postgres",
			host:     "myserver.postgres.database.azure.com",
			port:     5432,
			user:     "app",
			db:       "app",
			provider: "azure",
			sslMode:  "require",
		},
		{
			name:    "mockup_example",
			in:      "postgres://app@db.example.com:5432/shop?sslmode=require",
			engine:  "postgres",
			host:    "db.example.com",
			port:    5432,
			user:    "app",
			db:      "shop",
			sslMode: "require",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := store.ParseConnectionURL(tc.in)
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if got.Profile.Engine != tc.engine {
				t.Fatalf("engine = %q", got.Profile.Engine)
			}
			if got.Profile.Transport.Host != tc.host {
				t.Fatalf("host = %q want %q", got.Profile.Transport.Host, tc.host)
			}
			if got.Profile.Transport.Port != tc.port {
				t.Fatalf("port = %d want %d", got.Profile.Transport.Port, tc.port)
			}
			if got.Profile.User != tc.user {
				t.Fatalf("user = %q", got.Profile.User)
			}
			if got.Profile.Database != tc.db {
				t.Fatalf("db = %q", got.Profile.Database)
			}
			if got.Password != tc.password {
				t.Fatalf("password = %q want %q", got.Password, tc.password)
			}
			if tc.sslMode != "" && got.Profile.SSL.Mode != tc.sslMode {
				t.Fatalf("ssl = %q want %q", got.Profile.SSL.Mode, tc.sslMode)
			}
			if tc.provider != "" && got.Profile.Options.Provider != tc.provider {
				t.Fatalf("provider = %q want %q", got.Profile.Options.Provider, tc.provider)
			}
			if tc.warnSub != "" {
				found := false
				for _, w := range got.Warnings {
					if containsSub(w, tc.warnSub) {
						found = true
					}
				}
				if !found {
					t.Fatalf("warnings = %v want substring %q", got.Warnings, tc.warnSub)
				}
			}
		})
	}
}

func TestParseConnectionURLExtraParams(t *testing.T) {
	got, err := store.ParseConnectionURL("postgres://app@h/db?sslmode=disable&application_name=wisp&foo=bar")
	if err != nil {
		t.Fatal(err)
	}
	if got.Profile.Options.Extra["application_name"] != "wisp" {
		t.Fatalf("application_name = %v", got.Profile.Options.Extra)
	}
	if got.Profile.Options.Extra["foo"] != "bar" {
		t.Fatalf("foo = %v", got.Profile.Options.Extra)
	}
}

func TestParseConnectionURLInvalid(t *testing.T) {
	_, err := store.ParseConnectionURL("hello world")
	if err == nil {
		t.Fatal("expected error")
	}
}

func containsSub(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && stringIndex(s, sub) >= 0)
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
