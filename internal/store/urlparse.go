package store

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ParseURLResult is a connection draft from pasted text.
type ParseURLResult struct {
	Profile  Profile
	Password string
	Warnings []string
}

// ParseConnectionURL parses postgres/postgresql/mysql/mariadb URLs, libpq strings, or .env lines.
func ParseConnectionURL(raw string) (ParseURLResult, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ParseURLResult{}, errors.New("empty input")
	}
	s = unwrapEnvLine(s)
	if isLibpqConninfo(s) {
		return parseLibpq(s)
	}
	u, scheme, err := parseURLScheme(s)
	if err != nil {
		return ParseURLResult{}, err
	}
	switch scheme {
	case "postgres", "postgresql":
		return parsePostgresURL(u)
	case "mysql", "mariadb":
		return parseMySQLURL(u, scheme)
	default:
		return ParseURLResult{}, errors.New("unsupported connection URL scheme")
	}
}

func unwrapEnvLine(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "export ") {
		s = strings.TrimSpace(s[len("export "):])
	}
	const key = "DATABASE_URL="
	if idx := strings.Index(s, key); idx >= 0 {
		s = strings.TrimSpace(s[idx+len(key):])
	}
	s = strings.Trim(s, `"'`)
	return s
}

func isLibpqConninfo(s string) bool {
	if strings.Contains(s, "://") {
		return false
	}
	return strings.Contains(s, "=") && (strings.Contains(s, "host=") || strings.Contains(s, "dbname="))
}

func parseURLScheme(s string) (*url.URL, string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, "", err
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme == "" {
		return nil, "", errors.New("missing URL scheme")
	}
	return u, scheme, nil
}

func parsePostgresURL(u *url.URL) (ParseURLResult, error) {
	var warnings []string
	host, port, err := splitHostPort(u.Host, 5432, &warnings)
	if err != nil {
		return ParseURLResult{}, err
	}
	user := ""
	pass := ""
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}
	db := strings.TrimPrefix(u.Path, "/")
	q := u.Query()
	sslMode := q.Get("sslmode")
	if sslMode == "" {
		sslMode = "prefer"
	}
	extra := map[string]string{}
	for k, vals := range q {
		if len(vals) == 0 {
			continue
		}
		switch k {
		case "sslmode":
			sslMode = vals[0]
		case "sslrootcert", "sslcert", "sslkey":
			extra[k] = vals[0]
		default:
			extra[k] = vals[0]
		}
	}
	prof := baseProfile("postgres", host, port, user, db, sslMode, extra)
	applyProvider(&prof, host)
	return ParseURLResult{Profile: prof, Password: pass, Warnings: warnings}, nil
}

func parseMySQLURL(u *url.URL, scheme string) (ParseURLResult, error) {
	var warnings []string
	engine := "mysql"
	if scheme == "mariadb" {
		engine = "mariadb"
	}
	host, port, err := splitHostPort(u.Host, 3306, &warnings)
	if err != nil {
		return ParseURLResult{}, err
	}
	user := ""
	pass := ""
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}
	db := strings.TrimPrefix(u.Path, "/")
	q := u.Query()
	sslMode := "prefer"
	extra := map[string]string{}
	for k, vals := range q {
		if len(vals) == 0 {
			continue
		}
		lk := strings.ToLower(k)
		switch lk {
		case "ssl-mode", "sslmode", "tls":
			sslMode = normalizeMySQLSSL(vals[0])
		default:
			extra[k] = vals[0]
		}
	}
	prof := baseProfile(engine, host, port, user, db, sslMode, extra)
	applyProvider(&prof, host)
	return ParseURLResult{Profile: prof, Password: pass, Warnings: warnings}, nil
}

func parseLibpq(s string) (ParseURLResult, error) {
	kv, err := parseKeyValueWords(s)
	if err != nil {
		return ParseURLResult{}, err
	}
	host := kv["host"]
	if host == "" {
		host = "localhost"
	}
	port := 5432
	if p := kv["port"]; p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			return ParseURLResult{}, fmt.Errorf("invalid port")
		}
		port = n
	}
	db := kv["dbname"]
	if db == "" {
		db = kv["database"]
	}
	user := kv["user"]
	pass := kv["password"]
	sslMode := kv["sslmode"]
	if sslMode == "" {
		sslMode = "prefer"
	}
	extra := map[string]string{}
	for k, v := range kv {
		switch k {
		case "host", "port", "dbname", "database", "user", "password", "sslmode":
			continue
		default:
			extra[k] = v
		}
	}
	prof := baseProfile("postgres", host, port, user, db, sslMode, extra)
	applyProvider(&prof, host)
	if ct := kv["connect_timeout"]; ct != "" {
		if prof.Options.Extra == nil {
			prof.Options.Extra = map[string]string{}
		}
		prof.Options.Extra["connect_timeout"] = ct
	}
	return ParseURLResult{Profile: prof, Password: pass}, nil
}

