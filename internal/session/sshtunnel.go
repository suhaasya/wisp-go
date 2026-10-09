package session

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	gossh "golang.org/x/crypto/ssh"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/store"
	"github.com/suhaasya/wisp-go/internal/transport/ssh"
)

var sshPool = ssh.NewPool()

type sshSecrets struct {
	password   []byte
	passphrase []byte
}

func sshEnabled(c *store.SSHConfig) bool {
	if c == nil {
		return false
	}
	return c.Enabled || strings.TrimSpace(c.Host) != ""
}

func attachSSHTunnel(ctx context.Context, p store.Profile, sec sshSecrets) (dial func(context.Context, string, string) (net.Conn, error), release func(), via bool, err error) {
	release = func() {}
	if !sshEnabled(p.SSH) {
		return nil, release, false, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, release, false, err
	}
	configPath := filepath.Join(home, ".ssh", "config")
	blocks, err := ssh.ParseConfigFile(configPath)
	if err != nil {
		return nil, release, false, err
	}
	hops, err := profileHops(p, sec, blocks)
	if err != nil {
		return nil, release, false, err
	}
	knownHosts := filepath.Join(home, ".ssh", "known_hosts")
	key := sshPoolKey(hops, knownHosts)
	client, err := sshPool.Acquire(ctx, key, hops, knownHosts)
	if err != nil {
		return nil, release, false, err
	}
	release = func() { sshPool.Release(key) }
	dial = func(dctx context.Context, network, address string) (net.Conn, error) {
		if dctx.Err() != nil {
			return nil, dctx.Err()
		}
		host, portStr, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
			return nil, err
		}
		return clientDialThrough(dctx, client, host, port)
	}
	return dial, release, true, nil
}

func clientDialThrough(ctx context.Context, client *gossh.Client, host string, port int) (net.Conn, error) {
	type res struct {
		c   net.Conn
		err error
	}
	ch := make(chan res, 1)
	go func() {
		c, err := client.Dial("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
		ch <- res{c, err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		return r.c, r.err
	}
}

func profileHops(p store.Profile, sec sshSecrets, blocks []ssh.ConfigHost) ([]ssh.Hop, error) {
	if p.SSH == nil {
		return nil, errors.New("ssh: config missing")
	}
	var hops []ssh.Hop
	for _, j := range p.SSH.Jumps {
		hops = append(hops, jumpToHop(j, sec))
	}
	port := p.SSH.Port
	if port == 0 {
		port = 22
	}
	hops = append(hops, ssh.Hop{
		Host:       p.SSH.Host,
		Port:       port,
		User:       p.SSH.User,
		ConfigHost: p.SSH.ConfigHost,
		Auth:       sshAuth(p.SSH.Auth, p.SSH.KeyPath, sec),
	})
	return ssh.BuildHopChain(hops, blocks)
}

func jumpToHop(j store.SSHJump, sec sshSecrets) ssh.Hop {
	port := j.Port
	if port == 0 {
		port = 22
	}
	return ssh.Hop{
		Host:       j.Host,
		Port:       port,
		User:       j.User,
		ConfigHost: j.ConfigHost,
		Auth:       sshAuth(j.Auth, j.KeyPath, sec),
	}
}

func sshAuth(method, keyPath string, sec sshSecrets) ssh.AuthInput {
	return ssh.AuthInput{
		Method:     method,
		KeyPath:    keyPath,
		Passphrase: sec.passphrase,
		Password:   sec.password,
	}
}

func sshPoolKey(hops []ssh.Hop, knownHosts string) string {
	var b strings.Builder
	for _, h := range hops {
		_, _ = fmt.Fprintf(&b, "%s@%s:%d|", h.User, h.Host, h.Port)
	}
	b.WriteString(knownHosts)
	return b.String()
}

func applySSHToParams(ctx context.Context, params *driver.ConnectParams, p store.Profile, sec sshSecrets) (release func(), err error) {
	dial, rel, via, err := attachSSHTunnel(ctx, p, sec)
	if err != nil {
		return func() {}, err
	}
	if !via {
		return func() {}, nil
	}
	params.ViaSSH = true
	params.DialContext = dial
	return rel, nil
}
