package ssh_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/transport/ssh"
)

func TestBuildHopChainMaxThree(t *testing.T) {
	hops := []ssh.Hop{
		{Host: "h1", Port: 22, User: "u", Auth: ssh.AuthInput{Method: "password", Password: []byte("x")}},
		{Host: "h2", Port: 22, User: "u", Auth: ssh.AuthInput{Method: "password", Password: []byte("x")}},
		{Host: "h3", Port: 22, User: "u", Auth: ssh.AuthInput{Method: "password", Password: []byte("x")}},
		{Host: "h4", Port: 22, User: "u", Auth: ssh.AuthInput{Method: "password", Password: []byte("x")}},
	}
	_, err := ssh.BuildHopChain(hops, nil)
	if err == nil {
		t.Fatal("expected hop limit error")
	}
}
