package session

import (
	"context"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mysql"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/store"
)

// DefaultConnector opens real database drivers.
type DefaultConnector struct{}

func (DefaultConnector) Connect(ctx context.Context, p store.Profile, password []byte, sshSec SSHConnectSecrets) (driver.Conn, driver.ServerInfo, func(), error) {
	params := driver.ConnectParams{
		Host:                    p.Transport.Host,
		Port:                    p.Transport.Port,
		DB:                      p.Database,
		User:                    p.User,
		Password:                password,
		ConnectTimeout:          10 * time.Second,
		AllowPublicKeyRetrieval: p.Options.AllowPublicKeyRetrieval,
		TLS:                     TLSParamsFromProfile(p),
	}
	if p.Advanced.StatementTimeoutSec > 0 {
		params.StatementTimeout = time.Duration(p.Advanced.StatementTimeoutSec) * time.Second
	}
	release, err := applySSHToParams(ctx, &params, p, sshSecrets{password: sshSec.Password, passphrase: sshSec.Passphrase})
	if err != nil {
		return nil, driver.ServerInfo{}, func() {}, err
	}
	var d driver.Driver
	switch p.Engine {
	case "postgres":
		d = postgres.Driver{}
	case "mysql", "mariadb":
		d = mysql.Driver{}
	default:
		release()
		return nil, driver.ServerInfo{}, func() {}, errBadEngine
	}
	conn, err := d.Open(ctx, params)
	if err != nil {
		release()
		return nil, driver.ServerInfo{}, func() {}, err
	}
	if sql := readOnlySessionSQL(p.Engine); sql != "" && p.ReadOnly {
		if _, err := conn.Exec(ctx, sql); err != nil {
			_ = conn.Close()
			release()
			return nil, driver.ServerInfo{}, func() {}, err
		}
	}
	info, err := conn.ServerInfo(ctx)
	if err != nil {
		_ = conn.Close()
		release()
		return nil, driver.ServerInfo{}, func() {}, err
	}
	info.ViaSSH = params.ViaSSH
	return conn, info, release, nil
}
