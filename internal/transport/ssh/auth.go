package ssh

import (
	"context"
	"errors"
	"fmt"
	"os"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// AuthInput holds credentials for one hop (no logging of secrets).
type AuthInput struct {
	Method     string // agent, key, password, config
	KeyPath    string
	Passphrase []byte
	Password   []byte
}

// AuthMethods returns signers/password in PRD order: agent → key → password.
func AuthMethods(ctx context.Context, in AuthInput) ([]gossh.AuthMethod, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	method := in.Method
	if method == "" || method == "agent" {
		if m, ok := agentAuth(); ok {
			return []gossh.AuthMethod{m}, nil
		}
		if method == "agent" {
			return nil, errors.New("ssh: ssh-agent not available")
		}
	}
	if method == "key" || (method == "" && in.KeyPath != "") {
		m, err := keyAuth(in.KeyPath, in.Passphrase)
		if err != nil {
			return nil, err
		}
		return []gossh.AuthMethod{m}, nil
	}
	if method == "password" || len(in.Password) > 0 {
		return []gossh.AuthMethod{gossh.Password(string(in.Password))}, nil
	}
	return nil, errors.New("ssh: no authentication method configured")
}

func agentAuth() (gossh.AuthMethod, bool) {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return nil, false
	}
	ag, err := openSSHAgent(sock)
	if err != nil {
		return nil, false
	}
	return gossh.PublicKeysCallback(func() ([]gossh.Signer, error) {
		return ag.Signers()
	}), true
}

func openSSHAgent(sock string) (agent.Agent, error) {
	conn, err := platformAgentConn(sock)
	if err != nil {
		return nil, err
	}
	return agent.NewClient(conn), nil
}

func keyAuth(path string, passphrase []byte) (gossh.AuthMethod, error) {
	if path == "" {
		return nil, errors.New("ssh: key path required")
	}
	path = expandHome(path)
	pem, err := os.ReadFile(path) //nolint:gosec // G304: user-chosen key path from profile
	if err != nil {
		return nil, fmt.Errorf("ssh: read key: %w", err)
	}
	var signer gossh.Signer
	if len(passphrase) > 0 {
		signer, err = gossh.ParsePrivateKeyWithPassphrase(pem, passphrase)
	} else {
		signer, err = gossh.ParsePrivateKey(pem)
	}
	if err != nil {
		return nil, fmt.Errorf("ssh: parse key: %w", err)
	}
	return gossh.PublicKeys(signer), nil
}

func expandHome(p string) string {
	if len(p) >= 2 && p[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return home + p[1:]
	}
	return p
}
