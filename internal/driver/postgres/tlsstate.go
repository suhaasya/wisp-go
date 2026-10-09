package postgres

import (
	"crypto/tls"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/suhaasya/wisp-go/internal/driver"
)

func postgresTLSLine(pg *pgx.Conn) string {
	if pg == nil {
		return ""
	}
	st, ok := pgTLSState(pg.PgConn())
	if !ok {
		return ""
	}
	return driver.FormatTLSLine(&st)
}

func pgTLSState(pc any) (tls.ConnectionState, bool) {
	if pc == nil {
		return tls.ConnectionState{}, false
	}
	v := reflect.ValueOf(pc)
	if !v.IsValid() {
		return tls.ConnectionState{}, false
	}
	m := v.MethodByName("TLSConnectionState")
	if !m.IsValid() {
		m = v.MethodByName("ConnectionState")
	}
	if !m.IsValid() {
		return tls.ConnectionState{}, false
	}
	res := m.Call(nil)
	if len(res) == 2 {
		st, _ := res[0].Interface().(tls.ConnectionState)
		ok, _ := res[1].Interface().(bool)
		return st, ok && st.HandshakeComplete
	}
	if len(res) == 1 {
		st, _ := res[0].Interface().(tls.ConnectionState)
		return st, st.HandshakeComplete
	}
	return tls.ConnectionState{}, false
}
