package ssh

import (
	"context"
	"fmt"
	"net"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

// DialDatabase opens a TCP connection to the database through the SSH hop chain.
func DialDatabase(ctx context.Context, hops []Hop, knownHostsPath, dbHost string, dbPort int) (net.Conn, *gossh.Client, error) {
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}
	chain, err := BuildHopChain(hops, nil)
	if err != nil {
		return nil, nil, err
	}
	client, err := connectChain(ctx, chain, knownHostsPath)
	if err != nil {
		return nil, nil, err
	}
	addr := net.JoinHostPort(dbHost, fmt.Sprintf("%d", dbPort))
	dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var conn net.Conn
	done := make(chan error, 1)
	go func() {
		c, err := client.Dial("tcp", addr)
		if err != nil {
			done <- err
			return
		}
		conn = c
		done <- nil
	}()
	select {
	case <-dialCtx.Done():
		return nil, client, dialCtx.Err()
	case err := <-done:
		if err != nil {
			return nil, client, err
		}
		return conn, client, nil
	}
}

func connectChain(ctx context.Context, hops []Hop, knownHostsPath string) (*gossh.Client, error) {
	var client *gossh.Client
	for i, hop := range hops {
		if ctx.Err() != nil {
			if client != nil {
				_ = client.Close()
			}
			return nil, ctx.Err()
		}
		next, err := dialHop(ctx, client, hop, knownHostsPath)
		if err != nil {
			if client != nil {
				_ = client.Close()
			}
			return nil, err
		}
		if client != nil {
			_ = client.Close()
		}
		client = next
		if i == len(hops)-1 {
			return client, nil
		}
	}
	return client, nil
}

func dialHop(ctx context.Context, via *gossh.Client, hop Hop, knownHostsPath string) (*gossh.Client, error) {
	hostKeyCB, err := HostKeyCallback(knownHostsPath)
	if err != nil {
		return nil, err
	}
	auths, err := AuthMethods(ctx, hop.Auth)
	if err != nil {
		return nil, err
	}
	cfg := &gossh.ClientConfig{
		User:            hop.User,
		Auth:            auths,
		HostKeyCallback: hostKeyCB,
		Timeout:         15 * time.Second,
	}
	addr := net.JoinHostPort(hop.Host, fmt.Sprintf("%d", hop.Port))
	var conn net.Conn
	if via == nil {
		d := net.Dialer{Timeout: 15 * time.Second}
		conn, err = d.DialContext(ctx, "tcp", addr)
	} else {
		conn, err = via.Dial("tcp", addr)
	}
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := gossh.NewClientConn(conn, addr, cfg)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return gossh.NewClient(c, chans, reqs), nil
}
