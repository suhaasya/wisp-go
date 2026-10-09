package ssh

import (
	"context"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

// Pool shares ssh.Client instances per hop chain.
type Pool struct {
	mu    sync.Mutex
	items map[string]*pooledClient
}

type pooledClient struct {
	client   *gossh.Client
	refs     int
	keepStop chan struct{}
}

// NewPool returns an empty SSH client pool.
func NewPool() *Pool {
	return &Pool{items: make(map[string]*pooledClient)}
}

// Acquire returns a shared client for key; release with Release(key).
func (p *Pool) Acquire(ctx context.Context, key string, hops []Hop, knownHostsPath string) (*gossh.Client, error) {
	p.mu.Lock()
	if item, ok := p.items[key]; ok && item.client != nil {
		item.refs++
		p.mu.Unlock()
		return item.client, nil
	}
	p.mu.Unlock()

	client, err := connectChain(ctx, hops, knownHostsPath)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if item, ok := p.items[key]; ok && item.client != nil {
		_ = client.Close()
		item.refs++
		return item.client, nil
	}
	item := &pooledClient{client: client, refs: 1, keepStop: make(chan struct{})}
	p.items[key] = item
	go keepaliveLoop(item.client, item.keepStop)
	return client, nil
}

// Release decrements refs and closes the client when the last session releases it.
func (p *Pool) Release(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	item, ok := p.items[key]
	if !ok {
		return
	}
	item.refs--
	if item.refs <= 0 {
		close(item.keepStop)
		_ = item.client.Close()
		delete(p.items, key)
	}
}

func keepaliveLoop(client *gossh.Client, stop <-chan struct{}) {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			ok, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
			if err != nil || !ok {
				return
			}
		}
	}
}
