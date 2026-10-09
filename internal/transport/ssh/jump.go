package ssh

import (
	"errors"
	"fmt"
	"strings"
)

const maxHops = 3

// Hop is one SSH target in a jump chain.
type Hop struct {
	Host       string
	Port       int
	User       string
	Auth       AuthInput
	ConfigHost string
}

// BuildHopChain validates hop count and expands ProxyJump from config when set.
func BuildHopChain(hops []Hop, configBlocks []ConfigHost) ([]Hop, error) {
	if len(hops) == 0 {
		return nil, errors.New("ssh: at least one hop required")
	}
	out := make([]Hop, 0, len(hops))
	for _, h := range hops {
		expanded, err := expandHop(h, configBlocks)
		if err != nil {
			return nil, err
		}
		out = append(out, expanded...)
	}
	if len(out) > maxHops {
		return nil, fmt.Errorf("ssh: at most %d jump hops allowed", maxHops)
	}
	return out, nil
}

func expandHop(h Hop, blocks []ConfigHost) ([]Hop, error) {
	alias := h.ConfigHost
	if alias == "" {
		alias = h.Host
	}
	if cfg, ok := ResolveHost(blocks, alias); ok {
		if cfg.ProxyJump != "" {
			parts := strings.Split(cfg.ProxyJump, ",")
			var chain []Hop
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				sub, err := expandHop(Hop{Host: p, Port: 22, Auth: h.Auth, ConfigHost: p}, blocks)
				if err != nil {
					return nil, err
				}
				chain = append(chain, sub...)
			}
			h = mergeHop(h, cfg)
			chain = append(chain, h)
			return chain, nil
		}
		h = mergeHop(h, cfg)
	}
	if h.Host == "" {
		return nil, errors.New("ssh: hop host required")
	}
	if h.Port == 0 {
		h.Port = 22
	}
	return []Hop{h}, nil
}

func mergeHop(h Hop, cfg ConfigHost) Hop {
	if cfg.HostName != "" {
		h.Host = cfg.HostName
	}
	if cfg.User != "" && h.User == "" {
		h.User = cfg.User
	}
	if cfg.Port != 0 {
		h.Port = cfg.Port
	}
	if cfg.IdentityFile != "" && h.Auth.KeyPath == "" {
		h.Auth.KeyPath = cfg.IdentityFile
		h.Auth.Method = "key"
	}
	return h
}
