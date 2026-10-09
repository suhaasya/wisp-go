package session

import (
	"context"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mysql"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/store"
)

// SSHConnectSecrets holds optional SSH credentials for a test connect.
type SSHConnectSecrets struct {
	Password   []byte
	Passphrase []byte
}

// TestConnect opens a connection from a profile draft and pings the server.
func TestConnect(ctx context.Context, p store.Profile, password []byte, sshSec SSHConnectSecrets) (driver.Conn, time.Duration, driver.ServerInfo, error) {
	params := driver.ConnectParams{
		DB:                      p.Database,
		User:                    p.User,
		Password:                password,
		ConnectTimeout:          10 * time.Second,
		AllowPublicKeyRetrieval: p.Options.AllowPublicKeyRetrieval,
		TLS:                     TLSParamsFromProfile(p),
	}
	if err := driver.ApplyProfileTransport(&params, p.Engine, p.Transport.Kind, p.Transport.Host, p.Transport.Path, p.Transport.Port); err != nil {
		return nil, 0, driver.ServerInfo{}, err
	}
	if p.Advanced.StatementTimeoutSec > 0 {
		params.StatementTimeout = time.Duration(p.Advanced.StatementTimeoutSec) * time.Second
	}
	sshRelease, err := applySSHToParams(ctx, &params, p, sshSecrets{password: sshSec.Password, passphrase: sshSec.Passphrase})
	if err != nil {
		return nil, 0, driver.ServerInfo{}, err
	}
	defer sshRelease()
	var d driver.Driver
	switch p.Engine {
	case "postgres":
		d = postgres.Driver{}
	case "mysql", "mariadb":
		d = mysql.Driver{}
	default:
		return nil, 0, driver.ServerInfo{}, errBadEngine
	}
	conn, err := d.Open(ctx, params)
	if err != nil {
		return nil, 0, driver.ServerInfo{}, err
	}
	ping, err := conn.Ping(ctx)
	if err != nil {
		_ = conn.Close()
		return nil, 0, driver.ServerInfo{}, err
	}
	info, err := conn.ServerInfo(ctx)
	if err != nil {
		_ = conn.Close()
		return nil, 0, driver.ServerInfo{}, err
	}
	info.ViaSSH = params.ViaSSH
	return conn, ping, info, nil
}

var errBadEngine = &engineError{}

type engineError struct{}

func (e *engineError) Error() string { return "engine must be postgres, mysql, or mariadb" }
