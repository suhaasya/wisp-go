//go:build integration

package ssh_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	gossh "golang.org/x/crypto/ssh"

	"github.com/suhaasya/wisp-go/internal/bench"
	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mysql"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/transport/ssh"
)

func ciSSHDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../../ci/ssh"))
}

func clientKeyPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(ciSSHDir(t), "client", "id_ed25519")
}

func ssh1Endpoint(t *testing.T) (host string, port int) {
	t.Helper()
	host = envOr("WISP_SSH1_HOST", "127.0.0.1")
	port = envIntOr("WISP_SSH1_PORT", 2221)
	return host, port
}

func hostPubKey(t *testing.T, rel string) gossh.PublicKey {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(ciSSHDir(t), rel))
	if err != nil {
		t.Fatal(err)
	}
	pub, _, _, _, err := gossh.ParseAuthorizedKey(b)
	if err != nil {
		t.Fatal(err)
	}
	return pub
}

func knownHostsFile(t *testing.T) string {
	t.Helper()
	host, port := ssh1Endpoint(t)
	path := filepath.Join(t.TempDir(), "known_hosts")
	k1 := hostPubKey(t, "hostkeys/ssh1/ssh_host_ed25519_key.pub")
	k2 := hostPubKey(t, "hostkeys/ssh2/ssh_host_ed25519_key.pub")
	if err := ssh.AppendKnownHost(path, host, port, k1); err != nil {
		t.Fatal(err)
	}
	if err := ssh.AppendKnownHost(path, "ssh2", 22, k2); err != nil {
		t.Fatal(err)
	}
	return path
}

func hopAuth(t *testing.T) ssh.AuthInput {
	t.Helper()
	return ssh.AuthInput{Method: "key", KeyPath: clientKeyPath(t)}
}

func oneHop(t *testing.T) []ssh.Hop {
	t.Helper()
	host, port := ssh1Endpoint(t)
	return []ssh.Hop{{Host: host, Port: port, User: "wisp", Auth: hopAuth(t)}}
}

func twoHop(t *testing.T) []ssh.Hop {
	t.Helper()
	host, port := ssh1Endpoint(t)
	return []ssh.Hop{
		{Host: host, Port: port, User: "wisp", Auth: hopAuth(t)},
		{Host: "ssh2", Port: 22, User: "wisp", Auth: hopAuth(t)},
	}
}

func sshDialer(t *testing.T, hops []ssh.Hop) func(context.Context, string, string) (net.Conn, error) {
	t.Helper()
	known := knownHostsFile(t)
	pool := ssh.NewPool()
	key := t.Name()
	client, err := pool.Acquire(t.Context(), key, hops, known)
	if err != nil {
		t.Skipf("ssh bastion not available (docker compose -f ci/docker-compose.yml up -d ssh1 ssh2 pg16 mysql84): %v", err)
	}
	t.Cleanup(func() { pool.Release(key) })
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		type res struct {
			c   net.Conn
			err error
		}
		ch := make(chan res, 1)
		go func() {
			c, err := client.Dial("tcp", address)
			ch <- res{c, err}
		}()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-ch:
			return r.c, r.err
		}
	}
}

func pgParamsViaSSH(t *testing.T, dial func(context.Context, string, string) (net.Conn, error)) driver.ConnectParams {
	t.Helper()
	host := envOr("WISP_PG_VIA_SSH_HOST", "pg16")
	port := envIntOr("WISP_PG_VIA_SSH_PORT", 5432)
	return driver.ConnectParams{
		Host:           host,
		Port:           port,
		DB:             envOr("WISP_PG_DB", "wisp_bench"),
		User:           envOr("WISP_PG_USER", "wisp"),
		Password:       []byte(envOr("WISP_PG_PASSWORD", "wisp")),
		ConnectTimeout: 10 * time.Second,
		DialContext:    dial,
		ViaSSH:         true,
	}
}

func mysqlParamsViaSSH(t *testing.T, dial func(context.Context, string, string) (net.Conn, error)) driver.ConnectParams {
	t.Helper()
	host := envOr("WISP_MYSQL_VIA_SSH_HOST", "mysql84")
	port := envIntOr("WISP_MYSQL_VIA_SSH_PORT", 3306)
	return driver.ConnectParams{
		Host:           host,
		Port:           port,
		DB:             envOr("WISP_MYSQL_DB", "wisp_bench"),
		User:           envOr("WISP_MYSQL_USER", "wisp"),
		Password:       []byte(envOr("WISP_MYSQL_PASSWORD", "wisp")),
		ConnectTimeout: 10 * time.Second,
		DialContext:    dial,
		ViaSSH:         true,
	}
}

func TestIntegrationPostgresOneHop(t *testing.T) {
	dial := sshDialer(t, oneHop(t))
	conn, err := postgres.Driver{}.Open(t.Context(), pgParamsViaSSH(t, dial))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationMySQLOneHop(t *testing.T) {
	dial := sshDialer(t, oneHop(t))
	conn, err := mysql.Driver{}.Open(t.Context(), mysqlParamsViaSSH(t, dial))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationPostgresTwoHop(t *testing.T) {
	dial := sshDialer(t, twoHop(t))
	conn, err := postgres.Driver{}.Open(t.Context(), pgParamsViaSSH(t, dial))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationHostKeyMismatch(t *testing.T) {
	host, port := ssh1Endpoint(t)
	path := filepath.Join(t.TempDir(), "known_hosts")
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := gossh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := ssh.AppendKnownHost(path, host, port, signer.PublicKey()); err != nil {
		t.Fatal(err)
	}
	_, err = ssh.NewPool().Acquire(t.Context(), "mismatch", oneHop(t), path)
	if err == nil {
		t.Fatal("expected host key mismatch")
	}
	var mismatch *ssh.HostKeyMismatch
	if errors.As(err, &mismatch) {
		return
	}
	if isConnRefused(err) {
		t.Skipf("ssh bastion not available: %v", err)
	}
	t.Fatalf("want HostKeyMismatch, got %v", err)
}

func isConnRefused(err error) bool {
	var op *net.OpError
	if errors.As(err, &op) {
		return strings.Contains(strings.ToLower(op.Err.Error()), "connection refused")
	}
	return strings.Contains(strings.ToLower(err.Error()), "connection refused")
}

func TestIntegrationTunnelRSSBudget(t *testing.T) {
	before, err := bench.RSSMegabytes()
	if err != nil {
		t.Skipf("rss: %v", err)
	}
	dial := sshDialer(t, oneHop(t))
	conn, err := postgres.Driver{}.Open(t.Context(), pgParamsViaSSH(t, dial))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
	after, err := bench.RSSMegabytes()
	if err != nil {
		t.Skipf("rss after: %v", err)
	}
	delta := after - before
	if delta > 1.0 {
		t.Fatalf("tunnel RSS delta %.2f MB exceeds 1 MB budget", delta)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envIntOr(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			return n
		}
	}
	return def
}