func baseProfile(engine, host string, port int, user, db, sslMode string, extra map[string]string) Profile {
	if extra == nil {
		extra = map[string]string{}
	}
	name := db
	if name == "" {
		name = host
	}
	return Profile{
		Name:         name,
		Engine:       engine,
		Transport:    Transport{Kind: "tcp", Host: host, Port: port},
		User:         user,
		Database:     db,
		PasswordMode: "keychain",
		SSL:          SSLConfig{Mode: sslMode},
		Env:          guessEnv(host),
		Options:      ConnOptions{Extra: extra},
	}
}

func guessEnv(host string) string {
	h := strings.ToLower(host)
	if h == "localhost" || h == "127.0.0.1" || h == "::1" {
		return "local"
	}
	return "development"
}

func applyProvider(p *Profile, host string) {
	h := strings.ToLower(host)
	type rule struct {
		suffix   string
		provider string
		ssl      string
	}
	rules := []rule{
		{".supabase.co", "supabase", "require"},
		{".neon.tech", "neon", "require"},
		{".rds.amazonaws.com", "rds", "require"},
		{".psdb.cloud", "planetscale", "require"},
		{".sql.cloud.google.com", "cloudsql", "require"},
		{".postgres.database.azure.com", "azure", "require"},
	}
	for _, r := range rules {
		if strings.HasSuffix(h, r.suffix) {
			p.Options.Provider = r.provider
			if p.SSL.Mode == "" || p.SSL.Mode == "prefer" || p.SSL.Mode == "disable" {
				p.SSL.Mode = r.ssl
			}
			return
		}
	}
}

func splitHostPort(hostport string, defaultPort int, warnings *[]string) (string, int, error) {
	if hostport == "" {
		return "localhost", defaultPort, nil
	}
	if strings.Contains(hostport, ",") {
		parts := strings.Split(hostport, ",")
		hostport = parts[0]
		*warnings = append(*warnings, "Multiple hosts listed; using the first host only.")
	}
	if strings.HasPrefix(hostport, "[") {
		h, p, err := net.SplitHostPort(hostport)
		if err != nil {
			return "", 0, err
		}
		port, err := strconv.Atoi(p)
		if err != nil {
			return "", 0, err
		}
		return strings.Trim(h, "[]"), port, nil
	}
	if strings.Count(hostport, ":") > 1 && !strings.Contains(hostport, "]") {
		// bare IPv6 without brackets
		return hostport, defaultPort, nil
	}
	host, portStr, err := net.SplitHostPort(hostport)
	if err != nil {
		if strings.Contains(err.Error(), "missing port") {
			return hostport, defaultPort, nil
		}
		return hostport, defaultPort, nil
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func normalizeMySQLSSL(v string) string {
	switch strings.ToLower(v) {
	case "required", "require":
		return "require"
	case "disabled", "disable":
		return "disable"
	case "verify_ca", "verify-ca":
		return "verify-ca"
	case "verify_identity", "verify-full":
		return "verify-full"
	default:
		return strings.ToLower(v)
	}
}

func parseKeyValueWords(s string) (map[string]string, error) {
	out := map[string]string{}
	for _, w := range splitConninfoWords(s) {
		eq := strings.IndexByte(w, '=')
		if eq <= 0 {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(w[:eq]))
		v := strings.TrimSpace(w[eq+1:])
		v = strings.Trim(v, `"'`)
		out[k] = v
	}
	return out, nil
}

func splitConninfoWords(s string) []string {
	var words []string
	var b strings.Builder
	inQuote := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
				continue
			}
			b.WriteByte(c)
			continue
		}
		if c == '\'' || c == '"' {
			inQuote = c
			continue
		}
		if c == ' ' || c == '\t' {
			if b.Len() > 0 {
				words = append(words, b.String())
				b.Reset()
			}
			continue
		}
		b.WriteByte(c)
	}
	if b.Len() > 0 {
		words = append(words, b.String())
	}
	return words
}
