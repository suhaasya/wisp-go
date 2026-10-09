package session

import (
	"database/sql/driver"
	"errors"
	"strings"
)

func isConnError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, driver.ErrBadConn) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection is already closed") ||
		strings.Contains(msg, "bad connection") ||
		strings.Contains(msg, "invalid connection")
}
